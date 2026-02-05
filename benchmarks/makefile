GO=$(shell which go)

.PHONY: default
default: build test

.PHONY: all
all: build test benchmark

.PHONY: build
build:
	go build -o bin/stefstats cmd/stats/main.go
	go build -o bin/stefbench cmd/stefbench/main.go
	go build -o bin/otlp2stef cmd/otlp2stef/main.go

.PHONY: testcompat
testcompat: # Test compatibility of current implementation with old format files
	# Generate old test files
	cd scripts && ./genoldtestfiles.sh
	# Run compatibility tests
	go test -v -run TestCopy
	# Restore/generate current test files
	cd scripts && ./gentestfiles.sh

.PHONY: gentestfiles
gentestfiles:
	cd scripts && ./gentestfiles.sh

.PHONY: test
test: gentestfiles
	go test -v ./...

benchmark: gentestfiles
	go test ./... -run notest -bench . -benchtime 100ms

.PHONY: build-ci
build-ci:
	make testcompat
	make all
