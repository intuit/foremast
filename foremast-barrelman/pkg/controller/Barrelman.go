/*
Copyright 2018 The Kubernetes Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"fmt"
	"foremast.ai/foremast/foremast-barrelman/pkg/apis/deployment/v1alpha1"
	a "foremast.ai/foremast/foremast-barrelman/pkg/client/analyst"
	m "foremast.ai/foremast/foremast-barrelman/pkg/client/metrics"
	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"os"
	"strings"

	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	gocache "github.com/pmylund/go-cache"

	clientset "foremast.ai/foremast/foremast-barrelman/pkg/client/clientset/versioned"
	appsinformers "k8s.io/client-go/informers/apps/v1"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	appslisters "k8s.io/client-go/listers/apps/v1"
	deploymentutil "k8s.io/kubernetes/pkg/controller/deployment/util"
)

const barrelman = "barrelman"

const watchTime = 10

const waitUntilMax = 30

const DeploymentName = "deployment.kubernetes.io/name"

const Strategy = "deployment.foremast.ai/strategy"

const ForemastAnotation = "foremast.ai/monitoring"

const CanarySuffic = "-foremast-canary"

const (
	// SuccessSynced is used as part of the Event 'reason' when a Foo is synced
	SuccessSynced = "Synced"
	// ErrResourceExists is used as part of the Event 'reason' when a Foo fails
	// to sync due to a Deployment of the same name already existing.
	ErrResourceExists = "ErrResourceExists"

	// MessageResourceExists is the message used for Events when a resource
	// fails to sync due to a Deployment already existing
	MessageResourceExists = "Resource %q already exists and is not managed by an Foremast-enabled deployment"
	// MessageResourceSynced is the message used for an Event fired when a Foo
	// is synced successfully
	MessageResourceSynced = "Foremast-barrelman-enabled resource synced successfully"
)

// AddToManagerFuncs is a list of functions to add all Controllers to the Manager
var AddToManagerFuncs []func(manager.Manager) error

// AddToManager adds all Controllers to the Manager
func AddToManager(m manager.Manager) error {
	for _, f := range AddToManagerFuncs {
		if err := f(m); err != nil {
			return err
		}
	}
	return nil
}

var (
	namespaceBlacklist = map[string]bool{
		"kube-public": false,
		"kube-system": false,
		"opa":         false,
		"monitoring":  false,
	}
	namespaceCache = gocache.New(5*time.Minute, 30*time.Second)
	metadataCache  = gocache.New(1*time.Minute, 10*time.Second)
)

// Barrelman is the controller implementation for watching deployment changes
type Barrelman struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface

	foremastClientset clientset.Interface

	deploymentsLister appslisters.DeploymentLister

	deploymentsSynced cache.InformerSynced

	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder

	quit chan bool
}

func EnvArrayEquals(a []corev1.EnvVar, b []corev1.EnvVar) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v.Name != b[i].Name || v.Value != b[i].Value {
			return false
		}
	}
	return true
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

func (c *Barrelman) monitorContinuously(monitor *v1alpha1.DeploymentMonitor) error {
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

	c.monitorNewDeployment(appName, depl, depl, deploymentMetadata, monitor, false, m.StrategyContinuous)
	return nil
}

func (c *Barrelman) monitorDeployment(appName string, oldDepl, newDepl *appsv1.Deployment, deploymentMetadata *v1alpha1.DeploymentMetadata) {
	//Try to get the metadata by "app" name, if it doesn't existing, try to search by "appType"
	deploymentMetadata, err := c.getDeploymentMetadata(newDepl.Namespace, appName, newDepl)
	if err != nil {
		return
	}

	//get old and new lists of containers
	oldNumContainers := len(oldDepl.Spec.Template.Spec.Containers)
	newNumContainers := len(newDepl.Spec.Template.Spec.Containers)

	//skip if number of containers is different - this would cause problems in the next step
	if oldNumContainers != newNumContainers {
		glog.V(4).Infof("~~~~Number of containers in deployment changed from %v to %v~~~~", oldNumContainers, newNumContainers)
		c.handleObject(newDepl)
		return
	}

	//if any corresponding container has changed its image between old and new, there has been a new release
	for i := 0; i < newNumContainers; i++ {
		oldContainer := oldDepl.Spec.Template.Spec.Containers[i]
		newContainer := newDepl.Spec.Template.Spec.Containers[i]
		oldImage := oldContainer.Image
		newImage := newContainer.Image

		oldEnv := oldContainer.Env
		newEnv := newContainer.Env

		//if images are different we may need to start monitoring the new deployment metrics
		if newImage != oldImage || !EnvArrayEquals(oldEnv, newEnv) {
			glog.V(4).Infof("~~~~Image of container %v changed from %v to %v~~~~", i, oldImage, newImage)

			//Pretend the looping after rollback
			oldMonitor, err := c.foremastClientset.DeploymentV1alpha1().DeploymentMonitors(newDepl.Namespace).Get(newDepl.Name, metav1.GetOptions{})
			var monitorNotFound = true
			if err == nil {
				if oldMonitor.Spec.Continuous {
					glog.Infof("The deployment is watching continuously:%s", newDepl.Name)
					return
				}
				monitorNotFound = false
				newRevision, err := deploymentutil.Revision(newDepl)
				if err == nil && newRevision > 0 && newRevision == oldMonitor.Spec.RollbackRevision {
					glog.V(9).Infof("The new deployment is a rollback deployment:%s", appName)
					return
				}
				if oldDepl.Annotations["deprecated.deployment.rollback.to"] != "" {
					glog.V(9).Infof("The new deployment is a rollback deployment:%s", oldDepl.Annotations["deprecated.deployment.rollback.to"])
					return
				}
			}

			//begin monitoring new deployment
			if !oldMonitor.Spec.Continuous { //Only generates job when the continuous monitoring is not running
				glog.Info("Starting to monitor new deployment...")
				go c.monitorNewDeployment(appName, oldDepl, newDepl, deploymentMetadata, oldMonitor, monitorNotFound, m.StrategyRollingUpdate)
			}

			c.handleObject(newDepl)
			return
		}
	}
}

// NewBarrelman returns a new sample controller
func NewBarrelman(
	kubeclientset kubernetes.Interface,
	foremastClientset clientset.Interface,
	deploymentInformer appsinformers.DeploymentInformer) *Barrelman {

	// Create event broadcaster
	glog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: barrelman})

	//lastDeployments := make(map[string]string)

	//glog.Info("Connecting to Prometheus client")
	//promClient, err := prometheusClient.NewClient(prometheusClient.Config{Address: "http://127.0.0.1:9090"})
	//if err != nil {
	//	glog.Infof("Failed to set up prometheus client: %v", err)
	//}
	//prometheus := prometheusv1.NewAPI(promClient)
	quit := make(chan bool)

	controller := &Barrelman{
		kubeclientset:     kubeclientset,
		foremastClientset: foremastClientset,
		deploymentsLister: deploymentInformer.Lister(),
		deploymentsSynced: deploymentInformer.Informer().HasSynced,
		workqueue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "barrelman"),
		recorder:          recorder,
		quit:              quit,
	}

	glog.Info("Setting up event handlers")

	// Set up an event handler for when Deployment resources change. More info on this pattern:
	// https://github.com/kubernetes/community/blob/8cafef897a22026d42f5e5bb3f104febe7e29830/contributors/devel/controllers.md
	deploymentInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			//get deployment object
			newDepl := obj.(*appsv1.Deployment)
			var newApp string
			var ok bool
			var deploymentName = newDepl.Name

			if newApp, ok = newDepl.Labels["app"]; !ok || newApp == "" {
				glog.V(5).Infof("no app label found on new deployment, skipping deployment %v", newDepl.Name)
				return
			}

			if !controller.isMonitoring(newDepl.Namespace) {
				glog.V(5).Infof("Ignore namespace %s", newDepl.Namespace)
				return
			}

			deploymentMetadata, err := controller.getDeploymentMetadata(newDepl.Namespace, newApp, newDepl)
			if err != nil {
				return
			}

			var deploymentStrategy = m.StrategyRollingUpdate
			if strings.HasSuffix(deploymentName, CanarySuffic) {
				deploymentStrategy = m.StrategyCanary
			}

			var t = time.Now()
			var start = t.Format(time.RFC3339)
			var waitUntil = t.Add(waitUntilMax * time.Minute).Format(time.RFC3339)

			oldMonitor, err := foremastClientset.DeploymentV1alpha1().DeploymentMonitors(newDepl.Namespace).Get(newDepl.Name, metav1.GetOptions{})

			var remediationAction v1alpha1.RemediationAction
			var create = false
			var continuous = false
			if err != nil {
				oldMonitor = &v1alpha1.DeploymentMonitor{
					ObjectMeta: metav1.ObjectMeta{
						Name:      newDepl.Name,
						Namespace: newDepl.Namespace,
						Annotations: map[string]string{
							deploymentName: newDepl.Name,
						},
					},
				}
				remediationAction = v1alpha1.RemediationAction{
					Option: v1alpha1.RemediationNone,
				}

				create = true
			} else {
				remediationAction = oldMonitor.Spec.Remediation
				continuous = oldMonitor.Spec.Continuous
			}

			oldMonitor.Spec = v1alpha1.DeploymentMonitorSpec{
				Selector:         newDepl.Spec.Selector,
				Analyst:          deploymentMetadata.Spec.Analyst,
				StartTime:        start,
				WaitUntil:        waitUntil,
				Metrics:          deploymentMetadata.Spec.Metrics,
				Logs:             deploymentMetadata.Spec.Logs,
				Remediation:      remediationAction,
				Continuous:       continuous,
				RollbackRevision: 0,
			}
			oldMonitor.Status = v1alpha1.DeploymentMonitorStatus{
				JobId: "",
				Phase: v1alpha1.MonitorPhaseHealthy,
			}
			oldMonitor.SetGroupVersionKind(schema.GroupVersionKind{
				Group:   "deployment.foremast.ai",
				Kind:    "DeploymentMonitor",
				Version: "v1alpha1",
			})

			var monitors = foremastClientset.DeploymentV1alpha1().DeploymentMonitors(newDepl.Namespace)
			if create {
				_, err = monitors.Create(oldMonitor)
				if err != nil {
					glog.V(1).Infof("Creating DeploymentMonitor error: %v", err)
				}
			} else {
				_, err = monitors.Update(oldMonitor)
				if err != nil {
					glog.V(1).Infof("Updating DeploymentMonitor error: %v", err)
				}
			}

			//If the strategy is canary, it should be monitored also
			if deploymentStrategy == m.StrategyCanary {
				var l = len(deploymentName) - len(CanarySuffic)
				var oldDeploymentName = deploymentName[0:l]
				oldDepl, err := kubeclientset.AppsV1().Deployments(newDepl.Namespace).Get(oldDeploymentName, metav1.GetOptions{})
				if err != nil {
					glog.V(1).Infof("Not able to find the old deployment:%s error: %v", oldDeploymentName, err)
					return
				}
				controller.monitorDeployment(newApp, oldDepl, newDepl, deploymentMetadata)
			}
		},
		UpdateFunc: func(old, new interface{}) {

			//get old and new deployment objects
			newDepl := new.(*appsv1.Deployment)
			oldDepl := old.(*appsv1.Deployment)

			if !controller.isMonitoring(newDepl.Namespace) {
				glog.V(5).Infof("Ignore namespace %s", newDepl.Namespace)
				return
			}

			//skip if not marked to be tracked for ACA
			var newApp string
			var oldApp string
			var ok bool

			if newApp, ok = newDepl.Labels["app"]; !ok || newApp == "" {
				glog.V(5).Infof("no app label found on new deployment, skipping deployment %v", newDepl.Name)
				return
			}
			if oldApp, ok = oldDepl.Labels["app"]; !ok || oldApp == "" {
				glog.V(5).Infof("no app label found on old deployment, skipping deployment %v", oldDepl.Name)
				return
			}

			if newApp != oldApp {
				glog.V(5).Infof("The app names are not sample in two deployments, skipping deployment old[%v:%v] new[%v:%v",
					oldDepl.Name, oldApp, newDepl.Name, newApp)
				return
			}

			controller.monitorDeployment(newApp, old.(*appsv1.Deployment), new.(*appsv1.Deployment), nil)
		},
		DeleteFunc: func(obj interface{}) {
			//get deployment object
			depl := obj.(*appsv1.Deployment)

			if !controller.isMonitoring(depl.Namespace) {
				glog.V(5).Infof("Ignore namespace %s", depl.Namespace)
				return
			}

			//skip if not marked to be tracked for ACA
			if val, ok := depl.Annotations["aca"]; !ok || val != "true" {
				glog.V(5).Info("no aca=true annotation found, skipping deployment")
				return
			}

			//construct key from namespace and name of deployment
			//key := "" + depl.Namespace + "." + depl.Name

			//remove entry from map of last deployments
			//delete(controller.lastDeployments, key)
			//glog.V(4).Infof("Deleted object with image: %v", obj.(*appsv1.Deployment).Spec.Template.Spec.Containers[0].Image)
			//controller.handleObject(obj)

			// Delete DeploymentMonitor
			err := foremastClientset.DeploymentV1alpha1().DeploymentMetadatas(depl.Namespace).Delete(depl.Name, &metav1.DeleteOptions{})
			glog.V(4).Infof("Deleting DeployementMonitor %v error:  %v", depl.Name, err)
		},
	})

	ticker := time.NewTicker(10 * time.Second)
	go func() {
		for range ticker.C {
			controller.checkRunningStatus(kubeclientset, foremastClientset)
		}
	}()

	return controller
}

func (c *Barrelman) isMonitoring(namespace string) bool {
	_, ok := namespaceBlacklist[namespace]
	if ok {
		return false
	}

	if cached, found := namespaceCache.Get(namespace); found {
		return cached == true
	}

	ns, err := c.kubeclientset.CoreV1().Namespaces().Get(namespace, metav1.GetOptions{})
	if err != nil {
		return false
	}
	var result = ns.Annotations[ForemastAnotation] != "false"
	namespaceCache.Set(namespace, result, gocache.DefaultExpiration)
	return result
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
				} else if item.Spec.Continuous {
					if item.Status.Phase == v1alpha1.MonitorPhaseUnhealthy {
						d, err := time.Parse(time.RFC3339, item.Status.Timestamp)
						if err == nil && (time.Now().Unix()-d.Unix()) > 60 {
							//Trigger a continuous monitoring
							go c.monitorContinuously(&item)
						}
					} else {
						go c.monitorContinuously(&item)
					}
				}
			}
		}
	}

}

func convertToAnomaly(anomalyInfo map[string]a.AnomalyInfo) (v1alpha1.Anomaly, error) {
	var anomaly = v1alpha1.Anomaly{
		AnomalousMetrics: []v1alpha1.AnomalousMetric{},
	}

	for key, value := range anomalyInfo {
		var anomalousMetric = v1alpha1.AnomalousMetric{
			Name:   key,
			Tags:   value.Tags,
			Values: []v1alpha1.AnomalousMetricValue{},
		}

		var point v1alpha1.AnomalousMetricValue
		for index, v := range value.Values {
			if index%2 == 1 {
				point.Value = v
				anomalousMetric.Values = append(anomalousMetric.Values, point)
			} else {
				point = v1alpha1.AnomalousMetricValue{
					Time: int64(v),
				}
			}
		}
		anomaly.AnomalousMetrics = append(anomaly.AnomalousMetrics, anomalousMetric)
	}

	return anomaly, nil
}

// Contains tells whether a contains x.
func contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
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
	if strategy != m.StrategyContinuous {
		podNames, err = c.getPodNames(oldDepl, newDepl, c.kubeclientset)
		if err != nil {
			glog.Infof("Get pod names error %v %s", newDepl, err)
			return
		}

		glog.Infof("Found pod names: %s, %d", podNames[0], len(podNames))
	}

	analystClient, err := a.NewClient(nil, deploymentMetadata.Spec.Analyst.Endpoint)
	if err != nil {
		glog.Infof("Creating judgement error %v %s", newDepl, err)
		return
	}

	var jobId string
	jobId, err = analystClient.StartAnalyzing(newDepl.Namespace, appName, podNames, deploymentMetadata.Spec.Metrics.Endpoint, deploymentMetadata.Spec.Metrics, watchTime, strategy)

	if err != nil {
		glog.Infof("Starting analyzing error, try it again: %v", err)
		jobId, err = analystClient.StartAnalyzing(newDepl.Namespace, appName, podNames, deploymentMetadata.Spec.Metrics.Endpoint, deploymentMetadata.Spec.Metrics, watchTime, strategy)
		if err != nil {
			glog.Infof("Tried twice to analyzing error: %v", err)
			return
		}
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
		var oldRevision, _ = deploymentutil.Revision(oldDepl)

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
			RollbackRevision: oldRevision,
		}

		//Set default Remediation to RemediationAutoRollback
		if monitorNotFound {
			oldMonitor.Spec.Remediation = v1alpha1.RemediationAction{
				Option: v1alpha1.RemediationAutoRollback,
			}
		}

		oldMonitor.Status = v1alpha1.DeploymentMonitorStatus{
			JobId:     jobId,
			Phase:     v1alpha1.MonitorPhaseRunning,
			Timestamp: start,
			//CurrentRevision: newRevision,
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

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
// this is leftover from the sample controller, but seems to do a good job of setting things up
func (c *Barrelman) Run(threadiness int, stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	defer c.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	glog.Info("Starting foremast controller")

	// Wait for the caches to be synced before starting workers
	glog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.deploymentsSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	glog.Info("Starting workers")
	// Launch two workers to process Foo resources
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	glog.Info("Started workers")
	<-stopCh
	glog.Info("Shutting down workers")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *Barrelman) runWorker() {
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
// left over from sample controller, still processes objects from queue but not doing much with them
func (c *Barrelman) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer c.workqueue.Done(obj)
		var key string
		var ok bool
		// We expect strings to come off the workqueue. These are of the
		// form namespace/name. We do this as the delayed nature of the
		// workqueue means the items in the informer cache may actually be
		// more up to date that when the item was initially put onto the
		// workqueue.
		if key, ok = obj.(string); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.workqueue.Forget(obj)
			runtime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		// Run the syncHandler, passing it the namespace/name string of the
		// Foo resource to be synced.
		if err := c.syncHandler(key); err != nil {
			return fmt.Errorf("error syncing '%s': %s", key, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)
		glog.Infof("Successfully synced '%s'", key)

		return nil
	}(obj)

	if err != nil {
		runtime.HandleError(err)
		return true
	}

	return true
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the Foo resource
// with the current status of the resource.
// stripped down version of sample controller method, most things no longer needed
func (c *Barrelman) syncHandler(key string) error {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Get the deployment with the name specified
	deployment, err := c.deploymentsLister.Deployments(namespace).Get(name)

	// If an error occurs during Get, we'll requeue the item so we can
	// attempt processing again later. This could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		return err
	}

	c.recorder.Event(deployment, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	return nil
}

// enqueue takes a resource and converts it into a namespace/name
// string which is then put onto the work queue
// identical to sample controller's enqueueFoo method
func (c *Barrelman) enqueue(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}
	c.workqueue.AddRateLimited(key)
}

// handleObject will take any resource implementing metav1.Object
// It then enqueues that deployment resource to be processed.
// adapted/stripped down from sample controller, most handling done in handlers
func (c *Barrelman) handleObject(obj interface{}) {
	var object metav1.Object
	var ok bool
	if object, ok = obj.(metav1.Object); !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("error decoding object, invalid type"))
			return
		}
		object, ok = tombstone.Obj.(metav1.Object)
		if !ok {
			runtime.HandleError(fmt.Errorf("error decoding object tombstone, invalid type"))
			return
		}
		glog.V(4).Infof("Recovered deleted object '%s' from tombstone", object.GetName())
	}
	glog.V(4).Infof("Processing object: %s", object.GetName())

	deployment, err := c.deploymentsLister.Deployments(object.GetNamespace()).Get(object.GetName())
	if err != nil {
		glog.V(4).Infof("ignoring orphaned object '%s'", object.GetSelfLink())
		return
	}
	c.enqueue(deployment)
	return
}
