.PHONY: default
default: build

.PHONY: all
all: generate default

.PHONY: build
build:
	go test -v ./...

.PHONY: generate
generate:
	cd ../../stefc && make build
	../../stefc/bin/stefc --lang=go --outdir=internal ints.stef

.PHONY: benchmark
benchmark:
	go test -run noname -bench . -count 4 | tee results.txt && benchstat -row /file -col /format results.txt
