# STEF Compiler

stefc is a command-line compiler and code generator that takes STEF schema files as
input and generates type-safe serialization and deserialization code for multiple
programming languages. For usage instructions, please refer to
[STEFC documentation](https://www.stefdata.net/stefc.html).

## Design

stefc is implemented in Go and uses Go [text/template](https://pkg.go.dev/text/template)
for code generation. There is a set of templates for each target language:

- [Templates for Go](./templates/go)
- [Templates for Java](./templates/java)

stefc parses the input STEF schema file, then generates code by applying the parsed schema
to the appropriate templates for the specified target language. See generator
implementation in [generator](./generator) package.

Template files contain templated code to represent in-memory and to serialize and
deserialize supported types (structs, oneofs, enums, arrays, maps).

## Building and Testing

Pre-requisites:

- make
- Go 1.25 or newer

To build stefc, run `make build`. To run tests, use `make test`.

[generator/testdata](./generator/testdata) contains STEF schema files that are used in
unit tests for the code generator. For each of these schema files, a test is run that
generates Go and Java code from the schema, then runs generated tests that verify
correctness of serialization and deserialization. Generated tests use randomized input
from an initial seed of a random number generator. "seeds" directory contains recorded
seeds for each test case that previously resulted in detected bugs. These seeds are used
to avoid regressions.

If you make any changes to template files or to the generator code, rerun `make all`
to ensure that all generated code is up to date and all tests are run.

stefc is also used for generating code in [examples](../examples) directory and
in [otel](../go/otel) directory.
