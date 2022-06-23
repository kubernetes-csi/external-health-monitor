# Release notes for v0.6.0

[Documentation](https://kubernetes-csi.github.io/)

# Changelog since 0.5.0

## Changes by Kind

### Feature

- Remove the logic of external-health-monitor-agent because the logic is moved to kubelet. ([#62](https://github.com/kubernetes-csi/external-health-monitor/pull/62), [@fengzixu](https://github.com/fengzixu))

## Dependencies

### Added
- github.com/armon/go-socks5: [e753329](https://github.com/armon/go-socks5/tree/e753329)
- github.com/cncf/xds/go: [fbca930](https://github.com/cncf/xds/go/tree/fbca930)
- github.com/getkin/kin-openapi: [v0.76.0](https://github.com/getkin/kin-openapi/tree/v0.76.0)
- github.com/google/gnostic: [v0.5.7-v3refs](https://github.com/google/gnostic/tree/v0.5.7-v3refs)
- github.com/gorilla/mux: [v1.8.0](https://github.com/gorilla/mux/tree/v1.8.0)
- github.com/josharian/intern: [v1.0.0](https://github.com/josharian/intern/tree/v1.0.0)
- github.com/kubernetes-csi/csi-test/v4: [v4.3.0](https://github.com/kubernetes-csi/csi-test/v4/tree/v4.3.0)
- sigs.k8s.io/json: 9f7c6b3

### Changed
- cloud.google.com/go: v0.65.0 → v0.81.0
- github.com/emicklei/go-restful: [ff4f55a → v2.9.5+incompatible](https://github.com/emicklei/go-restful/compare/ff4f55a...v2.9.5)
- github.com/envoyproxy/go-control-plane: [668b12f → 63b5d3c](https://github.com/envoyproxy/go-control-plane/compare/668b12f...63b5d3c)
- github.com/evanphx/json-patch: [v4.11.0+incompatible → v4.12.0+incompatible](https://github.com/evanphx/json-patch/compare/v4.11.0...v4.12.0)
- github.com/go-logr/logr: [v0.4.0 → v1.2.0](https://github.com/go-logr/logr/compare/v0.4.0...v1.2.0)
- github.com/go-openapi/jsonpointer: [v0.19.3 → v0.19.5](https://github.com/go-openapi/jsonpointer/compare/v0.19.3...v0.19.5)
- github.com/go-openapi/jsonreference: [v0.19.3 → v0.19.5](https://github.com/go-openapi/jsonreference/compare/v0.19.3...v0.19.5)
- github.com/go-openapi/swag: [v0.19.5 → v0.19.14](https://github.com/go-openapi/swag/compare/v0.19.5...v0.19.14)
- github.com/golang/mock: [v1.4.4 → v1.5.0](https://github.com/golang/mock/compare/v1.4.4...v1.5.0)
- github.com/google/martian/v3: [v3.0.0 → v3.1.0](https://github.com/google/martian/v3/compare/v3.0.0...v3.1.0)
- github.com/google/pprof: [1a94d86 → cbba55b](https://github.com/google/pprof/compare/1a94d86...cbba55b)
- github.com/ianlancetaylor/demangle: [5e5cf60 → 28f6c0f](https://github.com/ianlancetaylor/demangle/compare/5e5cf60...28f6c0f)
- github.com/json-iterator/go: [v1.1.11 → v1.1.12](https://github.com/json-iterator/go/compare/v1.1.11...v1.1.12)
- github.com/kubernetes-csi/csi-lib-utils: [v0.10.0 → v0.11.0](https://github.com/kubernetes-csi/csi-lib-utils/compare/v0.10.0...v0.11.0)
- github.com/mailru/easyjson: [b2ccc51 → v0.7.6](https://github.com/mailru/easyjson/compare/b2ccc51...v0.7.6)
- github.com/modern-go/reflect2: [v1.0.1 → v1.0.2](https://github.com/modern-go/reflect2/compare/v1.0.1...v1.0.2)
- github.com/munnerz/goautoneg: [a547fc6 → a7dc8b6](https://github.com/munnerz/goautoneg/compare/a547fc6...a7dc8b6)
- github.com/nxadm/tail: [v1.4.4 → v1.4.5](https://github.com/nxadm/tail/compare/v1.4.4...v1.4.5)
- github.com/onsi/ginkgo: [v1.14.0 → v1.14.2](https://github.com/onsi/ginkgo/compare/v1.14.0...v1.14.2)
- github.com/onsi/gomega: [v1.10.1 → v1.10.4](https://github.com/onsi/gomega/compare/v1.10.1...v1.10.4)
- go.opencensus.io: v0.22.4 → v0.23.0
- golang.org/x/crypto: 5ea612d → 8634188
- golang.org/x/net: 37e1c6a → cd36cc0
- golang.org/x/oauth2: 08078c5 → d3ed0bb
- golang.org/x/sys: 59db8d7 → 3681064
- golang.org/x/term: 6a3ed07 → 03fcf44
- golang.org/x/text: v0.3.6 → v0.3.7
- golang.org/x/time: 1f47c86 → 90d013b
- golang.org/x/tools: v0.1.2 → v0.1.5
- google.golang.org/api: v0.30.0 → v0.43.0
- google.golang.org/grpc: v1.38.0 → v1.40.0
- google.golang.org/protobuf: v1.26.0 → v1.27.1
- k8s.io/api: v0.22.0 → v0.24.0
- k8s.io/apimachinery: v0.22.0 → v0.24.0
- k8s.io/client-go: v0.22.0 → v0.24.0
- k8s.io/gengo: 3a45101 → 485abfe
- k8s.io/klog/v2: v2.9.0 → v2.60.1
- k8s.io/kube-openapi: 9528897 → 3ee0da9
- k8s.io/utils: 4b05e18 → 3a6ce19
- sigs.k8s.io/structured-merge-diff/v4: v4.1.2 → v4.2.1

### Removed
- github.com/kubernetes-csi/csi-test/v3: [1899199](https://github.com/kubernetes-csi/csi-test/v3/tree/1899199)
- github.com/robertkrimen/otto: [c382bd3](https://github.com/robertkrimen/otto/tree/c382bd3)
- gopkg.in/sourcemap.v1: v1.0.5
