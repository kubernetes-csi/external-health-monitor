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
	"time"

	"k8s.io/klog"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	coreinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"

	"github.com/kubernetes-csi/external-health-monitor/pkg/util"
)

const (
	// DefaultWatcherResyncPeriod is the resync period of node watcher
	DefaultWatcherResyncPeriod = 30 * time.Minute

	// DefaultNodeNotReadyTimeDuration is the default time interval we need to consider node broken if it keeps NotReady
	DefaultNodeNotReadyTimeDuration = 300 * time.Second
)

// marking event related const vars
const (
	UpdatePVRetryCount = 5

	UpdatePVInterval = 5 * time.Second
)

// NodeWatcher watches nodes conditions
type NodeWatcher struct {
	driverName string

	monitorInterval time.Duration
	client          kubernetes.Interface
	recorder        record.EventRecorder

	nodeQueue workqueue.Interface

	nodeLister       corelisters.NodeLister
	nodeListerSynced cache.InformerSynced

	volumeLister corelisters.PersistentVolumeLister
	pvcLister    corelisters.PersistentVolumeClaimLister

	// mark the time when we first found the node is broken
	nodeFirstBrokenMap map[string]time.Time

	nodeEverMarkedDown map[string]bool

	pvcToPodsCache *util.PVCToPodsCache
}

// NewNodeWatcher creates a node watcher object that will watch the nodes
func NewNodeWatcher(driverName string, client kubernetes.Interface, volumeLister corelisters.PersistentVolumeLister,
	pvcLister corelisters.PersistentVolumeClaimLister, nodeInformer coreinformers.NodeInformer,
	recorder record.EventRecorder, pvcToPodsCache *util.PVCToPodsCache, monitorInterval time.Duration) *NodeWatcher {

	watcher := &NodeWatcher{
		driverName:         driverName,
		monitorInterval:    monitorInterval,
		client:             client,
		recorder:           recorder,
		volumeLister:       volumeLister,
		pvcLister:          pvcLister,
		nodeQueue:          workqueue.NewNamed("nodes"),
		nodeFirstBrokenMap: make(map[string]time.Time),
		nodeEverMarkedDown: make(map[string]bool),
		pvcToPodsCache:     pvcToPodsCache,
	}

	nodeInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) { watcher.enqueueWork(obj) },
			UpdateFunc: func(oldObj, newObj interface{}) {
				watcher.enqueueWork(newObj)
			},
			DeleteFunc: func(obj interface{}) {
				watcher.enqueueWork(obj)
			},
		},
	)
	watcher.nodeLister = nodeInformer.Lister()
	watcher.nodeListerSynced = nodeInformer.Informer().HasSynced

	return watcher
}

// enqueueWork adds node to given work queue.
func (watcher *NodeWatcher) enqueueWork(obj interface{}) {
	// Beware of "xxx deleted" events
	if unknown, ok := obj.(cache.DeletedFinalStateUnknown); ok && unknown.Obj != nil {
		obj = unknown.Obj
	}
	objName, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		klog.Errorf("failed to get key from object: %v", err)
		return
	}
	klog.V(6).Infof("enqueued %q for sync", objName)
	watcher.nodeQueue.Add(objName)
}

// resync supplements short resync period of shared informers - we don't want
// all consumers of Node shared informer to have a short resync period,
// therefore we do our own.
func (watcher *NodeWatcher) resync() {
	klog.V(4).Infof("resyncing Node watcher")

	nodes, err := watcher.nodeLister.List(labels.NewSelector())
	if err != nil {
		klog.Warningf("cannot list nodes: %s", err)
		return
	}
	for _, node := range nodes {
		watcher.enqueueWork(node)
	}
}

// Run starts all of this controller's control loops
func (watcher *NodeWatcher) Run(stopCh <-chan struct{}) {
	defer watcher.nodeQueue.ShutDown()
	if !cache.WaitForCacheSync(stopCh, watcher.nodeListerSynced) {
		klog.Errorf("Cannot sync caches")
		return
	}

	//go watcher.WatchNodes()
	go wait.Until(watcher.resync, DefaultWatcherResyncPeriod, stopCh) // TODO: do we need this?
	go wait.Until(watcher.WatchNodes, watcher.monitorInterval, stopCh)
	<-stopCh
}

