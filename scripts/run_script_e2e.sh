#!/usr/bin/env bash

shout() {
  set +x
  echo -e "\n!!!!!!!!!!!!!!!!!!!!\n${1}\n!!!!!!!!!!!!!!!!!!!!\n"
  set -x
}

# Prep for integration/e2e
shout "Building odo binaries"
make bin

# copy built odo to GOBIN
cp -avf ./odo $GOBIN_TEMP/

# Integration tests
shout "Testing against 4x cluster"

shout "Logging into 4x cluster for some setup (logs hidden)"
shout $CLUSTER_TYPE
set +x
if [[ $CLUSTER_TYPE == "PSI" ]]
then
    #PSI cluster login
    oc login -u kubeadmin -p ${OCP4X_KUBEADMIN_PASSWORD} --insecure-skip-tls-verify  ${OCP4X_API_URL}
else
    # Login to cluster in IBM Cloud using cluster API key
    oc login -u $IBM_OC_LOGIN_USER -p $IBMC_ADMIN_OCLOGIN_APIKEY --server=$IBMC_OCP47_SERVER

fi
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
for i in `echo "$REDHAT_OPENJDK11_RHEL8_PROJECT $REDHAT_NODEJS12_RHEL7_PROJECT $REDHAT_NODEJS12_UBI8_PROJECT $REDHAT_OPENJDK11_UBI8_PROJECT $REDHAT_NODEJS14_UBI8_PROJECT"`; do
    # create the namespace
    oc new-project $i
    # Applying pull secret to the namespace which will be used for pulling images from authenticated registry
    oc get secret pull-secret -n openshift-config -o yaml | sed "s/openshift-config/$i/g" | oc apply -f -
    # Let developer user have access to the project
    oc adm policy add-role-to-user edit developer
done

shout "Logging into 4x cluster as developer (logs hidden)"
set +x
if [[ $CLUSTER_TYPE == "PSI" ]]
then
    #PSI cluster login
    oc login -u developer -p ${OCP4X_DEVELOPER_PASSWORD} --insecure-skip-tls-verify ${OCP4X_API_URL}
else

    # Login to cluster in IBM Cloud using cluster API key
    oc login -u $IBM_OC_LOGIN_USER -p $IBMC_DEVELOPER_OCLOGIN_APIKEY --server=$IBMC_OCP47_SERVER
  
fi

set -x

# # Integration tests
shout "Running integration Tests"
make test-operator-hub || error=true
make test-integration || error=true
make test-integration-devfile || error=true 
make test-cmd-login-logout || error=true
make test-cmd-project || error=true


# E2e tests
shout "Running e2e tests"
make test-e2e-devfile || error=true


shout "copying artifacts"
cp -r reports tests/reports $ARTIFACT_DIR

shout "cleaning up post tests"
shout "Logging into 4x cluster for cleanup (logs hidden)"
set +x
if [[ $CLUSTER_TYPE == "PSI" ]]
then
    #PSI cluster login
    oc login -u kubeadmin -p ${OCP4X_KUBEADMIN_PASSWORD} --insecure-skip-tls-verify  ${OCP4X_API_URL}
else
    # Login to cluster in IBM Cloud using cluster API key
    oc login -u $IBM_OC_LOGIN_USER -p $IBMC_ADMIN_OCLOGIN_APIKEY --server=$IBMC_OCP47_SERVER
    
fi
set -x

shout "Cleaning up some leftover projects"

set +x
for i in $(oc projects -q); do
    if [[ $i == "${SCRIPT_IDENTITY}"* ]]; then
        oc delete project $i
    fi
done
set -x

# Fail the build if there is any error while test execution
if [ $error ]; then 
    exit -1
fi