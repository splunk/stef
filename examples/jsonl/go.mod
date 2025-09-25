module github.com/splunk/stef/examples/jsonl

go 1.23

require (
	github.com/golang/protobuf v1.5.0
	github.com/klauspost/compress v1.18.0
	github.com/splunk/stef/go/pkg v0.0.7
	github.com/stretchr/testify v1.9.0
	google.golang.org/protobuf v1.36.9
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/splunk/stef/go/pkg => ../../go/pkg
