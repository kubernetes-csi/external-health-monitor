package pv_monitor_controller

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2/ktesting"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/kubernetes-csi/external-health-monitor/pkg/metrics"
	"github.com/kubernetes-csi/external-health-monitor/pkg/mock"
	"github.com/stretchr/testify/assert"
)

func csiVolume(id string) *csi.Volume {
	return &csi.Volume{VolumeId: id}
}

func abnormalMockVolume() *mock.MockVolume {
	return &mock.MockVolume{
		CSIVolume: &mock.CSIVolume{
			Volume: csiVolume("abnormalVolume1"),
			Health: mock.AbnormalVolumeHealth("abnormalVolume1"),
		},
		NativeVolume:      mock.CreatePV(2, "pvc", "pv", mock.DefaultNS, "abnormalVolume1", "pvcuid", &mock.BlockVolumeMode, v1.VolumeBound),
		NativeVolumeClaim: mock.CreatePVC(1, 2, "pvc", "pvcuid", mock.DefaultNS, "pv", v1.ClaimBound),
	}
}

func healthyMockVolume() *mock.MockVolume {
	return &mock.MockVolume{
		CSIVolume: &mock.CSIVolume{
			Volume: csiVolume("normalVolume1"),
			Health: mock.HealthyVolumeHealth("normalVolume1"),
		},
		NativeVolume:      mock.CreatePV(2, "pvc", "pv", mock.DefaultNS, "normalVolume1", "pvcuid", &mock.BlockVolumeMode, v1.VolumeBound),
		NativeVolumeClaim: mock.CreatePVC(1, 2, "pvc", "pvcuid", mock.DefaultNS, "pv", v1.ClaimBound),
	}
}

func Test_AbnormalVolumeWithListVolumeHealth(t *testing.T) {
	runTest(t, &testCase{
		supportListVolumeHealth: true,
		fakeNativeObjects:       &fakeNativeObjects{MockVolume: abnormalMockVolume()},
		wantAbnormalPatch:       true,
	})
}

func Test_NormalVolumeWithListVolumeHealth(t *testing.T) {
	runTest(t, &testCase{
		supportListVolumeHealth: true,
		fakeNativeObjects:       &fakeNativeObjects{MockVolume: healthyMockVolume()},
		wantAbnormalPatch:       false,
	})
}

func Test_AbnormalVolumeWithGetVolumeHealth(t *testing.T) {
	runTest(t, &testCase{
		supportListVolumeHealth: false,
		fakeNativeObjects:       &fakeNativeObjects{MockVolume: abnormalMockVolume()},
		wantAbnormalPatch:       true,
	})
}

func Test_NormalVolumeWithGetVolumeHealth(t *testing.T) {
	runTest(t, &testCase{
		supportListVolumeHealth: false,
		fakeNativeObjects:       &fakeNativeObjects{MockVolume: healthyMockVolume()},
		wantAbnormalPatch:       false,
	})
}

// newGetModeFixture builds a Get-mode controller whose queue and workers are
// driven manually by the test; the informer factory is never started.
func newGetModeFixture(t *testing.T, pv *v1.PersistentVolume, pvc *v1.PersistentVolumeClaim) (*PVMonitorController, *mock.FakeCSIDriver, cache.Store) {
	t.Helper()
	ctrl, drv, factory := newGetModeFixtureFor(t, 1*time.Minute, []*v1.PersistentVolume{pv}, []*v1.PersistentVolumeClaim{pvc})
	return ctrl, drv, factory.Core().V1().PersistentVolumes().Informer().GetStore()
}

