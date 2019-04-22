package controller

import (
	"foremast.ai/foremast/foremast-barrelman/pkg/apis/deployment/v1alpha1"
	a "foremast.ai/foremast/foremast-barrelman/pkg/client/analyst"
	m "foremast.ai/foremast/foremast-barrelman/pkg/client/metrics"
	"github.com/golang/glog"
	gocache "github.com/pmylund/go-cache"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"os"
	"strings"
	"time"

	clientset "foremast.ai/foremast/foremast-barrelman/pkg/client/clientset/versioned"
	deploymentutil "k8s.io/kubernetes/pkg/controller/deployment/util"
)

const FOREMAST = "foremast"

const MODE_HPA_ONLY = "hpa_only"
const MODE_HPA_AND_HEALTHY_MONITORING = "hpa_and_healthy_monitoring"

// HPA score strategy
// If the deployment has HPA object, generates score
const HPA_STRATEGY_HPA_EXISTS = "hpa_exists"

// Generates HPA SCORE any way
const HPA_STRATEGY_ANYWAY = "anyway"

const HPA_SCORE_TEMPLATE_DEFAULT = "cpu_bound"

// DeploymentController is the controller implementation for watching deployment changes
type Barrelman struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface

	foremastClientset clientset.Interface

	mode string

	hpaStrategy string
}

func NewBarrelman(
	kubeclientset kubernetes.Interface,
	foremastClientset clientset.Interface,
	mode string,
	hpaStrategy string) *Barrelman {

	controller := &Barrelman{
		kubeclientset:     kubeclientset,
		foremastClientset: foremastClientset,
		mode:              mode,
		hpaStrategy:       hpaStrategy,
	}

	ticker := time.NewTicker(10 * time.Second)
	go func() {
		for range ticker.C {
			controller.checkRunningStatus(kubeclientset, foremastClientset)
		}
	}()

	return controller
}

func (c *Barrelman) hasHPA() bool {
	return strings.Contains(c.mode, "hpa")
}

func (c *Barrelman) hasHealthyMonitoring() bool {
	return strings.Contains(c.mode, "healthy_monitoring")
}

func (c *Barrelman) getReplicaSet(newUid, oldUid string, namespace string, kubeclientset kubernetes.Interface) ([]appsv1.ReplicaSet, error) {
	var selectedRsList = []appsv1.ReplicaSet{}
	rsList, err := kubeclientset.AppsV1().ReplicaSets(namespace).List(metav1.ListOptions{})

	if err == nil {
		for _, rs := range rsList.Items {
			if len(rs.OwnerReferences) > 0 {
				var rsUid = string(rs.OwnerReferences[0].UID)
				if (rsUid == newUid || rsUid == oldUid) && (*rs.Spec.Replicas > 0 || rs.Status.Replicas > 0) {
					selectedRsList = append(selectedRsList, rs)
				}
			}
		}
		return selectedRsList, nil
	}
	return nil, err
}

