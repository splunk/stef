module github.com/splunk/stef/stef-otel

go 1.23.2

require (
	github.com/splunk/stef/stef-go v0.0.1
	github.com/splunk/stef/stef-gogrpc v0.0.1
	github.com/stretchr/testify v1.10.0
	google.golang.org/grpc v1.68.0
)

require (
	github.com/klauspost/compress v1.17.8 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	golang.org/x/net v0.29.0 // indirect
	golang.org/x/sys v0.25.0 // indirect
	golang.org/x/text v0.18.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240903143218-8af14fe29dc1 // indirect
	google.golang.org/protobuf v1.35.2 // indirect
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
	github.com/splunk/stef/stef-go => ../stef-go
	github.com/splunk/stef/stef-gogrpc => ../stef-gogrpc
)
