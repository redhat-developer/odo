#!/bin/sh

# fail if some commands fails
set -e
# show commands
set -x

export CI="openshift"
make configure-installer-tests-cluster
make bin
mkdir -p $GOPATH/bin
go get -u github.com/onsi/ginkgo/ginkgo
export PATH="$PATH:$(pwd):$GOPATH/bin"
export ARTIFACTS_DIR="/tmp/artifacts"
export CUSTOM_HOMEDIR=$ARTIFACTS_DIR


echo $KUBECONFIG
oc whoami
oc config view

# Integration tests
# make test-integration
make test-integration-devfile
# make test-cmd-login-logout
# make test-cmd-project
# make test-operator-hub

odo logout
