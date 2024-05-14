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
	"context"
	"time"

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
	"k8s.io/klog/v2"

	"github.com/kubernetes-csi/external-health-monitor/pkg/util"
)

const (
	// DefaultNodeNotReadyTimeDuration is the default time interval we need to consider node broken if it keeps NotReady
	DefaultNodeNotReadyTimeDuration = 5 * time.Minute
)

// NodeWatcher watches nodes conditions
type NodeWatcher struct {
	driverName string
	client     kubernetes.Interface
	recorder   record.EventRecorder

	nodeQueue workqueue.Interface

	nodeLister       corelisters.NodeLister
	nodeListerSynced cache.InformerSynced

	volumeLister corelisters.PersistentVolumeLister
	pvcLister    corelisters.PersistentVolumeClaimLister

	// mark the time when we first found the node is broken
	nodeFirstBrokenMap map[string]time.Time

	// nodeEverMarkedDown stores all nodes which are marked down
	// if nodes recover, they will be removed from here
	nodeEverMarkedDown map[string]bool

	// pvcToPodsCache stores PVC/Pods mapping info, we can get all pods using one specific PVC more efficiently by this
	pvcToPodsCache *util.PVCToPodsCache

	// Time interval for executing node worker goroutines
	nodeWorkerExecuteInterval time.Duration
	// Time interval for listing nodess and add them to queue
	nodeListAndAddInterval time.Duration
}

// NewNodeWatcher creates a node watcher object that will watch the nodes
func NewNodeWatcher(
	logger klog.Logger,
	driverName string,
	client kubernetes.Interface,
	volumeLister corelisters.PersistentVolumeLister,
	pvcLister corelisters.PersistentVolumeClaimLister,
	nodeInformer coreinformers.NodeInformer,
	recorder record.EventRecorder,
	pvcToPodsCache *util.PVCToPodsCache,
	nodeWorkerExecuteInterval time.Duration,
	nodeListAndAddInterval time.Duration,
) *NodeWatcher {

	watcher := &NodeWatcher{
		driverName:                driverName,
		nodeWorkerExecuteInterval: nodeWorkerExecuteInterval,
		nodeListAndAddInterval:    nodeListAndAddInterval,
		client:                    client,
		recorder:                  recorder,
		volumeLister:              volumeLister,
		pvcLister:                 pvcLister,
		nodeQueue:                 workqueue.NewNamed("nodes"),
		nodeFirstBrokenMap:        make(map[string]time.Time),
		nodeEverMarkedDown:        make(map[string]bool),
		pvcToPodsCache:            pvcToPodsCache,
	}

	nodeInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) { watcher.enqueueWork(logger, obj) },
			UpdateFunc: func(oldObj, newObj interface{}) {
				watcher.enqueueWork(logger, newObj)
			},
			DeleteFunc: func(obj interface{}) {
				watcher.enqueueWork(logger, obj)
			},
		},
	)
	watcher.nodeLister = nodeInformer.Lister()
	watcher.nodeListerSynced = nodeInformer.Informer().HasSynced

	return watcher
}

// enqueueWork adds node to given work queue.
func (watcher *NodeWatcher) enqueueWork(logger klog.Logger, obj interface{}) {
	// Beware of "xxx deleted" events
	if unknown, ok := obj.(cache.DeletedFinalStateUnknown); ok && unknown.Obj != nil {
		obj = unknown.Obj
	}
	objName, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		logger.Error(err, "Failed to get key from object")
		return
	}
	logger.V(6).Info("Enqueued ObjectName for sync", "objectName", objName)
	watcher.nodeQueue.Add(objName)
}

// addNodesToQueue adds all existing nodes to queue periodically
func (watcher *NodeWatcher) addNodesToQueue(ctx context.Context) {
	logger := klog.FromContext(ctx)
	logger.V(4).Info("Resyncing Node watcher")

	nodes, err := watcher.nodeLister.List(labels.NewSelector())
	if err != nil {
		logger.Info("Cannot list nodes", "err", err)
		return
	}
	for _, node := range nodes {
		watcher.enqueueWork(logger, node)
	}
}

// Run starts all of this controller's control loops
func (watcher *NodeWatcher) Run(ctx context.Context) {
	logger := klog.FromContext(ctx)
	defer watcher.nodeQueue.ShutDown()
	if !cache.WaitForCacheSync(ctx.Done(), watcher.nodeListerSynced) {
		logger.Error(nil, "Cannot sync cache")
		return
	}

	go wait.UntilWithContext(ctx, watcher.addNodesToQueue, watcher.nodeListAndAddInterval)
	go wait.UntilWithContext(ctx, watcher.WatchNodes, watcher.nodeWorkerExecuteInterval)
	<-ctx.Done()
}

