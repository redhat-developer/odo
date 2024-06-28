#!/bin/sh

# This file is used for InterOP testing, i.e. testing odo with unreleased OpenShift versions.

# fail if some commands fails
set -e
# show commands
set -x

ARCH=$(uname -m)
export CI="openshift"
if [ "${ARCH}" == "s390x" ]; then
    make configure-installer-tests-cluster-s390x
elif [ "${ARCH}" == "ppc64le" ]; then
    make configure-installer-tests-cluster-ppc64le
else
    make configure-installer-tests-cluster
fi
make bin
mkdir -p $GOPATH/bin
export PATH="$PATH:$(pwd):$GOPATH/bin"
export CUSTOM_HOMEDIR=$ARTIFACT_DIR

# Copy kubeconfig to temporary kubeconfig file
# Read and Write permission to temporary kubeconfig file
TMP_DIR=$(mktemp -d)
cp $KUBECONFIG $TMP_DIR/kubeconfig
chmod 640 $TMP_DIR/kubeconfig
export KUBECONFIG=$TMP_DIR/kubeconfig

# Login as developer
oc login -u developer -p password@123 --insecure-skip-tls-verify
# Check login user name for debugging purpose
oc whoami

# We want to use a stable Devfile registry for InterOP testing, and so we use the custom Devfile Registry setup on IBM cloud
source ./scripts/openshiftci-config.sh

# Integration tests
make test-integration-openshift || error=true

# Login again (in case the token expires for some reason)
oc login -u developer -p password@123 --insecure-skip-tls-verify || true
oc whoami

# E2e tests
make test-e2e || error=true

# Fail the build if there is any error while test execution
if [ $error ]; then
    exit -1
fi

if [ ! -z "$ARTIFACT_DIR" ]
then
    #copy artifact to $ARTIFACT_DIR if ARTIFACT_DIR var is exposed
    cp -vr test-*.xml $ARTIFACT_DIR || true
fi

oc logout
