module github.com/splunk/stef/go/otel

go 1.24.0

require (
	github.com/splunk/stef/go/grpc v0.0.9
	github.com/splunk/stef/go/pkg v0.0.9
	github.com/stretchr/testify v1.11.1
	google.golang.org/grpc v1.77.0
)

require (
	github.com/klauspost/compress v1.18.2 // indirect
	golang.org/x/net v0.46.1-0.20251013234738-63d1a5100f82 // indirect
	golang.org/x/sys v0.37.0 // indirect
	golang.org/x/text v0.30.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251022142026-3a174f9686a8 // indirect
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
