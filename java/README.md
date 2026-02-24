This directory contains Java implementation of STEF.

# Prerequisites

- Install Java 8 or newer
- Install Gradle
- Install [Go 1.25 or newer](https://go.dev/doc/install)
- Run `cd ../benchmarks && make gentestfiles`. This will generate test files that are used by Java tests.

# Building and Running Tests

To build and run all tests do `./gradlew build`.

To run tests on variety of test schemas first run `cd ../stefc && make test`.
This will generate serializers/deserializers for test schemas in Java and will place
them in [src/test/java/com/example/gentest](./src/test/java/com/example/gentest) directory.
Then run `./gradlew test` to test generated code.

To generate the JMH performance benchmarks run `./gradlew jmh`.
