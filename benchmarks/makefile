GO=$(shell which go)

BENCH_COUNT ?= 6
CI_BENCHMARKS=STEF\|/STEF

MASTER_BRANCH ?= main
REF_NAME ?= $(shell git symbolic-ref -q --short HEAD || git describe --tags --exact-match)
SHELL := /bin/bash
VERSION=1.1
BENCHSTAT=$(shell which benchstat)

.PHONY: default
default: build test

.PHONY: all
all: build test benchmark

.PHONY: build
build:
	go build -o bin/stefstats cmd/stats/main.go
	go build -o bin/stefbench cmd/stefbench/main.go
	go build -o bin/otlp2tef cmd/otlp2tef/main.go
	go build -o bin/diff cmd/diff/main.go

.PHONY: test
test:
	go test -v ./...

benchmark:
	go test ./... -run notest -bench . -benchtime 100ms

.PHONY: benchmark-ci
benchmark-ci: bench-run bench-stat-diff bench-stat

.PHONY: bench-stat-cli
bench-stat-cli:
	@test -s $(GOPATH)/bin/benchstat || sudo GOFLAGS= GOBIN=$(GOPATH)/bin $(GO) install golang.org/x/perf/cmd/benchstat@latest

.PHONY: bench-run
bench-run:
	@set -o pipefail && $(GO) test -test.benchmem -bench=$(CI_BENCHMARKS) -count=$(BENCH_COUNT) -run=^a  ./... | tee bench-$(REF_NAME).txt

.PHONY: bench-stat-diff
bench-stat-diff: bench-stat-cli
	@test ! -e bench-$(MASTER_BRANCH).txt || benchstat bench-$(MASTER_BRANCH).txt bench-$(REF_NAME).txt

.PHONY: bench-stat
bench-stat: bench-stat-cli
	benchstat bench-$(REF_NAME).txt

.PHONY: update-charts
update-charts:
	UPDATE_BENCH_HTML=1 go test -run TestMetricsSize\|TestMetricsMultipart -bench alizeNative
