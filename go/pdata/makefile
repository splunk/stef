.PHONY: default
default: test

.PHONY: all
all: default

.PHONY: test
test:
	go test -v ./...

.PHONY: fuzz
fuzz:
	go test -fuzz FuzzReader