# Release notes for v0.15.0

[Documentation](https://kubernetes-csi.github.io/)

# Changelog since 0.14.0

## Changes by Kind

### Feature

- Add a new `--automaxprocs` flag to set the `GOMAXPROCS` environment variable to match the configured Linux container CPU quota. ([#325](https://github.com/kubernetes-csi/external-health-monitor/pull/325), [@nixpanic](https://github.com/nixpanic))
- Update kubernetes dependencies to v1.33.0 ([#324](https://github.com/kubernetes-csi/external-health-monitor/pull/324), [@Aishwarya-Hebbar](https://github.com/Aishwarya-Hebbar))

## Dependencies

### Added
- github.com/prashantv/gostub: [v1.1.0](https://github.com/prashantv/gostub/tree/v1.1.0)
- go.uber.org/automaxprocs: v1.6.0
- sigs.k8s.io/randfill: v1.0.0

### Changed
- github.com/google/btree: [v1.0.1 → v1.1.3](https://github.com/google/btree/compare/v1.0.1...v1.1.3)
- github.com/google/go-cmp: [v0.6.0 → v0.7.0](https://github.com/google/go-cmp/compare/v0.6.0...v0.7.0)
- github.com/google/gofuzz: [v1.2.0 → v1.0.0](https://github.com/google/gofuzz/compare/v1.2.0...v1.0.0)
- github.com/gorilla/websocket: [v1.5.0 → e064f32](https://github.com/gorilla/websocket/compare/v1.5.0...e064f32)
- github.com/grpc-ecosystem/grpc-gateway/v2: [v2.20.0 → v2.24.0](https://github.com/grpc-ecosystem/grpc-gateway/v2/compare/v2.20.0...v2.24.0)
- github.com/klauspost/compress: [v1.17.11 → v1.18.0](https://github.com/klauspost/compress/compare/v1.17.11...v1.18.0)
- github.com/kubernetes-csi/csi-lib-utils: [v0.20.0 → v0.22.0](https://github.com/kubernetes-csi/csi-lib-utils/compare/v0.20.0...v0.22.0)
- github.com/prometheus/client_golang: [v1.20.5 → v1.22.0](https://github.com/prometheus/client_golang/compare/v1.20.5...v1.22.0)
- github.com/prometheus/common: [v0.61.0 → v0.62.0](https://github.com/prometheus/common/compare/v0.61.0...v0.62.0)
- go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp: v0.53.0 → v0.58.0
- go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc: v1.27.0 → v1.33.0
- go.opentelemetry.io/otel/exporters/otlp/otlptrace: v1.28.0 → v1.33.0
- go.opentelemetry.io/otel/sdk: v1.31.0 → v1.33.0
- go.opentelemetry.io/proto/otlp: v1.3.1 → v1.4.0
- golang.org/x/crypto: v0.30.0 → v0.36.0
- golang.org/x/net: v0.32.0 → v0.38.0
- golang.org/x/oauth2: v0.24.0 → v0.27.0
- golang.org/x/sync: v0.10.0 → v0.12.0
- golang.org/x/sys: v0.28.0 → v0.31.0
- golang.org/x/term: v0.27.0 → v0.30.0
- golang.org/x/text: v0.21.0 → v0.23.0
- golang.org/x/time: v0.8.0 → v0.9.0
- google.golang.org/genproto/googleapis/api: 796eee8 → e6fa225
- google.golang.org/grpc: v1.69.0 → v1.69.2
- google.golang.org/protobuf: v1.36.0 → v1.36.5
- k8s.io/api: v0.32.0 → v0.33.0
- k8s.io/apimachinery: v0.32.0 → v0.33.0
- k8s.io/client-go: v0.32.0 → v0.33.0
- k8s.io/component-base: v0.32.0 → v0.33.0
- k8s.io/kube-openapi: 2c72e55 → c8a335a
- sigs.k8s.io/structured-merge-diff/v4: v4.5.0 → v4.6.0

### Removed
- github.com/go-kit/log: [v0.2.1](https://github.com/go-kit/log/tree/v0.2.1)
- github.com/go-logfmt/logfmt: [v0.5.1](https://github.com/go-logfmt/logfmt/tree/v0.5.1)
