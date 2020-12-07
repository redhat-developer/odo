#!/usr/bin/env bash

shout() {
  set +x
  echo -e "\n!!!!!!!!!!!!!!!!!!!!\n${1}\n!!!!!!!!!!!!!!!!!!!!\n"
  set -x
}

set -ex

shout "Setting up"

mkdir bin
GOBIN="`pwd`/bin"
KUBECONFIG="`pwd`/config"
SCRIPT_IDENTITY=${SCRIPT_IDENTITY:-"def-id"}
export SKIP_USER_LOGIN_TESTS="true"
export GINKGO_TEST_ARGS="--noColor"

shout "Getting oc binary"
if [[ $BASE_OS == "linux"  ]]; then
    set +x
    curl --connect-timeout 150 --max-time 2048 -k ${OCP4X_DOWNLOAD_URL}/${ARCH}/${BASE_OS}/oc.tar -o ./oc.tar
    set -x
    tar -C $GOBIN -xvf ./oc.tar && rm -rf ./oc.tar
else
    set +x
    curl --connect-timeout 150 --max-time 2048 -k ${OCP4X_DOWNLOAD_URL}/${ARCH}/${BASE_OS}/oc.zip -o ./oc.zip
    set -x
    if [[ $BASE_OS == "windows" ]]; then
        GOBIN="$(cygpath -pw $GOBIN)"
        CURRDIR="$(cygpath -pw $WORKDIR)"
        powershell -Command "Expand-Archive -Path $CURRDIR\oc.zip  -DestinationPath $GOBIN"
        chmod +x $GOBIN/*
    fi
    if [[ $BASE_OS == "mac" ]]; then
        unzip ./oc.zip -d $GOBIN && rm -rf ./oc.zip && chmod +x $GOBIN/oc
        PATH="$PATH:/usr/local/bin:/usr/local/go/bin"
    fi
fi

PATH=$PATH:$GOBIN

#-----------------------------------------------------------------------------

shout "Testing"

# Run unit tests
GOFLAGS='-mod=vendor' make test

# Prep for int
shout "Building"
make bin
cp -avrf ./odo $GOBIN/
shout "getting ginkgo"
GOBIN="$GOBIN" make goget-ginkgo

# Integration tests
shout "Testing against 4x cluster"

shout "Logging into 4x cluster for some setup (logs hidden)"
set +x
oc login -u kubeadmin -p ${OCP4X_KUBEADMIN_PASSWORD} --insecure-skip-tls-verify  ${OCP4X_API_URL}
set -x

shout "Doing some presetup"

for i in $(oc projects -q); do
    if [[ $i == "${SCRIPT_IDENTITY}"* ]]; then
        oc delete project $i
    fi
done

export REDHAT_OPENJDK11_RHEL8_PROJECT="${SCRIPT_IDENTITY}$(cat /dev/urandom | tr -dc 'a-z0-9' | fold -w 4 | head -n 1)"
export REDHAT_OPENJDK11_UBI8_PROJECT="${SCRIPT_IDENTITY}$(cat /dev/urandom | tr -dc 'a-z0-9' | fold -w 4 | head -n 1)"
export REDHAT_NODEJS12_RHEL7_PROJECT="${SCRIPT_IDENTITY}$(cat /dev/urandom | tr -dc 'a-z0-9' | fold -w 4 | head -n 1)"
export REDHAT_NODEJS12_UBI8_PROJECT="${SCRIPT_IDENTITY}$(cat /dev/urandom | tr -dc 'a-z0-9' | fold -w 4 | head -n 1)"

# Create the namespace for e2e image test apply pull secret to the namespace
for i in `echo "$REDHAT_OPENJDK11_RHEL8_PROJECT $REDHAT_NODEJS12_RHEL7_PROJECT $REDHAT_NODEJS12_UBI8_PROJECT $REDHAT_OPENJDK11_UBI8_PROJECT"`; do
    # create the namespace
    oc new-project $i
    # Applying pull secret to the namespace which will be used for pulling images from authenticated registry
    oc get secret pull-secret -n openshift-config -o yaml | sed "s/openshift-config/$i/g" | oc apply -f -
    # Let developer user have access to the project
    oc adm policy add-role-to-user edit developer
done

shout "Logging into 4x cluster as developer (logs hidden)"
set +x
oc login -u developer -p ${OCP4X_DEVELOPER_PASSWORD} --insecure-skip-tls-verify ${OCP4X_API_URL}
set -x

shout "Running integration Tests"
make test-integration
make test-integration-devfile	
make test-cmd-login-logout	
make test-cmd-project	
make test-operator-hub

shout "Running e2e tests"
make test-e2e-all

shout "cleanup"
shout "Logging into 4x cluster for cleanup (logs hidden)"
set +x
oc login -u kubeadmin -p ${OCP4X_KUBEADMIN_PASSWORD} --insecure-skip-tls-verify ${OCP4X_API_URL}
set -x

shout "Cleaning up some leftover projects"

set +x
for i in $(oc projects -q); do
    if [[ $i == "${SCRIPT_IDENTITY}"* ]]; then
        oc delete project $i
    fi
done
set -x
