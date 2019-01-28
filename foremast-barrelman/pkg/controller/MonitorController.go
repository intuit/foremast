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
	d "foremast.ai/foremast/foremast-barrelman/pkg/apis/deployment/v1alpha1"
	informers "foremast.ai/foremast/foremast-barrelman/pkg/client/informers/externalversions/deployment/v1alpha1"
	"github.com/golang/glog"
	"k8s.io/api/extensions/v1beta1"
	"strconv"

	//appsv1 "k8s.io/api/apps/v1"
	clientset "foremast.ai/foremast/foremast-barrelman/pkg/client/clientset/versioned"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	deploymentutil "k8s.io/kubernetes/pkg/controller/deployment/util"
	"time"
)

const MonitorControllerName = "monitorController"

type MonitorController struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface

	foremastClientset clientset.Interface

	monitorInformer informers.DeploymentMonitorInformer
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder

	barrelman *Barrelman

	remediationOptions *RemediationOptions
}

// NewBarrelman returns a new sample controller
func NewController(kubeclientset kubernetes.Interface, foremastClientset clientset.Interface,
	monitorInformer informers.DeploymentMonitorInformer, barrelman *Barrelman) *MonitorController {

	// Create event broadcaster
	glog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: MonitorControllerName})

	controller := &MonitorController{
		kubeclientset:     kubeclientset,
		foremastClientset: foremastClientset,
		monitorInformer:   monitorInformer,
		recorder:          recorder,
		barrelman:         barrelman,
	}

	controller.remediationOptions = &RemediationOptions{
		rollback: controller.rollback,
		pause:    controller.pause,
		auto:     nil,
	}

	glog.Info("Setting up event handlers for DeploymentMonitor")

	monitorInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {

		},
		UpdateFunc: func(old, new interface{}) {
			//get old and new deployment objects
			newMonitor := new.(*d.DeploymentMonitor)
			oldMonitor := old.(*d.DeploymentMonitor)

			//skip if not marked to be tracked for ACA
			var newPhase = newMonitor.Status.Phase
			var oldPhase = ""

			var continuous = oldMonitor.Spec.Continuous
			var newContinuous = newMonitor.Spec.Continuous
			var continuousChange = continuous != newContinuous

			if newPhase == oldPhase {
				if continuousChange {
					if newContinuous && newPhase != d.MonitorPhaseRunning { //Create a new continuous job
						go barrelman.monitorContinuously(newMonitor)
						return
					}
				} else {
					glog.V(10).Infof("There is no status change, skipping this event[%v:%v] new[%v:%v",
						oldMonitor.Name, newPhase, newMonitor.Name, oldPhase)
					//glog.Infof("No valid phase, skipping deployment monitor change %v", newMonitor.Name)
					return
				}
			}

			if newPhase == d.MonitorPhaseUnhealthy && !newMonitor.Status.RemediationTaken {
				//if !newMonitor.Spec.Continuous {
				var action func(monitor *d.DeploymentMonitor) error
				switch newMonitor.Spec.Remediation.Option {
				case d.RemediationAutoRollback:
					action = controller.remediationOptions.rollback
					break
				case d.RemediationAutoPause:
					action = controller.remediationOptions.pause
					break
				case d.RemediationAuto:
					action = controller.remediationOptions.auto
					break
				}
				if action != nil {
					newMonitor.Status.RemediationTaken = true

					controller.foremastClientset.DeploymentV1alpha1().DeploymentMonitors(newMonitor.Namespace).Update(newMonitor)

					go action(newMonitor)
					return
				}
				//}
			}

			// Got a newPhase
			if newContinuous && newPhase != d.MonitorPhaseRunning { //Create a new continuous job
				if newPhase == d.MonitorPhaseUnhealthy {
					dura, err := time.Parse(time.RFC3339, newMonitor.Status.Timestamp)
					if err == nil && (time.Now().Unix()-dura.Unix()) > 60 { //Make sure the new job w
						go barrelman.monitorContinuously(newMonitor)
					}
				} else {
					go barrelman.monitorContinuously(newMonitor)
				}
			}
		},
	})

	return controller
}

//Trigger a rollback if error occured
//RemediationAutoRollback = "AutoRollback"
//Trigger a pause only to reduce the error rate
//RemediationAutoPause = "AutoPause"
//Let foremast take care everything for you
//RemediationAuto = "Auto"
type RemediationOptions struct {
	rollback func(monitor *d.DeploymentMonitor) error

	pause func(monitor *d.DeploymentMonitor) error

	auto func(monitor *d.DeploymentMonitor) error
}

const AnnotationDeploymentRollbackMessage = "deployment.foremast.ai/rollbackMessage"