// WatchNodes periodically checks if nodes break down
func (watcher *NodeWatcher) WatchNodes(ctx context.Context) {
	logger := klog.FromContext(ctx)
	workFunc := func() bool {
		keyObj, quit := watcher.nodeQueue.Get()
		if quit {
			return true
		}
		defer watcher.nodeQueue.Done(keyObj)
		key := keyObj.(string)
		logger.V(4).Info("WatchNode", "node", key)

		_, name, err := cache.SplitMetaNamespaceKey(key)
		if err != nil {
			logger.Error(err, "Error getting name of node from informer", "node", key)
			return false
		}
		node, err := watcher.nodeLister.Get(name)
		if err == nil {
			// The node still exists in informer cache, the event must have
			// been add/update/sync
			watcher.updateNode(logger, node)
			return false
		}
		if !errors.IsNotFound(err) {
			logger.V(2).Info("Error getting node from informer", "node", key, "err", err)
			return false
		}

		// The node is not in informer cache, the event must be "delete"
		watcher.deleteNode(logger, node)
		return false
	}
	for {
		if quit := workFunc(); quit {
			logger.Info("Volume worker queue shutting down")
			return
		}
	}
}

func (watcher *NodeWatcher) updateNode(logger klog.Logger, node *v1.Node) {
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
			err := watcher.cleanNodeFailureConditionForPVC(logger, node)
			if err == nil {
				// when node recovers and send recovery event successfully, remove the node from the map
				delete(watcher.nodeEverMarkedDown, node.Name)
			} else {
				logger.Error(err, "Clean node failure message error")
			}
		}
		return
	}

	if watcher.isNodeBroken(logger, node) {
		logger.Info("Node is broken", "node", node.Name)
		// mark all PVCs/Pods on this node
		err := watcher.markPVCsAndPodsOnUnhealthyNode(logger, node)
		if err != nil {
			logger.Error(err, "Mark PVCs on not ready node failed, re-enqueue")
			// if error happened, re-enqueue
			watcher.enqueueWork(logger, node)
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

func (watcher *NodeWatcher) isNodeBroken(logger klog.Logger, node *v1.Node) bool {
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
				}
				logger.V(6).Info("Node is not ready, but less than 5 minutes", "node", node.Name)
				return false
			}

			// first time to mark the node NotReady
			watcher.nodeFirstBrokenMap[objName] = now
			return false
		}
	}

	return false
}

func (watcher *NodeWatcher) deleteNode(logger klog.Logger, node *v1.Node) {
	logger.Info("Node is deleted, so mark the PVs on the node", "node", node.Name)

	// mark all PVs on this node
	err := watcher.markPVCsAndPodsOnUnhealthyNode(logger, node)
	if err != nil {
		logger.Error(err, "Marking PVs failed")
		// must re-enqueue here, because we can not get this from informer(node-lister) any more
		watcher.enqueueWork(logger, node)
	}
}

func (watcher *NodeWatcher) cleanNodeFailureConditionForPVC(logger klog.Logger, node *v1.Node) error {
	pvs, err := watcher.volumeLister.List(labels.NewSelector())
	if err != nil {
		logger.Info("Cannot list pvs", "err", err)
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
		if len(pods) == 0 {
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

		// TODO: add events to Pods instead
		pvc, err := watcher.pvcLister.PersistentVolumeClaims(pv.Spec.ClaimRef.Namespace).Get(pv.Spec.ClaimRef.Name)
		if err != nil {
			logger.Error(err, "Get PVC from PVC lister error", "pvc", klog.KRef(pv.Spec.ClaimRef.Namespace, pv.Spec.ClaimRef.Name))
			return err
		}

		pvcClone := pvc.DeepCopy()
		message := "Node: " + node.Name + " recovered"
		watcher.recorder.Event(pvcClone, v1.EventTypeWarning, "NodeRecovered", message)

	}
	return nil
}

func (watcher *NodeWatcher) markPVCsAndPodsOnUnhealthyNode(logger klog.Logger, node *v1.Node) error {
	pvs, err := watcher.volumeLister.List(labels.NewSelector())
	if err != nil {
		logger.Info("Cannot list pvs", "err", err)
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
		if len(pods) == 0 {
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
			logger.Error(err, "Get PVC from PVC lister error", "pvc", klog.KRef(pv.Spec.ClaimRef.Namespace, pv.Spec.ClaimRef.Name))
			return err
		}

		// TODO: add events to Pods instead
		pvcClone := pvc.DeepCopy()

		message := "Pods: [ "
		for _, pod := range podsOnThatNode {
			message = message + pod.Name + " "
		}
		message += "]" + " consuming PVC: " + pvcClone.Name + " in namespace: " + pvcClone.Namespace + " are now on a failed node: " + node.Name

		watcher.recorder.Event(pvcClone, v1.EventTypeWarning, "NodeFailed", message)
	}
	return nil
}
