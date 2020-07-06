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

package util

import (
	"sync"

	v1 "k8s.io/api/core/v1"
)

// PodSet is a pod map, whose key is podname + "/" + podname
type PodSet map[string]*v1.Pod

// PVCToPodsCache stores PVCs/Pods mapping info
// we can get all pods using one specific PVC more efficiently by this
type PVCToPodsCache struct {
	sync.Mutex

	// caches pvc/pods mapping info, key is pvc namespace + pvc name
	pvcToPodsMap map[string]PodSet
}

// NewPVCToPodsCache creates a new PVCToPodsCache
func NewPVCToPodsCache() *PVCToPodsCache {
	return &PVCToPodsCache{
		pvcToPodsMap: make(map[string]PodSet),
	}
}

// TODO: support inline csi volumes
func (cache *PVCToPodsCache) AddPod(pod *v1.Pod) {
	cache.Lock()
	defer cache.Unlock()

	for _, volume := range pod.Spec.Volumes {
		if volume.PersistentVolumeClaim != nil {
			if cache.pvcToPodsMap[pod.Namespace+"/"+volume.PersistentVolumeClaim.ClaimName] == nil {
				cache.pvcToPodsMap[pod.Namespace+"/"+volume.PersistentVolumeClaim.ClaimName] = make(PodSet)
			}
			cache.pvcToPodsMap[pod.Namespace+"/"+volume.PersistentVolumeClaim.ClaimName][pod.Namespace+"/"+pod.Name] = pod
		}
	}
}

// TODO: support inline csi volumes
func (cache *PVCToPodsCache) DeletePod(pod *v1.Pod) {
	cache.Lock()
	defer cache.Unlock()

	for _, volume := range pod.Spec.Volumes {
		if volume.PersistentVolumeClaim != nil {
			if cache.pvcToPodsMap[pod.Namespace+"/"+volume.PersistentVolumeClaim.ClaimName] == nil {
				return
			}
			delete(cache.pvcToPodsMap[pod.Namespace+"/"+volume.PersistentVolumeClaim.ClaimName], pod.Namespace+"/"+pod.Name)
			if len(cache.pvcToPodsMap[pod.Namespace+"/"+volume.PersistentVolumeClaim.ClaimName]) == 0 {
				delete(cache.pvcToPodsMap, pod.Namespace+"/"+volume.PersistentVolumeClaim.ClaimName)
			}
		}
	}
}

func (cache *PVCToPodsCache) GetPodsByPVC(pvcNamespace, pvcName string) PodSet {
	cache.Lock()
	defer cache.Unlock()

	return cache.pvcToPodsMap[pvcNamespace+"/"+pvcName]
}
