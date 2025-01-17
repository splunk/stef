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
