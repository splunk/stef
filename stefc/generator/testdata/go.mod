module github.com/splunk/stef/stefc/generator/testdata

go 1.23.2

require (
	github.com/splunk/stef/go/pkg v0.0.8
	github.com/stretchr/testify v1.11.1
	modernc.org/b/v2 v2.1.10
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/splunk/stef/go/pkg => ../../../go/pkg
