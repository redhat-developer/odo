#!/bin/sh

# fail if some commands fails
set -e
# show commands
set -x

export CI="openshift"
export TIMEOUT="30m"
make configure-installer-tests-cluster
make bin
go get -u github.com/onsi/ginkgo/ginkgo
export PATH="$PATH:$(pwd):$GOPATH/bin"
export CUSTOM_HOMEDIR="/tmp/artifacts"

make test-e2e-beta
make test-e2e-java
make test-e2e-source
odo logout
