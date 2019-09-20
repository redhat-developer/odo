#!/bin/sh

# fail if some commands fails
set -e
# show commands
set -x

export CI="openshift"
export TIMEOUT="30m"
export GINKGO_VERBOSE_MODE="-v"
make configure-installer-tests-cluster
make bin
mkdir -p $GOPATH/bin
go get -u github.com/onsi/ginkgo/ginkgo
export PATH="$PATH:$(pwd):$GOPATH/bin"
export CUSTOM_HOMEDIR="/tmp/artifacts"

env

make test-benchmark
odo logout
