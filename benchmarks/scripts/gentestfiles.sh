#!/bin/bash

# This script generates .stefz files from .zst files using the otlp2stef tool
# Generated files are placed in testdata/generated/.

set -euo pipefail

# Build otlp2stef tool (see makefile for how to do it)
echo
echo "Building otlp2stef tool"
go build -o ../bin/otlp2stef ../cmd/otlp2stef/main.go

# Copy .zst files to a temp directory
echo
echo "Converting .zst files to .stefz format using otlp2stef tool"
TMPDIR=$(mktemp -d)
cp ../testdata/astronomy-otelmetrics.zst "$TMPDIR/"
cp ../testdata/hipstershop-otelmetrics.zst "$TMPDIR/"
cp ../testdata/hostandcollector-otelmetrics.zst "$TMPDIR/"

# Use otlp2stef tool to convert these files to stefz format
for f in astronomy-otelmetrics hipstershop-otelmetrics hostandcollector-otelmetrics; do
  ../bin/otlp2stef \
    -compression=zstd \
    --input="$TMPDIR/$f.zst"
done

# Copy stefz files from temp directory into testdata directory, overwriting if they exist
echo
echo "Copying converted .stefz files to ../testdata/generated/ directory"
mkdir -p "../testdata/generated/"
cp "$TMPDIR/astronomy-otelmetrics.stefz" ../testdata/generated/
cp "$TMPDIR/hipstershop-otelmetrics.stefz" ../testdata/generated/
cp "$TMPDIR/hostandcollector-otelmetrics.stefz" ../testdata/generated/

echo "Test files generated."
