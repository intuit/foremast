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
	d "foremast.ai/foremast/foremast-barrelman/pkg/apis/deployment/v1alpha1"
)

// DeploymentMonitorEventHandler can handle notifications for events that happen to a
// change on Deployment. The events are informational only, so you can't return an
// error.
//  * OnAdd is called when an object is added.
//  * OnUpdate is called when an object is modified. Note that oldObj is the
//      last known state of the object-- it is possible that several changes
//      were combined together, so you can't use this to see every single
//      change. OnUpdate is also called when a re-list happens, and it will
//      get called even if nothing changed. This is useful for periodically
//      evaluating or syncing something.
//  * OnDelete will get the final state of the item if it is known, otherwise
//      it will get an object of type DeletedFinalStateUnknown. This can
//      happen if the watch is closed and misses the delete event and we don't
//      notice the deletion until the subsequent re-list.

type DeploymentMonitorEventHandler interface {
	OnAdd(obj *d.DeploymentMonitor)
	OnUpdate(oldObj, newObj *d.DeploymentMonitor)
	OnDelete(obj *d.DeploymentMonitor)
}

type DeploymentMonitorEventFuns struct {
	addFunc func(obj *d.DeploymentMonitor)

	updateFunc func(oldObj, newObj *d.DeploymentMonitor)

	deleteFunc func(obj *d.DeploymentMonitor)
}
