package csi_handler

import (
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	informerV1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/tools/record"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/mock/gomock"
	"github.com/kubernetes-csi/csi-test/v3/driver"
	"github.com/kubernetes-csi/external-health-monitor/pkg/mock"
	"github.com/kubernetes-csi/external-health-monitor/pkg/util"
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
			pvcLister:    informer.Core().V1().PersistentVolumeClaims().Lister(),
			pvLister:     informer.Core().V1().PersistentVolumes().Lister(),
			csiPVHandler: handler,
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
		name      string
		pvc       *v1.PersistentVolumeClaim
		pv        *v1.PersistentVolume
		volumeId  string
		wantErr   bool
		wantEvent bool
	}{
		{
			name:      "VolumeConditionAbnormal Case",
			pvc:       mock.CreatePVC(1, 2, "pvc", "uid", mock.DefaultNS, "pv", v1.ClaimBound),
			pv:        mock.CreatePV(2, "pvc", "pv", mock.DefaultNS, "1", "uid", &mock.FSVolumeMode, v1.VolumeBound),
			wantEvent: true,
			volumeId:  "1",
		},
		{
			name:      "VolumeConditionNormal Case",
			pvc:       mock.CreatePVC(1, 2, "pvc", "uid", mock.DefaultNS, "pv", v1.ClaimBound),
			pv:        mock.CreatePV(2, "pvc", "pv", mock.DefaultNS, "2", "uid", &mock.FSVolumeMode, v1.VolumeBound),
			wantEvent: false,
			volumeId:  "2",
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

			checker.csiControllerServer.EXPECT().ListVolumes(gomock.Any(), in).Return(out, nil).Times(1)
			if err := checker.pvHealthConditionChecker.CheckControllerListVolumeStatuses(); (err != nil) != tt.wantErr {
				t.Errorf("PVHealthConditionChecker.CheckControllerListVolumeStatuses() error = %v", err)
			}

			event, err := mock.WatchEvent(tt.wantEvent, checker.eventStore)
			if tt.wantEvent {
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
		name      string
		pv        *v1.PersistentVolume
		pvc       *v1.PersistentVolumeClaim
		volumeId  string
		wantErr   bool
		wantEvent bool
	}{
		{
			name:      "VolumeConditionAbnormal Case",
			pvc:       mock.CreatePVC(1, 2, "pvc", "uid", mock.DefaultNS, "pv", v1.ClaimBound),
			pv:        mock.CreatePV(2, "pvc", "pv", mock.DefaultNS, "1", "uid", &mock.FSVolumeMode, v1.VolumeBound),
			volumeId:  "1",
			wantEvent: true,
		},
		{
			name:      "VolumeConditionNormal Case",
			pvc:       mock.CreatePVC(1, 2, "pvc", "uid", mock.DefaultNS, "pv", v1.ClaimBound),
			pv:        mock.CreatePV(2, "pvc", "pv", mock.DefaultNS, "2", "uid", &mock.FSVolumeMode, v1.VolumeBound),
			wantEvent: false,
			volumeId:  "2",
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

			checker.csiControllerServer.EXPECT().ControllerGetVolume(gomock.Any(), in).Return(out, nil).Times(1)
			if err := checker.pvHealthConditionChecker.CheckControllerVolumeStatus(tt.pv); (err != nil) != tt.wantErr {
				t.Errorf("PVHealthConditionChecker.CheckControllerVolumeStatus() error = %v, wantErr %v", err, tt.wantErr)
			}

			event, err := mock.WatchEvent(tt.wantEvent, checker.eventStore)
			if tt.wantEvent {
				assert.Nil(err)
				assert.EqualValues(event, mock.AbnormalEvent)
			} else {
				assert.EqualValues(mock.ErrorWatchTimeout.Error(), err.Error())
			}
		})
	}
}