func newGetModeFixtureFor(t *testing.T, interval time.Duration, pvs []*v1.PersistentVolume, pvcs []*v1.PersistentVolumeClaim) (*PVMonitorController, *mock.FakeCSIDriver, informers.SharedInformerFactory) {
	t.Helper()

	objects := make([]runtime.Object, 0, len(pvs)+len(pvcs))
	for _, pv := range pvs {
		objects = append(objects, pv)
	}
	for _, pvc := range pvcs {
		objects = append(objects, pvc)
	}
	client := fake.NewSimpleClientset(objects...)
	factory := informers.NewSharedInformerFactory(client, 0)
	pvStore := factory.Core().V1().PersistentVolumes().Informer().GetStore()
	pvcStore := factory.Core().V1().PersistentVolumeClaims().Informer().GetStore()

	drv, csiConn := mock.StartFakeDriver(t)
	option := &PVMonitorOptions{
		DriverName:               mock.DriverName,
		ContextTimeout:           15 * time.Second,
		ListVolumesInterval:      5 * time.Minute,
		PVWorkerExecuteInterval:  interval,
		VolumeListAndAddInterval: 5 * time.Minute,
		SupportListVolumeHealth:  false,
	}
	ctrl := NewPVMonitorController(client, csiConn, factory, metrics.New(), record.NewFakeRecorder(100), option)
	t.Cleanup(ctrl.pvQueue.ShutDown)

	for _, pv := range pvs {
		if err := pvStore.Add(pv); err != nil {
			t.Fatal(err)
		}
	}
	for _, pvc := range pvcs {
		if err := pvcStore.Add(pvc); err != nil {
			t.Fatal(err)
		}
	}
	return ctrl, drv, factory
}

func boundVolumes(n int) ([]*v1.PersistentVolume, []*v1.PersistentVolumeClaim) {
	pvs := make([]*v1.PersistentVolume, 0, n)
	pvcs := make([]*v1.PersistentVolumeClaim, 0, n)
	for i := 0; i < n; i++ {
		pvName := fmt.Sprintf("pv-%d", i)
		pvcName := fmt.Sprintf("pvc-%d", i)
		pvcUID := fmt.Sprintf("pvcuid-%d", i)
		volumeID := fmt.Sprintf("vol-%d", i)
		pvs = append(pvs, mock.CreatePV(2, pvcName, pvName, mock.DefaultNS, volumeID, types.UID(pvcUID), &mock.BlockVolumeMode, v1.VolumeBound))
		pvcs = append(pvcs, mock.CreatePVC(1, 2, pvcName, pvcUID, mock.DefaultNS, pvName, v1.ClaimBound))
	}
	return pvs, pvcs
}

func getRequestTimesPerVolume(drv *mock.FakeCSIDriver) map[string][]time.Time {
	reqs := drv.Controller.GetVolumeHealthRequests()
	times := drv.Controller.GetVolumeHealthTimes()
	perVolume := map[string][]time.Time{}
	for i, req := range reqs {
		perVolume[req.GetVolumeId()] = append(perVolume[req.GetVolumeId()], times[i])
	}
	return perVolume
}

func Test_PendingPVIsMonitoredOnceBound(t *testing.T) {
	assert := assert.New(t)

	pv := mock.CreatePV(2, "pvc", "pv", mock.DefaultNS, "vol-1", "pvcuid", &mock.BlockVolumeMode, v1.VolumePending)
	pvc := mock.CreatePVC(1, 2, "pvc", "pvcuid", mock.DefaultNS, "pv", v1.ClaimBound)
	ctrl, drv, pvStore := newGetModeFixture(t, pv, pvc)
	drv.Controller.SetGetVolumeHealth("vol-1", &csi.ControllerGetVolumeHealthResponse{VolumeHealth: mock.AbnormalVolumeHealth("vol-1")}, nil)

	// While pending, the periodic pass must leave the PV alone.
	assert.Nil(ctrl.AddPVsToQueue())
	assert.Equal(0, ctrl.pvQueue.Len())
	assert.False(ctrl.pvEnqueued["pv"])

	bound := pv.DeepCopy()
	bound.Status.Phase = v1.VolumeBound
	assert.Nil(pvStore.Update(bound))

	assert.Nil(ctrl.AddPVsToQueue())
	assert.Equal(1, ctrl.pvQueue.Len())

	_, ctx := ktesting.NewTestContext(t)
	ctrl.checkPVWorker(ctx)

	assert.Len(drv.Controller.GetVolumeHealthRequests(), 1, "bound PV must be health-checked")
	assert.True(ctrl.pvEnqueued["pv"], "checked PV must stay tracked for its delayed re-check")
	assert.Equal(0, ctrl.pvQueue.Len(), "checked PV must not be re-enqueued before the monitor interval")
}

