#!/bin/bash

# This script generates .stefz files from .zst files using the otlp2stef tool
# built from a specific base branch. It checks out the base branch, builds the tool,
# converts the files, and then checks out the original branch again.

set -euo pipefail

# Remember the current branch or commit hash if in detached HEAD state.
CUR_BRANCH=$(git symbolic-ref -q --short HEAD || git rev-parse HEAD)
echo "Remembering current branch/commit hash as $CUR_BRANCH"

TMPDIR=$(mktemp -d)
cp ./gentestfiles.sh "$TMPDIR/"

# Checkout the base branch to compare to
BASE_BRANCH=tigran/oneofcodec # Change this to the main branch commit after tigran/oneofcodec is merged.
git -c advice.detachedHead=false checkout $BASE_BRANCH

# Convert/generate the files.
$TMPDIR/gentestfiles.sh

# Checkout the previously remembered branch
git checkout "$CUR_BRANCH"

