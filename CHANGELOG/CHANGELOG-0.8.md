# Release notes for v0.8.0

[Documentation](https://kubernetes-csi.github.io/)

# Changelog since 0.7.0

## Changes by Kind

### Uncategorized

## Dependencies

### Added
- github.com/cenkalti/backoff/v4: [v4.1.3](https://github.com/cenkalti/backoff/v4/tree/v4.1.3)
- github.com/go-logr/stdr: [v1.2.2](https://github.com/go-logr/stdr/tree/v1.2.2)
- github.com/grpc-ecosystem/grpc-gateway/v2: [v2.7.0](https://github.com/grpc-ecosystem/grpc-gateway/v2/tree/v2.7.0)
- go.opentelemetry.io/otel/exporters/otlp/internal/retry: v1.10.0
- go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc: v1.10.0
- go.opentelemetry.io/otel/exporters/otlp/otlptrace: v1.10.0

### Changed
- github.com/felixge/httpsnoop: [v1.0.1 → v1.0.3](https://github.com/felixge/httpsnoop/compare/v1.0.1...v1.0.3)
- github.com/inconshreveable/mousetrap: [v1.0.0 → v1.0.1](https://github.com/inconshreveable/mousetrap/compare/v1.0.0...v1.0.1)
- github.com/kubernetes-csi/csi-lib-utils: [v0.11.0 → v0.12.0](https://github.com/kubernetes-csi/csi-lib-utils/compare/v0.11.0...v0.12.0)
- github.com/moby/term: [3f7ff69 → 39b0c02](https://github.com/moby/term/compare/3f7ff69...39b0c02)
- github.com/onsi/ginkgo/v2: [v2.1.6 → v2.4.0](https://github.com/onsi/ginkgo/v2/compare/v2.1.6...v2.4.0)
- github.com/onsi/gomega: [v1.20.1 → v1.23.0](https://github.com/onsi/gomega/compare/v1.20.1...v1.23.0)
- github.com/prometheus/client_golang: [v1.13.1 → v1.14.0](https://github.com/prometheus/client_golang/compare/v1.13.1...v1.14.0)
- github.com/spf13/cobra: [v1.4.0 → v1.6.0](https://github.com/spf13/cobra/compare/v1.4.0...v1.6.0)
- go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp: v0.20.0 → v0.35.0
- go.opentelemetry.io/otel/metric: v0.20.0 → v0.31.0
- go.opentelemetry.io/otel/sdk: v0.20.0 → v1.10.0
- go.opentelemetry.io/otel/trace: v0.20.0 → v1.10.0
- go.opentelemetry.io/otel: v0.20.0 → v1.10.0
- go.opentelemetry.io/proto/otlp: v0.7.0 → v0.19.0
- golang.org/x/crypto: 3147a52 → 089bfa5
- golang.org/x/lint: 6edffad → 738671d
- golang.org/x/net: v0.1.0 → v0.4.0
- golang.org/x/sys: v0.1.0 → v0.3.0
- golang.org/x/term: v0.1.0 → v0.3.0
- golang.org/x/text: v0.4.0 → v0.5.0
- google.golang.org/grpc: v1.50.1 → v1.51.0
- k8s.io/api: v0.25.3 → v0.26.0
- k8s.io/apimachinery: v0.25.3 → v0.26.0
- k8s.io/client-go: v0.25.3 → v0.26.0
- k8s.io/component-base: v0.25.3 → v0.26.0
- k8s.io/utils: 61b03e2 → 1a15be2

### Removed
- cloud.google.com/go/firestore: v1.1.0
- github.com/Azure/go-autorest/autorest/adal: [v0.9.20](https://github.com/Azure/go-autorest/autorest/adal/tree/v0.9.20)
- github.com/Azure/go-autorest/autorest/date: [v0.3.0](https://github.com/Azure/go-autorest/autorest/date/tree/v0.3.0)
- github.com/Azure/go-autorest/autorest/mocks: [v0.4.1](https://github.com/Azure/go-autorest/autorest/mocks/tree/v0.4.1)
- github.com/Azure/go-autorest/autorest: [v0.11.27](https://github.com/Azure/go-autorest/autorest/tree/v0.11.27)
- github.com/Azure/go-autorest/logger: [v0.2.1](https://github.com/Azure/go-autorest/logger/tree/v0.2.1)
- github.com/Azure/go-autorest/tracing: [v0.6.0](https://github.com/Azure/go-autorest/tracing/tree/v0.6.0)
- github.com/Azure/go-autorest: [v14.2.0+incompatible](https://github.com/Azure/go-autorest/tree/v14.2.0)
- github.com/armon/circbuf: [bbbad09](https://github.com/armon/circbuf/tree/bbbad09)
- github.com/armon/go-metrics: [f0300d1](https://github.com/armon/go-metrics/tree/f0300d1)
- github.com/armon/go-radix: [7fddfc3](https://github.com/armon/go-radix/tree/7fddfc3)
- github.com/benbjohnson/clock: [v1.0.3](https://github.com/benbjohnson/clock/tree/v1.0.3)
- github.com/bgentry/speakeasy: [v0.1.0](https://github.com/bgentry/speakeasy/tree/v0.1.0)
- github.com/bketelsen/crypt: [5cbc8cc](https://github.com/bketelsen/crypt/tree/5cbc8cc)
- github.com/blang/semver: [v3.5.1+incompatible](https://github.com/blang/semver/tree/v3.5.1)
- github.com/coreos/bbolt: [v1.3.2](https://github.com/coreos/bbolt/tree/v1.3.2)
- github.com/coreos/etcd: [v3.3.13+incompatible](https://github.com/coreos/etcd/tree/v3.3.13)
- github.com/coreos/go-semver: [v0.3.0](https://github.com/coreos/go-semver/tree/v0.3.0)
- github.com/coreos/go-systemd: [95778df](https://github.com/coreos/go-systemd/tree/95778df)
- github.com/coreos/pkg: [399ea9e](https://github.com/coreos/pkg/tree/399ea9e)
- github.com/cpuguy83/go-md2man/v2: [v2.0.0](https://github.com/cpuguy83/go-md2man/v2/tree/v2.0.0)
- github.com/creack/pty: [v1.1.11](https://github.com/creack/pty/tree/v1.1.11)
- github.com/dgrijalva/jwt-go: [v3.2.0+incompatible](https://github.com/dgrijalva/jwt-go/tree/v3.2.0)
- github.com/dgryski/go-sip13: [e10d5fe](https://github.com/dgryski/go-sip13/tree/e10d5fe)
- github.com/emicklei/go-restful: [ff4f55a](https://github.com/emicklei/go-restful/tree/ff4f55a)
- github.com/fatih/color: [v1.7.0](https://github.com/fatih/color/tree/v1.7.0)
- github.com/form3tech-oss/jwt-go: [v3.2.3+incompatible](https://github.com/form3tech-oss/jwt-go/tree/v3.2.3)
- github.com/golang-jwt/jwt/v4: [v4.2.0](https://github.com/golang-jwt/jwt/v4/tree/v4.2.0)
- github.com/googleapis/gnostic: [v0.5.5](https://github.com/googleapis/gnostic/tree/v0.5.5)
- github.com/gopherjs/gopherjs: [0766667](https://github.com/gopherjs/gopherjs/tree/0766667)
- github.com/gorilla/websocket: [v1.4.2](https://github.com/gorilla/websocket/tree/v1.4.2)
- github.com/grpc-ecosystem/go-grpc-middleware: [v1.0.0](https://github.com/grpc-ecosystem/go-grpc-middleware/tree/v1.0.0)
- github.com/grpc-ecosystem/go-grpc-prometheus: [v1.2.0](https://github.com/grpc-ecosystem/go-grpc-prometheus/tree/v1.2.0)
- github.com/hashicorp/consul/api: [v1.1.0](https://github.com/hashicorp/consul/api/tree/v1.1.0)
- github.com/hashicorp/consul/sdk: [v0.1.1](https://github.com/hashicorp/consul/sdk/tree/v0.1.1)
- github.com/hashicorp/errwrap: [v1.0.0](https://github.com/hashicorp/errwrap/tree/v1.0.0)
- github.com/hashicorp/go-cleanhttp: [v0.5.1](https://github.com/hashicorp/go-cleanhttp/tree/v0.5.1)
- github.com/hashicorp/go-immutable-radix: [v1.0.0](https://github.com/hashicorp/go-immutable-radix/tree/v1.0.0)
- github.com/hashicorp/go-msgpack: [v0.5.3](https://github.com/hashicorp/go-msgpack/tree/v0.5.3)
- github.com/hashicorp/go-multierror: [v1.0.0](https://github.com/hashicorp/go-multierror/tree/v1.0.0)
- github.com/hashicorp/go-rootcerts: [v1.0.0](https://github.com/hashicorp/go-rootcerts/tree/v1.0.0)
- github.com/hashicorp/go-sockaddr: [v1.0.0](https://github.com/hashicorp/go-sockaddr/tree/v1.0.0)
- github.com/hashicorp/go-syslog: [v1.0.0](https://github.com/hashicorp/go-syslog/tree/v1.0.0)
- github.com/hashicorp/go-uuid: [v1.0.1](https://github.com/hashicorp/go-uuid/tree/v1.0.1)
- github.com/hashicorp/go.net: [v0.0.1](https://github.com/hashicorp/go.net/tree/v0.0.1)
- github.com/hashicorp/hcl: [v1.0.0](https://github.com/hashicorp/hcl/tree/v1.0.0)
- github.com/hashicorp/logutils: [v1.0.0](https://github.com/hashicorp/logutils/tree/v1.0.0)
- github.com/hashicorp/mdns: [v1.0.0](https://github.com/hashicorp/mdns/tree/v1.0.0)
- github.com/hashicorp/memberlist: [v0.1.3](https://github.com/hashicorp/memberlist/tree/v0.1.3)
- github.com/hashicorp/serf: [v0.8.2](https://github.com/hashicorp/serf/tree/v0.8.2)
- github.com/jonboulle/clockwork: [v0.1.0](https://github.com/jonboulle/clockwork/tree/v0.1.0)
- github.com/jtolds/gls: [v4.20.0+incompatible](https://github.com/jtolds/gls/tree/v4.20.0)
- github.com/magiconair/properties: [v1.8.1](https://github.com/magiconair/properties/tree/v1.8.1)
- github.com/mattn/go-colorable: [v0.0.9](https://github.com/mattn/go-colorable/tree/v0.0.9)
- github.com/mattn/go-isatty: [v0.0.3](https://github.com/mattn/go-isatty/tree/v0.0.3)
- github.com/miekg/dns: [v1.0.14](https://github.com/miekg/dns/tree/v1.0.14)
- github.com/mitchellh/cli: [v1.0.0](https://github.com/mitchellh/cli/tree/v1.0.0)
- github.com/mitchellh/go-homedir: [v1.1.0](https://github.com/mitchellh/go-homedir/tree/v1.1.0)
- github.com/mitchellh/go-testing-interface: [v1.0.0](https://github.com/mitchellh/go-testing-interface/tree/v1.0.0)
- github.com/mitchellh/gox: [v0.4.0](https://github.com/mitchellh/gox/tree/v0.4.0)
- github.com/mitchellh/iochan: [v1.0.0](https://github.com/mitchellh/iochan/tree/v1.0.0)
- github.com/oklog/ulid: [v1.3.1](https://github.com/oklog/ulid/tree/v1.3.1)
- github.com/pascaldekloe/goe: [57f6aae](https://github.com/pascaldekloe/goe/tree/57f6aae)
- github.com/pelletier/go-toml: [v1.2.0](https://github.com/pelletier/go-toml/tree/v1.2.0)
- github.com/posener/complete: [v1.1.1](https://github.com/posener/complete/tree/v1.1.1)
- github.com/prometheus/tsdb: [v0.7.1](https://github.com/prometheus/tsdb/tree/v0.7.1)
- github.com/russross/blackfriday/v2: [v2.0.1](https://github.com/russross/blackfriday/v2/tree/v2.0.1)
- github.com/ryanuber/columnize: [9b3edd6](https://github.com/ryanuber/columnize/tree/9b3edd6)
- github.com/sean-/seed: [e2103e2](https://github.com/sean-/seed/tree/e2103e2)
- github.com/shurcooL/sanitized_anchor_name: [v1.0.0](https://github.com/shurcooL/sanitized_anchor_name/tree/v1.0.0)
- github.com/smartystreets/assertions: [b2de0cb](https://github.com/smartystreets/assertions/tree/b2de0cb)
- github.com/smartystreets/goconvey: [v1.6.4](https://github.com/smartystreets/goconvey/tree/v1.6.4)
- github.com/soheilhy/cmux: [v0.1.4](https://github.com/soheilhy/cmux/tree/v0.1.4)
- github.com/spf13/afero: [v1.2.2](https://github.com/spf13/afero/tree/v1.2.2)
- github.com/spf13/cast: [v1.3.0](https://github.com/spf13/cast/tree/v1.3.0)
- github.com/spf13/jwalterweatherman: [v1.0.0](https://github.com/spf13/jwalterweatherman/tree/v1.0.0)
- github.com/spf13/viper: [v1.7.0](https://github.com/spf13/viper/tree/v1.7.0)
- github.com/subosito/gotenv: [v1.2.0](https://github.com/subosito/gotenv/tree/v1.2.0)
- github.com/tmc/grpc-websocket-proxy: [0ad062e](https://github.com/tmc/grpc-websocket-proxy/tree/0ad062e)
- github.com/xiang90/probing: [43a291a](https://github.com/xiang90/probing/tree/43a291a)
- go.etcd.io/bbolt: v1.3.2
- go.opentelemetry.io/contrib: v0.20.0
- go.opentelemetry.io/otel/exporters/otlp: v0.20.0
- go.opentelemetry.io/otel/oteltest: v0.20.0
- go.opentelemetry.io/otel/sdk/export/metric: v0.20.0
- go.opentelemetry.io/otel/sdk/metric: v0.20.0
- gopkg.in/ini.v1: v1.51.0
- gopkg.in/resty.v1: v1.12.0
