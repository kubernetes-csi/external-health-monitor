package util

import (
	"reflect"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	pvcToPodsCache = NewPVCToPodsCache()
	pod1           = &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "pod1",
		},
		Spec: v1.PodSpec{
			Volumes: []v1.Volume{
				{
					VolumeSource: v1.VolumeSource{
						PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
							ClaimName: "pvc1",
						},
					},
				},
			},
		},
	}
)

func TestPVCToPodsCache_AddPod(t *testing.T) {
	type args struct {
		pod *v1.Pod
	}
	tests := []struct {
		name    string
		args    args
		pvcName string
		want    PodSet
	}{
		{
			name: "case1",
			args: args{
				pod: pod1,
			},
			pvcName: "pvc1",
			want: PodSet{
				pod1.Namespace + "/" + pod1.Name: pod1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pvcToPodsCache.AddPod(tt.args.pod)
			got := pvcToPodsCache.GetPodsByPVC(pod1.Namespace, tt.pvcName)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PVCToPodCache.AddPod() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPVCToPodsCache_DeletePod(t *testing.T) {
	type args struct {
		pod *v1.Pod
	}

	var w PodSet
	tests := []struct {
		name    string
		args    args
		want    PodSet
		pvcName string
	}{
		{
			name: "case1",
			args: args{
				pod: pod1,
			},
			want:    w,
			pvcName: "pvc1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pvcToPodsCache.DeletePod(tt.args.pod)
			got := pvcToPodsCache.GetPodsByPVC(pod1.Namespace, tt.pvcName)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PVCToPodCache.DeletePod() = %v, want %v", got, tt.want)
			}
		})
	}
}
