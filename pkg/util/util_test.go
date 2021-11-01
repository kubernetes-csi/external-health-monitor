package util

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
)

var (
	kubeletRootDir = "/var/lib/kubelet"
)

func TestEscapeQualifiedName(t *testing.T) {
	assert := assert.New(t)
	originalName := "kubernetes.io/empty-dir"
	expectedPVName := "kubernetes.io~empty-dir"
	actualPVName := EscapeQualifiedName(originalName)
	assert.Equal(expectedPVName, actualPVName)
}

func TestGetVolumePath(t *testing.T) {
	assert := assert.New(t)
	originalPVName := "kubernetes.io/empty-dir"
	podUID := "b84244b0-dc4a-4f11-ba7f-0ce5b6c6f831"
	expectedVolumePath := filepath.Join(kubeletRootDir, "pods", podUID, "volumes", EscapeQualifiedName(CSIPluginName), "kubernetes.io~empty-dir", "mount")
	actualVolumePath := GetVolumePath(kubeletRootDir, originalPVName, podUID, false)
	assert.Equal(expectedVolumePath, actualVolumePath)

	expectedBlockVolumePath := filepath.Join(kubeletRootDir, "pods", podUID, "volumeDevices", EscapeQualifiedName(CSIPluginName), "kubernetes.io~empty-dir")
	actualBlockVolumePath := GetVolumePath(kubeletRootDir, originalPVName, podUID, true)
	assert.Equal(expectedBlockVolumePath, actualBlockVolumePath)
}

func TestMakeDeviceMountPath(t *testing.T) {
	assert := assert.New(t)
	// Negative Case: pvName is nil
	_, err := MakeDeviceMountPath("", &v1.PersistentVolume{})
	assert.NotNil(err)

	// Positive Case: pvName is "pv-test"
	expectedMountPath := "/var/lib/kubelet/plugins/kubernetes.io/csi/pv/pvc-431ceccf-7999-11ea-ab4a-fa163ffd8213/globalmount"
	pv := &v1.PersistentVolume{}
	pv.Name = "pvc-431ceccf-7999-11ea-ab4a-fa163ffd8213"
	actualMountPath, err := MakeDeviceMountPath(kubeletRootDir, pv)
	assert.Equal(expectedMountPath, actualMountPath)
}
