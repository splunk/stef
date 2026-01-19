# Dynamic, Schemaless JSON-like Data

This example demonstrates how JSONL data can be represented using STEF encoding and also
includes Protobuf encoding for JSONL data for reference. It compares size and speed of
JSONL vs STEF vs Protobuf encoding on some sample JSONL datasets.

This is about the worst case for STEF and Protobuf encodings since they have to model 
schemaless JSONL data structure using the available primitives and compete with native
hard-coded JSON implementation. The benchmark show how Protobuf struggles with this case.
Nevertheless, such JSON-like data is modeled in some production systems,
see for example this
[AnyValue](https://github.com/open-telemetry/opentelemetry-proto/blob/c30610041736aa5c0077b156f27b09e878b797ea/opentelemetry/proto/common/v1/common.proto#L28)
Protobuf schema used by OpenTelemetry.

Results below show STEF is able to outperform both native JSONL and Protobuf in terms of 
size and 
speed on all datasets.

## Size Comparison

The following table shows the size of the datasets in bytes. The numbers in parentheses
show the size factor compared to the original JSONL size.

| File                           | JSONL | Protobuf      | STEF         |
|--------------------------------|-------|---------------|--------------|
| currencies_historical          | 4197  | 4393 (1.05x)  | 1082 (0.26x) |
| macosx_releases                | 2449  | 2604 (1.06x)  | 935 (0.38x)  |
| programming_languages_keywords | 10681 | 11912 (1.12x) | 8253 (0.77x) |


## Deserialization Time Comparison

The times below show the duration of deserializing of one record (one line).

| Dataset                        | JSONL (sec/record) | STEF (sec/record) | STEF vs JSONL | Protobuf (sec/record) | Protobuf vs JSONL |
|--------------------------------|--------------------|-------------------|---------------|-----------------------|-------------------|
| currencies_historical          | 1147.5n            | 479.7n            | -58.20%       | 1445.5n               | +25.97%           |
| macosx_releases                | 1219.5n            | 748.0n            | -38.66%       | 1763.0n               | +44.57%           |
| programming_languages_keywords | 6.215µ             | 4.462µ            | -28.21%       | 9.636µ                | +55.06%           |

## How to run

- Prereqs: Go 1.24+, `protoc`.
- Build/tests: `make build`
- Regenerate STEF + Protobuf bindings: `make generate`
  - Invokes `stefc` for Go bindings under `internal/jsonstef`
  - Invokes `protoc` to emit JSONL protobuf types under `internal/jsonpb`
- Benchmarks: `make benchmark`

Sample data lives in `testdata/`; generated bindings stay under `internal/`.
