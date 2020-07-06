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

package pv_monitor_agent

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog"
)

func (agent *PVMonitorAgent) addPodToQueue(pod *v1.Pod) {
	if len(pod.Spec.NodeName) != 0 && pod.Spec.NodeName == agent.nodeName {
		// TODO: support inline csi volumes
		for _, volume := range pod.Spec.Volumes {
			if volume.PersistentVolumeClaim != nil {
				pvc, err := agent.pvcLister.PersistentVolumeClaims(pod.Namespace).Get(volume.PersistentVolumeClaim.ClaimName)
				if err != nil || pvc.Status.Phase != v1.ClaimBound {
					klog.Warningf("get pvc error or pvc is not bound. pvc: %s", pod.Namespace+"/"+volume.PersistentVolumeClaim.ClaimName)
					continue
				}

				pv, err := agent.pvLister.Get(pvc.Spec.VolumeName)
				if err != nil {
					klog.Warningf("get pv: %s error", pvc.Spec.VolumeName)
					continue
				}

				if pv.Spec.CSI != nil && pv.Spec.CSI.Driver == agent.driverName {
					item := &PodWithPVItem{
						podNameSpace: pod.Namespace,
						podName:      pod.Name,
						pvName:       pv.Name,
					}
					agent.podWithPVItemQueue.Add(item)
				}
			}
		}
	}
}

func (agent *PVMonitorAgent) podAdded(obj interface{}) {
	// If the NodeName is set, handle the pod ADD events so that the cache can recover from agent crash
	agent.addPodToQueue(obj.(*v1.Pod))
}

func (agent *PVMonitorAgent) podUpdated(oldObj, newObj interface{}) {
	oldPod := oldObj.(*v1.Pod)
	newPod := newObj.(*v1.Pod)

	if newPod.DeletionTimestamp != nil {
		// pod is being deleted, skip it
		// do not delete item from queue because the item should be deleted in Pod Deletion event handler function
		return
	}

	// pod is bound to this node
	if len(oldPod.Spec.NodeName) == 0 && len(newPod.Spec.NodeName) != 0 && newPod.Spec.NodeName == agent.nodeName {
		agent.addPodToQueue(newPod)
	}

}
