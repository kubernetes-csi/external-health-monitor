package csi_handler

import (
	csitestutil "github.com/kubernetes-csi/csi-test/v5/utils"
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"
	informerV1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2/ktesting"
	_ "k8s.io/klog/v2/ktesting/init"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/mock/gomock"
	"github.com/kubernetes-csi/csi-test/v5/driver"
	"github.com/kubernetes-csi/external-health-monitor/pkg/mock"
	"github.com/stretchr/testify/assert"
)

type MockPVHealthConditionChecker struct {
	pvHealthConditionChecker *PVHealthConditionChecker
	pvcInformer              informerV1.PersistentVolumeClaimInformer
	pvInformer               informerV1.PersistentVolumeInformer
	eventStore               chan string
	csiControllerServer      *driver.MockControllerServer
	csiNodeServer            *driver.MockNodeServer
}

func createMockPVHealthConditionChecker(t *testing.T) *MockPVHealthConditionChecker {
	k8sClient, informer := mock.FakeK8s()
	_, _, _, controllerServer, nodeServer, csiConn, err := mock.CreateMockServer(t)
	if err != nil {
		t.Fatal(err)
	}

	handler := NewCSIPVHandler(csiConn)
	eventStore := make(chan string, 1)
	return &MockPVHealthConditionChecker{
		pvHealthConditionChecker: &PVHealthConditionChecker{
			driverName: mock.DriverName,
			csiConn:    csiConn,
			timeout:    15 * time.Second,
			k8sClient:  k8sClient,
			eventRecorder: &record.FakeRecorder{
				Events: eventStore,
			},
			eventInformer: informer.Core().V1().Events(),
			pvcLister:     informer.Core().V1().PersistentVolumeClaims().Lister(),
			pvLister:      informer.Core().V1().PersistentVolumes().Lister(),
			csiPVHandler:  handler,
		},
		pvcInformer:         informer.Core().V1().PersistentVolumeClaims(),
		pvInformer:          informer.Core().V1().PersistentVolumes(),
		csiControllerServer: controllerServer,
		csiNodeServer:       nodeServer,
		eventStore:          eventStore,
	}
}

func TestPVHealthConditionChecker_CheckControllerListVolumeStatuses(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		name              string
		pvc               *v1.PersistentVolumeClaim
		pv                *v1.PersistentVolume
		volumeId          string
		wantErr           bool
		wantAbnormalEvent bool
	}{
		{
			name:              "VolumeConditionAbnormal Case",
			pvc:               mock.CreatePVC(1, 2, "pvc", "uid", mock.DefaultNS, "pv", v1.ClaimBound),
			pv:                mock.CreatePV(2, "pvc", "pv", mock.DefaultNS, "1", "uid", &mock.FSVolumeMode, v1.VolumeBound),
			wantAbnormalEvent: true,
			volumeId:          "1",
		},
		{
			name:              "VolumeConditionNormal Case",
			pvc:               mock.CreatePVC(1, 2, "pvc", "uid", mock.DefaultNS, "pv", v1.ClaimBound),
			pv:                mock.CreatePV(2, "pvc", "pv", mock.DefaultNS, "2", "uid", &mock.FSVolumeMode, v1.VolumeBound),
			wantAbnormalEvent: false,
			volumeId:          "2",
		},
		{
			name:     "PV without CSI driver Case",
			pvc:      mock.CreatePVC(1, 2, "pvc", "uid", mock.DefaultNS, "pv", v1.ClaimBound),
			pv:       mock.CreatePVWithoutCSIDriver(2, "pvc", "pv", mock.DefaultNS, "1", "uid", v1.VolumeBound, &mock.FSVolumeMode),
			volumeId: "1",
		},
	}

	var (
		in  *csi.ListVolumesRequest
		out *csi.ListVolumesResponse
	)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := createMockPVHealthConditionChecker(t)
			if err := checker.pvInformer.Informer().GetStore().Add(tt.pv); err != nil {
				t.Errorf("PVHealthConditionChecker.CheckControllerListVolumeStatuses() error = %v", err)
			}

			if err := checker.pvcInformer.Informer().GetStore().Add(tt.pvc); err != nil {
				t.Errorf("PVHealthConditionChecker.CheckControllerListVolumeStatuses() error = %v", err)
			}

			in = &csi.ListVolumesRequest{
				StartingToken: "",
			}
			out = &csi.ListVolumesResponse{
				Entries: []*csi.ListVolumesResponse_Entry{
					{
						Volume: volumeMap[tt.volumeId].Volume,
						Status: &csi.ListVolumesResponse_VolumeStatus{
							VolumeCondition: volumeMap[tt.volumeId].Condition,
						},
					},
				},
				NextToken: "",
			}

			_, ctx := ktesting.NewTestContext(t)
			checker.csiControllerServer.EXPECT().ListVolumes(gomock.Any(), csitestutil.Protobuf(in)).Return(out, nil).Times(1)
			if err := checker.pvHealthConditionChecker.CheckControllerListVolumeStatuses(ctx); (err != nil) != tt.wantErr {
				t.Errorf("PVHealthConditionChecker.CheckControllerListVolumeStatuses() error = %v", err)
			}

			event, err := mock.WatchEvent(tt.wantAbnormalEvent, checker.eventStore)
			if tt.wantAbnormalEvent {
				assert.Nil(err)
				assert.EqualValues(event, mock.AbnormalEvent)
			} else {
				assert.EqualValues(mock.ErrorWatchTimeout.Error(), err.Error())
			}
		})
	}
}

