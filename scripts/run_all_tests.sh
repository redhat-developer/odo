#!/usr/bin/env bash
set -ex

# Run unit tests
GOFLAGS='-mod=vendor' make test

GOBIN="`pwd`/bin"
if [[ $BASE_OS == "windows" ]]; then
    GOBIN="$(cygpath -pw $GOBIN)"
fi

PATH=$PATH:$GOBIN

# Prep for int
echo "getting ginkgo"
GOBIN="$GOBIN" make goget-ginkgo

set +e
ls -a $GOBIN
ginkgo version
run_all=$?
set -e
# Integration tests
if [[ $run_all -eq 0 ]]; then
    make test-e2e-all
else
    echo "Ginkgo does not exist, skipping integration/e2e"
fi

