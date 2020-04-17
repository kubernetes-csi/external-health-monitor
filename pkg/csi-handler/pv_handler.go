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

	"google.golang.org/grpc"

	"github.com/container-storage-interface/spec/lib/go/csi"
)

var _ CSIHandler = &csiPVHandler{}

type csiPVHandler struct {
	conn *grpc.ClientConn
}

func NewCSIPVHandler(conn *grpc.ClientConn) CSIHandler {
	return &csiPVHandler{
		conn: conn,
	}
}

type VolumeConditionResult struct {
	abnormal bool
	message  string
}

func (vcr *VolumeConditionResult) GetAbnormal() bool {
	return vcr.abnormal
}

func (vcr *VolumeConditionResult) GetMessage() string {
	return vcr.message
}

func (handler *csiPVHandler) ControllerListVolumeConditions(ctx context.Context) (map[string]*VolumeConditionResult, error) {
	client := csi.NewControllerClient(handler.conn)

	p := map[string]*VolumeConditionResult{}

	token := ""
	for {
		rsp, err := client.ListVolumes(ctx, &csi.ListVolumesRequest{
			StartingToken: token,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list volumes: %v", err)
		}

		for _, e := range rsp.Entries {
			p[e.GetVolume().VolumeId] = &VolumeConditionResult{
				abnormal: e.GetStatus().GetVolumeCondition().GetAbnormal(),
				message:  e.GetStatus().GetVolumeCondition().GetMessage(),
			}
		}
		token = rsp.NextToken

		if len(token) == 0 {
			break
		}
	}
	return p, nil
}

func (handler *csiPVHandler) ControllerGetVolumeCondition(ctx context.Context, volumeID string) (*VolumeConditionResult, error) {
	client := csi.NewControllerClient(handler.conn)

	req := csi.ControllerGetVolumeRequest{
		VolumeId: volumeID,
	}

	res, err := client.ControllerGetVolume(ctx, &req)
	if err != nil {
		// if there is an error, do not return abnormal status
		// wait for another call
		return nil, err
	}

	// We reach here only when VOLUME_CONDITION controller capability is supported
	// so the Status in ControllerGetVolumeResponse must not be nil

	return &VolumeConditionResult{abnormal: res.GetStatus().GetVolumeCondition().GetAbnormal(), message: res.GetStatus().GetVolumeCondition().GetMessage()}, nil
}

func (handler *csiPVHandler) NodeGetVolumeCondition(ctx context.Context, volumeID string, volumePath string, volumeStagingPath string) (*VolumeConditionResult, error) {
	client := csi.NewNodeClient(handler.conn)

	req := csi.NodeGetVolumeStatsRequest{
		VolumeId:          volumeID,
		VolumePath:        volumePath,
		StagingTargetPath: volumeStagingPath,
	}

	res, err := client.NodeGetVolumeStats(ctx, &req)
	if err != nil {
		// if there is an error, do not return abnormal status
		// wait for another call
		return nil, err
	}

	return &VolumeConditionResult{abnormal: res.GetVolumeCondition().GetAbnormal(), message: res.GetVolumeCondition().GetMessage()}, nil
}
