#!/usr/bin/env bash

shout() {
  set +x
  echo -e "\n!!!!!!!!!!!!!!!!!!!!\n${1}\n!!!!!!!!!!!!!!!!!!!!\n"
  set -x
}


set -ex

shout "Setting up some stuff"

# Create a bin directory whereever script runs. This will be where all binaries that need to be in PATH will reside.
mkdir bin artifacts
# Change the default location of go's bin directory (without affecting GOPATH). This is where compiled binaries will end up by default
# for eg go get ginkgo later on will produce ginkgo binary in GOBIN
export GOBIN="`pwd`/bin"
# Set kubeconfig to current dir. This ensures no clashes with other test runs
export KUBECONFIG="`pwd`/config"
export ARTIFACTS_DIR="`pwd`/artifacts"
export CUSTOM_HOMEDIR=$ARTIFACT_DIR
LIBDIR="./scripts/configure-cluster"
LIBCOMMON="$LIBDIR/common"

# This si one of the variables injected by ci-firewall. Its purpose is to allow scripts to handle uniqueness as needed
SCRIPT_IDENTITY=${SCRIPT_IDENTITY:-"def-id"}
export SKIP_USER_LOGIN_TESTS="true"

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

# Add GOBIN which is the bin dir we created earlier to PATH so any binaries there are automatically available in PATH
export PATH=$PATH:$GOBIN

#-----------------------------------------------------------------------------

shout "Running unit tests"

# Run unit tests
GOFLAGS='-mod=vendor' make test

# Prep for integration/e2e
shout "Building odo binaries"
make bin

# copy built odo to GOBIN
cp -avrf ./odo $GOBIN/
shout "getting ginkgo"
make goget-ginkgo

# Integration tests
shout "Testing against 4x cluster"

shout "Logging into 4x cluster for some setup (logs hidden)"
set +x
oc login -u kubeadmin -p ${OCP4X_KUBEADMIN_PASSWORD} --insecure-skip-tls-verify  ${OCP4X_API_URL}
set -x

shout "Doing some presetup"

# Delete any projects with SCRIPT_IDENTITY PREFIX. This is GC from previous runs which fail before end of script cleanup
for i in $(oc projects -q); do
    if [[ $i == "${SCRIPT_IDENTITY}"* ]]; then
        oc delete project $i
    fi
done

# Generate random project names to some tests
export REDHAT_OPENJDK11_RHEL8_PROJECT="${SCRIPT_IDENTITY}$(cat /dev/urandom | tr -dc 'a-z0-9' | fold -w 4 | head -n 1)"
export REDHAT_OPENJDK11_UBI8_PROJECT="${SCRIPT_IDENTITY}$(cat /dev/urandom | tr -dc 'a-z0-9' | fold -w 4 | head -n 1)"
export REDHAT_NODEJS12_RHEL7_PROJECT="${SCRIPT_IDENTITY}$(cat /dev/urandom | tr -dc 'a-z0-9' | fold -w 4 | head -n 1)"
export REDHAT_NODEJS12_UBI8_PROJECT="${SCRIPT_IDENTITY}$(cat /dev/urandom | tr -dc 'a-z0-9' | fold -w 4 | head -n 1)"
export REDHAT_NODEJS14_UBI8_PROJECT="${SCRIPT_IDENTITY}$(cat /dev/urandom | tr -dc 'a-z0-9' | fold -w 4 | head -n 1)"

# Create the namespace for e2e image test apply pull secret to the namespace
for i in `echo "$REDHAT_OPENJDK11_RHEL8_PROJECT $REDHAT_NODEJS12_RHEL7_PROJECT $REDHAT_NODEJS12_UBI8_PROJECT $REDHAT_OPENJDK11_UBI8_PROJECT $REDHAT_NODEJS14_UBI8_PROJECT $REDHAT_POSTGRES_OPERATOR_PROJECT"`; do
    # create the namespace
    oc new-project $i
    # Applying pull secret to the namespace which will be used for pulling images from authenticated registry
    oc get secret pull-secret -n openshift-config -o yaml | sed "s/openshift-config/$i/g" | oc apply -f -
    # Let developer user have access to the project
    oc adm policy add-role-to-user edit developer
done

#---------------------------------------------------------------------

shout "Logging into 4x cluster as developer (logs hidden)"
set +x
oc login -u developer -p ${OCP4X_DEVELOPER_PASSWORD} --insecure-skip-tls-verify ${OCP4X_API_URL}
set -x

# Integration tests
shout "Running integration Tests"
make test-operator-hub || error=true
make test-integration || error=true
make test-integration-devfile || error=true
make test-cmd-login-logout || error=true
make test-cmd-project || error=true

# E2e tests
shout "Running e2e tests"
make test-e2e-all || error=true
# Fail the build if there is any error while test execution
if [ $error ]; then 
    exit -1
fi

shout "cleaning up post tests"
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
