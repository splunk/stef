module github.com/splunk/stef/otelcol

go 1.24.0

require (
	github.com/spf13/pflag v1.0.9 // indirect
	github.com/splunk/stef/go/grpc v0.0.8
	github.com/splunk/stef/go/pdata v0.0.0
	go.opentelemetry.io/collector v0.136.0 // indirect
	go.opentelemetry.io/collector/confmap v1.42.0
	go.opentelemetry.io/collector/pdata v1.42.0
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0
	golang.org/x/sys v0.36.0
)

require (
	github.com/gogo/protobuf v1.3.2
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/fileexporter v0.136.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/splunkhecexporter v0.136.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourcedetectionprocessor v0.136.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/filelogreceiver v0.136.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver v0.136.0
	github.com/open-telemetry/otel-arrow v0.24.0
	github.com/splunk/stef/go/otel v0.0.8
	github.com/splunk/stef/go/pkg v0.0.8
	github.com/stretchr/testify v1.11.1
	go.opentelemetry.io/collector/component v1.42.0
	go.opentelemetry.io/collector/component/componentstatus v0.136.0
	go.opentelemetry.io/collector/component/componenttest v0.136.0
	go.opentelemetry.io/collector/config/configgrpc v0.136.0
	go.opentelemetry.io/collector/config/confignet v1.42.0
	go.opentelemetry.io/collector/confmap/provider/envprovider v1.26.0
	go.opentelemetry.io/collector/confmap/provider/fileprovider v1.42.0
	go.opentelemetry.io/collector/connector v0.136.0
	go.opentelemetry.io/collector/consumer v1.42.0
	go.opentelemetry.io/collector/consumer/consumererror v0.136.0
	go.opentelemetry.io/collector/consumer/consumertest v0.136.0
	go.opentelemetry.io/collector/exporter v1.42.0
	go.opentelemetry.io/collector/exporter/debugexporter v0.136.0
	go.opentelemetry.io/collector/exporter/exporterhelper v0.136.0
	go.opentelemetry.io/collector/exporter/exportertest v0.136.0
	go.opentelemetry.io/collector/exporter/otlpexporter v0.136.0
	go.opentelemetry.io/collector/exporter/otlphttpexporter v0.136.0
	go.opentelemetry.io/collector/extension v1.42.0
	go.opentelemetry.io/collector/extension/zpagesextension v0.136.0
	go.opentelemetry.io/collector/otelcol v0.136.0
	go.opentelemetry.io/collector/processor v1.42.0
	go.opentelemetry.io/collector/receiver v1.42.0
	go.opentelemetry.io/collector/receiver/otlpreceiver v0.136.0
	go.opentelemetry.io/collector/receiver/receivertest v0.136.0
	google.golang.org/grpc v1.75.1
)

