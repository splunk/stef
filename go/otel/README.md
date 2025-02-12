## Otel/STEF Format Definition

Otel/STEF is a representation of OpenTelemetry data model in
[STEF](/stef-spec/specification.md) format.

OTEL/STEF schema definition is [here](oteltef.wire.json). Reader/Writer Go code is generated from the
schema and is placed in [oteltef](oteltef) directory.

To re-generate Reader/Writer run `make generate`.
