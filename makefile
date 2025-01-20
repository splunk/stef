.PHONY: default
default:
	cd stefgen && make
	cd go/pkg && make
	cd go/grpc && make
	cd go/otel && make
	cd go/pdata && make
	cd otelcol && make
	cd benchmarks && make

all:
	cd stefgen && make all
	cd go/pkg && make all
	cd go/grpc && make all
	cd go/otel && make all
	cd go/pdata && make all
	cd otelcol && make all
	cd benchmarks && make all

prepver:
ifndef VERSION
		@echo "VERSION is unset or set to the empty string"
		@exit 1
endif
	echo Updating to version ${VERSION}
	cd go/grpc     && go mod edit -require=github.com/splunk/stef/go/pkg@${VERSION} && go mod tidy
	cd go/otel     && go mod edit -require=github.com/splunk/stef/go/pkg@${VERSION} \
				   && go mod edit -require=github.com/splunk/stef/go/grpc@${VERSION} && go mod tidy
	cd go/pdata    && go mod edit -require=github.com/splunk/stef/go/pkg@${VERSION} \
				   && go mod edit -require=github.com/splunk/stef/go/otel@${VERSION} && go mod tidy
	cd otelcol     && go mod tidy
	cd benchmarks  && go mod tidy
