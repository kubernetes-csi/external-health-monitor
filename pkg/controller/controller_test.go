package pv_monitor_controller

import (
	"testing"

	v1 "k8s.io/api/core/v1"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/kubernetes-csi/external-health-monitor/pkg/mock"
)

func Test_AbnormalVolumeWithoutNodeWatcher(t *testing.T) {
	abnormalVolume := &mock.MockVolume{
		CSIVolume: &mock.CSIVolume{
			Volume: &csi.Volume{
				VolumeId: "abnormalVolume1",
			},
			Condition: &csi.VolumeCondition{
				Abnormal: true,
				Message:  "Volume not found",
			},
		},
		NativeVolume:      mock.CreatePV(2, "pvc", "pv", mock.DefaultNS, "abnormalVolume1", "pvcuid", &mock.FSVolumeMode, v1.VolumeBound),
		NativeVolumeClaim: mock.CreatePVC(1, 2, "pvc", "pvcuid", mock.DefaultNS, "pv", v1.ClaimBound),
	}

	testCase := &testCase{
		name:               "abnormal_volume_case1",
		enableNodeWatcher:  false,
		supportListVolumes: true,
		fakeNativeObjects: &fakeNativeObjects{
			MockVolume: abnormalVolume,
		},
		wantAbnormalEvent: true,
	}

	runTest(t, testCase)
}

func Test_AbnormalVolumeWithNodeWatcher(t *testing.T) {
	abnormalVolume := &mock.MockVolume{
		CSIVolume: &mock.CSIVolume{
			Volume: &csi.Volume{
				VolumeId: "abnormalVolume1",
			},
			Condition: &csi.VolumeCondition{
				Abnormal: true,
				Message:  "Volume not found",
			},
		},
		NativeVolume:      mock.CreatePV(2, "pvc", "pv", mock.DefaultNS, "abnormalVolume1", "pvcuid", &mock.FSVolumeMode, v1.VolumeBound),
		NativeVolumeClaim: mock.CreatePVC(1, 2, "pvc", "pvcuid", mock.DefaultNS, "pv", v1.ClaimBound),
	}

	abnormalNodes := &mock.MockNode{
		NativeNode: mock.CreateNode("node1", ""),
	}

	testCase := &testCase{
		name:               "abnormal_volume_case1",
		enableNodeWatcher:  true,
		supportListVolumes: true,
		fakeNativeObjects: &fakeNativeObjects{
			MockVolume: abnormalVolume,
			MockNode:   abnormalNodes,
		},
		wantAbnormalEvent: true,
	}

	runTest(t, testCase)
}

func Test_NormalVolumeWithoutNodeWatcher(t *testing.T) {
	normalVolume := &mock.MockVolume{
		CSIVolume: &mock.CSIVolume{
			Volume: &csi.Volume{
				VolumeId: "normalVolume1",
			},
			Condition: &csi.VolumeCondition{
				Abnormal: false,
				Message:  "",
			},
		},
		NativeVolume:      mock.CreatePV(2, "pvc", "pv", mock.DefaultNS, "normalVolume1", "pvcuid", &mock.FSVolumeMode, v1.VolumeBound),
		NativeVolumeClaim: mock.CreatePVC(1, 2, "pvc", "pvcuid", mock.DefaultNS, "pv", v1.ClaimBound),
	}

	testCase := &testCase{
		name:               "normal_volume_case1",
		enableNodeWatcher:  false,
		supportListVolumes: true,
		fakeNativeObjects: &fakeNativeObjects{
			MockVolume: normalVolume,
		},
		wantAbnormalEvent: false,
	}

	runTest(t, testCase)
}

func Test_AbnormalVolumeWithoutNodeWatcherAndListVolume(t *testing.T) {
	abnormalVolume := &mock.MockVolume{
		CSIVolume: &mock.CSIVolume{
			Volume: &csi.Volume{
				VolumeId: "abnormalVolume1",
			},
			Condition: &csi.VolumeCondition{
				Abnormal: true,
				Message:  "Volume not found",
			},
		},
		NativeVolume:      mock.CreatePV(2, "pvc", "pv", mock.DefaultNS, "abnormalVolume1", "pvcuid", &mock.FSVolumeMode, v1.VolumeBound),
		NativeVolumeClaim: mock.CreatePVC(1, 2, "pvc", "pvcuid", mock.DefaultNS, "pv", v1.ClaimBound),
	}

	testCase := &testCase{
		name:               "abnormal_volume_case1",
		enableNodeWatcher:  false,
		supportListVolumes: false,
		fakeNativeObjects: &fakeNativeObjects{
			MockVolume: abnormalVolume,
		},
		wantAbnormalEvent: true,
	}

	runTest(t, testCase)
}

