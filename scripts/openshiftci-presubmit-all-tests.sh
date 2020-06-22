#!/bin/sh

# fail if some commands fails
set -e
# show commands
set -x

export GIT_LOG_FILE="/tmp/git-log"
export CI="openshift"

git log master.. > $GIT_LOG_FILE
cat $GIT_LOG_FILE | grep "skip ci"

if [[ $? -ne 0 ]]; then
        echo "Keyword 'skip ci' detected Skipping CI run"
	exit 0
fi

make configure-installer-tests-cluster
make bin
mkdir -p $GOPATH/bin
go get -u github.com/onsi/ginkgo/ginkgo
export PATH="$PATH:$(pwd):$GOPATH/bin"
export ARTIFACTS_DIR="/tmp/artifacts"
export CUSTOM_HOMEDIR=$ARTIFACTS_DIR

# Integration tests
make test-integration
make test-integration-devfile
make test-cmd-login-logout
make test-cmd-project
make test-operator-hub

# E2e tests
make test-e2e-all

odo logout