// WatchNodes periodically checks if nodes break down
func (watcher *NodeWatcher) WatchNodes() {
	workFunc := func() bool {
		keyObj, quit := watcher.nodeQueue.Get()
		if quit {
			return true
		}
		defer watcher.nodeQueue.Done(keyObj)
		key := keyObj.(string)
		klog.V(4).Infof("WatchNode: %s", key)

		_, name, err := cache.SplitMetaNamespaceKey(key)
		if err != nil {
			klog.Errorf("error getting name of node %q from informer: %v", key, err)
			return false
		}
		node, err := watcher.nodeLister.Get(name)
		if err == nil {
			// The node still exists in informer cache, the event must have
			// been add/update/sync
			watcher.updateNode(key, node)
			return false
		}
		if !errors.IsNotFound(err) {
			klog.V(2).Infof("error getting node %q from informer: %v", key, err)
			return false
		}

		// The node is not in informer cache, the event must be
		// "delete"
		watcher.deleteNode(key, node)
		return false
	}
	for {
		if quit := workFunc(); quit {
			klog.Infof("volume worker queue shutting down")
			return
		}
	}
}

func (watcher *NodeWatcher) updateNode(key string, node *v1.Node) {
	// TODO: if node is ready, check if node was ever marked down, if yes, reset it
	if watcher.isNodeReady(node) {
		// The node status is ok, but if it was marked before, remove the mark
		_, ok := watcher.nodeFirstBrokenMap[node.Name]
		if ok {
			delete(watcher.nodeFirstBrokenMap, node.Name)
		}

		// if the node was ever marked down, reset PVCs status on it
		if watcher.nodeEverMarkedDown[node.Name] {
			// TODO: reset PVCs status on the node
			err := watcher.cleanNodeFailureConditionForPVC(node)
			if err == nil {
				// when node recovers and send recovery event successfully, remove the node from the map
				delete(watcher.nodeEverMarkedDown, node.Name)
			} else {
				klog.Errorf("clean node failure message error: %+v", err)
			}
		}

		return
	}

	// TODO: need to revisit this later
	if watcher.isNodeBroken(node) {
		klog.Infof("node: %s is broken", node.Name)
		// mark all PVCs/Pods on this node
		// try several times again
		var err error
		for i := 0; i < UpdatePVRetryCount; i++ {
			err = watcher.markPVCsAndPodsOnUnhealthyNode(node)
			if err != nil {
				time.Sleep(UpdatePVInterval)
				continue
			}
			break
		}
		if err != nil {
			klog.Infof("mark PVCs on not ready node failed, re-enqueue")
			// if error happened, re-enqueue
			watcher.enqueueWork(node)
			return
		}

		// node is broken and PVCs on it are marked, remove it from map
		_, ok := watcher.nodeFirstBrokenMap[node.Name]
		if ok {
			delete(watcher.nodeFirstBrokenMap, node.Name)
		}

		watcher.nodeEverMarkedDown[node.Name] = true
	}

}

func (watcher *NodeWatcher) isNodeReady(node *v1.Node) bool {
	for _, condition := range node.Status.Conditions {
		if condition.Type == v1.NodeReady && condition.Status == v1.ConditionTrue {
			return true
		}
	}

	return false
}

func (watcher *NodeWatcher) isNodeBroken(node *v1.Node) bool {
	if node.Status.Phase == v1.NodeTerminated {
		return true
	}
	objName := node.Name
	for _, condition := range node.Status.Conditions {
		if condition.Type == v1.NodeReady && condition.Status != v1.ConditionTrue {
			now := time.Now()
			firstMarkTime, ok := watcher.nodeFirstBrokenMap[objName]
			if ok {
				timeInterval := now.Sub(firstMarkTime)
				if timeInterval.Seconds() > DefaultNodeNotReadyTimeDuration.Seconds() {
					return true
				} else {
					klog.V(6).Infof("node:%s is not ready, but less than 5 minutes, re-enqueue", node.Name)
					// NotReady status lasts less than 5 minutes
					// re-enqueue
					watcher.enqueueWork(node)
					return false
				}
			} else {
				// first time to mark the node NotReady
				watcher.nodeFirstBrokenMap[objName] = now
				watcher.enqueueWork(node)
				return false
			}
		}
	}

	return false
}

func (watcher *NodeWatcher) deleteNode(key string, node *v1.Node) {
	klog.Infof("node:%s is deleted, so mark the local PVs on it", node.Name)

	// mark all local PVs on this node
	// try several times again
	for i := 0; i < UpdatePVRetryCount; i++ {
		err := watcher.markPVCsAndPodsOnUnhealthyNode(node)
		if err != nil {
			klog.V(4).Infof("marking local PVs failed: %v", err)
			time.Sleep(UpdatePVInterval)
			continue
		}
		return
	}

	// if we reach here, it means that marking PVCs/Pods failed, re-enqueue
	watcher.enqueueWork(node)
}

