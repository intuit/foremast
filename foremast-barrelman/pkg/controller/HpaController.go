package controller

import (
	"bytes"
	d "foremast.ai/foremast/foremast-barrelman/pkg/apis/deployment/v1alpha1"
	clientset "foremast.ai/foremast/foremast-barrelman/pkg/client/clientset/versioned"
	"github.com/golang/glog"
	asv2 "k8s.io/api/autoscaling/v2beta2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	hpainformers "k8s.io/client-go/informers/autoscaling/v2beta2"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"sort"
	"strconv"
	"text/template"
	"time"
)

const ForemastHPA = "foremastHpa"

// Watch HPA object,
type HpaController struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface

	foremastClientset clientset.Interface

	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder

	barrelman *Barrelman
}

// HpaAlertContent use to prepare data for alert template
type HpaAlertContent struct {
	Timestamp   string
	Application string
	Namespace   string
	Action      string
	Old         string
	New         string
	HpaLogEntry []d.HpaLogEntry
}

// Define alert template
const letter = `
At {{.Timestamp}} {{.Application}} at {{.Namespace}} was scaled {{.Action}} from {{.Old}} to {{.New}} pods. This is because
{{range $index, $l := .HpaLogEntry}}{{range $i, $d := $l.HpaLog.Details}}
{{$d.MetricAlias}} {{$l.Timestamp}} value {{$d.Current}} is out of normal range ({{$d.Lower}}, {{$d.Upper}}){{end}}{{end}}

If you have any question, Please refer to IKS HPA Doc
IKS Teams
`

// NewDeploymentController returns a new sample controller
func NewHpaController(kubeclientset kubernetes.Interface, foremastClientset clientset.Interface,
	hpaInfomer hpainformers.HorizontalPodAutoscalerInformer, barrelman *Barrelman) *HpaController {

	// Create event broadcaster
	glog.V(4).Info("Creating event broadcaster:" + ForemastHPA)
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: ForemastHPA})

	// Create event broadcaster
	controller := &HpaController{
		kubeclientset:     kubeclientset,
		foremastClientset: foremastClientset,
		recorder:          recorder,
		barrelman:         barrelman,
	}

	glog.Info("Setting up event handlers for HpaController")

	glog.Info("Creating HpaController")

	hpaInfomer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			newHpa := obj.(*asv2.HorizontalPodAutoscaler)
			controller.updateDeploymentMonitor(newHpa)
		},
		UpdateFunc: func(old, new interface{}) {
			oldHpa := old.(*asv2.HorizontalPodAutoscaler)
			newHpa := new.(*asv2.HorizontalPodAutoscaler)
			controller.updateDeploymentMonitor(newHpa)

			// If there is desiredReplicas changed
			if oldHpa.Status.DesiredReplicas != newHpa.Status.DesiredReplicas && len(newHpa.Spec.Metrics) > 0 {
				var l = len(newHpa.Spec.Metrics)
				for i := 0; i < l; i++ {
					var m = newHpa.Spec.Metrics[i]
					if m.Type == "Object" && m.Object.Metric.Name == "namespace_app_pod_hpa_score" {
						//TODO check the status from CRD and get the hpalog from CRD
						monitor := controller.getDeploymentMonitor(newHpa)
						alertContent := HpaAlertContent{}
						alertContent.Timestamp = time.Now().Format(time.RFC1123)
						alertContent.Application = monitor.Annotations[DeploymentName]
						alertContent.Namespace = monitor.Namespace
						alertContent.Action = "up"
						alertContent.Old = strconv.Itoa(int(oldHpa.Status.CurrentReplicas))
						hpaEntries := []d.HpaLogEntry{}

						sort.Slice(monitor.Status.HpaLogs, func(i, j int) bool { // desc
							return monitor.Status.HpaLogs[i].Timestamp > monitor.Status.HpaLogs[j].Timestamp
						})
						current := strconv.FormatInt(time.Now().Unix(), 10)
						logCount := 4 // default scaling up log count
						if newHpa.Status.DesiredReplicas < oldHpa.Status.CurrentReplicas {
							// find most recently 4 scaling down logs
							logCount = 6
							alertContent.Action = "down"
						}
						for _, l := range monitor.Status.HpaLogs {
							if current >= l.Timestamp {
								logTime, _ := strconv.ParseFloat(l.Timestamp, 10)
								l.Timestamp = time.Unix(int64(logTime), 0).Format(time.RFC1123)
								hpaEntries = append(hpaEntries, l)
								if alertContent.New == "" {
									alertContent.New = strconv.Itoa(int(newHpa.Status.DesiredReplicas))
								}
								if len(hpaEntries) >= logCount {
									break
								}
							}
						}
						alertContent.HpaLogEntry = hpaEntries
						// write to log
						tmpl := template.Must(template.New("letter").Parse(letter))
						var buf bytes.Buffer
						tmpl.Execute(&buf, alertContent)
						glog.Infof("%v", buf.String())
					}
				}
			}

		},
		DeleteFunc: func(obj interface{}) {
			hpa := obj.(*asv2.HorizontalPodAutoscaler)
			if hpa == nil {
				return
			}
			monitor := controller.getDeploymentMonitor(hpa)
			if monitor != nil {
				monitor.Spec.HpaScoreTemplate = ""
				controller.foremastClientset.DeploymentV1alpha1().DeploymentMonitors(hpa.Namespace).Update(monitor)
				glog.V(4).Infof("Updating deployment monitor error, while HPA got deleted: %s", monitor.GetName())
				//TODO foremast-brain
			}
		},
	})

	return controller
}

