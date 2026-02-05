module github.com/splunk/stef/examples/profile

go 1.24.0

require (
	github.com/google/go-cmp v0.7.0
	github.com/google/pprof v0.0.0-20250630185457-6e76a2b096b5
	github.com/klauspost/compress v1.18.3
	github.com/splunk/stef/go/pkg v0.1.0
	github.com/stretchr/testify v1.11.1
	modernc.org/b/v2 v2.1.10
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/splunk/stef/go/pkg => ../../go/pkg
