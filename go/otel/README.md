## Otel/STEF Format Definition

Otel/STEF is a representation of OpenTelemetry data model in
[STEF](/stef-spec/specification.md) format.

OTEL/STEF schema definition is [here](otel.stef). Reader/Writer Go code is generated from the
schema and is placed in [otelstef](otelstef) directory.

To re-generate Reader/Writer run `make generate`.
