.PHONY: default
default: test

.PHONY: all
all: test build

.PHONY: test
test:
	go test -v ./... -bench . -benchtime 1ms

build:
	go build -o bin/stefgen main.go
	