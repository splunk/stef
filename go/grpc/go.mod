module github.com/splunk/stef/go/grpc

go 1.22.7

toolchain go1.23.2

require (
	github.com/splunk/stef/go/pkg v0.0.6
	github.com/stretchr/testify v1.10.0
	google.golang.org/grpc v1.68.0
	google.golang.org/protobuf v1.35.2
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/klauspost/compress v1.17.8 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/net v0.29.0 // indirect
	golang.org/x/sys v0.25.0 // indirect
	golang.org/x/text v0.18.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240903143218-8af14fe29dc1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/splunk/stef/go/pkg => ../pkg
