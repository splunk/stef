.PHONY: default
default: build

.PHONY: all
all: generate default

.PHONY: build
build:
	go test ./... -bench . -benchtime 1ms

generate:
	cd ../../stefgen && make build
	../../stefgen/bin/stefgen --lang=go otel.stef
	../../stefgen/bin/stefgen --lang=java --outdir=../../java/src/main/java --testoutdir=../../java/src/test/java otel.stef
