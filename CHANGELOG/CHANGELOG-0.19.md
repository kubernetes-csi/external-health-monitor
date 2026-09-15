# Release notes for v0.19.0

[Documentation](https://kubernetes-csi.github.io/)

# Changelog since 0.18.0

## Changes by Kind

### Feature

- The controller now records reported volume health on the PersistentVolumeClaim `status.healthStatus` field instead of emitting events. ([#376](https://github.com/kubernetes-csi/external-health-monitor/pull/376), [@torredil](https://github.com/torredil))

### Uncategorized

- Upgrade dependencies ([#381](https://github.com/kubernetes-csi/external-health-monitor/pull/381), [@torredil](https://github.com/torredil))
- Update release-tools to fix Go CVEs and spelling-check argument injection ([#374](https://github.com/kubernetes-csi/external-health-monitor/pull/374), [@humblec](https://github.com/humblec))
- Fix release on leader election code for health monitor ([#372](https://github.com/kubernetes-csi/external-health-monitor/pull/372), [@gnufied](https://github.com/gnufied))

## Dependencies

### Added
- github.com/go-openapi/swag/pools: [v0.29.2](https://github.com/go-openapi/swag/tree/v0.29.2)

### Changed
- cel.dev/expr: v0.25.2 → v0.25.3
- github.com/container-storage-interface/spec: [v1.12.0 → v1.13.0](https://github.com/container-storage-interface/spec/compare/v1.12.0...v1.13.0)
- github.com/felixge/httpsnoop: [v1.0.4 → v1.1.0](https://github.com/felixge/httpsnoop/compare/v1.0.4...v1.1.0)
- github.com/fxamacker/cbor/v2: [v2.9.2 → v2.9.3](https://github.com/fxamacker/cbor/compare/v2.9.2...v2.9.3)
- github.com/go-logr/logr: [v1.4.3 → v1.4.4](https://github.com/go-logr/logr/compare/v1.4.3...v1.4.4)
- github.com/go-openapi/jsonpointer: [v0.23.1 → v1.0.1](https://github.com/go-openapi/jsonpointer/compare/v0.23.1...v1.0.1)
- github.com/go-openapi/jsonreference: [v0.21.5 → v1.0.2](https://github.com/go-openapi/jsonreference/compare/v0.21.5...v1.0.2)
- github.com/go-openapi/swag: [v0.26.0 → v0.29.2](https://github.com/go-openapi/swag/compare/v0.26.0...v0.29.2)
- github.com/go-openapi/swag/cmdutils: [v0.26.0 → v0.29.2](https://github.com/go-openapi/swag/compare/v0.26.0...v0.29.2)
- github.com/go-openapi/swag/conv: [v0.26.0 → v0.29.2](https://github.com/go-openapi/swag/compare/v0.26.0...v0.29.2)
- github.com/go-openapi/swag/fileutils: [v0.26.0 → v0.29.2](https://github.com/go-openapi/swag/compare/v0.26.0...v0.29.2)
- github.com/go-openapi/swag/jsonutils: [v0.26.0 → v0.29.2](https://github.com/go-openapi/swag/compare/v0.26.0...v0.29.2)
- github.com/go-openapi/swag/loading: [v0.26.0 → v0.29.2](https://github.com/go-openapi/swag/compare/v0.26.0...v0.29.2)
- github.com/go-openapi/swag/mangling: [v0.26.0 → v0.29.2](https://github.com/go-openapi/swag/compare/v0.26.0...v0.29.2)
- github.com/go-openapi/swag/netutils: [v0.26.0 → v0.29.2](https://github.com/go-openapi/swag/compare/v0.26.0...v0.29.2)
- github.com/go-openapi/swag/stringutils: [v0.26.0 → v0.29.2](https://github.com/go-openapi/swag/compare/v0.26.0...v0.29.2)
- github.com/go-openapi/swag/typeutils: [v0.26.0 → v0.29.2](https://github.com/go-openapi/swag/compare/v0.26.0...v0.29.2)
- github.com/go-openapi/swag/yamlutils: [v0.26.0 → v0.29.2](https://github.com/go-openapi/swag/compare/v0.26.0...v0.29.2)
- github.com/google/cel-go: [v0.28.1 → v0.31.0](https://github.com/google/cel-go/compare/v0.28.1...v0.31.0)
- github.com/grpc-ecosystem/go-grpc-middleware/v2: [v2.3.3 → v2.3.4](https://github.com/grpc-ecosystem/go-grpc-middleware/compare/v2.3.3...v2.3.4)
- github.com/grpc-ecosystem/grpc-gateway/v2: [v2.29.0 → v2.30.0](https://github.com/grpc-ecosystem/grpc-gateway/compare/v2.29.0...v2.30.0)
- github.com/kubernetes-csi/csi-lib-utils: [v0.24.0 → v0.25.0](https://github.com/kubernetes-csi/csi-lib-utils/compare/v0.24.0...v0.25.0)
- github.com/prometheus/client_golang: [v1.23.2 → v1.24.1](https://github.com/prometheus/client_golang/compare/v1.23.2...v1.24.1)
- github.com/prometheus/client_model: [v0.6.2 → v0.6.3](https://github.com/prometheus/client_model/compare/v0.6.2...v0.6.3)
- github.com/prometheus/common: [v0.67.5 → v0.71.0](https://github.com/prometheus/common/compare/v0.67.5...v0.71.0)
- github.com/prometheus/procfs: [v0.20.1 → v0.22.0](https://github.com/prometheus/procfs/compare/v0.20.1...v0.22.0)
- github.com/stretchr/testify: [v1.11.1 → v1.12.1](https://github.com/stretchr/testify/compare/v1.11.1...v1.12.1)
- go.etcd.io/etcd/api/v3: v3.6.11 → v3.7.1
- go.etcd.io/etcd/client/pkg/v3: v3.6.11 → v3.7.1
- go.etcd.io/etcd/client/v3: v3.6.11 → v3.7.1
- go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc: v0.68.0 → v0.71.0
- go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp: v0.68.0 → v0.71.0
- go.opentelemetry.io/otel: v1.43.0 → v1.46.0
- go.opentelemetry.io/otel/exporters/otlp/otlptrace: v1.43.0 → v1.46.0
- go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc: v1.43.0 → v1.46.0
- go.opentelemetry.io/otel/metric: v1.43.0 → v1.46.0
- go.opentelemetry.io/otel/sdk: v1.43.0 → v1.46.0
- go.opentelemetry.io/otel/trace: v1.43.0 → v1.46.0
- go.opentelemetry.io/proto/otlp: v1.10.0 → v1.11.0
- go.yaml.in/yaml/v3: v3.0.4 → v3.0.5
- golang.org/x/crypto: v0.49.0 → v0.57.0
- golang.org/x/exp: v0.0.0-20251219203646-944ab1f22d93 → v0.0.0-20260824195058-e88cd73687aa
- golang.org/x/net: v0.52.0 → v0.59.0
- golang.org/x/oauth2: v0.36.0 → v0.37.0
- golang.org/x/sync: v0.20.0 → v0.23.0
- golang.org/x/sys: v0.42.0 → v0.48.0
- golang.org/x/term: v0.41.0 → v0.46.0
- golang.org/x/text: v0.36.0 → v0.42.0
- golang.org/x/time: v0.15.0 → v0.16.0
- google.golang.org/genproto/googleapis/api: v0.0.0-20260414002931-afd174a4e478 → v0.0.0-20260831171406-18b4a7587f8a
- google.golang.org/genproto/googleapis/rpc: v0.0.0-20260414002931-afd174a4e478 → v0.0.0-20260831171406-18b4a7587f8a
- google.golang.org/grpc: [v1.81.1 → v1.83.2](https://github.com/grpc/grpc-go/compare/v1.81.1...v1.83.2)
- google.golang.org/protobuf: v1.36.12-0.20260120151049-f2248ac996af → v1.36.12
- k8s.io/api: v0.36.1 → v0.37.0
- k8s.io/apimachinery: v0.36.1 → v0.37.0
- k8s.io/apiserver: v0.36.1 → v0.37.0
- k8s.io/client-go: v0.36.1 → v0.37.0
- k8s.io/component-base: v0.36.1 → v0.37.0
- k8s.io/kube-openapi: v0.0.0-20260317180543-43fb72c5454a → v0.0.0-20260821135717-be32def86098
- k8s.io/streaming: v0.36.1 → v0.37.0
- k8s.io/utils: v0.0.0-20260210185600-b8788abfbbc2 → v0.0.0-20260707023825-cf1189d6abe3
- sigs.k8s.io/apiserver-network-proxy/konnectivity-client: v0.35.0 → v0.36.0
- sigs.k8s.io/structured-merge-diff/v6: [v6.4.0 → v6.4.2](https://github.com/kubernetes-sigs/structured-merge-diff/compare/v6.4.0...v6.4.2)

### Removed
- github.com/go-openapi/swag/jsonname: [v0.26.0](https://github.com/go-openapi/swag/tree/v0.26.0)
- github.com/gogo/protobuf: [v1.3.2](https://github.com/gogo/protobuf/tree/v1.3.2)
- github.com/golang/mock: [v1.6.0](https://github.com/golang/mock/tree/v1.6.0)
- github.com/kubernetes-csi/csi-test/v5: [v5.4.0](https://github.com/kubernetes-csi/csi-test/tree/v5.4.0)
- gopkg.in/yaml.v3: v3.0.1
