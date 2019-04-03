package controller

import (
	d "foremast.ai/foremast/foremast-barrelman/pkg/apis/deployment/v1alpha1"
	clientset "foremast.ai/foremast/foremast-barrelman/pkg/client/clientset/versioned"
	"github.com/golang/glog"
	asv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	hpainformers "k8s.io/client-go/informers/autoscaling/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
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
			newHpa := obj.(*asv1.HorizontalPodAutoscaler)
			controller.updateDeploymentMonitor(newHpa)
		},
		UpdateFunc: func(old, new interface{}) {
			newHpa := new.(*asv1.HorizontalPodAutoscaler)
			controller.updateDeploymentMonitor(newHpa)
		},
		DeleteFunc: func(obj interface{}) {
			hpa := obj.(*asv1.HorizontalPodAutoscaler)
			if hpa == nil {
				return
			}
			monitor, err := controller.getDeploymentMonitor(hpa)
			if err == nil {
				monitor.Spec.HpaScoreTemplate = ""
				_, err = controller.foremastClientset.DeploymentV1alpha1().DeploymentMonitors(hpa.Namespace).Update(monitor)
				glog.V(4).Infof("Updating deployment monitor error, while HPA got deleted: %s", monitor.GetName())
			}
		},
	})

	return controller
}

func (c *HpaController) getDeploymentMonitor(hpa *asv1.HorizontalPodAutoscaler) (*d.DeploymentMonitor, error) {
	return c.foremastClientset.DeploymentV1alpha1().DeploymentMonitors(hpa.Namespace).Get(hpa.Name, metav1.GetOptions{})
}

func (c *HpaController) updateDeploymentMonitor(hpa *asv1.HorizontalPodAutoscaler) {
	monitor, err := c.getDeploymentMonitor(hpa)
	if err == nil {
		var hpaStrategy = c.barrelman.hpaStrategy
		if hpaStrategy == HPA_STRATEGY_ANYWAY || hpaStrategy == HPA_STRATEGY_SPEC_EXISTS {
			if monitor.Spec.HpaScoreTemplate == "" || monitor.Spec.HpaScoreTemplate == "disabled" {
				monitor.Spec.HpaScoreTemplate = HPA_SCORE_TEMPLATE_DEFAULT
			}
		} else if hpaStrategy == HPA_STRATEGY_ENABLED_ONLY {
			if *hpa.Spec.MinReplicas > 0 {

			}
		}
		monitor.Spec.HpaScoreTemplate = ""
		_, err = c.foremastClientset.DeploymentV1alpha1().DeploymentMonitors(hpa.Namespace).Update(monitor)
		glog.V(4).Infof("Updating deployment monitor error, while HPA got deleted: %s", monitor.GetName())
	}
}