func TestPVHealthConditionChecker_GetVolumeHandle(t *testing.T) {
	tests := []struct {
		name    string
		pv      *v1.PersistentVolume
		wantErr bool
		want    string
	}{
		{
			name:    "VolumeConditionNormal Case",
			pv:      mock.CreatePV(2, "pvc", "pv", mock.DefaultNS, "2", "uid", &mock.FSVolumeMode, v1.VolumeBound),
			wantErr: false,
			want:    "2",
		},
		{
			name:    "PV without CSI driver Case",
			pv:      mock.CreatePVWithoutCSIDriver(2, "pvc", "pv", mock.DefaultNS, "1", "uid", v1.VolumeBound, &mock.FSVolumeMode),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := createMockPVHealthConditionChecker(t)
			got, err := checker.pvHealthConditionChecker.GetVolumeHandle(tt.pv)
			if (err != nil) != tt.wantErr {
				t.Errorf("PVHealthConditionChecker.GetVolumeHandle() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("PVHealthConditionChecker.GetVolumeHandle() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPVHealthConditionChecker_CheckControllerVolumeStatus(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		name              string
		pv                *v1.PersistentVolume
		pvc               *v1.PersistentVolumeClaim
		volumeId          string
		wantErr           bool
		wantAbnormalEvent bool
	}{
		{
			name:              "VolumeConditionAbnormal Case",
			pvc:               mock.CreatePVC(1, 2, "pvc", "uid", mock.DefaultNS, "pv", v1.ClaimBound),
			pv:                mock.CreatePV(2, "pvc", "pv", mock.DefaultNS, "1", "uid", &mock.FSVolumeMode, v1.VolumeBound),
			volumeId:          "1",
			wantAbnormalEvent: true,
		},
		{
			name:              "VolumeConditionNormal Case",
			pvc:               mock.CreatePVC(1, 2, "pvc", "uid", mock.DefaultNS, "pv", v1.ClaimBound),
			pv:                mock.CreatePV(2, "pvc", "pv", mock.DefaultNS, "2", "uid", &mock.FSVolumeMode, v1.VolumeBound),
			wantAbnormalEvent: false,
			volumeId:          "2",
		},
		{
			name:     "PV without CSI driver Case",
			pvc:      mock.CreatePVC(1, 2, "pvc", "uid", mock.DefaultNS, "pv", v1.ClaimBound),
			pv:       mock.CreatePVWithoutCSIDriver(2, "pvc", "pv", mock.DefaultNS, "1", "uid", v1.VolumeBound, &mock.FSVolumeMode),
			volumeId: "1",
			wantErr:  true,
		},
		{
			name:     "PV isn't in VolumeBound state",
			pvc:      mock.CreatePVC(1, 2, "pvc", "uid", mock.DefaultNS, "pv", v1.ClaimBound),
			pv:       mock.CreatePV(2, "pvc", "pv", mock.DefaultNS, "1", "uid", &mock.FSVolumeMode, v1.VolumePending),
			volumeId: "1",
			wantErr:  true,
		},
		{
			name:     "PV with nil VolumeHandle",
			pvc:      mock.CreatePVC(1, 2, "pvc", "uid", mock.DefaultNS, "pv", v1.ClaimBound),
			pv:       mock.CreatePVWithNilVolumeHandle(2, "pvc", "pv", mock.DefaultNS, "1", "uid", v1.VolumeBound, &mock.FSVolumeMode),
			volumeId: "1",
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := createMockPVHealthConditionChecker(t)
			if err := checker.pvInformer.Informer().GetStore().Add(tt.pv); err != nil {
				t.Errorf("PVHealthConditionChecker.CheckControllerVolumeStatus() error = %v", err)
			}

			if err := checker.pvcInformer.Informer().GetStore().Add(tt.pvc); err != nil {
				t.Errorf("PVHealthConditionChecker.CheckControllerVolumeStatus() error = %v", err)
			}

			in := &csi.ControllerGetVolumeRequest{
				VolumeId: tt.volumeId,
			}

			out := &csi.ControllerGetVolumeResponse{
				Volume: volumeMap[tt.volumeId].Volume,
				Status: &csi.ControllerGetVolumeResponse_VolumeStatus{
					VolumeCondition: volumeMap[tt.volumeId].Condition,
				},
			}

			_, ctx := ktesting.NewTestContext(t)
			checker.csiControllerServer.EXPECT().ControllerGetVolume(gomock.Any(), csitestutil.Protobuf(in)).Return(out, nil).Times(1)
			if err := checker.pvHealthConditionChecker.CheckControllerVolumeStatus(ctx, tt.pv); (err != nil) != tt.wantErr {
				t.Errorf("PVHealthConditionChecker.CheckControllerVolumeStatus() error = %v, wantErr %v", err, tt.wantErr)
			}

			event, err := mock.WatchEvent(tt.wantAbnormalEvent, checker.eventStore)
			if tt.wantAbnormalEvent {
				assert.Nil(err)
				assert.EqualValues(event, mock.AbnormalEvent)
			} else {
				assert.EqualValues(mock.ErrorWatchTimeout.Error(), err.Error())
			}
		})
	}
}
