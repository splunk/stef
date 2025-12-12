.PHONY: default
default: test

.PHONY: all
all: all-test build

.PHONY: test
all-test:
	STEF_ENABLE_SLOW_TESTS=1 make test

.PHONY: test
test:
	go test -v ./... -bench . -benchtime 1ms

build:
	go build -o bin/stefc main.go
	