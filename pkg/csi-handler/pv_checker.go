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

package csi_handler

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	coreinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"

	"github.com/kubernetes-csi/external-health-monitor/pkg/util"
)

// PVHealthConditionChecker is for checking pv health condition
type PVHealthConditionChecker struct {
	driverName string

	csiConn       *grpc.ClientConn
	timeout       time.Duration
	k8sClient     kubernetes.Interface
	eventRecorder record.EventRecorder

	pvcLister corelisters.PersistentVolumeClaimLister
	pvLister  corelisters.PersistentVolumeLister

	eventInformer coreinformers.EventInformer

	csiPVHandler CSIHandler
}

// NewPVHealthConditionChecker returns an instance of PVHealthConditionChecker
func NewPVHealthConditionChecker(
	name string,
	conn *grpc.ClientConn,
	kClient kubernetes.Interface,
	timeout time.Duration,
	pvcLister corelisters.PersistentVolumeClaimLister,
	pvLister corelisters.PersistentVolumeLister,
	eventInformer coreinformers.EventInformer,
	recorder record.EventRecorder,
) *PVHealthConditionChecker {
	return &PVHealthConditionChecker{
		driverName:    name,
		csiConn:       conn,
		k8sClient:     kClient,
		eventRecorder: recorder,
		pvcLister:     pvcLister,
		pvLister:      pvLister,
		timeout:       timeout,
		eventInformer: eventInformer,
		csiPVHandler:  NewCSIPVHandler(conn),
	}
}

// CheckControllerListVolumeStatuses checks volumes health condition by ListVolumes
func (checker *PVHealthConditionChecker) CheckControllerListVolumeStatuses(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, checker.timeout)
	defer cancel()

	result, err := checker.csiPVHandler.ControllerListVolumeConditions(ctx)
	if err != nil {
		return err
	}

	pvs, err := checker.pvLister.List(labels.Everything())
	if err != nil {
		return err
	}

	logger := klog.FromContext(ctx)
	for _, pv := range pvs {
		if pv.Spec.CSI == nil || pv.Spec.CSI.Driver != checker.driverName {
			logger.Info("CSI source is nil or the volume is not managed by this checker/monitor")
			continue
		}

		if pv.Status.Phase != v1.VolumeBound {
			logger.Info("PV status is not bound", "pv", pv.Name)
			continue
		}

		volumeHandle, err := checker.GetVolumeHandle(pv)
		if err != nil {
			logger.Error(err, "Get volume handle error")
			continue
		}

		volumeCondition := result[volumeHandle]
		if volumeCondition == nil {
			continue
		}

		pvc, err := checker.pvcLister.PersistentVolumeClaims(pv.Spec.ClaimRef.Namespace).Get(pv.Spec.ClaimRef.Name)
		if err != nil {
			logger.Error(err, "Get PVC error")
			continue
		}

		if volumeCondition.GetAbnormal() {
			// Since pv status is bound, we believe PV controller, do not check pv.Spec.ClaimRef here.
			checker.eventRecorder.Event(pvc, v1.EventTypeWarning, "VolumeConditionAbnormal", volumeCondition.GetMessage())
		} else {
			// Send recovery event if the abnormal event was sent and unexpired
			checker.sendRecoveryEventToPVC(logger, pvc)
		}
	}

	return nil
}

// GetVolumeHandle returns the volume handle of the pv
func (checker *PVHealthConditionChecker) GetVolumeHandle(pv *v1.PersistentVolume) (string, error) {
	if pv.Spec.CSI == nil {
		return "", fmt.Errorf("csi source is nil")
	}

	return pv.Spec.CSI.VolumeHandle, nil
}

// CheckControllerVolumeStatus checks volume status in controller side
func (checker *PVHealthConditionChecker) CheckControllerVolumeStatus(ctx context.Context, pv *v1.PersistentVolume) error {
	if pv.Spec.CSI == nil || pv.Spec.CSI.Driver != checker.driverName {
		return fmt.Errorf("csi source is nil or the volume is not managed by this checker/monitor")
	}

	if pv.Status.Phase != v1.VolumeBound {
		return fmt.Errorf("PV: %s status is not bound", pv.Name)
	}

	ctx, cancel := context.WithTimeout(ctx, checker.timeout)
	defer cancel()

	logger := klog.FromContext(ctx)
	volumeHandle, err := checker.GetVolumeHandle(pv)
	if err != nil {
		logger.Error(err, "Get volume handle error")
		return err
	}

	if len(volumeHandle) == 0 {
		return fmt.Errorf("volume handle in csi source is empty")
	}

	volumeCondition, err := checker.csiPVHandler.ControllerGetVolumeCondition(ctx, volumeHandle)
	if err != nil {
		return err
	}

	pvc, err := checker.pvcLister.PersistentVolumeClaims(pv.Spec.ClaimRef.Namespace).Get(pv.Spec.ClaimRef.Name)
	if err != nil {
		return err
	}

	// At the first stage, we just send PVC events
	if volumeCondition.GetAbnormal() {
		// Since pv status is bound, we believe PV controller, do not check pv.Spec.ClaimRef here.
		checker.eventRecorder.Event(pvc, v1.EventTypeWarning, "VolumeConditionAbnormal", volumeCondition.GetMessage())
	} else {
		// Send recovery event if the abnormal event was sent and unexpired
		checker.sendRecoveryEventToPVC(logger, pvc)
	}

	return nil
}

// sendRecoveryEventToPVC sends the recovery event to the pvc
// If the volume condition is normal and abnormal event wasn't expired,
// PVHealthConditionChecker should send recovery event.
func (checker *PVHealthConditionChecker) sendRecoveryEventToPVC(logger klog.Logger, pvc *v1.PersistentVolumeClaim) {
	pvcUID := string(pvc.ObjectMeta.GetUID())
	key := fmt.Sprintf("%s:%s:%s", pvcUID, v1.EventTypeWarning, "VolumeConditionAbnormal")
	events, err := checker.eventInformer.Informer().GetIndexer().ByIndex(util.DefaultEventIndexerName, key)
	if err != nil {
		logger.Info("Get abnormal event from indexer failed", "err", err)
	}

	if len(events) > 0 {
		checker.eventRecorder.Event(pvc, v1.EventTypeNormal, "VolumeConditionNormal", util.DefaultRecoveryEventMessage)
	}
}