func Test_WorkerCountDoesNotChangePerPVCheckCadence(t *testing.T) {
	const interval = 300 * time.Millisecond

	for _, tc := range []struct {
		name    string
		workers int
		volumes int
	}{
		{name: "more workers than volumes", workers: 10, volumes: 1},
		{name: "more volumes than workers", workers: 1, volumes: 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			pvs, pvcs := boundVolumes(tc.volumes)
			ctrl, drv, factory := newGetModeFixtureFor(t, interval, pvs, pvcs)
			for _, pv := range pvs {
				id := pv.Spec.CSI.VolumeHandle
				drv.Controller.SetGetVolumeHealth(id, &csi.ControllerGetVolumeHealthResponse{VolumeHealth: mock.HealthyVolumeHealth(id)}, nil)
			}

			_, ctx := ktesting.NewTestContext(t)
			ctx, cancel := context.WithCancel(ctx)
			factory.Start(ctx.Done())
			var wg sync.WaitGroup
			wg.Go(func() { ctrl.Run(ctx, tc.workers, &wg) })
			t.Cleanup(func() {
				cancel()
				wg.Wait()
			})

			assert.Eventually(func() bool {
				times := getRequestTimesPerVolume(drv)
				for _, pv := range pvs {
					if len(times[pv.Spec.CSI.VolumeHandle]) < 3 {
						return false
					}
				}
				return true
			}, 10*time.Second, 5*time.Millisecond, "every volume must be checked at least three times")

			for volumeID, times := range getRequestTimesPerVolume(drv) {
				for i := 1; i < len(times); i++ {
					gap := times[i].Sub(times[i-1])
					assert.GreaterOrEqual(gap, interval, "%s was checked again before the monitor interval elapsed", volumeID)
					assert.Less(gap, 2*interval, "%s was not checked again within the monitor interval", volumeID)
				}
			}
		})
	}
}

func Test_RequeuePVRequiresTracking(t *testing.T) {
	for _, tc := range []struct {
		name    string
		forget  bool
		wantLen int
	}{
		{
			name:    "tracked",
			wantLen: 1,
		},
		{
			name:    "forgotten",
			forget:  true,
			wantLen: 0,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := &PVMonitorController{
				pvEnqueued: make(map[string]bool),
				pvQueue:    workqueue.NewTypedDelayingQueue[string](),
			}
			t.Cleanup(ctrl.pvQueue.ShutDown)

			ctrl.enqueuePV("pv")
			key, quit := ctrl.pvQueue.Get()
			assert.False(t, quit)
			if tc.forget {
				ctrl.forgetPV("pv")
			}

			ctrl.requeuePV("pv")
			ctrl.pvQueue.Done(key)

			assert.Equal(t, tc.wantLen, ctrl.pvQueue.Len())
		})
	}
}

