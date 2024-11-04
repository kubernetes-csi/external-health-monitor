package csi_handler

import (
	"context"
	"reflect"
	"testing"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/mock/gomock"
	"github.com/kubernetes-csi/csi-test/v5/utils"
	"github.com/kubernetes-csi/external-health-monitor/pkg/mock"
)

// Test Data
var (
	volume1 = &csi.Volume{
		VolumeId: "1",
	}

	volume2 = &csi.Volume{
		VolumeId: "2",
	}

	abnormalVolumeCondition = &csi.VolumeCondition{
		Abnormal: true,
		Message:  "Volume not found",
	}

	normalVolumeCondition = &csi.VolumeCondition{
		Abnormal: false,
		Message:  "",
	}

	volumeMap = map[string]VolumeSample{
		"1": {
			Volume:    volume1,
			Condition: abnormalVolumeCondition,
		},
		"2": {
			Volume:    volume2,
			Condition: normalVolumeCondition,
		},
	}
)

type VolumeSample struct {
	Volume    *csi.Volume
	Condition *csi.VolumeCondition
}

func Test_csiPVHandler_ControllerListVolumeConditions(t *testing.T) {
	mockController, driver, _, controllerServer, _, csiConn, err := mock.CreateMockServer(t)
	if err != nil {
		t.Fatal(err)
	}
	defer mockController.Finish()
	defer driver.Stop()

	handler := NewCSIPVHandler(csiConn)
	in := &csi.ListVolumesRequest{
		StartingToken: "",
	}
	out := &csi.ListVolumesResponse{
		Entries: []*csi.ListVolumesResponse_Entry{
			{
				Volume: volume1,
				Status: &csi.ListVolumesResponse_VolumeStatus{
					VolumeCondition: abnormalVolumeCondition,
				},
			},
			{
				Volume: volume2,
				Status: &csi.ListVolumesResponse_VolumeStatus{
					VolumeCondition: normalVolumeCondition,
				},
			},
		},
		NextToken: "",
	}

	controllerServer.EXPECT().ListVolumes(gomock.Any(), utils.Protobuf(in)).Return(out, nil).Times(1)
	tests := []struct {
		name    string
		want    map[string]*VolumeConditionResult
		wantErr bool
	}{
		{
			name: "case1",
			want: map[string]*VolumeConditionResult{
				"1": {
					abnormal: true,
					message:  "Volume not found",
				},
				"2": {
					abnormal: false,
					message:  "",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := handler.ControllerListVolumeConditions(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("csiPVHandler.ControllerListVolumeConditions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("csiPVHandler.ControllerListVolumeConditions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_csiPVHandler_ControllerGetVolumeCondition(t *testing.T) {
	mockController, driver, _, controllerServer, _, csiConn, err := mock.CreateMockServer(t)
	if err != nil {
		t.Fatal(err)
	}
	defer mockController.Finish()
	defer driver.Stop()

	handler := NewCSIPVHandler(csiConn)
	tests := []struct {
		name     string
		want     *VolumeConditionResult
		volumeId string
		wantErr  bool
	}{
		{
			name:     "AbnormalCase",
			volumeId: "1",
			want: &VolumeConditionResult{
				abnormal: true,
				message:  "Volume not found",
			},
			wantErr: false,
		},
		{
			name:     "NormalCase",
			volumeId: "2",
			want: &VolumeConditionResult{
				abnormal: false,
				message:  "",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := &csi.ControllerGetVolumeRequest{
				VolumeId: tt.volumeId,
			}
			out := &csi.ControllerGetVolumeResponse{
				Volume: volumeMap[tt.volumeId].Volume,
				Status: &csi.ControllerGetVolumeResponse_VolumeStatus{
					VolumeCondition: volumeMap[tt.volumeId].Condition,
				},
			}
			controllerServer.EXPECT().ControllerGetVolume(gomock.Any(), utils.Protobuf(in)).Return(out, nil).Times(1)
			got, err := handler.ControllerGetVolumeCondition(context.Background(), tt.volumeId)
			if (err != nil) != tt.wantErr {
				t.Errorf("csiPVHandler.ControllerGetVolumeCondition() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("csiPVHandler.ControllerGetVolumeCondition() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_csiPVHandler_NodeGetVolumeCondition(t *testing.T) {
	mockController, driver, _, _, nodeServer, csiConn, err := mock.CreateMockServer(t)
	if err != nil {
		t.Fatal(err)
	}
	defer mockController.Finish()
	defer driver.Stop()

	handler := NewCSIPVHandler(csiConn)
	type args struct {
		volumeID          string
		volumePath        string
		volumeStagingPath string
		ctx               context.Context
	}
	tests := []struct {
		name    string
		want    *VolumeConditionResult
		wantErr bool
		args    args
	}{
		{
			name: "AbnormalCase",
			want: &VolumeConditionResult{
				abnormal: true,
				message:  "Volume not found",
			},
			args: args{
				volumeID:          "1",
				volumePath:        "",
				volumeStagingPath: "",
				ctx:               context.Background(),
			},
			wantErr: false,
		},
		{
			name: "NormalCase",
			want: &VolumeConditionResult{
				abnormal: false,
				message:  "",
			},
			args: args{
				volumeID:          "2",
				volumePath:        "",
				volumeStagingPath: "",
				ctx:               context.Background(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := &csi.NodeGetVolumeStatsRequest{
				VolumeId:          tt.args.volumeID,
				VolumePath:        "",
				StagingTargetPath: "",
			}
			out := &csi.NodeGetVolumeStatsResponse{
				VolumeCondition: volumeMap[tt.args.volumeID].Condition,
			}
			nodeServer.EXPECT().NodeGetVolumeStats(gomock.Any(), utils.Protobuf(in)).Return(out, nil).Times(1)
			got, err := handler.NodeGetVolumeCondition(tt.args.ctx, tt.args.volumeID, tt.args.volumePath, tt.args.volumeStagingPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("csiPVHandler.NodeGetVolumeCondition() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("csiPVHandler.NodeGetVolumeCondition() = %v, want %v", got, tt.want)
			}
		})
	}
}
