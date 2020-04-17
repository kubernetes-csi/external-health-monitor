module github.com/kubernetes-csi/external-health-monitor

go 1.12

require (
	github.com/container-storage-interface/spec v1.2.0-rc1.0.20200415022618-e129a75169c1
	github.com/golang/protobuf v1.3.5 // indirect
	github.com/imdario/mergo v0.3.9 // indirect
	github.com/kubernetes-csi/csi-lib-utils v0.7.0
	golang.org/x/lint v0.0.0-20190313153728-d0100b6bd8b3 // indirect
	golang.org/x/tools v0.0.0-20190524140312-2c0ae7006135 // indirect
	google.golang.org/grpc v1.28.0
	honnef.co/go/tools v0.0.0-20190523083050-ea95bdfd59fc // indirect
	k8s.io/api v0.17.0
	k8s.io/apimachinery v0.17.1-beta.0
	k8s.io/client-go v0.17.0
	k8s.io/klog v1.0.0
	k8s.io/utils v0.0.0-20200327001022-6496210b90e8 // indirect
)

replace k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.17.0

replace k8s.io/apiserver => k8s.io/apiserver v0.17.0