func Test_AbnormalVolumeWithNodeWatcherNoListVolume(t *testing.T) {
	abnormalVolume := &mock.MockVolume{
		CSIVolume: &mock.CSIVolume{
			Volume: &csi.Volume{
				VolumeId: "abnormalVolume1",
			},
			Condition: &csi.VolumeCondition{
				Abnormal: true,
				Message:  "Volume not found",
			},
		},
		NativeVolume:      mock.CreatePV(2, "pvc", "pv", mock.DefaultNS, "abnormalVolume1", "pvcuid", &mock.FSVolumeMode, v1.VolumeBound),
		NativeVolumeClaim: mock.CreatePVC(1, 2, "pvc", "pvcuid", mock.DefaultNS, "pv", v1.ClaimBound),
	}

	abnormalNodes := &mock.MockNode{
		NativeNode: mock.CreateNode("node1", ""),
	}

	testCase := &testCase{
		name:               "abnormal_volume_case1",
		enableNodeWatcher:  true,
		supportListVolumes: false,
		fakeNativeObjects: &fakeNativeObjects{
			MockVolume: abnormalVolume,
			MockNode:   abnormalNodes,
		},
		wantAbnormalEvent: true,
	}

	runTest(t, testCase)
}

func Test_NormalVolumeWithoutNodeWatcherAndListVolume(t *testing.T) {
	normalVolume := &mock.MockVolume{
		CSIVolume: &mock.CSIVolume{
			Volume: &csi.Volume{
				VolumeId: "normalVolume1",
			},
			Condition: &csi.VolumeCondition{
				Abnormal: false,
				Message:  "",
			},
		},
		NativeVolume:      mock.CreatePV(2, "pvc", "pv", mock.DefaultNS, "normalVolume1", "pvcuid", &mock.FSVolumeMode, v1.VolumeBound),
		NativeVolumeClaim: mock.CreatePVC(1, 2, "pvc", "pvcuid", mock.DefaultNS, "pv", v1.ClaimBound),
	}

	testCase := &testCase{
		name:               "normal_volume_case1",
		enableNodeWatcher:  false,
		supportListVolumes: false,
		fakeNativeObjects: &fakeNativeObjects{
			MockVolume: normalVolume,
		},
		wantAbnormalEvent: false,
	}

	runTest(t, testCase)
}

func Test_RecoveryEventWithListVolume(t *testing.T) {
	normalVolume := &mock.MockVolume{
		CSIVolume: &mock.CSIVolume{
			Volume: &csi.Volume{
				VolumeId: "normalVolume1",
			},
			Condition: &csi.VolumeCondition{
				Abnormal: false,
				Message:  "Volume is healthy",
			},
		},
		NativeVolume:      mock.CreatePV(2, "pvc", "pv", mock.DefaultNS, "normalVolume1", "pvcuid", &mock.FSVolumeMode, v1.VolumeBound),
		NativeVolumeClaim: mock.CreatePVC(1, 2, "pvc", "pvcuid", mock.DefaultNS, "pv", v1.ClaimBound),
	}

	oldAbnormalEvent := &mock.MockEvent{
		NativeEvent: mock.CreateEvent("event", "", "pvcuid", v1.EventTypeWarning, "VolumeConditionAbnormal"),
	}
	testCase := &testCase{
		name:               "normal_volume_recovery_event",
		enableNodeWatcher:  false,
		supportListVolumes: true,
		fakeNativeObjects: &fakeNativeObjects{
			MockVolume: normalVolume,
			MockEvent:  oldAbnormalEvent,
		},
		wantAbnormalEvent: false,
		hasRecoveryEvent:  true,
	}

	runTest(t, testCase)
}

func Test_RecoveryEventWithoutListVolume(t *testing.T) {
	normalVolume := &mock.MockVolume{
		CSIVolume: &mock.CSIVolume{
			Volume: &csi.Volume{
				VolumeId: "normalVolume1",
			},
			Condition: &csi.VolumeCondition{
				Abnormal: false,
				Message:  "Volume is healthy",
			},
		},
		NativeVolume:      mock.CreatePV(2, "pvc", "pv", mock.DefaultNS, "normalVolume1", "pvcuid", &mock.FSVolumeMode, v1.VolumeBound),
		NativeVolumeClaim: mock.CreatePVC(1, 2, "pvc", "pvcuid", mock.DefaultNS, "pv", v1.ClaimBound),
	}

	oldAbnormalEvent := &mock.MockEvent{
		NativeEvent: mock.CreateEvent("event", "", "pvcuid", v1.EventTypeWarning, "VolumeConditionAbnormal"),
	}
	testCase := &testCase{
		name:               "normal_volume_recovery_event",
		enableNodeWatcher:  false,
		supportListVolumes: false,
		fakeNativeObjects: &fakeNativeObjects{
			MockVolume: normalVolume,
			MockEvent:  oldAbnormalEvent,
		},
		wantAbnormalEvent: false,
		hasRecoveryEvent:  true,
	}

	runTest(t, testCase)
}
