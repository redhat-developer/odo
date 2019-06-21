#!/bin/sh

# fail if some commands fails
set -e
# show commands
set -x

export CI="openshift"
export TIMEOUT="30m"
make configure-installer-tests-cluster
make bin
mkdir -p $GOPATH/bin
go get -u github.com/onsi/ginkgo/ginkgo
export PATH="$PATH:$(pwd):$GOPATH/bin"
export CUSTOM_HOMEDIR="/tmp/artifacts"
make test-generic
make test-odo-login-e2e
make test-json-format-output
make test-java-e2e
make test-source-e2e
make test-cmp-e2e
make test-cmp-sub-e2e
make test-odo-config
make test-watch-e2e
make test-cmd-storage
make test-cmd-app