require (
	cloud.google.com/go/compute/metadata v0.7.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/detectors/gcp v1.29.0 // indirect
	github.com/HdrHistogram/hdrhistogram-go v1.1.2 // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/Microsoft/hcsshim v0.12.5 // indirect
	github.com/Showmax/go-fqdn v1.0.0 // indirect
	github.com/apache/arrow/go/v16 v16.1.0 // indirect
	github.com/armon/go-metrics v0.4.1 // indirect
	github.com/aws/aws-sdk-go-v2 v1.37.0 // indirect
	github.com/aws/aws-sdk-go-v2/config v1.30.1 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.18.1 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.18.0 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.4.0 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.7.0 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.237.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.13.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.26.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.31.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.35.0 // indirect
	github.com/aws/smithy-go v1.22.5 // indirect
	github.com/axiomhq/hyperloglog v0.0.0-20230201085229-3ddf4bad03dc // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bmatcuk/doublestar/v4 v4.9.1 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cenkalti/backoff/v5 v5.0.3 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/containerd/containerd v1.7.25 // indirect
	github.com/containerd/errdefs v1.0.0 // indirect
	github.com/containerd/errdefs/pkg v0.3.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dgryski/go-metro v0.0.0-20180109044635-280f6062b5bc // indirect
	github.com/distribution/reference v0.6.0 // indirect
	github.com/docker/docker v28.4.0+incompatible // indirect
	github.com/docker/go-connections v0.6.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/ebitengine/purego v0.8.4 // indirect
	github.com/elastic/lunes v0.1.0 // indirect
	github.com/emicklei/go-restful/v3 v3.11.0 // indirect
	github.com/expr-lang/expr v1.17.6 // indirect
	github.com/fatih/color v1.16.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/foxboron/go-tpm-keyfiles v0.0.0-20250903184740-5d135037bd4d // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/fxamacker/cbor/v2 v2.7.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/go-openapi/jsonpointer v0.21.0 // indirect
	github.com/go-openapi/jsonreference v0.20.2 // indirect
	github.com/go-openapi/swag v0.23.0 // indirect
	github.com/go-resty/resty/v2 v2.16.5 // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/golang/snappy v1.0.0 // indirect
	github.com/google/flatbuffers v24.3.25+incompatible // indirect
	github.com/google/gnostic-models v0.6.8 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/go-tpm v0.9.6 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.27.1 // indirect
	github.com/hashicorp/consul/api v1.32.1 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-hclog v1.5.0 // indirect
	github.com/hashicorp/go-immutable-radix v1.3.1 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-rootcerts v1.0.2 // indirect
	github.com/hashicorp/go-version v1.7.0 // indirect
	github.com/hashicorp/golang-lru v1.0.2 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/hashicorp/serf v0.10.1 // indirect
	github.com/hetznercloud/hcloud-go/v2 v2.23.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jonboulle/clockwork v0.5.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/klauspost/cpuid/v2 v2.2.7 // indirect
	github.com/knadh/koanf v1.5.0 // indirect
	github.com/knadh/koanf/v2 v2.3.0 // indirect
	github.com/leodido/go-syslog/v4 v4.2.0 // indirect
	github.com/leodido/ragel-machinery v0.0.0-20190525184631-5f46317e436b // indirect
	github.com/linode/go-metadata v0.2.2 // indirect
	github.com/lufia/plan9stats v0.0.0-20211012122336-39d0f177ccd0 // indirect
	github.com/magefile/mage v1.15.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/mapstructure v1.5.1-0.20231216201459-8508981c8b6c // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/moby/docker-image-spec v1.3.1 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.3-0.20250322232337-35a7c28c31ee // indirect
	github.com/mostynb/go-grpc-compression v1.2.3 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/aws/ecsutil v0.136.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/common v0.136.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal v0.136.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/filter v0.136.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/gopsutilenv v0.136.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/k8sconfig v0.136.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/metadataproviders v0.136.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/sharedcomponent v0.136.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/splunk v0.136.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/batchperresourceattr v0.136.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/experimentalmetricmetadata v0.136.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza v0.136.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/winperfcounters v0.136.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.1 // indirect
	github.com/openshift/api v3.9.0+incompatible // indirect
	github.com/openshift/client-go v0.0.0-20241203091221-452dfb8fa071 // indirect
	github.com/pierrec/lz4/v4 v4.1.22 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/power-devops/perfstat v0.0.0-20240221224432-82ca36839d55 // indirect
	github.com/prometheus/client_golang v1.23.2 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.66.1 // indirect
	github.com/prometheus/procfs v0.17.0 // indirect
	github.com/rs/cors v1.11.1 // indirect
	github.com/scaleway/scaleway-sdk-go v1.0.0-beta.35 // indirect
	github.com/shirou/gopsutil/v4 v4.25.8 // indirect
	github.com/spf13/cobra v1.10.1 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/tilinna/clock v1.1.0 // indirect
	github.com/tklauser/go-sysconf v0.3.15 // indirect
	github.com/tklauser/numcpus v0.10.0 // indirect
	github.com/valyala/fastjson v1.6.4 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	github.com/zeebo/xxh3 v1.0.2 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/collector/client v1.42.0 // indirect
	go.opentelemetry.io/collector/config/configauth v0.136.0 // indirect
	go.opentelemetry.io/collector/config/configcompression v1.42.0 // indirect
	go.opentelemetry.io/collector/config/confighttp v0.136.0 // indirect
	go.opentelemetry.io/collector/config/configmiddleware v1.42.0 // indirect
	go.opentelemetry.io/collector/config/configopaque v1.42.0 // indirect
	go.opentelemetry.io/collector/config/configoptional v0.136.0 // indirect
	go.opentelemetry.io/collector/config/configretry v1.42.0 // indirect
	go.opentelemetry.io/collector/config/configtelemetry v0.136.0 // indirect
	go.opentelemetry.io/collector/config/configtls v1.42.0 // indirect
	go.opentelemetry.io/collector/confmap/xconfmap v0.136.0 // indirect
	go.opentelemetry.io/collector/connector/connectortest v0.136.0 // indirect
	go.opentelemetry.io/collector/connector/xconnector v0.136.0 // indirect
	go.opentelemetry.io/collector/consumer/consumererror/xconsumererror v0.136.0 // indirect
	go.opentelemetry.io/collector/consumer/xconsumer v0.136.0 // indirect
	go.opentelemetry.io/collector/exporter/exporterhelper/xexporterhelper v0.136.0 // indirect
	go.opentelemetry.io/collector/exporter/xexporter v0.136.0 // indirect
	go.opentelemetry.io/collector/extension/extensionauth v1.42.0 // indirect
	go.opentelemetry.io/collector/extension/extensioncapabilities v0.136.0 // indirect
	go.opentelemetry.io/collector/extension/extensionmiddleware v0.136.0 // indirect
	go.opentelemetry.io/collector/extension/extensiontest v0.136.0 // indirect
	go.opentelemetry.io/collector/extension/xextension v0.136.0 // indirect
	go.opentelemetry.io/collector/featuregate v1.42.0 // indirect
	go.opentelemetry.io/collector/filter v0.136.0 // indirect
	go.opentelemetry.io/collector/internal/fanoutconsumer v0.136.0 // indirect
	go.opentelemetry.io/collector/internal/sharedcomponent v0.136.0 // indirect
	go.opentelemetry.io/collector/internal/telemetry v0.136.0 // indirect
	go.opentelemetry.io/collector/pdata/pprofile v0.136.0 // indirect
	go.opentelemetry.io/collector/pdata/testdata v0.136.0 // indirect
	go.opentelemetry.io/collector/pdata/xpdata v0.136.0 // indirect
	go.opentelemetry.io/collector/pipeline v1.42.0 // indirect
	go.opentelemetry.io/collector/pipeline/xpipeline v0.136.0 // indirect
	go.opentelemetry.io/collector/processor/processorhelper v0.136.0 // indirect
	go.opentelemetry.io/collector/processor/processorhelper/xprocessorhelper v0.136.0 // indirect
	go.opentelemetry.io/collector/processor/processortest v0.136.0 // indirect
	go.opentelemetry.io/collector/processor/xprocessor v0.136.0 // indirect
	go.opentelemetry.io/collector/receiver/receiverhelper v0.136.0 // indirect
	go.opentelemetry.io/collector/receiver/xreceiver v0.136.0 // indirect
	go.opentelemetry.io/collector/scraper v0.136.0 // indirect
	go.opentelemetry.io/collector/scraper/scraperhelper v0.136.0 // indirect
	go.opentelemetry.io/collector/service v0.136.0 // indirect
	go.opentelemetry.io/collector/service/hostcapabilities v0.136.0 // indirect
	go.opentelemetry.io/contrib/bridges/otelzap v0.12.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.62.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.62.0 // indirect
	go.opentelemetry.io/contrib/otelconf v0.16.0 // indirect
	go.opentelemetry.io/contrib/propagators/b3 v1.36.0 // indirect
	go.opentelemetry.io/contrib/zpages v0.62.0 // indirect
	go.opentelemetry.io/otel v1.38.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc v0.13.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp v0.13.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v1.37.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp v1.37.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.37.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.37.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.37.0 // indirect
	go.opentelemetry.io/otel/exporters/prometheus v0.58.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdoutlog v0.13.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdoutmetric v1.37.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.37.0 // indirect
	go.opentelemetry.io/otel/log v0.14.0 // indirect
	go.opentelemetry.io/otel/metric v1.38.0 // indirect
	go.opentelemetry.io/otel/sdk v1.38.0 // indirect
	go.opentelemetry.io/otel/sdk/log v0.13.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.38.0 // indirect
	go.opentelemetry.io/otel/trace v1.38.0 // indirect
	go.opentelemetry.io/proto/otlp v1.7.0 // indirect
	go.yaml.in/yaml/v2 v2.4.2 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/crypto v0.42.0 // indirect
	golang.org/x/exp v0.0.0-20250305212735-054e65f0b394 // indirect
	golang.org/x/mod v0.27.0 // indirect
	golang.org/x/net v0.44.0 // indirect
	golang.org/x/oauth2 v0.30.0 // indirect
	golang.org/x/sync v0.17.0 // indirect
	golang.org/x/term v0.35.0 // indirect
	golang.org/x/text v0.29.0 // indirect
	golang.org/x/time v0.7.0 // indirect
	golang.org/x/tools v0.36.0 // indirect
	golang.org/x/xerrors v0.0.0-20231012003039-104605ab7028 // indirect
	gonum.org/v1/gonum v0.16.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250707201910-8d1bb00bc6a7 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250707201910-8d1bb00bc6a7 // indirect
	google.golang.org/protobuf v1.36.9 // indirect
	gopkg.in/evanphx/json-patch.v4 v4.12.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/api v0.32.3 // indirect
	k8s.io/apimachinery v0.32.3 // indirect
	k8s.io/client-go v0.32.3 // indirect
	k8s.io/klog/v2 v2.130.1 // indirect
	k8s.io/kube-openapi v0.0.0-20241105132330-32ad38e42d3f // indirect
	k8s.io/utils v0.0.0-20241104100929-3ea5e8cea738 // indirect
	modernc.org/b/v2 v2.1.0 // indirect
	sigs.k8s.io/json v0.0.0-20241010143419-9aa6b5e7a4b3 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.4.2 // indirect
	sigs.k8s.io/yaml v1.5.0 // indirect
)

