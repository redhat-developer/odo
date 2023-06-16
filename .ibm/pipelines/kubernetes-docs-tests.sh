#!/bin/bash

LOGFILE="pr-${GIT_PR_NUMBER}-kubernetes-docs-tests-${BUILD_NUMBER}"
TEST_NAME="Kubernetes Docs Tests"

source .ibm/pipelines/functions.sh

ibmcloud login --apikey "${API_KEY_QE}"
ibmcloud target -r "${IBM_REGION}"
ibmcloud ks cluster config --cluster "${IBM_KUBERNETES_ID}" --admin

cleanup_namespaces
(
    set -e
    make install
    make test-doc-automation
) |& tee "/tmp/${LOGFILE}"

RESULT=${PIPESTATUS[0]}

save_logs "${LOGFILE}" "${TEST_NAME}" ${RESULT}
save_results "${PWD}/test-doc-automation.xml" "${LOGFILE}" "${TEST_NAME}" "${BUILD_NUMBER}"

exit ${RESULT}
