#!/bin/bash
# Update the operator version to a new version at various places across the repository.
# Refer https://semver.org/

set -e
set -u

MANIFESTS_DIR="./../manifests-upstream"
NEW_VERSION=$1
filename="../Makefile"
OLD_VERSION=$(grep -m 1 OPERATOR_VERSION $filename | sed 's/^.*= //g')

function replace {
    LOCATION=$1
    if [ -e $LOCATION ] ; then
        sed -i -e 's/'${OLD_VERSION}'/'${NEW_VERSION}'/g' $LOCATION
    else
        echo ERROR: Failed to find $LOCATION
        exit 1 #terminate and indicate error
    fi
}
function move {
    OLD_LOCATION=$1
    NEW_LOCATION=$2
    if [ -e ${OLD_LOCATION} ] || [ -e ${NEW_LOCATION} ] ; then
        mv ${OLD_LOCATION} ${NEW_LOCATION}
    else
        echo ERROR: Failed to find file location
        exit 1 #terminate and indicate error
    fi
}
move ${MANIFESTS_DIR}/${OLD_VERSION} ${MANIFESTS_DIR}/${NEW_VERSION}
replace ${MANIFESTS_DIR}/service-binding-operator.package.yaml
move ${MANIFESTS_DIR}/${NEW_VERSION}/service-binding-operator.v${OLD_VERSION}.clusterserviceversion.yaml \
${MANIFESTS_DIR}/${NEW_VERSION}/service-binding-operator.v${NEW_VERSION}.clusterserviceversion.yaml
replace ${MANIFESTS_DIR}/${NEW_VERSION}/service-binding-operator.v${NEW_VERSION}.clusterserviceversion.yaml
replace ./../openshift-ci/Dockerfile.registry.build
replace ${filename}
echo -e "\n\033[0;32m \xE2\x9C\x94 Operator version upgraded from \
${OLD_VERSION} to ${NEW_VERSION} \033[0m\n"
