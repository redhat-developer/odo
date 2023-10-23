#!/bin/bash

LOGFILE="pr-${GIT_PR_NUMBER}-openshift-tests-${BUILD_NUMBER}"
TEST_NAME="OpenShift Tests"

source .ibm/pipelines/functions.sh

skip_if_only

ibmcloud login --apikey "${API_KEY_QE}"
ibmcloud target -r eu-de
ibmcloud oc cluster config -c "${CLUSTER_ID}"
oc login -u apikey -p "${API_KEY_QE}" "${IBM_OPENSHIFT_ENDPOINT}"

cleanup_namespaces

(
    set -e
    make install
    make test-integration-cluster
    make test-e2e
) |& tee "/tmp/${LOGFILE}"

RESULT=${PIPESTATUS[0]}

save_logs "${LOGFILE}" "${TEST_NAME}" ${RESULT}
save_results "${PWD}/test-integration.xml" "${LOGFILE}" "${TEST_NAME}" "${BUILD_NUMBER}"
save_results "${PWD}/test-e2e.xml" "${LOGFILE}" "${TEST_NAME}" "${BUILD_NUMBER}"

exit ${RESULT}
