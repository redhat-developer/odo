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
make test-cmd-login-logout
make test-json-format-output
make test-cmd-cmp
make test-cmd-cmp-sub
make test-cmd-pref-config
make test-cmd-watch
make test-cmd-storage
make test-cmd-app
make test-cmd-url
make test-cmd-push
odo logout