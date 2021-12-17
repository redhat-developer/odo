#!/bin/bash

LOGFILE="pr-${GIT_PR_NUMBER}-kubernetes-tests-${BUILD_NUMBER}"

source .ibm/pipelines/functions.sh

ibmcloud login --apikey "${API_KEY}" -r "${IBM_REGION}"
ibmcloud ks cluster config --cluster "${IBM_KUBERNETES_ID}" --admin

cleanup_namespaces

(
    set -e
    make install
    make test-integration-devfile
    make test-e2e-devfile
    make test-cmd-project
) |& tee "/tmp/${LOGFILE}"
RESULT=${PIPESTATUS[0]}

save_logs "${LOGFILE}" "Kubernetes Tests" ${RESULT}

exit ${RESULT}
