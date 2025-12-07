module github.com/splunk/stef/go/otel

go 1.24.0

require (
	github.com/splunk/stef/go/grpc v0.0.8
	github.com/splunk/stef/go/pkg v0.0.8
	github.com/stretchr/testify v1.11.1
	google.golang.org/grpc v1.75.1
)

require (
	github.com/klauspost/compress v1.18.2 // indirect
	golang.org/x/net v0.41.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.26.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250707201910-8d1bb00bc6a7 // indirect
	google.golang.org/protobuf v1.36.10 // indirect
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	modernc.org/b/v2 v2.1.10
)

replace (
	github.com/splunk/stef/go/grpc => ../grpc
	github.com/splunk/stef/go/pkg => ../pkg
)
