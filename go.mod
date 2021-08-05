module github.com/kubernetes-csi/external-health-monitor

go 1.16

require (
	github.com/container-storage-interface/spec v1.5.0
	github.com/golang/mock v1.4.4
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/imdario/mergo v0.3.11 // indirect
	github.com/kubernetes-csi/csi-lib-utils v0.9.0
	github.com/kubernetes-csi/csi-test/v3 v3.1.2-0.20200722022205-189919973123
	github.com/stretchr/testify v1.7.0
	golang.org/x/oauth2 v0.0.0-20201208152858-08078c50e5b5 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/grpc v1.38.0
	k8s.io/api v0.22.0
	k8s.io/apimachinery v0.22.0
	k8s.io/client-go v0.22.0
	k8s.io/component-base v0.22.0 // indirect
	k8s.io/klog/v2 v2.9.0
)
