This directory contains Java implementation of STEF.

# Prerequisites

- Install Java 8 or newer
- Install Gradle
- Install [Go 1.24 or newer](https://go.dev/doc/install)
- Run `cd ../benchmarks && make gentestfiles`. This will generate test files that are used by Java tests.

# Building and Running Tests

- To build and run all tests do `./gradlew build`.
- To generate the JMH performance benchmarks run `./gradlew jmh`.
