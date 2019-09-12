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
export ARTIFACTS_DIR="/tmp/artifacts"
export CUSTOM_HOMEDIR=$ARTIFACTS_DIR
make test-generic
make test-cmd-login-logout
make test-cmd-cmp
make test-cmd-cmp-sub
make test-cmd-pref-config
make test-cmd-watch
make test-cmd-storage
make test-cmd-app
make test-cmd-url
make test-cmd-push
odo logout

# upload the junit test reports
cp -r reports $ARTIFACTS_DIR