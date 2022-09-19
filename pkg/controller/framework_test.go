package pv_monitor_controller

import (
	"context"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/record"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/mock/gomock"
	"github.com/kubernetes-csi/csi-test/v5/driver"
	"github.com/kubernetes-csi/external-health-monitor/pkg/mock"
	"github.com/stretchr/testify/assert"
)

type fakeNativeObjects struct {
	MockVolume *mock.MockVolume
	MockNode   *mock.MockNode
	MockEvent  *mock.MockEvent
}

type testCase struct {
	name               string
	enableNodeWatcher  bool
	fakeNativeObjects  *fakeNativeObjects
	supportListVolumes bool
	wantAbnormalEvent  bool
	hasRecoveryEvent   bool
}

func runTest(t *testing.T, tc *testCase) {
	assert := assert.New(t)
	// Initialize native controller objects
	nativeObjects := []runtime.Object{
		tc.fakeNativeObjects.MockVolume.NativeVolume,
		tc.fakeNativeObjects.MockVolume.NativeVolumeClaim,
	}

	if tc.enableNodeWatcher {
		nativeObjects = append(nativeObjects, tc.fakeNativeObjects.MockNode.NativeNode)
	}

	if tc.hasRecoveryEvent {
		nativeObjects = append(nativeObjects, tc.fakeNativeObjects.MockEvent.NativeEvent)
	}

	client := fake.NewSimpleClientset(nativeObjects...)
	informers := informers.NewSharedInformerFactory(client, 0)
	pvInformer := informers.Core().V1().PersistentVolumes()
	pvcInformer := informers.Core().V1().PersistentVolumeClaims()
	podInformer := informers.Core().V1().Pods()
	nodeInformer := informers.Core().V1().Nodes()
	eventInformer := informers.Core().V1().Events()
	option := &PVMonitorOptions{
		DriverName:                "fake.csi.driver.io",
		ContextTimeout:            15 * time.Second,
		EnableNodeWatcher:         tc.enableNodeWatcher,
		ListVolumesInterval:       5 * time.Minute,
		PVWorkerExecuteInterval:   1 * time.Minute,
		VolumeListAndAddInterval:  5 * time.Minute,
		NodeWorkerExecuteInterval: 1 * time.Minute,
		NodeListAndAddInterval:    5 * time.Minute,
		SupportListVolume:         tc.supportListVolumes,
	}

	_, _, _, controllerServer, _, csiConn, err := mock.CreateMockServer(t)

	assert.Nil(err)

	eventStore := make(chan string, 1)
	eventRecorder := record.FakeRecorder{
		Events: eventStore,
	}

	var volumes []*mock.CSIVolume

	// Inject test cases
	volumes = append(volumes, tc.fakeNativeObjects.MockVolume.CSIVolume)
	err = pvInformer.Informer().GetStore().Add(tc.fakeNativeObjects.MockVolume.NativeVolume)
	assert.Nil(err)
	err = pvcInformer.Informer().GetStore().Add(tc.fakeNativeObjects.MockVolume.NativeVolumeClaim)
	assert.Nil(err)

	if tc.enableNodeWatcher {
		err = nodeInformer.Informer().GetStore().Add(tc.fakeNativeObjects.MockNode.NativeNode)
		assert.Nil(err)
	}

	mockCSIcontrollerServer(controllerServer, tc.supportListVolumes, volumes)
	pvMonitorController := NewPVMonitorController(client, csiConn, pvInformer, pvcInformer, podInformer, nodeInformer, eventInformer, &eventRecorder, option)
	assert.NotNil(pvMonitorController)

	if tc.hasRecoveryEvent {
		err = eventInformer.Informer().GetStore().Add(tc.fakeNativeObjects.MockEvent.NativeEvent)
		assert.Nil(err)
	}

	ctx, cancel := context.WithCancel(context.TODO())
	stopCh := ctx.Done()
	informers.Start(stopCh)
	go pvMonitorController.Run(1, stopCh)

	event, err := mock.WatchEvent(tc.wantAbnormalEvent, eventStore)
	if tc.wantAbnormalEvent {
		assert.Nil(err)
		assert.EqualValues(event, mock.AbnormalEvent)
	} else if tc.hasRecoveryEvent {
		assert.Nil(err)
		assert.EqualValues(event, mock.NormalEvent)
	} else {
		assert.EqualValues(mock.ErrorWatchTimeout.Error(), err.Error())
	}

	cancel()
}

func mockCSIcontrollerServer(csiControllerServer *driver.MockControllerServer, supportListVolume bool, objects []*mock.CSIVolume) {
	if supportListVolume {
		volumeResponseEntries := make([]*csi.ListVolumesResponse_Entry, len(objects))
		for index, volume := range objects {
			volumeResponseEntries[index] = &csi.ListVolumesResponse_Entry{
				Volume: volume.Volume,
				Status: &csi.ListVolumesResponse_VolumeStatus{
					VolumeCondition: volume.Condition,
				},
			}
		}

		in := &csi.ListVolumesRequest{
			StartingToken: "",
		}
		out := &csi.ListVolumesResponse{
			Entries:   volumeResponseEntries,
			NextToken: "",
		}
		csiControllerServer.EXPECT().ListVolumes(gomock.Any(), in).Return(out, nil).Times(100000)
	} else {
		for _, volume := range objects {

			in := &csi.ControllerGetVolumeRequest{
				VolumeId: volume.Volume.VolumeId,
			}
			out := &csi.ControllerGetVolumeResponse{
				Volume: volume.Volume,
				Status: &csi.ControllerGetVolumeResponse_VolumeStatus{
					VolumeCondition: volume.Condition,
				},
			}
			csiControllerServer.EXPECT().ControllerGetVolume(gomock.Any(), in).Return(out, nil).Times(100000)
		}
	}
}
