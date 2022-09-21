# Release notes for v0.7.0

[Documentation](https://kubernetes-csi.github.io/)

# Changelog since 0.6.0

## Changes by Kind

### Uncategorized

- The kubernetes client dependencies are updated to v1.25 ([#134](https://github.com/kubernetes-csi/external-health-monitor/pull/134), [@humblec](https://github.com/humblec))

## Dependencies

### Added
- github.com/emicklei/go-restful/v3: [v3.8.0](https://github.com/emicklei/go-restful/v3/tree/v3.8.0)
- github.com/go-task/slim-sprig: [348f09d](https://github.com/go-task/slim-sprig/tree/348f09d)
- github.com/golang-jwt/jwt/v4: [v4.2.0](https://github.com/golang-jwt/jwt/v4/tree/v4.2.0)
- github.com/kubernetes-csi/csi-test/v5: [v5.0.0](https://github.com/kubernetes-csi/csi-test/v5/tree/v5.0.0)
- github.com/onsi/ginkgo/v2: [v2.1.4](https://github.com/onsi/ginkgo/v2/tree/v2.1.4)

### Changed
- cloud.google.com/go: v0.81.0 → v0.97.0
- github.com/Azure/go-autorest/autorest/adal: [v0.9.13 → v0.9.20](https://github.com/Azure/go-autorest/autorest/adal/compare/v0.9.13...v0.9.20)
- github.com/Azure/go-autorest/autorest: [v0.11.18 → v0.11.27](https://github.com/Azure/go-autorest/autorest/compare/v0.11.18...v0.11.27)
- github.com/cncf/udpa/go: [5459f2c → 04548b0](https://github.com/cncf/udpa/go/compare/5459f2c...04548b0)
- github.com/cncf/xds/go: [fbca930 → cb28da3](https://github.com/cncf/xds/go/compare/fbca930...cb28da3)
- github.com/container-storage-interface/spec: [v1.5.0 → v1.6.0](https://github.com/container-storage-interface/spec/compare/v1.5.0...v1.6.0)
- github.com/emicklei/go-restful: [v2.9.5+incompatible → ff4f55a](https://github.com/emicklei/go-restful/compare/v2.9.5...ff4f55a)
- github.com/envoyproxy/go-control-plane: [63b5d3c → 49ff273](https://github.com/envoyproxy/go-control-plane/compare/63b5d3c...49ff273)
- github.com/go-logr/logr: [v1.2.0 → v1.2.3](https://github.com/go-logr/logr/compare/v1.2.0...v1.2.3)
- github.com/golang/mock: [v1.5.0 → v1.6.0](https://github.com/golang/mock/compare/v1.5.0...v1.6.0)
- github.com/google/go-cmp: [v0.5.5 → v0.5.8](https://github.com/google/go-cmp/compare/v0.5.5...v0.5.8)
- github.com/google/martian/v3: [v3.1.0 → v3.0.0](https://github.com/google/martian/v3/compare/v3.1.0...v3.0.0)
- github.com/google/pprof: [cbba55b → 94a9f03](https://github.com/google/pprof/compare/cbba55b...94a9f03)
- github.com/google/uuid: [v1.1.2 → v1.3.0](https://github.com/google/uuid/compare/v1.1.2...v1.3.0)
- github.com/nxadm/tail: [v1.4.5 → v1.4.8](https://github.com/nxadm/tail/compare/v1.4.5...v1.4.8)
- github.com/onsi/ginkgo: [v1.14.2 → v1.16.4](https://github.com/onsi/ginkgo/compare/v1.14.2...v1.16.4)
- github.com/onsi/gomega: [v1.10.4 → v1.20.0](https://github.com/onsi/gomega/compare/v1.10.4...v1.20.0)
- github.com/yuin/goldmark: [v1.3.5 → v1.4.1](https://github.com/yuin/goldmark/compare/v1.3.5...v1.4.1)
- go.opencensus.io: v0.23.0 → v0.22.4
- golang.org/x/crypto: 8634188 → 3147a52
- golang.org/x/mod: v0.4.2 → 9b9b3d8
- golang.org/x/net: cd36cc0 → 0bcc04d
- golang.org/x/sys: 3681064 → a90be44
- golang.org/x/tools: v0.1.5 → v0.1.10
- google.golang.org/api: v0.43.0 → v0.30.0
- google.golang.org/grpc: v1.40.0 → v1.48.0
- google.golang.org/protobuf: v1.27.1 → v1.28.0
- gopkg.in/yaml.v3: 496545a → v3.0.1
- k8s.io/api: v0.24.0 → v0.25.0
- k8s.io/apimachinery: v0.24.0 → v0.25.0
- k8s.io/client-go: v0.24.0 → v0.25.0
- k8s.io/klog/v2: v2.60.1 → v2.70.1
- k8s.io/kube-openapi: 3ee0da9 → 67bda5d
- k8s.io/utils: 3a6ce19 → ee6ede2
- sigs.k8s.io/json: 9f7c6b3 → f223a00
- sigs.k8s.io/structured-merge-diff/v4: v4.2.1 → v4.2.3

### Removed
- github.com/gorilla/mux: [v1.8.0](https://github.com/gorilla/mux/tree/v1.8.0)
- github.com/kubernetes-csi/csi-test/v4: [v4.3.0](https://github.com/kubernetes-csi/csi-test/v4/tree/v4.3.0)