// https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/12322#issuecomment-1185029670
// https://github.com/docker/go-connections/issues/99
replace github.com/docker/go-connections => github.com/docker/go-connections v0.4.0

// security updates
replace (
	github.com/Masterminds/goutils => github.com/Masterminds/goutils v1.1.1
	github.com/apache/thrift => github.com/apache/thrift v0.16.0
	github.com/containernetworking/plugins => github.com/containernetworking/plugins v1.1.1
	github.com/gin-gonic/gin => github.com/gin-gonic/gin v1.7.7
	github.com/go-kit/kit => github.com/go-kit/kit v0.12.0 // required to drop dependency on deprecated go.etcd.io/etcd
	github.com/nats-io/jwt/v2 => github.com/nats-io/jwt/v2 v2.2.0
	github.com/nats-io/nats-server/v2 => github.com/nats-io/nats-server/v2 v2.8.1
	github.com/nats-io/nats.go => github.com/nats-io/nats.go v1.14.0
	github.com/opencontainers/runc => github.com/opencontainers/runc v1.1.2
	github.com/spf13/viper => github.com/spf13/viper v1.11.0 // required to drop dependency on deprecated github.com/coreos/etcd and github.com/coreos/go-etcd
	github.com/valyala/fasthttp => github.com/valyala/fasthttp v1.36.0
	golang.org/x/crypto => golang.org/x/crypto v0.5.0
	k8s.io/apiserver => k8s.io/apiserver v0.24.1 // required to drop dependency on deprecated go.etcd.io/etcd
)

