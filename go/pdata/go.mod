module github.com/splunk/stef/go/pdata

go 1.22.7

toolchain go1.23.2

require (
	github.com/google/go-cmp v0.6.0
	github.com/splunk/stef/go/otel v0.0.3
	github.com/splunk/stef/go/pkg v0.0.3
	github.com/stretchr/testify v1.10.0
	go.opentelemetry.io/collector/pdata v1.16.0
	modernc.org/b/v2 v2.1.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.17.8 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/net v0.29.0 // indirect
	golang.org/x/sys v0.25.0 // indirect
	golang.org/x/text v0.18.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240903143218-8af14fe29dc1 // indirect
	google.golang.org/grpc v1.68.0 // indirect
	google.golang.org/protobuf v1.35.2 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	github.com/splunk/stef/go/grpc => ../grpc
	github.com/splunk/stef/go/otel => ../otel
	github.com/splunk/stef/go/pkg => ../pkg
)