// rollback will take an old and a new deployment object, and restore the new
// deployment object to the same specification as the old deployment
func (c *MonitorController) rollback(monitor *d.DeploymentMonitor) error {
	if monitor.Spec.RollbackRevision == 0 {
		return nil
	}

	var deploymentName = c.getDeploymentName(monitor)
	depl, err := c.kubeclientset.ExtensionsV1beta1().Deployments(monitor.Namespace).Get(deploymentName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	revision, err := deploymentutil.Revision(depl)
	var rollbackRevision = monitor.Spec.RollbackRevision
	if revision == monitor.Spec.RollbackRevision {
		glog.Infof("Rolled back already %d", monitor.Spec.RollbackRevision)
		return nil
	}

	var now = metav1.Time{
		Time: time.Now(),
	}

	depl.Status.Conditions = append(depl.Status.Conditions, v1beta1.DeploymentCondition{
		Type:               v1beta1.DeploymentProgressing,
		Status:             "True",
		LastUpdateTime:     now,
		LastTransitionTime: now,
		Reason:             "RollbackProgressing",
		Message:            "Foremast detected unhealthy, so roll it back automatically to revision:" + strconv.FormatInt(monitor.Spec.RollbackRevision, 10),
	})

	glog.Infof("Rolling back to %d", rollbackRevision)

	// Update messages
	if _, err := c.kubeclientset.ExtensionsV1beta1().Deployments(depl.Namespace).Update(depl); err != nil {
		glog.Infof("Updating existing deployment error %v %v", depl, err)
	}

	if depl.Spec.Paused {
		return fmt.Errorf("you cannot rollback a paused deployment; resume it first with 'kubectl rollout resume deployment/%s' and try again", deploymentName)
	}
	//TODO use the following code for now, since the kubectl.rollback has bug, in version 1.10.x
	deploymentRollback := &v1beta1.DeploymentRollback{
		Name: deploymentName,
		UpdatedAnnotations: map[string]string{
			AnnotationDeploymentRollbackMessage: "Foremast detected unhealthy, so roll it back automatically to revision:" + strconv.FormatInt(monitor.Spec.RollbackRevision, 10),
		},
		RollbackTo: v1beta1.RollbackConfig{
			Revision: rollbackRevision,
		},
	}

	glog.Infof("Rolling back to %d", rollbackRevision)
	// Do the rollback
	if err := c.kubeclientset.ExtensionsV1beta1().Deployments(depl.Namespace).Rollback(deploymentRollback); err != nil {
		glog.Infof("Rolling back existing deployment error %v %v", depl, err)
		return err
	}

	glog.Infof("Rolled back to %d", rollbackRevision)
	//rollbacker, err := kubectl.RollbackerFor(schema.GroupKind{
	//	Kind:  "Deployment",
	//	Group: "apps",
	//}, c.kubeclientset)
	//msg, err := rollbacker.Rollback(depl, nil, monitor.Spec.RollbackRevision, false)
	return err
}

func (c *MonitorController) getDeploymentName(monitor *d.DeploymentMonitor) string {
	var deploymentName = monitor.Annotations[DeploymentName]
	if deploymentName == "" {
		deploymentName = monitor.Name
	}
	return deploymentName
}

func (c *MonitorController) getDeployment(monitor *d.DeploymentMonitor) (*v1beta1.Deployment, error) {
	var deploymentName = c.getDeploymentName(monitor)
	return c.kubeclientset.ExtensionsV1beta1().Deployments(monitor.Namespace).Get(deploymentName, metav1.GetOptions{})
}

// pause the current deployment
func (c *MonitorController) pause(monitor *d.DeploymentMonitor) error {
	depl, err := c.getDeployment(monitor)
	if err != nil {
		return err
	}

	depl.Spec.Paused = true

	var now = metav1.Time{
		Time: time.Now(),
	}

	depl.Status.Conditions = append(depl.Status.Conditions, v1beta1.DeploymentCondition{
		Type:               v1beta1.DeploymentProgressing,
		Status:             "True",
		LastUpdateTime:     now,
		LastTransitionTime: now,
		Reason:             "ForemastPaused",
		Message:            "Foremast detected unhealthy, so paused this deployment",
	})

	// Update messages
	if _, err := c.kubeclientset.ExtensionsV1beta1().Deployments(depl.Namespace).Update(depl); err != nil {
		glog.Infof("Updating existing deployment error %v %v", depl, err)
		return err
	}
	return nil
}

func (c *MonitorController) auto(monitor *d.DeploymentMonitor) error {
	//TODO Determine by foremast
	return nil
}
