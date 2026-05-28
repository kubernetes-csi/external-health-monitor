# Release notes for v0.18.0

[Documentation](https://kubernetes-csi.github.io/)

# Changelog since 0.17.0

## Changes by Kind

### Uncategorized

- Bump k8s dependencies to v1.36.1 ([#360](https://github.com/kubernetes-csi/external-health-monitor/pull/360), [@dfajmon](https://github.com/dfajmon))
- Update release-tools with Go CVE fixes and bump Go to 1.26 ([#352](https://github.com/kubernetes-csi/external-health-monitor/pull/352), [@humblec](https://github.com/humblec))

## Dependencies

### Added
- github.com/cenkalti/backoff/v5: [v5.0.3](https://github.com/cenkalti/backoff/tree/v5.0.3)
- github.com/go-openapi/swag/cmdutils: [v0.26.0](https://github.com/go-openapi/swag/tree/v0.26.0)
- github.com/go-openapi/swag/conv: [v0.26.0](https://github.com/go-openapi/swag/tree/v0.26.0)
- github.com/go-openapi/swag/fileutils: [v0.26.0](https://github.com/go-openapi/swag/tree/v0.26.0)
- github.com/go-openapi/swag/jsonname: [v0.26.0](https://github.com/go-openapi/swag/tree/v0.26.0)
- github.com/go-openapi/swag/jsonutils: [v0.26.0](https://github.com/go-openapi/swag/tree/v0.26.0)
- github.com/go-openapi/swag/loading: [v0.26.0](https://github.com/go-openapi/swag/tree/v0.26.0)
- github.com/go-openapi/swag/mangling: [v0.26.0](https://github.com/go-openapi/swag/tree/v0.26.0)
- github.com/go-openapi/swag/netutils: [v0.26.0](https://github.com/go-openapi/swag/tree/v0.26.0)
- github.com/go-openapi/swag/stringutils: [v0.26.0](https://github.com/go-openapi/swag/tree/v0.26.0)
- github.com/go-openapi/swag/typeutils: [v0.26.0](https://github.com/go-openapi/swag/tree/v0.26.0)
- github.com/go-openapi/swag/yamlutils: [v0.26.0](https://github.com/go-openapi/swag/tree/v0.26.0)
- github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus: [v1.1.0](https://github.com/grpc-ecosystem/go-grpc-middleware/tree/v1.1.0)
- github.com/grpc-ecosystem/go-grpc-middleware/v2: [v2.3.3](https://github.com/grpc-ecosystem/go-grpc-middleware/tree/v2.3.3)
- gonum.org/v1/gonum: [v0.17.0](https://github.com/gonum/gonum/tree/v0.17.0)
- k8s.io/streaming: v0.36.1

### Changed
- cel.dev/expr: v0.24.0 → v0.25.2
- github.com/antlr4-go/antlr/v4: [v4.13.0 → v4.13.1](https://github.com/antlr4-go/antlr/compare/v4.13.0...v4.13.1)
- github.com/container-storage-interface/spec: [v1.11.0 → v1.12.0](https://github.com/container-storage-interface/spec/compare/v1.11.0...v1.12.0)
- github.com/coreos/go-systemd/v22: [v22.5.0 → v22.7.0](https://github.com/coreos/go-systemd/compare/v22.5.0...v22.7.0)
- github.com/emicklei/go-restful/v3: [v3.12.2 → v3.13.0](https://github.com/emicklei/go-restful/compare/v3.12.2...v3.13.0)
- github.com/fsnotify/fsnotify: [v1.9.0 → v1.10.1](https://github.com/fsnotify/fsnotify/compare/v1.9.0...v1.10.1)
- github.com/fxamacker/cbor/v2: [v2.9.0 → v2.9.2](https://github.com/fxamacker/cbor/compare/v2.9.0...v2.9.2)
- github.com/go-openapi/jsonpointer: [v0.21.0 → v0.23.1](https://github.com/go-openapi/jsonpointer/compare/v0.21.0...v0.23.1)
- github.com/go-openapi/jsonreference: [v0.21.0 → v0.21.5](https://github.com/go-openapi/jsonreference/compare/v0.21.0...v0.21.5)
- github.com/go-openapi/swag: [v0.23.0 → v0.26.0](https://github.com/go-openapi/swag/compare/v0.23.0...v0.26.0)
- github.com/google/cel-go: [v0.26.0 → v0.28.1](https://github.com/google/cel-go/compare/v0.26.0...v0.28.1)
- github.com/google/gnostic-models: [v0.7.0 → v0.7.1](https://github.com/google/gnostic-models/compare/v0.7.0...v0.7.1)
- github.com/grpc-ecosystem/grpc-gateway/v2: [v2.26.3 → v2.29.0](https://github.com/grpc-ecosystem/grpc-gateway/compare/v2.26.3...v2.29.0)
- github.com/kubernetes-csi/csi-lib-utils: [v0.23.2 → v0.24.0](https://github.com/kubernetes-csi/csi-lib-utils/compare/v0.23.2...v0.24.0)
- github.com/kubernetes-csi/csi-test/v5: [v5.3.1 → v5.4.0](https://github.com/kubernetes-csi/csi-test/compare/v5.3.1...v5.4.0)
- github.com/prometheus/common: [v0.66.1 → v0.67.5](https://github.com/prometheus/common/compare/v0.66.1...v0.67.5)
- github.com/prometheus/procfs: [v0.16.1 → v0.20.1](https://github.com/prometheus/procfs/compare/v0.16.1...v0.20.1)
- github.com/spf13/cobra: [v1.10.0 → v1.10.2](https://github.com/spf13/cobra/compare/v1.10.0...v1.10.2)
- github.com/spf13/pflag: [v1.0.9 → v1.0.10](https://github.com/spf13/pflag/compare/v1.0.9...v1.0.10)
- go.etcd.io/etcd/api/v3: v3.6.5 → v3.6.11
- go.etcd.io/etcd/client/pkg/v3: v3.6.5 → v3.6.11
- go.etcd.io/etcd/client/v3: v3.6.5 → v3.6.11
- go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc: v0.60.0 → v0.68.0
- go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp: v0.61.0 → v0.68.0
- go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc: v1.34.0 → v1.43.0
- go.opentelemetry.io/otel/exporters/otlp/otlptrace: v1.34.0 → v1.43.0
- go.opentelemetry.io/otel/metric: v1.38.0 → v1.43.0
- go.opentelemetry.io/otel/sdk: v1.36.0 → v1.43.0
- go.opentelemetry.io/otel/trace: v1.38.0 → v1.43.0
- go.opentelemetry.io/otel: v1.38.0 → v1.43.0
- go.opentelemetry.io/proto/otlp: v1.5.0 → v1.10.0
- go.uber.org/zap: [v1.27.0 → v1.28.0](https://github.com/uber-go/zap/compare/v1.27.0...v1.28.0)
- go.yaml.in/yaml/v2: v2.4.3 → v2.4.4
- golang.org/x/crypto: v0.45.0 → v0.49.0
- golang.org/x/exp: v0.0.0-20240719175910-8a7402abbf56 → v0.0.0-20251219203646-944ab1f22d93
- golang.org/x/net: v0.47.0 → v0.52.0
- golang.org/x/oauth2: v0.30.0 → v0.36.0
- golang.org/x/sync: v0.18.0 → v0.20.0
- golang.org/x/sys: v0.38.0 → v0.42.0
- golang.org/x/term: v0.37.0 → v0.41.0
- golang.org/x/text: v0.31.0 → v0.36.0
- golang.org/x/time: v0.9.0 → v0.15.0
- google.golang.org/genproto/googleapis/api: a0af3ef → afd174a4e478
- google.golang.org/genproto/googleapis/rpc: 200df99 → afd174a4e478
- google.golang.org/grpc: [v1.72.2 → v1.81.1](https://github.com/grpc/grpc-go/compare/v1.72.2...v1.81.1)
- google.golang.org/protobuf: v1.36.8 → v1.36.12-0.20260120151049-f2248ac996af
- k8s.io/api: v0.35.0 → v0.36.1
- k8s.io/apimachinery: v0.35.0 → v0.36.1
- k8s.io/apiserver: v0.35.0 → v0.36.1
- k8s.io/client-go: v0.35.0 → v0.36.1
- k8s.io/component-base: v0.35.0 → v0.36.1
- k8s.io/klog/v2: [v2.130.1 → v2.140.0](https://github.com/kubernetes/klog/compare/v2.130.1...v2.140.0)
- k8s.io/kube-openapi: f3f2b99 → 43fb72c5454a
- k8s.io/utils: bc988d5 → b8788abfbbc2
- sigs.k8s.io/apiserver-network-proxy/konnectivity-client: v0.31.2 → v0.35.0
- sigs.k8s.io/structured-merge-diff/v6: [v6.3.0 → v6.4.0](https://github.com/kubernetes-sigs/structured-merge-diff/compare/v6.3.0...v6.4.0)

### Removed
- github.com/cenkalti/backoff/v4: [v4.3.0](https://github.com/cenkalti/backoff/tree/v4.3.0)
- github.com/google/go-cmp: [v0.7.0](https://github.com/google/go-cmp/tree/v0.7.0)
- github.com/grpc-ecosystem/go-grpc-prometheus: [v1.2.0](https://github.com/grpc-ecosystem/go-grpc-prometheus/tree/v1.2.0)
- github.com/josharian/intern: [v1.0.0](https://github.com/josharian/intern/tree/v1.0.0)
- github.com/mailru/easyjson: [v0.9.1](https://github.com/mailru/easyjson/tree/v0.9.1)
- github.com/stoewer/go-strcase: [v1.3.1](https://github.com/stoewer/go-strcase/tree/v1.3.1)
