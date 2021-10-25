package pv_monitor_agent

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
	"github.com/kubernetes-csi/csi-test/v3/driver"
	"github.com/kubernetes-csi/external-health-monitor/pkg/mock"
	"github.com/stretchr/testify/assert"
)

type fakeNativeObjects struct {
	MockVolume *mock.MockVolume
	MockNode   *mock.MockNode
	MockPod    *mock.MockPod
	MockEvent  *mock.MockEvent
}

type testCase struct {
	name               string
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
		tc.fakeNativeObjects.MockPod.NativePod,
	}

	if tc.hasRecoveryEvent {
		nativeObjects = append(nativeObjects, tc.fakeNativeObjects.MockEvent.NativeEvent)
	}

	client := fake.NewSimpleClientset(nativeObjects...)
	informers := informers.NewSharedInformerFactory(client, 0)
	pvInformer := informers.Core().V1().PersistentVolumes()
	pvcInformer := informers.Core().V1().PersistentVolumeClaims()
	podInformer := informers.Core().V1().Pods()
	eventInformer := informers.Core().V1().Events()
	_, _, _, _, nodeServer, csiConn, err := mock.CreateMockServer(t)
	assert.Nil(err)

	eventStore := make(chan string, 1)
	eventRecorder := record.FakeRecorder{
		Events: eventStore,
	}

	var (
		volumes []*mock.CSIVolume
	)

	// Inject test cases
	volumes = append(volumes, tc.fakeNativeObjects.MockVolume.CSIVolume)
	err = pvInformer.Informer().GetStore().Add(tc.fakeNativeObjects.MockVolume.NativeVolume)
	assert.Nil(err)
	err = pvcInformer.Informer().GetStore().Add(tc.fakeNativeObjects.MockVolume.NativeVolumeClaim)
	assert.Nil(err)
	err = podInformer.Informer().GetStore().Add(tc.fakeNativeObjects.MockPod.NativePod)
	assert.Nil(err)

	mockCSInodeServer(nodeServer, volumes)
	pvMonitorAgent, err := NewPVMonitorAgent(client, mock.DriverName, csiConn, time.Second*600, 1*time.Minute, pvInformer, pvcInformer, podInformer, eventInformer, false, mock.DefaultKubeletPath, &eventRecorder)
	assert.Nil(err)
	assert.NotNil(pvMonitorAgent)

	if tc.hasRecoveryEvent {
		err = eventInformer.Informer().GetStore().Add(tc.fakeNativeObjects.MockEvent.NativeEvent)
		assert.Nil(err)
	}

	pvMonitorAgent.addPodToQueue(tc.fakeNativeObjects.MockPod.NativePod)

	ctx, cancel := context.WithCancel(context.TODO())
	stopCh := ctx.Done()
	informers.Start(stopCh)
	go pvMonitorAgent.Run(1, stopCh)

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

func mockCSInodeServer(nodeServer *driver.MockNodeServer, objects []*mock.CSIVolume) {
	for _, volume := range objects {
		volumePath := "/var/lib/kubelet/pods/uid/volumes/kubernetes.io~csi/pv/mount"
		if volume.IsBlock {
			volumePath = "/var/lib/kubelet/pods/uid/volumeDevices/kubernetes.io~csi/pv"
		}
		in := &csi.NodeGetVolumeStatsRequest{
			VolumeId:          volume.Volume.VolumeId,
			VolumePath:        volumePath,
			StagingTargetPath: "",
		}
		out := &csi.NodeGetVolumeStatsResponse{
			VolumeCondition: volume.Condition,
		}
		nodeServer.EXPECT().NodeGetVolumeStats(gomock.Any(), in).Return(out, nil).Times(10000)
	}
}
