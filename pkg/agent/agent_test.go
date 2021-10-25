package pv_monitor_agent

import (
	"os"
	"testing"

	v1 "k8s.io/api/core/v1"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/kubernetes-csi/external-health-monitor/pkg/mock"
	"github.com/kubernetes-csi/external-health-monitor/pkg/util"
)

func Test_AbnormalVolume(t *testing.T) {
	abnormalVolume := &mock.MockVolume{
		CSIVolume: &mock.CSIVolume{
			Volume: &csi.Volume{
				VolumeId: "abnormalVolume1",
			},
			Condition: &csi.VolumeCondition{
				Abnormal: true,
				Message:  "Volume not found",
			},
			IsBlock: true,
		},
		NativeVolume:      mock.CreatePV(2, "pvc", "pv", mock.DefaultNS, "abnormalVolume1", "pvcuid", &mock.FSVolumeMode, v1.VolumeBound),
		NativeVolumeClaim: mock.CreatePVC(1, 2, "pvc", "pvcuid", mock.DefaultNS, "pv", v1.ClaimBound),
	}

	pod := &mock.MockPod{
		NativePod: mock.CreatePod("pod", mock.DefaultNS, "pv", "pvc", "node", "uid", false),
	}

	testCase := &testCase{
		name: "abnormal_volume_case1",
		fakeNativeObjects: &fakeNativeObjects{
			MockVolume: abnormalVolume,
			MockPod:    pod,
		},
		wantAbnormalEvent: true,
	}

	os.Setenv("NODE_NAME", "node")
	runTest(t, testCase)
}

func Test_NormalVolume(t *testing.T) {
	normalVolume := &mock.MockVolume{
		CSIVolume: &mock.CSIVolume{
			Volume: &csi.Volume{
				VolumeId: "normalVolume1",
			},
			Condition: &csi.VolumeCondition{
				Abnormal: false,
				Message:  "",
			},
			IsBlock: true,
		},
		NativeVolume:      mock.CreatePV(2, "pvc", "pv", mock.DefaultNS, "normalVolume1", "pvcuid", &mock.FSVolumeMode, v1.VolumeBound),
		NativeVolumeClaim: mock.CreatePVC(1, 2, "pvc", "pvcuid", mock.DefaultNS, "pv", v1.ClaimBound),
	}

	pod := &mock.MockPod{
		NativePod: mock.CreatePod("pod", mock.DefaultNS, "pv", "pvc", "node", "uid", false),
	}

	testCase := &testCase{
		name: "normal_volume_case1",
		fakeNativeObjects: &fakeNativeObjects{
			MockVolume: normalVolume,
			MockPod:    pod,
		},
		wantAbnormalEvent: false,
	}

	os.Setenv("NODE_NAME", "node")
	runTest(t, testCase)
}

func Test_RecoveryEvent(t *testing.T) {
	normalVolume := &mock.MockVolume{
		CSIVolume: &mock.CSIVolume{
			Volume: &csi.Volume{
				VolumeId: "normalVolume1",
			},
			Condition: &csi.VolumeCondition{
				Abnormal: false,
				Message:  util.DefaultRecoveryEventMessage,
			},
			IsBlock: true,
		},
		NativeVolume:      mock.CreatePV(2, "pvc", "pv", mock.DefaultNS, "normalVolume1", "pvcuid", &mock.FSVolumeMode, v1.VolumeBound),
		NativeVolumeClaim: mock.CreatePVC(1, 2, "pvc", "pvcuid", mock.DefaultNS, "pv", v1.ClaimBound),
	}

	pod := &mock.MockPod{
		NativePod: mock.CreatePod("pod", mock.DefaultNS, "pv", "pvc", "node", "uid", false),
	}

	oldAbnormalEvent := &mock.MockEvent{
		NativeEvent: mock.CreateEvent("event", "", "uid", v1.EventTypeWarning, "VolumeConditionAbnormal"),
	}

	testCase := &testCase{
		name: "recovery_test_case1",
		fakeNativeObjects: &fakeNativeObjects{
			MockVolume: normalVolume,
			MockPod:    pod,
			MockEvent:  oldAbnormalEvent,
		},
		wantAbnormalEvent: false,
		hasRecoveryEvent:  true,
	}

	os.Setenv("NODE_NAME", "node")
	runTest(t, testCase)
}
