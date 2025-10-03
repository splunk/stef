module github.com/splunk/stef/go/otel

go 1.23.2

require (
	github.com/splunk/stef/go/grpc v0.0.8
	github.com/splunk/stef/go/pkg v0.0.8
	github.com/stretchr/testify v1.11.1
	google.golang.org/grpc v1.75.1
)

require (
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	golang.org/x/net v0.41.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.26.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250707201910-8d1bb00bc6a7 // indirect
	google.golang.org/protobuf v1.36.9 // indirect
	modernc.org/mathutil v1.5.0 // indirect
	modernc.org/strutil v1.1.3 // indirect
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	modernc.org/b/v2 v2.1.0
)

replace (
	github.com/splunk/stef/go/grpc => ../grpc
	github.com/splunk/stef/go/pkg => ../pkg
)
