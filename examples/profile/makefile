.PHONY: default
default: build

.PHONY: all
all: generate default

.PHONY: build
build:
	go test -v

.PHONY: generate
generate:
	cd ../../stefgen && make build
	../../stefgen/bin/stefgen --lang=go --outdir=internal profile.stef

.PHONY: benchmark
benchmark:
	go test -run noname -bench . -count 4 | tee results.txt && benchstat -row /file -col /format results.txt