func (c *Barrelman) getPodNames(oldDepl, newDepl *appsv1.Deployment, kubeclientset kubernetes.Interface) ([][]string, error) {
	var oldUid = string(oldDepl.UID)
	var newUid = string(newDepl.UID)

	var selectedRsList, err = c.getReplicaSet(newUid, oldUid, newDepl.Namespace, kubeclientset)

	var l = 0
	if err == nil {
		l = len(selectedRsList)
	}
	glog.Infof("ReplicaSet found %d", l)

	var oldPods = []string{}

	if l <= 1 { //Sleep 5 seconds and try again
		if l == 1 { //Get the pod before k8s terminates them
			podList, err := kubeclientset.CoreV1().Pods(newDepl.Namespace).List(metav1.ListOptions{
				LabelSelector: "pod-template-hash in (" + selectedRsList[0].Labels["pod-template-hash"] + ")",
			})
			if err == nil {
				for _, pod := range podList.Items {
					if !contains(oldPods, pod.Name) {
						oldPods = append(oldPods, pod.Name)
					}
				}
			}
		}
		glog.Infof("Can not find ReplicaSet or only one, sleeping 5 seconds: %s", newDepl.Name)
		//Sleep 5 seconds to make sure k8s will create newReplicaSet
		time.Sleep(5 * time.Second)
		selectedRsList, err = c.getReplicaSet(newUid, oldUid, newDepl.Namespace, kubeclientset)
		l = len(selectedRsList)

		glog.Infof("ReplicaSet found after sleeping %d", l)
	}

	if l == 2 { // Two
		var result = [][]string{}
		var oldRs = selectedRsList[1]
		var newRs = selectedRsList[0]

		var oldMessage = ""
		if oldDepl.Status.Conditions != nil {
			for _, cond := range oldDepl.Status.Conditions {
				if strings.HasPrefix(cond.Message, "ReplicaSet") {
					oldMessage = cond.Message
				}
			}
		}

		var oldMatch = false
		if oldMessage != "" && strings.Contains(oldMessage, oldRs.Name) {
			oldMatch = true
		}

		if !oldMatch { //Only exchange when they are not match
			var tmp = newRs
			newRs = oldRs
			oldRs = tmp
		}

		var retry int32 = 0
		result = append(result, []string{})
		result = append(result, []string{})

		for retry < 3 {
			podList, err := kubeclientset.CoreV1().Pods(newDepl.Namespace).List(metav1.ListOptions{
				LabelSelector: "pod-template-hash in (" + newRs.Labels["pod-template-hash"] + "," + oldRs.Labels["pod-template-hash"] + ")",
			})
			if err != nil {
				glog.V(1).Infof("Listing Pods error: %v %v", newDepl.Name, err)
				return nil, err
			}

			for _, pod := range podList.Items {
				if len(pod.OwnerReferences) > 0 && pod.OwnerReferences[0].UID == newRs.UID {
					if !contains(result[0], pod.Name) {
						result[0] = append(result[0], pod.Name)
					}
				} else {
					if !contains(result[1], pod.Name) {
						result[1] = append(result[1], pod.Name)
					}
				}
			}

			//Result[0] is current
			if len(result[1]) == 0 {
				if len(oldPods) > 0 {
					result[1] = oldPods
					return result, nil
				} else {
					var newResult = [][]string{}
					newResult = append(newResult, result[0])
					return newResult, nil
				}
			} else if len(result[0]) == 0 { //No current, wait couple seconds
				//return nil, errors.NewResourceExpired("No pod found")
				retry += 1
				time.Sleep(5 * time.Second)
			} else {
				break
			}
		}

		return result, nil
	} else if l == 1 { //Only one
		var newRs = selectedRsList[0]
		podList, err := kubeclientset.CoreV1().Pods(newDepl.Namespace).List(metav1.ListOptions{
			LabelSelector: "pod-template-hash = " + newRs.Labels["pod-template-hash"],
		})

		if err != nil {
			glog.V(1).Infof("Listing Pods error: %v %v", newDepl.Name, err)
			return nil, err
		}

		var result = [][]string{}
		result = append(result, []string{})
		if len(oldPods) > 0 {
			result = append(result, oldPods)
		}
		for _, pod := range podList.Items {
			result[0] = append(result[0], pod.Name)
		}
		return result, nil
	} else {
		glog.Infof("No ReplicaSet found %s", err)
		return nil, err
	}
}

