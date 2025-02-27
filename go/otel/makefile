.PHONY: default
default: build

.PHONY: all
all: generate default

.PHONY: build
build:
	go test ./... -bench . -benchtime 1ms

generate:
	cd ../../stefgen && make build
	../../stefgen/bin/stefgen otel.stef
