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
	"fmt"
	"sync"
	"time"

	"google.golang.org/grpc"

	v1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	handler "github.com/kubernetes-csi/external-health-monitor/pkg/csi-handler"
	"github.com/kubernetes-csi/external-health-monitor/pkg/features"
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
func NewPVMonitorController(
	logger klog.Logger,
	client kubernetes.Interface,
	conn *grpc.ClientConn,
	factory informers.SharedInformerFactory,
	eventRecorder record.EventRecorder,
	option *PVMonitorOptions,
) *PVMonitorController {
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
	ctrl.setupPVInformer(factory)
	ctrl.setupPVCInformer(factory)
	ctrl.setupEventInformer(factory)
	ctrl.setupPVChecker(factory, client, conn, option)
	ctrl.setupPodNodeInformersIfNecessary(factory, logger, option)
	return ctrl
}

func (ctrl *PVMonitorController) setupPVInformer(factory informers.SharedInformerFactory) {
	informer := factory.Core().V1().PersistentVolumes()
	informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: ctrl.pvAdded,
		// we do not care about PV changes, so do not need UpdateFunc here.
		// deleted PVs will not be readded to the queue, so do not need DeleteFunc here
	})
	ctrl.pvLister = informer.Lister()
	ctrl.pvListerSynced = informer.Informer().HasSynced
}

func (ctrl *PVMonitorController) setupPVCInformer(factory informers.SharedInformerFactory) {
	informer := factory.Core().V1().PersistentVolumeClaims()
	ctrl.pvcLister = informer.Lister()
	ctrl.pvcListerSynced = informer.Informer().HasSynced
}

func (ctrl *PVMonitorController) setupEventInformer(factory informers.SharedInformerFactory) {
	informer := factory.Core().V1().Events()
	informer.Informer().AddIndexers(cache.Indexers{
		util.DefaultEventIndexerName: func(obj interface{}) ([]string, error) {
			event := obj.(*v1.Event)
			if event != nil {
				key := fmt.Sprintf("%s:%s:%s", string(event.InvolvedObject.UID), event.Type, event.Reason)
				return []string{key}, nil
			} else {
				return nil, nil
			}
		},
	})
}

func (ctrl *PVMonitorController) setupPVChecker(
	factory informers.SharedInformerFactory,
	client kubernetes.Interface,
	conn *grpc.ClientConn,
	option *PVMonitorOptions,
) {
	ctrl.pvChecker = handler.NewPVHealthConditionChecker(
		option.DriverName,
		conn,
		client,
		option.ContextTimeout,
		ctrl.pvcLister,
		ctrl.pvLister,
		factory.Core().V1().Events(),
		ctrl.eventRecorder,
	)
}

func (ctrl *PVMonitorController) setupPodNodeInformersIfNecessary(factory informers.SharedInformerFactory, logger klog.Logger, option *PVMonitorOptions) {
	if ctrl.enableNodeWatcher {
		ctrl.setupPodInformer(factory)
		ctrl.setupNodeWatcher(factory, logger, option)
	}
}

func (ctrl *PVMonitorController) setupPodInformer(factory informers.SharedInformerFactory) {
	informer := factory.Core().V1().Pods()
	informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: ctrl.podAdded,
		// UpdateFunc: ctrl.podUpdated,  TODO: do we need this ?
		DeleteFunc: ctrl.podDeleted,
	})
	ctrl.podLister = informer.Lister()
	ctrl.podListerSynced = informer.Informer().HasSynced
}

func (ctrl *PVMonitorController) setupNodeWatcher(factory informers.SharedInformerFactory, logger klog.Logger, option *PVMonitorOptions) {
	ctrl.nodeWatcher = NewNodeWatcher(
		logger,
		ctrl.driverName,
		ctrl.client,
		ctrl.pvLister,
		ctrl.pvcLister,
		factory.Core().V1().Nodes(),
		ctrl.eventRecorder,
		ctrl.pvcToPodsCache,
		option.NodeWorkerExecuteInterval,
		option.NodeListAndAddInterval,
	)
}

