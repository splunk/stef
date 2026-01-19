.PHONY: default
default: build

.PHONY: all
all: default

.PHONY: build
build:
	go test -v ./...

.PHONY: generate
generate:
	cd ../../stefc && make build
	../../stefc/bin/stefc --lang=go --outdir=internal jsonl.stef
	mkdir -p internal/jsonpb
	protoc --go_out=internal/jsonpb --go_opt=paths=source_relative jsonl.proto

.PHONY: benchmark
benchmark:
	go test -run noname -bench . -count 4 | tee results.txt && benchstat -row /file -col /format results.txt
