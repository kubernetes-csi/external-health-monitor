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

	"k8s.io/klog"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/record"

	"github.com/kubernetes-csi/external-health-monitor/pkg/util"
)

// PVHealthConditionChecker is for checking pv health condition
type PVHealthConditionChecker struct {
	checkerName string

	csiConn       *grpc.ClientConn
	timeout       time.Duration
	k8sClient     kubernetes.Interface
	eventRecorder record.EventRecorder

	pvcLister corelisters.PersistentVolumeClaimLister
	pvLister  corelisters.PersistentVolumeLister

	csiPVHandler CSIHandler
}

// NewPVHealthConditionChecker return an instance of PVHealthConditionChecker
func NewPVHealthConditionChecker(name string, conn *grpc.ClientConn, kClient kubernetes.Interface, timeout time.Duration,
	pvcLister corelisters.PersistentVolumeClaimLister, pvLister corelisters.PersistentVolumeLister, recorder record.EventRecorder) *PVHealthConditionChecker {
	/*broadcaster := record.NewBroadcaster()
	broadcaster.StartRecordingToSink(&corev1.EventSinkImpl{Interface: kClient.CoreV1().Events(v1.NamespaceAll)})
	var eventRecorder record.EventRecorder
	eventRecorder = broadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: fmt.Sprintf("csi-pv-monitor %s", name)})
	*/
	return &PVHealthConditionChecker{
		checkerName:   name,
		csiConn:       conn,
		k8sClient:     kClient,
		eventRecorder: recorder,
		pvcLister:     pvcLister,
		pvLister:      pvLister,
		timeout:       timeout,

		csiPVHandler: NewCSIPVHandler(conn),
	}
}

// CheckControllerVolumesStatus checks volumes health condition by ListVolumes
func (checker *PVHealthConditionChecker) CheckControllerVolumesStatus() error {
	ctx, cancel := context.WithTimeout(context.Background(), checker.timeout)
	defer cancel()

	result, err := checker.csiPVHandler.ControllerVolumesChecking(ctx)
	if err != nil {
		return err
	}

	pvs, err := checker.pvLister.List(labels.Everything())
	if err != nil {
		return err
	}

	for _, pv := range pvs {
		if pv.Spec.CSI == nil || pv.Spec.CSI.Driver != checker.checkerName {
			klog.Infof("csi source is nil or the volume is not managed by this checker/monitor")
			continue
		}

		if pv.Status.Phase != v1.VolumeBound {
			klog.Infof("PV: %s status is not bound", pv.Name)
			continue
		}

		volumeHandle, err := checker.GetVolumeHandle(pv)
		if err != nil {
			klog.Errorf("Get volume handler error: %+v", err)
			continue
		}

		if result[volumeHandle] != nil && result[volumeHandle].GetAbnormal() {
			// Since pv status is bound, we believe PV controller, do not check pv.Spec.ClaimRef here.
			pvc, err := checker.pvcLister.PersistentVolumeClaims(pv.Spec.ClaimRef.Namespace).Get(pv.Spec.ClaimRef.Name)
			if err != nil {
				klog.Errorf("Get PVC error: %+v", err)
				continue
			}
			checker.eventRecorder.Event(pvc, v1.EventTypeWarning, "PV is not healthy(detected by monitor controller)", result[volumeHandle].GetMessage())
		}
	}

	return nil
}

// GetVolumeHandle return the volume handle of the pv
func (checker *PVHealthConditionChecker) GetVolumeHandle(pv *v1.PersistentVolume) (string, error) {
	if pv.Spec.CSI == nil {
		return "", fmt.Errorf("csi source is nil")
	}

	return pv.Spec.CSI.VolumeHandle, nil
}

// CheckControllerVolumeStatus checks volume status in controller side
func (checker *PVHealthConditionChecker) CheckControllerVolumeStatus(pv *v1.PersistentVolume) error {
	if pv.Spec.CSI == nil || pv.Spec.CSI.Driver != checker.checkerName {
		return fmt.Errorf("csi source is nil or the volume is not managed by this checker/monitor")
	}

	if pv.Status.Phase != v1.VolumeBound {
		return fmt.Errorf("PV: %s status is not bound", pv.Name)
	}

	ctx, cancel := context.WithTimeout(context.Background(), checker.timeout)
	defer cancel()

	volumeHandle, err := checker.GetVolumeHandle(pv)
	if err != nil {
		klog.Errorf("Get volume handler error: %+v", err)
		return err
	}

	if len(volumeHandle) == 0 {
		return fmt.Errorf("volume handle in csi source is empty")
	}

	volumeCondition, err := checker.csiPVHandler.ControllerVolumeChecking(ctx, volumeHandle)
	if err != nil {
		return err
	}

	// if error is not nil, the volumeCondition result can not be nil neither

	// At the first stage, we just send PVC events
	if volumeCondition.GetAbnormal() {
		// Since pv status is bound, we believe PV controller, do not check pv.Spec.ClaimRef here.
		pvc, err := checker.pvcLister.PersistentVolumeClaims(pv.Spec.ClaimRef.Namespace).Get(pv.Spec.ClaimRef.Name)
		if err != nil {
			return err
		}
		checker.eventRecorder.Event(pvc, v1.EventTypeWarning, "PV is not healthy(detected by monitor controller)", volumeCondition.GetMessage())
	}

	return nil
}

// CheckNodeVolumeStatus checks volume status in node side
func (checker *PVHealthConditionChecker) CheckNodeVolumeStatus(kubeletRootPath string, supportStageUnstage bool, pv *v1.PersistentVolume, pod *v1.Pod) error {
	if pv.Spec.CSI == nil || pv.Spec.CSI.Driver != checker.checkerName {
		return fmt.Errorf("csi source is nil or the volume is not managed by this checker/monitor")
	}

	if pv.Status.Phase != v1.VolumeBound {
		return fmt.Errorf("PV: %s status is not bound", pv.Name)
	}

	ctx, cancel := context.WithTimeout(context.Background(), checker.timeout)
	defer cancel()

	volumeHandle, err := checker.GetVolumeHandle(pv)
	if err != nil {
		klog.Errorf("Get volume handler error: %+v", err)
		return err
	}

	if len(volumeHandle) == 0 {
		return fmt.Errorf("volume handle in csi source is empty")
	}

	var volumePath, stagingTargetPath string

	volumePath = util.GetVolumePath(kubeletRootPath, pv.Name, string(pod.UID))

	if supportStageUnstage {
		stagingTargetPath, err = util.MakeDeviceMountPath(kubeletRootPath, pv)
		if err != nil {
			return err
		}
	}

	volumeCondition, err := checker.csiPVHandler.NodeVolumeChecking(ctx, volumeHandle, volumePath, stagingTargetPath)
	if err != nil {
		return err
	}

	// if error is not nil, the volumeCondition result can not be nil neither

	// At the first stage, we just send PVC events
	if volumeCondition.GetAbnormal() {
		// Since pv status is bound, we believe PV controller, do not check pv.Spec.ClaimRef here.
		pvc, err := checker.pvcLister.PersistentVolumeClaims(pv.Spec.ClaimRef.Namespace).Get(pv.Spec.ClaimRef.Name)
		if err != nil {
			return err
		}
		checker.eventRecorder.Event(pvc, v1.EventTypeWarning, "PV is not healthy(detected by monitor agent)", volumeCondition.GetMessage())
	}

	return nil
}
