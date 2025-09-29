module github.com/splunk/stef/benchmarks

go 1.24.0

require (
	github.com/go-echarts/go-echarts/v2 v2.6.3
	github.com/klauspost/compress v1.18.0
	github.com/parquet-go/parquet-go v0.25.1
	github.com/splunk/stef/go/otel v0.0.8
	github.com/splunk/stef/go/pdata v0.0.0
	github.com/splunk/stef/go/pkg v0.0.8
	github.com/stretchr/testify v1.11.1
	go.opentelemetry.io/collector/pdata v1.19.0
	golang.org/x/text v0.29.0
	google.golang.org/protobuf v1.36.9
	modernc.org/b/v2 v2.1.9
)

require (
	github.com/andybalholm/brotli v1.1.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pierrec/lz4/v4 v4.1.21 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/net v0.43.0 // indirect
	golang.org/x/sys v0.35.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240903143218-8af14fe29dc1 // indirect
	google.golang.org/grpc v1.68.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/splunk/stef/go/pkg => ../go/pkg

replace github.com/splunk/stef/go/grpc => ../go/grpc

replace github.com/splunk/stef/go/pdata => ../go/pdata

replace github.com/splunk/stef/go/otel => ../go/otel