func Test_PVUnboundAtPopIsForgottenAndPickedUpAgain(t *testing.T) {
	assert := assert.New(t)

	pv := mock.CreatePV(2, "pvc", "pv", mock.DefaultNS, "vol-1", "pvcuid", &mock.BlockVolumeMode, v1.VolumeBound)
	pvc := mock.CreatePVC(1, 2, "pvc", "pvcuid", mock.DefaultNS, "pv", v1.ClaimBound)
	ctrl, drv, pvStore := newGetModeFixture(t, pv, pvc)
	drv.Controller.SetGetVolumeHealth("vol-1", &csi.ControllerGetVolumeHealthResponse{VolumeHealth: mock.AbnormalVolumeHealth("vol-1")}, nil)

	// Enqueued while bound, but released again before the worker pops it.
	ctrl.pvAdded(pv)
	released := pv.DeepCopy()
	released.Status.Phase = v1.VolumeReleased
	assert.Nil(pvStore.Update(released))

	_, ctx := ktesting.NewTestContext(t)
	ctrl.checkPVWorker(ctx)

	assert.Empty(drv.Controller.GetVolumeHealthRequests(), "unbound PV must not be health-checked")
	assert.False(ctrl.pvEnqueued["pv"], "dropped PV must be forgotten so it can be re-added")

	// Once bound again, the periodic pass picks it up and it gets checked.
	assert.Nil(pvStore.Update(pv.DeepCopy()))
	assert.Nil(ctrl.AddPVsToQueue())
	assert.Equal(1, ctrl.pvQueue.Len())

	ctrl.checkPVWorker(ctx)
	assert.Len(drv.Controller.GetVolumeHealthRequests(), 1, "re-bound PV must be health-checked")
}

func Test_DeletingPVIsForgotten(t *testing.T) {
	assert := assert.New(t)

	pv := mock.CreatePV(2, "pvc", "pv", mock.DefaultNS, "vol-1", "pvcuid", &mock.BlockVolumeMode, v1.VolumeBound)
	pvc := mock.CreatePVC(1, 2, "pvc", "pvcuid", mock.DefaultNS, "pv", v1.ClaimBound)
	ctrl, drv, pvStore := newGetModeFixture(t, pv, pvc)

	ctrl.pvAdded(pv)
	deleting := pv.DeepCopy()
	now := metav1.Now()
	deleting.DeletionTimestamp = &now
	assert.Nil(pvStore.Update(deleting))

	_, ctx := ktesting.NewTestContext(t)
	ctrl.checkPVWorker(ctx)

	assert.Empty(drv.Controller.GetVolumeHealthRequests(), "deleting PV must not be health-checked")
	assert.False(ctrl.pvEnqueued["pv"], "deleting PV must not stay tracked")
	assert.Nil(ctrl.AddPVsToQueue())
	assert.Equal(0, ctrl.pvQueue.Len(), "deleting PV must not be re-enqueued by the periodic pass")
}

func Test_PVDeletionForgetsTracking(t *testing.T) {
	assert := assert.New(t)

	pv := mock.CreatePV(2, "pvc", "pv", mock.DefaultNS, "vol-1", "pvcuid", &mock.BlockVolumeMode, v1.VolumeBound)
	pvc := mock.CreatePVC(1, 2, "pvc", "pvcuid", mock.DefaultNS, "pv", v1.ClaimBound)
	ctrl, _, _ := newGetModeFixture(t, pv, pvc)

	ctrl.pvAdded(pv)
	assert.True(ctrl.pvEnqueued["pv"])
	ctrl.pvDeleted(pv)
	assert.False(ctrl.pvEnqueued["pv"], "deleted PV must be forgotten")

	// The informer may hand a tombstone instead of the object.
	ctrl.pvAdded(pv)
	assert.True(ctrl.pvEnqueued["pv"])
	ctrl.pvDeleted(cache.DeletedFinalStateUnknown{Key: "pv", Obj: pv})
	assert.False(ctrl.pvEnqueued["pv"], "tombstoned PV must be forgotten")

	// Deletion events for foreign objects must be ignored without panicking.
	ctrl.pvDeleted(cache.DeletedFinalStateUnknown{Key: "x", Obj: nil})
	ctrl.pvcDeleted(pvc)
	ctrl.pvcDeleted(cache.DeletedFinalStateUnknown{Key: "pvc", Obj: pvc})
}