func TestPVHealthConditionChecker_CheckNodeVolumeStatus(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		name                string
		pv                  *v1.PersistentVolume
		pvc                 *v1.PersistentVolumeClaim
		pod                 *v1.Pod
		wantErr             bool
		wantEvent           bool
		kubeletRootPath     string
		volumeId            string
		supportStageUnstage bool
	}{
		{
			name: "VolumeConditionNormal Case",
			pvc:  mock.CreatePVC(1, 2, "pvc", "uid", mock.DefaultNS, "pv", v1.ClaimBound),
			pv:   mock.CreatePV(2, "pvc", "pv", mock.DefaultNS, "2", "uid", &mock.FSVolumeMode, v1.VolumeBound),
			pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					UID: "1",
				},
			},
			volumeId: "2",
		},
		{
			name:    "PV without CSI driver Case",
			pvc:     mock.CreatePVC(1, 2, "pvc", "uid", mock.DefaultNS, "pv", v1.ClaimBound),
			pv:      mock.CreatePVWithoutCSIDriver(2, "pvc", "pv", mock.DefaultNS, "1", "uid", v1.VolumeBound, &mock.FSVolumeMode),
			wantErr: true,
			pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					UID: "1",
				},
			},
			volumeId: "1",
		},
		{
			name:    "PV isn't in VolumeBound state",
			pvc:     mock.CreatePVC(1, 2, "pvc", "uid", mock.DefaultNS, "pv", v1.ClaimBound),
			pv:      mock.CreatePV(2, "pvc", "pv", mock.DefaultNS, "1", "uid", &mock.FSVolumeMode, v1.VolumePending),
			wantErr: true,
			pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					UID: "1",
				},
			},
			volumeId: "1",
		},
		{
			name:    "PV with nil VolumeHandle",
			pvc:     mock.CreatePVC(1, 2, "pvc", "uid", mock.DefaultNS, "pv", v1.ClaimBound),
			pv:      mock.CreatePVWithNilVolumeHandle(2, "pvc", "pv", mock.DefaultNS, "1", "uid", v1.VolumePending, &mock.FSVolumeMode),
			wantErr: true,
			pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					UID: "1",
				},
			},
			volumeId: "1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := createMockPVHealthConditionChecker(t)
			if err := checker.pvInformer.Informer().GetStore().Add(tt.pv); err != nil {
				t.Errorf("PVHealthConditionChecker.CheckControllerListVolumeStatuses() error = %v", err)
			}

			if err := checker.pvcInformer.Informer().GetStore().Add(tt.pvc); err != nil {
				t.Errorf("PVHealthConditionChecker.CheckControllerListVolumeStatuses() error = %v", err)
			}
			in := &csi.NodeGetVolumeStatsRequest{
				VolumeId:          tt.volumeId,
				VolumePath:        util.GetVolumePath(tt.kubeletRootPath, tt.pv.Name, string(tt.pod.ObjectMeta.UID)),
				StagingTargetPath: "",
			}
			out := &csi.NodeGetVolumeStatsResponse{
				VolumeCondition: volumeMap[tt.volumeId].Condition,
			}
			checker.csiNodeServer.EXPECT().NodeGetVolumeStats(gomock.Any(), in).Return(out, nil).Times(1)

			if err := checker.pvHealthConditionChecker.CheckNodeVolumeStatus(tt.kubeletRootPath, tt.supportStageUnstage, tt.pv, tt.pod); (err != nil) != tt.wantErr {
				t.Errorf("PVHealthConditionChecker.CheckNodeVolumeStatus() error = %v, wantErr %v", err, tt.wantErr)
			}

			event, err := mock.WatchEvent(tt.wantEvent, checker.eventStore)
			if tt.wantEvent {
				assert.Nil(err)
				assert.EqualValues(event, mock.AbnormalEvent)
			} else {
				assert.EqualValues(mock.ErrorWatchTimeout.Error(), err.Error())
			}
		})
	}
}
