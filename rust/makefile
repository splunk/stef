.PHONY: default build test bench

default: build test

build:
	cargo build --workspace

test:
	cargo test --workspace

benchmark:
	RUSTC_BOOTSTRAP=1 RUSTFLAGS="--cfg bench_nightly" cargo bench --workspace
