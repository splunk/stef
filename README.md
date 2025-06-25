# STEF (Sequential Tabular Encoding Format)

STEF is a data format and network protocol optimized for small payload size
and fast serialization. See some
[benchmark results here](https://splunk.github.io/stef/benchmarks.html).

### Directories

- [benchmarks](./benchmarks): Benchmarks, tests, comparisons to other formats.
- [otelcol](./otelcol): Otel/STEF protocol Collector exporter implementation.
- [go/pkg](./go/pkg): STEF supporting libraries for Go.
- [go/grpc](./go/grpc): STEF/gRPC protocol implementation in Go.
- [go/otel](./go/otel): Otel/STEF protocol schema and generated code in Go.
- [go/pdata](./go/pdata): Collector pdata <-> Otel/STEF converters.
- [stef-spec](./stef-spec): STEF Specification and Protobuf definitions.
- [stefgen](./stefgen): Generates serializers from STEF schema.

## Splunk Copyright Notice

Copyright 2022 Splunk Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and limitations under the License.