func (c *HpaController) getDeploymentMonitor(hpa *asv2.HorizontalPodAutoscaler) *d.DeploymentMonitor {
	if hpa.Spec.ScaleTargetRef.Kind == "Deployment" {
		var deplName = hpa.Spec.ScaleTargetRef.Name
		if deplName != "" {
			monitor, _ := c.foremastClientset.DeploymentV1alpha1().DeploymentMonitors(hpa.Namespace).Get(deplName, metav1.GetOptions{})
			return monitor
		} else {
			return nil
		}
	}
	return nil
}

/*
apiVersion: autoscaling/v2beta2
kind: HorizontalPodAutoscaler
metadata:
  annotations:
  labels:
    app: hpa-samples
  name: hpa-samples
  namespace: dev-fm-foremast-examples-usw2-dev-dev
spec:
  maxReplicas: 10
  metrics:
  - object:
      metric:
        name: namespace_app_pod_http_server_requests_2xx
      target:
        type: Value
        value: 8
      describedObject:
        apiVersion: apps/v1beta2
        kind: Deployment
        name: hpa-samples
    type: Object
  minReplicas: 3
  scaleTargetRef:
    apiVersion: apps/v1beta2
    kind: Deployment
    name: hpa-samples
*/

func (c *HpaController) updateDeploymentMonitor(hpa *asv2.HorizontalPodAutoscaler) {
	monitor := c.getDeploymentMonitor(hpa)
	if monitor != nil {
		if monitor.Status.HpaScoreEnabled {
			return
		}
		var hpaStrategy = c.barrelman.hpaStrategy
		if hpaStrategy == HPA_STRATEGY_ANYWAY || hpaStrategy == HPA_STRATEGY_HPA_EXISTS {
			if monitor.Spec.HpaScoreTemplate == "" {
				monitor.Spec.HpaScoreTemplate = HPA_SCORE_TEMPLATE_DEFAULT
			}
		} else {
			monitor.Spec.HpaScoreTemplate = ""
		}

		glog.V(4).Infof("Updating deployment monitor: %s", monitor.GetName())
		monitor.Status.HpaScoreEnabled = true

		if monitor.Spec.HpaScoreTemplate != "" {
			glog.V(4).Infof("Notifying foremast service: %s", monitor.GetName())
			if c.barrelman.monitorHpa(monitor) != nil {
				monitor.Status.HpaScoreEnabled = false
			}
		}
	}
}
