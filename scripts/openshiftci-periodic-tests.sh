#!/bin/sh

# fail if some commands fails
set -e
# show commands
set -x

export CI="openshift"
make configure-installer-tests-cluster
make bin
mkdir -p $GOPATH/bin
make goget-ginkgo
export PATH="$PATH:$(pwd):$GOPATH/bin"
export ARTIFACTS_DIR="/tmp/artifacts"
export CUSTOM_HOMEDIR=$ARTIFACTS_DIR

# Copy kubeconfig to temporary kubeconfig file
# Read and Write permission to temporary kubeconfig file
TMP_DIR=$(mktemp -d)
cp $KUBECONFIG $TMP_DIR/kubeconfig
chmod 640 $TMP_DIR/kubeconfig
export KUBECONFIG=$TMP_DIR/kubeconfig

# Login as developer
odo login -u developer -p developer

# Check login user name for debugging purpose
oc whoami

# Integration tests
make test-integration
make test-integration-devfile
make test-cmd-login-logout
make test-cmd-project
make test-operator-hub

# E2e tests
make test-e2e-all

odo logout
