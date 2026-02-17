module github.com/splunk/stef/go/otel

go 1.24.0

require (
	github.com/splunk/stef/go/grpc v0.1.1
	github.com/splunk/stef/go/pkg v0.1.1
	github.com/stretchr/testify v1.11.1
	google.golang.org/grpc v1.79.1
)

require (
	github.com/klauspost/compress v1.18.4 // indirect
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/text v0.32.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251202230838-ff82c1b0f217 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
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
