#!/bin/bash

LOGFILE="pr-${GIT_PR_NUMBER}-openshift-tests-${BUILD_NUMBER}"

source .ibm/pipelines/functions.sh

install_oc

oc login -u apikey -p "${API_KEY}" "${IBM_OPENSHIFT_ENDPOINT}"

(
    set -e
    make install
    make test-integration
    make test-integration-devfile
    make test-operator-hub
    make test-cmd-login-logout
    make test-cmd-project
    make test-e2e-devfile
) |& tee "/tmp/${LOGFILE}"
RESULT=${PIPESTATUS[0]}

install_ibmcloud cloud-object-storage
install_gh
save_logs "${LOGFILE}" "OpenShift Tests"

exit ${RESULT}
