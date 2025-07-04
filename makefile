# Function to execute a command.
# Accepts command to execute as first parameter.
define exec-command
$(1)

endef

.PHONY: default
default:
	cd stefgen && make
	cd go/pkg && make
	cd go/grpc && make
	cd go/otel && make
	cd go/pdata && make
	cd otelcol && make
	cd examples && make
	cd benchmarks && make

.PHONY: all
all:
	cd stefgen && make all
	cd go/pkg && make all
	cd go/grpc && make all
	cd go/otel && make all
	cd go/pdata && make all
	cd otelcol && make all
	cd examples && make all
	cd benchmarks && make all

.PHONY: build-ci
build-ci:
	cd stefgen && make all
	cd go/pkg && make all
	cd go/grpc && make
	cd go/otel && make all
	cd go/pdata && make all
	cd otelcol && make all
	cd examples && make
	cd benchmarks && make all

.PHONY: verifyver
verifyver:
ifndef VERSION
		@echo "VERSION is unset or set to the empty string"
		@exit 1
endif

.PHONY: prepver
prepver: verifyver
	echo Updating to version ${VERSION}
	cd go/grpc     && go mod edit -require=github.com/splunk/stef/go/pkg@${VERSION} && go mod tidy
	cd go/otel     && go mod edit -require=github.com/splunk/stef/go/pkg@${VERSION} \
				   && go mod edit -require=github.com/splunk/stef/go/grpc@${VERSION} && go mod tidy
	cd go/pdata    && go mod edit -require=github.com/splunk/stef/go/pkg@${VERSION} \
				   && go mod edit -require=github.com/splunk/stef/go/otel@${VERSION} && go mod tidy
	cd stefgen     && go mod edit -require=github.com/splunk/stef/go/pkg@${VERSION} && go mod tidy
	cd examples/jsonl && go mod edit -require=github.com/splunk/stef/go/pkg@${VERSION} && go mod tidy
	cd otelcol     && go mod tidy
	cd benchmarks  && go mod tidy

MODULES := go/pkg go/grpc go/otel go/pdata

.PHONY: releasever
releasever: verifyver
	echo Tagging version $(VERSION)
	$(foreach gomod,$(MODULES),$(call exec-command,git tag $(gomod)/$(VERSION) && git push origin $(gomod)/$(VERSION)))
