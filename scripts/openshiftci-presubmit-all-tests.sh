#!/bin/sh

# fail if some commands fails
set -e
# show commands
set -x

export CI="openshift"
export GINKGO_VERBOSE_MODE="-v"
make configure-installer-tests-cluster
make bin
mkdir -p $GOPATH/bin
go get -u github.com/onsi/ginkgo/ginkgo
export PATH="$PATH:$(pwd):$GOPATH/bin"
export ARTIFACTS_DIR="/tmp/artifacts"
export CUSTOM_HOMEDIR=$ARTIFACTS_DIR

# Integration tests
make test-integration
make test-cmd-login-logout
make test-cmd-project

# E2e tests
make test-e2e-all

# Benchmark tests
make test-benchmark

odo logout
