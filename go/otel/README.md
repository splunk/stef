## Otel/STEF Format Definition

Otel/STEF is a representation of OpenTelemetry data model in
[STEF](/stef-spec/specification.md) format.

OTEL/STEF schema definition is [here](otel.stef). Reader/Writer Go code is generated from the
schema and is placed in [otelstef](otelstef) directory.

Generated Go reader/writer code is in `otelstef/`. Java bindings are emitted into
`java/src/main/java` by the makefile target.

## Regenerate bindings

- From this directory run: `make generate`
  - Generates Go code under `otelstef/`
  - Generates Java code/tests under `../../java/src/main/java` and `../../java/src/test/java`

## Develop and test

If you modify `otel.stef` or templates in `stefc`, rerun `make generate` here and commit
updated generated code in both Go and Java trees.
