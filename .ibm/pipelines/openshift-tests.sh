#!/bin/bash

LOGFILE="pr-${GIT_PR_NUMBER}-openshift-tests-${BUILD_NUMBER}"

source .ibm/pipelines/functions.sh

oc login -u apikey -p "${API_KEY}" "${IBM_OPENSHIFT_ENDPOINT}"

cleanup_namespaces

(
    set -e
    make install
    make test-integration
    make test-integration-devfile
    make test-cmd-login-logout
    make test-cmd-project
    make test-e2e-devfile
) |& tee "/tmp/${LOGFILE}"
RESULT=${PIPESTATUS[0]}

ibmcloud login --apikey "${API_KEY}" -r "${IBM_REGION}"
save_logs "${LOGFILE}" "OpenShift Tests" ${RESULT}

exit ${RESULT}