//monitorNewDeployment loops, checking the new deployment metrics against those from the old deployment, rolling back if need be
func (c *Barrelman) monitorNewDeployment(appName string, oldDepl, newDepl *appsv1.Deployment, deploymentMetadata *v1alpha1.DeploymentMetadata,
	oldMonitor *v1alpha1.DeploymentMonitor, monitorNotFound bool, strategy string) {
	glog.Infof("Beginning to monitor new deployment %v vs. old deployment %v, strategy %s",
		newDepl.Spec.Template.Spec.Containers[0].Image, oldDepl.Spec.Template.Spec.Containers[0].Image, strategy)
	glog.Infof("Beginning to monitor new env %v vs. old env %v",
		newDepl.Spec.Template.Spec.Containers[0].Env, oldDepl.Spec.Template.Spec.Containers[0].Env)

	var podNames [][]string
	var err error
	if !(strategy == m.StrategyContinuous || strategy == m.StrategyHpa) {
		podNames, err = c.getPodNames(oldDepl, newDepl, c.kubeclientset)
		if err != nil {
			glog.Infof("Get pod names error %v %s", newDepl, err)
			return
		}

		glog.Infof("Found pod names: %s, %d", podNames[0], len(podNames))
	}

	var jobId string
	var phase string
	var oldMonitor2, er = c.foremastClientset.DeploymentV1alpha1().DeploymentMonitors(newDepl.Namespace).Get(newDepl.Name, metav1.GetOptions{})
	if er == nil {
		oldMonitor = oldMonitor2
	}

	//begin monitoring new deployment
	if (strategy == m.StrategyContinuous || strategy == m.StrategyHpa) || (oldMonitor != nil && !oldMonitor.Spec.Continuous) { //Only generates job when the continuous monitoring is not running
		analystClient, err := a.NewClient(nil, deploymentMetadata.Spec.Analyst.Endpoint)
		if err != nil {
			glog.Infof("Creating foremast service request error %v %s", newDepl, err)
			return
		}

		var metrics = deploymentMetadata.Spec.Metrics
		var metricAliases []string = nil

		if strategy == m.StrategyHpa { //Select metrics by hpaScoreTemplate
			if oldMonitor.Spec.HpaScoreTemplate == "" {
				glog.Infof("No HpaScore Template ignore %v", oldMonitor)
				return
			}

			for _, t := range deploymentMetadata.Spec.HpaScoreTemplates {
				if oldMonitor.Spec.HpaScoreTemplate == t.Name {
					metricAliases = t.Metrics
					break
				}
			}
		}

		jobId, err = analystClient.StartAnalyzing(newDepl.Namespace, appName, podNames, deploymentMetadata.Spec.Metrics.Endpoint, metrics, watchTime, strategy, metricAliases)

		if err != nil {
			glog.Infof("Starting analyzing error, try it again: %v", err)
			jobId, err = analystClient.StartAnalyzing(newDepl.Namespace, appName, podNames, deploymentMetadata.Spec.Metrics.Endpoint, metrics, watchTime, strategy, metricAliases)
			if err != nil {
				glog.Infof("Tried twice to analyzing error: %v", err)
				return
			}
		}
		phase = v1alpha1.MonitorPhaseRunning
	} else { //Take the old jobId
		jobId = oldMonitor.Status.JobId
		phase = oldMonitor.Status.Phase
	}

	//Create deployment analysis object if it doesn't exist
	if oldMonitor != nil {

		var t = time.Now()
		var start = t.Format(time.RFC3339)
		t = t.Add(waitUntilMax * time.Minute)
		var waitUntil = t.Format(time.RFC3339)

		oldMonitor.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   "deployment.foremast.ai",
			Kind:    "DeploymentMonitor",
			Version: "v1alpha1",
		})
		oldMonitor.Namespace = newDepl.Namespace
		oldMonitor.Name = newDepl.Name
		if oldMonitor.Annotations == nil {
			oldMonitor.Annotations = map[string]string{}
		}

		// default is 0
		var oldRevision = oldMonitor.Spec.RollbackRevision
		if strategy == m.StrategyRollingUpdate {
			oldRevision, _ = deploymentutil.Revision(oldDepl)
		}

		oldMonitor.Annotations[DeploymentName] = newDepl.Name
		var remediationOption = v1alpha1.RemediationNone
		if oldMonitor.Spec.Remediation.Option != "" {
			remediationOption = oldMonitor.Spec.Remediation.Option
		}

		oldMonitor.Spec = v1alpha1.DeploymentMonitorSpec{
			Selector:  newDepl.Spec.Selector,
			Analyst:   deploymentMetadata.Spec.Analyst,
			StartTime: start,
			WaitUntil: waitUntil,
			Metrics:   deploymentMetadata.Spec.Metrics,
			Logs:      deploymentMetadata.Spec.Logs,
			Remediation: v1alpha1.RemediationAction{
				Option: remediationOption,
			},
			Continuous:       oldMonitor.Spec.Continuous,
			HpaScoreTemplate: oldMonitor.Spec.HpaScoreTemplate,
			RollbackRevision: oldRevision,
		}

		oldMonitor.Status = v1alpha1.DeploymentMonitorStatus{
			JobId:     jobId,
			Phase:     phase,
			Timestamp: start,
		}
	}
	if monitorNotFound { //Not found
		_, err = c.foremastClientset.DeploymentV1alpha1().DeploymentMonitors(newDepl.Namespace).Create(oldMonitor)
		if err != nil {
			glog.Infof("Creating DeploymentMonitor error: %v", err)
		} else {
			glog.Infof("Created DeploymentMonitor %v", oldMonitor)
		}
	} else {
		_, err = c.foremastClientset.DeploymentV1alpha1().DeploymentMonitors(newDepl.Namespace).Update(oldMonitor)
		if err != nil {
			glog.Infof("Updating DeploymentMonitor error: %v", err)
		} else {
			glog.Infof("Updated DeploymentMonitor %v", oldMonitor)
		}
	}
}

