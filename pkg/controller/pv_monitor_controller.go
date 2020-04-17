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
	"fmt"
	"sync"
	"time"

	"k8s.io/klog"

	"google.golang.org/grpc"

	v1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
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

// PVMonitorController is the struct of pv monitor controller containing all information to perform volumes health condition checking
type PVMonitorController struct {
	client             kubernetes.Interface
	driverName         string
	eventRecorder      record.EventRecorder
	supportListVolumes bool

	pvChecker *handler.PVHealthConditionChecker

	enableNodeWatcher bool
	nodeWatcher       *NodeWatcher

	csiConn *grpc.ClientConn

	pvLister       corelisters.PersistentVolumeLister
	pvListerSynced cache.InformerSynced

	pvcLister       corelisters.PersistentVolumeClaimLister
	pvcListerSynced cache.InformerSynced

	podLister       corelisters.PodLister
	podListerSynced cache.InformerSynced

	// used for updating pvEnqueue map
	sync.Mutex
	// pvEnqueued stores all CSI PVs which are enqueued
	pvEnqueued map[string]bool
	// pvcToPodsCache stores PVCs/Pods mapping info
	pvcToPodsCache *util.PVCToPodsCache
	// we get PVs from pvQueue to check their health conditions
	pvQueue workqueue.Interface

	// Time interval for calling ListVolumes RPC to check volumes' health condition
	ListVolumesInterval time.Duration
	// Time interval for executing pv worker goroutines
	PVWorkerExecuteInterval time.Duration
	// Time interval for listing volumes and add them to queue
	VolumeListAndAddInterval time.Duration
}

// PVMonitorOptions configures PV monitor
type PVMonitorOptions struct {
	ContextTimeout    time.Duration
	DriverName        string
	EnableNodeWatcher bool
	SupportListVolume bool

	ListVolumesInterval      time.Duration
	PVWorkerExecuteInterval  time.Duration
	VolumeListAndAddInterval time.Duration

	NodeWorkerExecuteInterval time.Duration
	NodeListAndAddInterval    time.Duration
}

// NewPVMonitorController creates PV monitor controller
func NewPVMonitorController(client kubernetes.Interface, conn *grpc.ClientConn, pvInformer coreinformers.PersistentVolumeInformer,
	pvcInformer coreinformers.PersistentVolumeClaimInformer, podInformer coreinformers.PodInformer, nodeInformer coreinformers.NodeInformer, option *PVMonitorOptions) *PVMonitorController {

	broadcaster := record.NewBroadcaster()
	broadcaster.StartRecordingToSink(&corev1.EventSinkImpl{Interface: client.CoreV1().Events(v1.NamespaceAll)})
	var eventRecorder record.EventRecorder
	eventRecorder = broadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: fmt.Sprintf("csi-pv-monitor-controller-%s", option.DriverName)})

	ctrl := &PVMonitorController{
		csiConn:            conn,
		eventRecorder:      eventRecorder,
		supportListVolumes: option.SupportListVolume,
		enableNodeWatcher:  option.EnableNodeWatcher,
		client:             client,
		driverName:         option.DriverName,
		pvQueue:            workqueue.NewNamed("csi-monitor-pv-queue"),

		pvcToPodsCache: util.NewPVCToPodsCache(),
		pvEnqueued:     make(map[string]bool),

		ListVolumesInterval:      option.ListVolumesInterval,
		PVWorkerExecuteInterval:  option.PVWorkerExecuteInterval,
		VolumeListAndAddInterval: option.VolumeListAndAddInterval,
	}

	// PV informer
	pvInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: ctrl.pvAdded,
		// we do not care about PV changes, so do not need UpdateFunc here.
		// deleted PVs will not be readded to the queue, so do not need DeleteFunc here
	})
	ctrl.pvLister = pvInformer.Lister()
	ctrl.pvListerSynced = pvInformer.Informer().HasSynced

	// PVC informer
	ctrl.pvcLister = pvcInformer.Lister()
	ctrl.pvListerSynced = pvcInformer.Informer().HasSynced

	// Pod informer
	podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: ctrl.podAdded,
		// UpdateFunc: ctrl.podUpdated,  TODO: do we need this ?
		DeleteFunc: ctrl.podDeleted,
	})
	ctrl.podLister = podInformer.Lister()
	ctrl.podListerSynced = podInformer.Informer().HasSynced

	if ctrl.enableNodeWatcher {
		ctrl.nodeWatcher = NewNodeWatcher(ctrl.driverName, ctrl.client, ctrl.pvLister, ctrl.pvcLister, nodeInformer, ctrl.eventRecorder, ctrl.pvcToPodsCache, option.NodeWorkerExecuteInterval, option.NodeListAndAddInterval)
	}

	ctrl.pvChecker = handler.NewPVHealthConditionChecker(option.DriverName, conn, client, option.ContextTimeout, ctrl.pvcLister, ctrl.pvLister, ctrl.eventRecorder)

	return ctrl
}

