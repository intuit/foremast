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
	"k8s.io/apimachinery/pkg/runtime/schema"
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

// DeploymentController is the controller implementation for watching deployment changes
type DeploymentController struct {
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

	barrelman *Barrelman
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

func (c *DeploymentController) monitorDeployment(appName string, oldDepl, newDepl *appsv1.Deployment, deploymentMetadata *v1alpha1.DeploymentMetadata) {
	//Try to get the metadata by "app" name, if it doesn't existing, try to search by "appType"
	deploymentMetadata, err := c.barrelman.getDeploymentMetadata(newDepl.Namespace, appName, newDepl)
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

			glog.Info("Starting to monitor new deployment...")
			go c.barrelman.monitorNewDeployment(appName, oldDepl, newDepl, deploymentMetadata, oldMonitor, monitorNotFound, m.StrategyRollingUpdate)

			c.handleObject(newDepl)
			return
		}
	}
}

// NewDeploymentController returns a new sample controller
func NewDeploymentController(
	kubeclientset kubernetes.Interface,
	foremastClientset clientset.Interface,
	deploymentInformer appsinformers.DeploymentInformer,
	barrelman *Barrelman) *DeploymentController {

	// Create event broadcaster
	glog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: FOREMAST})

	//lastDeployments := make(map[string]string)

	//glog.Info("Connecting to Prometheus client")
	//promClient, err := prometheusClient.NewClient(prometheusClient.Config{Address: "http://127.0.0.1:9090"})
	//if err != nil {
	//	glog.Infof("Failed to set up prometheus client: %v", err)
	//}
	//prometheus := prometheusv1.NewAPI(promClient)
	quit := make(chan bool)

	controller := &DeploymentController{
		kubeclientset:     kubeclientset,
		foremastClientset: foremastClientset,
		deploymentsLister: deploymentInformer.Lister(),
		deploymentsSynced: deploymentInformer.Informer().HasSynced,
		workqueue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "barrelman"),
		recorder:          recorder,
		quit:              quit,
		barrelman:         barrelman,
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

			deploymentMetadata, err := barrelman.getDeploymentMetadata(newDepl.Namespace, newApp, newDepl)
			if err != nil {
				return
			}

			var deploymentStrategy = ""
			if barrelman.hasHealthyMonitoring() {
				deploymentStrategy = m.StrategyRollingUpdate
			} else {
				deploymentStrategy = m.StrategyHpa
			}
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
			var hpaScoreTemplate = ""
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
				hpaScoreTemplate = oldMonitor.Spec.HpaScoreTemplate
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
				HpaScoreTemplate: hpaScoreTemplate,
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

	return controller
}

func (c *DeploymentController) isMonitoring(namespace string) bool {
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

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
// this is leftover from the sample controller, but seems to do a good job of setting things up
func (c *DeploymentController) Run(threadiness int, stopCh <-chan struct{}) error {
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
func (c *DeploymentController) runWorker() {
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
// left over from sample controller, still processes objects from queue but not doing much with them
func (c *DeploymentController) processNextWorkItem() bool {
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
func (c *DeploymentController) syncHandler(key string) error {
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
func (c *DeploymentController) enqueue(obj interface{}) {
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
func (c *DeploymentController) handleObject(obj interface{}) {
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
