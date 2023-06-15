#!/bin/bash

LOGFILE="pr-${GIT_PR_NUMBER}-kubernetes-tests-${BUILD_NUMBER}"
TEST_NAME="Kubernetes Tests"

source .ibm/pipelines/functions.sh

skip_if_only

ibmcloud login --apikey "${API_KEY_QE}"
ibmcloud target -r "${IBM_REGION}"
ibmcloud ks cluster config --cluster "${IBM_KUBERNETES_ID}" --admin

cleanup_namespaces
export SKIP_USER_LOGIN_TESTS=true
(
    set -e
    export DEVFILE_PROXY="$(kubectl get svc -n devfile-proxy nginx -o jsonpath='{.status.loadBalancer.ingress[0].hostname}' || true)"
    echo Using Devfile proxy: ${DEVFILE_PROXY}
    make install
    make test-integration-cluster
) |& tee "/tmp/${LOGFILE}"

RESULT=${PIPESTATUS[0]}

save_logs "${LOGFILE}" "${TEST_NAME}" ${RESULT}
save_results "${PWD}/test-integration.xml" "${LOGFILE}" "${TEST_NAME}" "${BUILD_NUMBER}"
exit ${RESULT}
