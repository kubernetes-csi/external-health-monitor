# Release notes for v0.11.0

[Documentation](https://kubernetes-csi.github.io/)

# Changelog since 0.10.0

## Changes by Kind

### Uncategorized

- Update kubernetes dependencies to v1.29.0 ([#226](https://github.com/kubernetes-csi/external-health-monitor/pull/226), [@sunnylovestiramisu](https://github.com/sunnylovestiramisu))

## Dependencies

### Added
- github.com/gorilla/websocket: [v1.5.0](https://github.com/gorilla/websocket/tree/v1.5.0)
- go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc: v0.44.0

### Changed
- cloud.google.com/go/compute: v1.21.0 → v1.23.0
- github.com/container-storage-interface/spec: [v1.8.0 → v1.9.0](https://github.com/container-storage-interface/spec/compare/v1.8.0...v1.9.0)
- github.com/emicklei/go-restful/v3: [v3.9.0 → v3.11.0](https://github.com/emicklei/go-restful/v3/compare/v3.9.0...v3.11.0)
- github.com/go-logr/logr: [v1.2.4 → v1.3.0](https://github.com/go-logr/logr/compare/v1.2.4...v1.3.0)
- github.com/golang/glog: [v1.1.0 → v1.1.2](https://github.com/golang/glog/compare/v1.1.0...v1.1.2)
- github.com/google/go-cmp: [v0.5.9 → v0.6.0](https://github.com/google/go-cmp/compare/v0.5.9...v0.6.0)
- github.com/google/uuid: [v1.3.1 → v1.4.0](https://github.com/google/uuid/compare/v1.3.1...v1.4.0)
- github.com/grpc-ecosystem/grpc-gateway/v2: [v2.7.0 → v2.16.0](https://github.com/grpc-ecosystem/grpc-gateway/v2/compare/v2.7.0...v2.16.0)
- github.com/kubernetes-csi/csi-lib-utils: [v0.14.0 → v0.17.0](https://github.com/kubernetes-csi/csi-lib-utils/compare/v0.14.0...v0.17.0)
- github.com/kubernetes-csi/csi-test/v5: [v5.1.0 → v5.2.0](https://github.com/kubernetes-csi/csi-test/v5/compare/v5.1.0...v5.2.0)
- github.com/onsi/ginkgo/v2: [v2.12.0 → v2.13.1](https://github.com/onsi/ginkgo/v2/compare/v2.12.0...v2.13.1)
- github.com/onsi/gomega: [v1.27.10 → v1.30.0](https://github.com/onsi/gomega/compare/v1.27.10...v1.30.0)
- github.com/yuin/goldmark: [v1.3.5 → v1.4.13](https://github.com/yuin/goldmark/compare/v1.3.5...v1.4.13)
- go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp: v0.35.1 → v0.44.0
- go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc: v1.10.0 → v1.19.0
- go.opentelemetry.io/otel/exporters/otlp/otlptrace: v1.10.0 → v1.19.0
- go.opentelemetry.io/otel/metric: v0.31.0 → v1.19.0
- go.opentelemetry.io/otel/sdk: v1.10.0 → v1.19.0
- go.opentelemetry.io/otel/trace: v1.10.0 → v1.19.0
- go.opentelemetry.io/otel: v1.10.0 → v1.19.0
- go.opentelemetry.io/proto/otlp: v0.19.0 → v1.0.0
- go.uber.org/goleak: v1.2.1 → v1.1.10
- golang.org/x/crypto: v0.12.0 → v0.15.0
- golang.org/x/net: v0.14.0 → v0.18.0
- golang.org/x/oauth2: v0.10.0 → v0.13.0
- golang.org/x/sync: v0.3.0 → v0.4.0
- golang.org/x/sys: v0.11.0 → v0.14.0
- golang.org/x/term: v0.11.0 → v0.14.0
- golang.org/x/text: v0.12.0 → v0.14.0
- golang.org/x/tools: v0.12.0 → v0.14.0
- google.golang.org/appengine: v1.6.7 → v1.6.8
- google.golang.org/genproto/googleapis/api: 782d3b1 → d307bd8
- google.golang.org/genproto/googleapis/rpc: 782d3b1 → bbf56f3
- google.golang.org/genproto: 782d3b1 → d783a09
- google.golang.org/grpc: v1.58.0 → v1.60.1
- k8s.io/api: v0.28.0 → v0.29.0
- k8s.io/apimachinery: v0.28.0 → v0.29.0
- k8s.io/client-go: v0.28.0 → v0.29.0
- k8s.io/component-base: v0.28.0 → v0.29.0
- k8s.io/gengo: 485abfe → 9cce18d
- k8s.io/klog/v2: v2.100.1 → v2.110.1
- k8s.io/kube-openapi: 2695361 → 2dd684a
- k8s.io/utils: d93618c → 3b25d92
- sigs.k8s.io/structured-merge-diff/v4: v4.2.3 → v4.4.1

### Removed
- github.com/google/gnostic: [v0.5.7-v3refs](https://github.com/google/gnostic/tree/v0.5.7-v3refs)
- go.opentelemetry.io/otel/exporters/otlp/internal/retry: v1.10.0