func (watcher *NodeWatcher) cleanNodeFailureConditionForPVC(node *v1.Node) error {
	pvs, err := watcher.volumeLister.List(labels.NewSelector())
	if err != nil {
		klog.Warningf("cannot list pvs: %s", err)
		return err
	}

	for _, pv := range pvs {
		if pv.Spec.CSI == nil || pv.Spec.CSI.Driver != watcher.driverName {
			continue
		}

		if pv.Status.Phase != v1.VolumeBound {
			continue
		}

		pods := watcher.pvcToPodsCache.GetPodsByPVC(pv.Spec.ClaimRef.Namespace, pv.Spec.ClaimRef.Name)
		if pods == nil || len(pods) == 0 {
			continue
		}

		podsOnThatNode := make([]*v1.Pod, 0)
		for _, pod := range pods {
			if pod.Spec.NodeName == node.Name {
				podsOnThatNode = append(podsOnThatNode, pod)
			}
		}
		if len(podsOnThatNode) == 0 {
			continue
		}

		pvc, err := watcher.pvcLister.PersistentVolumeClaims(pv.Spec.ClaimRef.Namespace).Get(pv.Spec.ClaimRef.Name)
		if err != nil {
			klog.Errorf("get PVC[%s] from PVC lister error: %+v", pv.Spec.ClaimRef.Namespace+"/"+pv.Spec.ClaimRef.Name, err)
			return err
		}

		pvcClone := pvc.DeepCopy()

		message := "Node: " + node.Name + " recovered"
		watcher.recorder.Event(pvcClone, v1.EventTypeWarning, "NodeRecovered", message)

		// TODO: handle PVC conditions for the node recovery?
		/*nodeFailureExist := false
		for _, pvcCondition := range pvcClone.Status.Conditions {
			if pvcCondition.Type == util.NodeFailed + "d" {
				pvcCondition.Message = "Node:"
				pvcCondition.Type = util.NodeRecovered
				nodeFailureExist = true
				break
			}
		}

		if !nodeFailureExist {
			continue
		}

		_, err = watcher.client.CoreV1().PersistentVolumeClaims(pvcClone.Namespace).UpdateStatus(pvcClone)
		if err != nil {
			klog.Errorf("update PVC[%s] error: %+v", pvcClone.Namespace + "/" + pvcClone.Name, err)
			return err
		}*/

	}
	return nil
}

func (watcher *NodeWatcher) markPVCsAndPodsOnUnhealthyNode(node *v1.Node) error {
	pvs, err := watcher.volumeLister.List(labels.NewSelector())
	if err != nil {
		klog.Warningf("cannot list pvs: %s", err)
		return err
	}

	for _, pv := range pvs {
		if pv.Spec.CSI == nil || pv.Spec.CSI.Driver != watcher.driverName {
			continue
		}

		if pv.Status.Phase != v1.VolumeBound {
			continue
		}

		pods := watcher.pvcToPodsCache.GetPodsByPVC(pv.Spec.ClaimRef.Namespace, pv.Spec.ClaimRef.Name)
		if pods == nil || len(pods) == 0 {
			continue
		}

		podsOnThatNode := make([]*v1.Pod, 0)
		for _, pod := range pods {
			if pod.Spec.NodeName == node.Name {
				podsOnThatNode = append(podsOnThatNode, pod)
			}
		}
		if len(podsOnThatNode) == 0 {
			continue
		}

		pvc, err := watcher.pvcLister.PersistentVolumeClaims(pv.Spec.ClaimRef.Namespace).Get(pv.Spec.ClaimRef.Name)
		if err != nil {
			klog.Errorf("get PVC[%s] from PVC lister error: %+v", pv.Spec.ClaimRef.Namespace+"/"+pv.Spec.ClaimRef.Name, err)
			return err
		}

		pvcClone := pvc.DeepCopy()

		message := "Pods: [ "
		for _, pod := range podsOnThatNode {
			message = message + pod.Name + " "
		}
		message += "]" + " consuming PVC: " + pvcClone.Name + " in namespace: " + pvcClone.Namespace + " are now on a failed node: " + node.Name

		watcher.recorder.Event(pvcClone, v1.EventTypeWarning, "NodeFailed", message)

		// TODO: store the message in PVC conditions ?
		/*nodeFailureExist := false
		for _, pvcCondition := range pvcClone.Status.Conditions {
			if pvcCondition.Type == util.NodeFailed || pvcCondition.Type == util.NodeRecovered {
				pvcCondition.Message = message
				nodeFailureExist = true
				break
			}
		}

		if !nodeFailureExist {
			condition := v1.PersistentVolumeClaimCondition{
				Type:util.NodeFailed,
				Message:message,
			}
			pvcClone.Status.Conditions = append(pvcClone.Status.Conditions, condition)
		}

		_, err = watcher.client.CoreV1().PersistentVolumeClaims(pvcClone.Namespace).UpdateStatus(pvcClone)
		if err != nil {
			klog.Errorf("update PVC[%s] error: %+v", pvcClone.Namespace + "/" + pvcClone.Name, err)
			return err
		}*/

	}
	return nil
}
