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
					}
				} else {
					glog.V(10).Infof("There is no status change, skipping this event[%v:%v] new[%v:%v",
						oldMonitor.Name, newPhase, newMonitor.Name, oldPhase)
					//glog.Infof("No valid phase, skipping deployment monitor change %v", newMonitor.Name)
					return
				}
			}

			if newPhase == d.MonitorPhaseUnhealthy && !newMonitor.Status.RemediationTaken {
				if newMonitor.Spec.Remediation.Option == d.RemediationAutoRollback {
					newMonitor.Status.RemediationTaken = true
					controller.foremastClientset.DeploymentV1alpha1().DeploymentMonitors(newMonitor.Namespace).Update(newMonitor)

					go controller.rollback(*newMonitor)
				}
			}

			// Got a newPhase
			if newContinuous && newPhase != d.MonitorPhaseRunning { //Create a new continuous job
				go barrelman.monitorContinuously(newMonitor)
			}
		},
	})

	return controller
}

const AnnotationDeploymentRollbackMessage = "deployment.foremast.ai/rollbackMessage"

// rollback will take an old and a new deployment object, and restore the new
// deployment object to the same specification as the old deployment
func (c *MonitorController) rollback(monitor d.DeploymentMonitor) error {
	if monitor.Spec.RollbackRevision == 0 {
		return nil
	}

	var deploymentName = monitor.Annotations[DeploymentName]
	if deploymentName == "" {
		deploymentName = monitor.Name
	}

	depl, err := c.kubeclientset.ExtensionsV1beta1().Deployments(monitor.Namespace).Get(deploymentName, metav1.GetOptions{})

	if err != nil {
		return err
	}

	revision, err := deploymentutil.Revision(depl)
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

	glog.Infof("Rolling back to %d", monitor.Spec.RollbackRevision)

	// Update messages
	if _, err := c.kubeclientset.ExtensionsV1beta1().Deployments(depl.Namespace).Update(depl); err != nil {
		glog.Infof("Updating existing deployment error %v %v", depl, err)
		return err
	}

	//rollbacker, err := kubectl.RollbackerFor(schema.GroupKind{
	//	Kind:  "Deployment",
	//	Group: "apps",
	//}, c.kubeclientset)

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
			Revision: monitor.Spec.RollbackRevision,
		},
	}

	glog.Infof("Rolling back to %d", monitor.Spec.RollbackRevision)
	// Do the rollback
	if err := c.kubeclientset.ExtensionsV1beta1().Deployments(depl.Namespace).Rollback(deploymentRollback); err != nil {
		glog.Infof("Rolling back existing deployment error %v %v", depl, err)
		return err
	}

	glog.Infof("Rolled back to %d", monitor.Spec.RollbackRevision)
	//msg, err := rollbacker.Rollback(depl, nil, monitor.Spec.RollbackRevision, false)
	return err
}
