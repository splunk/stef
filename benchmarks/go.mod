module github.com/splunk/stef/benchmarks

go 1.24.0

require (
	github.com/go-echarts/go-echarts/v2 v2.6.2
	github.com/klauspost/compress v1.18.0
	github.com/open-telemetry/otel-arrow v0.31.0
	github.com/parquet-go/parquet-go v0.25.1
	github.com/splunk/stef/go/otel v0.0.8
	github.com/splunk/stef/go/pdata v0.0.0
	github.com/splunk/stef/go/pkg v0.0.8
	github.com/stretchr/testify v1.11.1
	go.opentelemetry.io/collector/pdata v1.19.0
	golang.org/x/text v0.29.0
	google.golang.org/protobuf v1.36.9
	modernc.org/b/v2 v2.1.0
)

require (
	github.com/HdrHistogram/hdrhistogram-go v1.1.2 // indirect
	github.com/andybalholm/brotli v1.1.0 // indirect
	github.com/apache/arrow/go/v17 v17.0.0 // indirect
	github.com/axiomhq/hyperloglog v0.0.0-20230201085229-3ddf4bad03dc // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-metro v0.0.0-20180109044635-280f6062b5bc // indirect
	github.com/fxamacker/cbor/v2 v2.4.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/goccy/go-json v0.10.3 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/flatbuffers v24.3.25+incompatible // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/cpuid/v2 v2.2.8 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pierrec/lz4/v4 v4.1.21 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/zeebo/xxh3 v1.0.2 // indirect
	go.opentelemetry.io/collector/config/configtelemetry v0.114.0 // indirect
	go.opentelemetry.io/otel v1.31.0 // indirect
	go.opentelemetry.io/otel/metric v1.31.0 // indirect
	go.opentelemetry.io/otel/trace v1.31.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/exp v0.0.0-20240506185415-9bf2ced13842 // indirect
	golang.org/x/mod v0.27.0 // indirect
	golang.org/x/net v0.43.0 // indirect
	golang.org/x/sync v0.17.0 // indirect
	golang.org/x/sys v0.35.0 // indirect
	golang.org/x/tools v0.36.0 // indirect
	golang.org/x/xerrors v0.0.0-20231012003039-104605ab7028 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240903143218-8af14fe29dc1 // indirect
	google.golang.org/grpc v1.68.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/splunk/stef/go/pkg => ../go/pkg

replace github.com/splunk/stef/go/grpc => ../go/grpc

replace github.com/splunk/stef/go/pdata => ../go/pdata

replace github.com/splunk/stef/go/otel => ../go/otel