func (c *Barrelman) monitorHpa(monitor *v1alpha1.DeploymentMonitor) error {
	return c.monitorInternal(monitor, m.StrategyHpa)
}

func (c *Barrelman) monitorContinuously(monitor *v1alpha1.DeploymentMonitor) error {
	return c.monitorInternal(monitor, m.StrategyContinuous)
}

func (c *Barrelman) getDeploymentMetadata(namespace string, appName string, depl *appsv1.Deployment) (*v1alpha1.DeploymentMetadata, error) {

	var key = namespace + ":" + appName
	if cached, found := metadataCache.Get(key); found {
		switch cached.(type) {
		case *v1alpha1.DeploymentMetadata:
			return cached.(*v1alpha1.DeploymentMetadata), nil
		default:
			return nil, cached.(error)
		}
	}

	//Try to get the metadata by "app" name, if it doesn't existing, try to search by "appType"
	metadatas := c.foremastClientset.DeploymentV1alpha1().DeploymentMetadatas(depl.Namespace)
	deploymentMetadata, err := metadatas.Get(appName, metav1.GetOptions{})
	if err != nil {
		newAppType, hasAppType := depl.Labels["appType"]
		if hasAppType {
			deploymentMetadata, err = metadatas.Get(newAppType, metav1.GetOptions{})
			if err != nil {
				var curerntNamespace = os.Getenv("NAMESPACE")
				deploymentMetadata, err = c.foremastClientset.DeploymentV1alpha1().DeploymentMetadatas(curerntNamespace).Get(newAppType, metav1.GetOptions{})
				if err != nil {
					metadataCache.Set(key, err, gocache.DefaultExpiration)
					glog.Infof("Getting deployment metadata error by appType:%s, in either namespace %s or %s:", newAppType, depl.Namespace, curerntNamespace)
					return nil, err
				}
			}
		} else {
			metadataCache.Set(key, err, gocache.DefaultExpiration)
			return nil, err
		}
	}
	metadataCache.Set(key, deploymentMetadata, gocache.DefaultExpiration)
	return deploymentMetadata, nil
}

func (c *Barrelman) monitorInternal(monitor *v1alpha1.DeploymentMonitor, strategy string) error {
	var deploymentName = monitor.Annotations[DeploymentName]
	if deploymentName == "" {
		deploymentName = monitor.Name
	}
	depl, err := c.kubeclientset.AppsV1().Deployments(monitor.Namespace).Get(deploymentName, metav1.GetOptions{})

	if err != nil {
		return err
	}

	var appName string
	var ok bool

	if appName, ok = depl.Labels["app"]; !ok || appName == "" {
		glog.V(5).Infof("no app label found on new deployment, skipping deployment %v", deploymentName)
		return errors.NewBadRequest("no app label found on new deployment, skipping deployment " + deploymentName)
	}

	//Try to get the metadata by "app" name, if it doesn't existing, try to search by "appType"
	deploymentMetadata, err := c.getDeploymentMetadata(depl.Namespace, appName, depl)
	if err != nil {
		return err
	}

	c.monitorNewDeployment(appName, depl, depl, deploymentMetadata, monitor, false, strategy)
	return nil
}

