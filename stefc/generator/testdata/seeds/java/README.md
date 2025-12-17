This directory stores **random seeds that previously caused failures in Java-generated schema tests** (e.g. `*WriterTest`).

## Why this exists

- When a randomized Java writer/reader test fails, it prints the seed and **appends it** to the appropriate `*_seeds.txt` file in this directory.
- On subsequent test runs, the seeds in these files are **replayed first** to ensure the bug does not regress.

## File naming convention
Each file is named:
`<javaPackage>_<RootStruct>_seeds.txt`
Example:
`com.example.otelstef_Metrics_seeds.txt`

## Important
- Only add seeds that were found by **Java** failures. Go/other-language seeds are not portable because RNGs differ.