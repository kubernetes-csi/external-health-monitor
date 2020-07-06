/*
Copyright 2020 The Kubernetes Authors.

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

package pv_monitor_controller

import (
	v1 "k8s.io/api/core/v1"
)

func (ctrl *PVMonitorController) pvAdded(obj interface{}) {
	pv := obj.(*v1.PersistentVolume)
	if pv.Status.Phase != v1.VolumeBound || pv.Spec.CSI == nil || pv.Spec.CSI.Driver != ctrl.driverName {
		return
	}

	ctrl.Lock()
	defer ctrl.Unlock()

	ctrl.pvQueue.Add(pv.Name)
	ctrl.pvEnqueued[pv.Name] = true
}

func (ctrl *PVMonitorController) podAdded(obj interface{}) {
	ctrl.pvcToPodsCache.AddPod(obj.(*v1.Pod))
}

func (ctrl *PVMonitorController) podDeleted(obj interface{}) {
	ctrl.pvcToPodsCache.DeletePod(obj.(*v1.Pod))
}
