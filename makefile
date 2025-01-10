.PHONY: default
default:
	cd stefgen && make
	cd stef-go && make
	cd stef-gogrpc && make
	cd stef-otel && make
	cd stef-pdata && make
	cd otelcol && make
	cd benchmarks && make

all:
	cd stefgen && make all
	cd stef-go && make all
	cd stef-gogrpc && make all
	cd stef-otel && make all
	cd stef-pdata && make all
	cd otelcol && make all
	cd benchmarks && make all
