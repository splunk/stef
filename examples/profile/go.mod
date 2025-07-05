module github.com/splunk/stef/examples/profile

go 1.23.2

require (
	github.com/splunk/stef/go/pkg v0.0.6
	github.com/stretchr/testify v1.10.0
	modernc.org/b/v2 v2.1.4
)

require (
	github.com/chzyer/readline v1.5.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/google/pprof v0.0.0-20250630185457-6e76a2b096b5 // indirect
	github.com/ianlancetaylor/demangle v0.0.0-20250417193237-f615e6bd150b // indirect
	github.com/klauspost/compress v1.17.8 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/sys v0.32.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/splunk/stef/go/pkg => ../../go/pkg