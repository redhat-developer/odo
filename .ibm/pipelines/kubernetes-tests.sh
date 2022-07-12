#!/bin/bash

LOGFILE="pr-${GIT_PR_NUMBER}-kubernetes-tests-${BUILD_NUMBER}"

source .ibm/pipelines/functions.sh

ibmcloud login --apikey "${API_KEY_QE}"
ibmcloud target -r "${IBM_REGION}"
ibmcloud ks cluster config --cluster "${IBM_KUBERNETES_ID}" --admin

cleanup_namespaces
export SKIP_USER_LOGIN_TESTS=true
(
    set -e
    make install
    make test-integration
    make test-e2e
) |& tee "/tmp/${LOGFILE}"

RESULT=${PIPESTATUS[0]}

save_logs "${LOGFILE}" "Kubernetes Tests" ${RESULT}

exit ${RESULT}