// Run runs the volume health condition checking method
func (ctrl *PVMonitorController) Run(ctx context.Context, workers int, wg *sync.WaitGroup) {
	defer ctrl.pvQueue.ShutDown()

	logger := klog.FromContext(ctx)
	logger.Info("Starting CSI External PV Health Monitor Controller")
	defer logger.Info("Shutting down CSI External PV Health Monitor Controller")

	if !waitForCacheSyncSucceed(ctx, ctrl) {
		logger.Error(nil, "Cannot sync cache")
		return
	}

	if ctrl.enableNodeWatcher {
		go ctrl.nodeWatcher.Run(ctx)
	}

	// TODO: we need to cache the PVs info and get the diff so that we can identify the NotFound error
	// if storage support List Volumes RPC, ListVolumes is preferred for performance reasons
	if ctrl.supportListVolumes {
		if utilfeature.DefaultFeatureGate.Enabled(features.ReleaseLeaderElectionOnExit) {
			wg.Add(1)
			go func() {
				defer wg.Done()
				wait.UntilWithContext(ctx, ctrl.checkPVsHealthConditionByListVolumes, ctrl.ListVolumesInterval)
			}()
		} else {
			go wait.UntilWithContext(ctx, ctrl.checkPVsHealthConditionByListVolumes, ctrl.ListVolumesInterval)
		}
	} else {
		if utilfeature.DefaultFeatureGate.Enabled(features.ReleaseLeaderElectionOnExit) {
			for i := 0; i < workers; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					wait.UntilWithContext(ctx, ctrl.checkPVWorker, ctrl.PVWorkerExecuteInterval)
				}()
			}
			wg.Add(1)
			go func() {
				defer wg.Done()
				wait.UntilWithContext(ctx, func(ctx context.Context) {
					logger := klog.FromContext(ctx)
					err := ctrl.AddPVsToQueue()
					if err != nil {
						logger.Error(err, "Failed to reconcile volumes")
					}
				}, ctrl.VolumeListAndAddInterval)
			}()
		} else {
			for i := 0; i < workers; i++ {
				go wait.UntilWithContext(ctx, ctrl.checkPVWorker, ctrl.PVWorkerExecuteInterval)
			}

			go wait.UntilWithContext(ctx, func(ctx context.Context) {
				logger := klog.FromContext(ctx)
				err := ctrl.AddPVsToQueue()
				if err != nil {
					logger.Error(err, "Failed to reconcile volumes")
				}
			}, ctrl.VolumeListAndAddInterval)
		}
	}

	<-ctx.Done()
}

func waitForCacheSyncSucceed(ctx context.Context, ctrl *PVMonitorController) bool {
	return cache.WaitForCacheSync(ctx.Done(), ctrl.pvListerSynced, ctrl.pvcListerSynced) &&
		(!ctrl.enableNodeWatcher || cache.WaitForCacheSync(ctx.Done(), ctrl.podListerSynced))
}

func (ctrl *PVMonitorController) checkPVsHealthConditionByListVolumes(ctx context.Context) {
	logger := klog.FromContext(ctx)
	err := ctrl.pvChecker.CheckControllerListVolumeStatuses(ctx)
	if err != nil {
		logger.Error(err, "Check controller volume status error")
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

func (ctrl *PVMonitorController) checkPVWorker(ctx context.Context) {
	key, quit := ctrl.pvQueue.Get()
	if quit {
		return
	}
	defer ctrl.pvQueue.Done(key)

	logger := klog.FromContext(ctx)
	pvName := key.(string)
	logger.V(4).Info("Started PV processing", "pv", pvName)

	// get PV to process
	pv, err := ctrl.pvLister.Get(pvName)
	if err != nil {
		if apierrs.IsNotFound(err) {
			// PV was deleted in the meantime, ignore.
			ctrl.Lock()
			// delete pv from cache here so that we do not need to handle pv deletion events
			delete(ctrl.pvEnqueued, pvName)
			ctrl.Unlock()
			logger.V(3).Info("PV deleted, ignoring", "pv", pvName)
			return
		}
		logger.Error(err, "Error getting PersistentVolume", "pv", pvName)
		ctrl.pvQueue.Add(pvName)
		return
	}

	if pv.DeletionTimestamp != nil {
		logger.Info("PV is being deleted now, skip checking health condition", "pv", pv.Name)
		return
	}

	if pv.Status.Phase != v1.VolumeBound {
		logger.Info("PV status is not bound, remove it from the queue", "pv", pv.Name)
		return
	}

	err = ctrl.pvChecker.CheckControllerVolumeStatus(ctx, pv)
	if err != nil {
		logger.Error(err, "Check controller volume status error")
	}

	// re-enqueue anyway
	ctrl.pvQueue.Add(pvName)
}