// this is the version that doesn't suffer from https://github.com/mattn/go-ieproxy/issues/45
replace github.com/mattn/go-ieproxy => github.com/mattn/go-ieproxy v0.0.1

// vault has invalid requirements https://github.com/hashicorp/vault/pull/13321
replace (
	github.com/hashicorp/vault/api/auth/approle => github.com/hashicorp/vault/api/auth/approle v0.1.2-0.20211223174530-3688d63348b3
	github.com/hashicorp/vault/api/auth/userpass => github.com/hashicorp/vault/api/auth/userpass v0.1.1-0.20211223174530-3688d63348b3
)

// https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/8081
replace github.com/googleapis/gnostic v0.5.6 => github.com/googleapis/gnostic v0.5.5

// required to drop dependency on deprecated git.apache.org/thrift.git
exclude go.opencensus.io v0.19.1

// pin version until https://github.com/splunk/stefcol/pull/2418 is resolved
replace github.com/testcontainers/testcontainers-go => github.com/testcontainers/testcontainers-go v0.15.0

replace (
	github.com/splunk/stef/go/grpc => ../go/grpc
	github.com/splunk/stef/go/otel => ../go/otel
	github.com/splunk/stef/go/pdata => ../go/pdata
	github.com/splunk/stef/go/pkg => ../go/pkg
)
