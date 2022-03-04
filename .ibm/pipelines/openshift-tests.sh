#!/bin/bash

LOGFILE="pr-${GIT_PR_NUMBER}-openshift-tests-${BUILD_NUMBER}"

source .ibm/pipelines/functions.sh

oc login -u apikey -p "${API_KEY_QE}" "${IBM_OPENSHIFT_ENDPOINT}"

cleanup_namespaces

(
    set -e
    make install
    make test-integration
    make test-integration-interactive
    make test-integration-devfile
    make test-cmd-login-logout
    make test-cmd-project
    make test-e2e-devfile
) |& tee "/tmp/${LOGFILE}"
RESULT=${PIPESTATUS[0]}

save_logs "${LOGFILE}" "OpenShift Tests" ${RESULT}

exit ${RESULT}
