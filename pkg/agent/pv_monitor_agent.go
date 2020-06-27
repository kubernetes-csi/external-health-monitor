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
	"fmt"
	"time"

	"k8s.io/klog"

	"google.golang.org/grpc"

	v1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	coreinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"

	handler "github.com/kubernetes-csi/external-health-monitor/pkg/csi-handler"
	"github.com/kubernetes-csi/external-health-monitor/pkg/util"
)

// PodWithPVItem contains pv name as well as pod information the pv is attached to
type PodWithPVItem struct {
	podNameSpace string
	podName      string
	pvName       string
}

// PVMonitorAgent is the struct of pv monitor agent containing all information to perform volumes health condition checking
type PVMonitorAgent struct {
	client          kubernetes.Interface
	driverName      string
	nodeName        string
	monitorInterval time.Duration
	eventRecorder   record.EventRecorder

	csiConn *grpc.ClientConn

	pvcToPodsCache *util.PVCToPodsCache

	pvLister       corelisters.PersistentVolumeLister
	pvListerSynced cache.InformerSynced

	pvcLister       corelisters.PersistentVolumeClaimLister
	pvcListerSynced cache.InformerSynced

	podLister       corelisters.PodLister
	podListerSynced cache.InformerSynced

	supportStageUnstage bool
	kubeletRootPath     string
	pvChecker           *handler.PVHealthConditionChecker

	podWithPVItemQueue workqueue.Interface
}

// NewPVMonitorAgent create pv monitor agent
func NewPVMonitorAgent(client kubernetes.Interface, driverName string, conn *grpc.ClientConn, timeout time.Duration, monitorInterval time.Duration, pvInformer coreinformers.PersistentVolumeInformer,
	pvcInformer coreinformers.PersistentVolumeClaimInformer, podInformer coreinformers.PodInformer, supportStageUnstage bool, kubeletRootPath string) *PVMonitorAgent {
	broadcaster := record.NewBroadcaster()
	broadcaster.StartRecordingToSink(&corev1.EventSinkImpl{Interface: client.CoreV1().Events(v1.NamespaceAll)})
	var eventRecorder record.EventRecorder
	eventRecorder = broadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: fmt.Sprintf("csi-pv-monitor-controller-%s", driverName)})

	agent := &PVMonitorAgent{
		supportStageUnstage: supportStageUnstage,
		kubeletRootPath:     kubeletRootPath,
		csiConn:             conn,
		eventRecorder:       eventRecorder,
		client:              client,
		driverName:          driverName,
		monitorInterval:     monitorInterval,
		podWithPVItemQueue:  workqueue.NewNamed("csi-monitor-pod-pv-queue"),
	}

	// PV lister
	agent.pvLister = pvInformer.Lister()
	agent.pvListerSynced = pvInformer.Informer().HasSynced

	// PVC informer
	agent.pvcLister = pvcInformer.Lister()
	agent.pvcListerSynced = pvcInformer.Informer().HasSynced

	// Pod informer
	podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    agent.podAdded,
		UpdateFunc: agent.podUpdated,
		// deleted pods will not be readded to the queue, do not need DeleteFunc here
	})
	agent.podLister = podInformer.Lister()
	agent.podListerSynced = podInformer.Informer().HasSynced

	agent.pvChecker = handler.NewPVHealthConditionChecker(driverName, conn, client, timeout, agent.pvcLister, agent.pvLister, agent.eventRecorder)

	return agent
}

// Run runs volume health condition checking method
func (agent *PVMonitorAgent) Run(workers int, stopCh <-chan struct{}) {
	defer agent.podWithPVItemQueue.ShutDown()

	klog.Infof("Starting CSI External PV Health Monitor Agent")
	defer klog.Infof("Shutting down CSI External PV Health Monitor Agent")

	if !cache.WaitForCacheSync(stopCh, agent.pvcListerSynced, agent.pvListerSynced, agent.podListerSynced) {
		klog.Errorf("Cannot sync cache")
		return
	}

	for i := 0; i < workers; i++ {
		go wait.Until(agent.checkPVWorker, agent.monitorInterval, stopCh)
	}

	<-stopCh
}

func (agent *PVMonitorAgent) isPodWithPVItemInValid(podWithPVItem *PodWithPVItem) bool {
	if len(podWithPVItem.podNameSpace) == 0 || len(podWithPVItem.podName) == 0 || len(podWithPVItem.pvName) == 0 {
		return true
	}

	return false
}

func (agent *PVMonitorAgent) checkPVWorker() {
	key, quit := agent.podWithPVItemQueue.Get()
	if quit {
		return
	}
	defer agent.podWithPVItemQueue.Done(key)

	podWithPV, ok := key.(*PodWithPVItem)
	if !ok || agent.isPodWithPVItemInValid(podWithPV) {
		klog.Errorf("error item type or PodWithPVItem is invalid(there are empty fileds in it)")
	}

	klog.V(6).Infof("Started PV processing %q", podWithPV.pvName)
	// get PV to process
	pv, err := agent.pvLister.Get(podWithPV.pvName)
	if err != nil {
		if apierrs.IsNotFound(err) {
			klog.V(3).Infof("PV %q deleted, ignoring", podWithPV.pvName)
			return
		}
		klog.Errorf("Error getting PersistentVolume %q: %v", podWithPV.pvName, err)
		agent.podWithPVItemQueue.Add(podWithPV)
		return
	}

	pod, err := agent.podLister.Pods(podWithPV.podNameSpace).Get(podWithPV.podName)
	if err != nil {
		if apierrs.IsNotFound(err) {
			klog.V(3).Infof("Pod %q deleted, ignoring", podWithPV.podName)
			return
		}
		klog.Errorf("Error getting PersistentVolume %q: %v", podWithPV.podName, err)
		agent.podWithPVItemQueue.Add(podWithPV)
		return
	}

	if pv.DeletionTimestamp != nil {
		klog.Infof("PV: %s is being deleted now, skip checking health condition", pv.Name)
		return
	}

	if pv.Status.Phase != v1.VolumeBound {
		klog.Infof("PV: %s status is not bound, remove it from the queue", pv.Name)
		return
	}

	err = agent.pvChecker.CheckNodeVolumeStatus(agent.kubeletRootPath, agent.supportStageUnstage, pv, pod)
	if err != nil {
		klog.Errorf("check node volume status error: %+v", err)
	}

	// re-enqueue anyway
	agent.podWithPVItemQueue.Add(podWithPV)
}