func (c *Barrelman) checkRunningStatus(kubeclientset kubernetes.Interface, foremastClientset clientset.Interface) {

	ni := kubeclientset.CoreV1().Namespaces()
	nl, err := ni.List(metav1.ListOptions{})
	if err != nil {
		glog.Infof("Listing deployment monitors error: %v", err)
		return
	}

	for _, n := range nl.Items {
		list, err := foremastClientset.DeploymentV1alpha1().DeploymentMonitors(n.Name).List(metav1.ListOptions{})
		if err != nil {
			glog.Infof("Listing deployment monitors error: %v", err)
		}

		if list != nil && len(list.Items) > 0 {
			for _, item := range list.Items {

				if item.Status.Phase == v1alpha1.MonitorPhaseRunning {
					//Expire
					var changed = false
					if !item.Status.Expired {
						analystClient, err := a.NewClient(nil, item.Spec.Analyst.Endpoint)
						if err != nil {
							glog.Infof("Creating backend client error: %v, %s", item.Namespace+":"+item.Name, err)
							continue
						}

						if item.Status.JobId == "" {
							item.Status.Expired = true
							item.Status.Phase = v1alpha1.MonitorPhaseHealthy
							item.Status.Timestamp = time.Now().Format(time.RFC3339)
							changed = true
						} else {
							statusResponse, err := analystClient.GetStatus(item.Status.JobId)
							if err != nil {
								glog.Infof("Getting deployment monitor status error: %v, %s", item.Namespace+":"+item.Name, err)
								continue
							}

							var oldPhase = item.Status.Phase

							item.Status.Phase = statusResponse.Status
							if statusResponse.Anomaly != nil {
								anomaly, err := convertToAnomaly(statusResponse.Anomaly)
								if err == nil {
									item.Status.Anomaly = anomaly
									changed = true
								}
							}

							if item.Status.Phase != oldPhase {
								changed = true
							}
						}

						item.Status.Timestamp = time.Now().Format(time.RFC3339)
					}

					//If it still running then check whether is expired
					if item.Status.Phase == v1alpha1.MonitorPhaseRunning && item.Spec.WaitUntil != "" {
						until, err := time.Parse(time.RFC3339, item.Spec.WaitUntil)
						if err == nil {
							if until.Unix() < time.Now().Unix() {
								item.Status.Phase = v1alpha1.MonitorPhaseHealthy
								item.Status.Expired = true
								item.Status.Timestamp = time.Now().Format(time.RFC3339)
							}
						}
					}

					if changed {
						item.Status.RemediationTaken = false
						_, err = foremastClientset.DeploymentV1alpha1().DeploymentMonitors(item.Namespace).Update(&item)
						if err != nil {
							glog.Infof("Updating deployment monitor error: %v, %s", item.Namespace+":"+item.Name, err)
						} else {
							glog.Infof("Updated deployment monitor %v", item)
						}
					}
				} else if item.Spec.Continuous || item.Spec.HpaScoreTemplate != "" {
					if c.hasHealthyMonitoring() && item.Status.Phase == v1alpha1.MonitorPhaseUnhealthy {
						d, err := time.Parse(time.RFC3339, item.Status.Timestamp)
						if err == nil && (time.Now().Unix()-d.Unix()) > 60 {
							//Trigger a continuous monitoring
							go c.monitorContinuously(&item)
						}
					} else {
						if c.hasHealthyMonitoring() && item.Spec.Continuous {
							go c.monitorContinuously(&item)
						} else if item.Spec.HpaScoreTemplate != "" {
							go c.monitorHpa(&item)
						}
					}
				}
			}
		}
	}

}
