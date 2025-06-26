module github.com/splunk/stef/stefgen

go 1.22.7

require (
	github.com/splunk/stef/go/pkg v0.0.6
	github.com/stretchr/testify v1.9.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/klauspost/compress v1.17.8 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/splunk/stef/go/pkg => ../go/pkg
