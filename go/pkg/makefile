.PHONY: default
default: test

.PHONY: all
all: test

.PHONY: test
test:
	go test -v ./... -bench . -benchtime 1ms

# make generate re-generates the lookup tables.
.PHONY: generate
generate:
	cd tablegen && go install .
	go generate