// Run runs the volume health condition checking method
func (ctrl *PVMonitorController) Run(workers int, stopCh <-chan struct{}) {
	defer ctrl.pvQueue.ShutDown()

	klog.Infof("Starting CSI External PV Health Monitor Controller")
	defer klog.Infof("Shutting down CSI External PV Health Monitor Controller")

	if !cache.WaitForCacheSync(stopCh, ctrl.pvcListerSynced, ctrl.pvListerSynced, ctrl.podListerSynced) {
		klog.Errorf("Cannot sync cache")
		return
	}

	if ctrl.enableNodeWatcher {
		go ctrl.nodeWatcher.Run(stopCh)
	}

	// TODO: we need to cache the PVs info and get the diff so that we can identify the NotFound error
	// if storage support List Volumes RPC, ListVolumes is preferred for performance reasons
	if ctrl.supportListVolumes {
		go wait.Until(ctrl.checkPVsHealthConditionByListVolumes, ctrl.ListVolumesInterval, stopCh)
	} else {
		for i := 0; i < workers; i++ {
			go wait.Until(ctrl.checkPVWorker, ctrl.PVWorkerExecuteInterval, stopCh)
		}

		go wait.Until(func() {
			err := ctrl.AddPVsToQueue()
			if err != nil {
				klog.Errorf("Failed to reconcile volumes: %v", err)
			}
		}, ctrl.VolumeListAndAddInterval, stopCh)
	}

	<-stopCh
}

func (ctrl *PVMonitorController) checkPVsHealthConditionByListVolumes() {
	err := ctrl.pvChecker.CheckControllerListVolumeStatuses()
	if err != nil {
		klog.Errorf("check controller volume status error: %+v", err)
	}
}

// AddPVsToQueue adds PVs to queue periodically
func (ctrl *PVMonitorController) AddPVsToQueue() error {
	// TODO: add PV filters when listing
	// for example: only return CSI PVs
	pvs, err := ctrl.pvLister.List(labels.Everything())
	if err != nil {
		return err
	}

	for _, pv := range pvs {
		if pv.Spec.CSI == nil || pv.Spec.CSI.Driver != ctrl.driverName {
			continue
		}
		if !ctrl.pvEnqueued[pv.Name] {
			ctrl.Lock()
			ctrl.pvEnqueued[pv.Name] = true
			ctrl.pvQueue.Add(pv.Name)
			ctrl.Unlock()
		}
	}

	return nil
}

func (ctrl *PVMonitorController) checkPVWorker() {
	key, quit := ctrl.pvQueue.Get()
	if quit {
		return
	}
	defer ctrl.pvQueue.Done(key)

	pvName := key.(string)
	klog.V(4).Infof("Started PV processing PV %q", pvName)

	// get PV to process
	pv, err := ctrl.pvLister.Get(pvName)
	if err != nil {
		if apierrs.IsNotFound(err) {
			// PV was deleted in the meantime, ignore.
			ctrl.Lock()
			// delete pv from cache here so that we do not need to handle pv deletion events
			delete(ctrl.pvEnqueued, pv.Name)
			ctrl.Unlock()
			klog.V(3).Infof("PV %q deleted, ignoring", pvName)
			return
		}
		klog.Errorf("Error getting PersistentVolume %q: %v", pvName, err)
		ctrl.pvQueue.Add(pvName)
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

	err = ctrl.pvChecker.CheckControllerVolumeStatus(pv)
	if err != nil {
		klog.Errorf("check controller volume status error: %+v", err)
	}

	// re-enqueue anyway
	ctrl.pvQueue.Add(pvName)
}
