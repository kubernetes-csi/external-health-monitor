module github.com/kubernetes-csi/external-health-monitor

go 1.13

require (
	github.com/container-storage-interface/spec v1.3.0
	github.com/golang/mock v1.3.1
	github.com/imdario/mergo v0.3.9 // indirect
	github.com/kubernetes-csi/csi-lib-utils v0.9.0
	github.com/kubernetes-csi/csi-test/v3 v3.1.2-0.20200722022205-189919973123
	github.com/stretchr/testify v1.5.1
	google.golang.org/grpc v1.29.0
	k8s.io/api v0.19.0
	k8s.io/apimachinery v0.19.0
	k8s.io/client-go v0.19.0
	k8s.io/klog/v2 v2.2.0
)

replace k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.19.0

replace k8s.io/apiserver => k8s.io/apiserver v0.19.0
