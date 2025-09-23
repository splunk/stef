.PHONY: default
default: build

.PHONY: all
all: generate default

.PHONY: build
build:
	go test ./... -bench . -benchtime 1ms

generate:
	cd ../../stefc && make build
	../../stefc/bin/stefc --lang=go otel.stef
	../../stefc/bin/stefc --lang=java --outdir=../../java/src/main/java --testoutdir=../../java/src/test/java otel.stef
