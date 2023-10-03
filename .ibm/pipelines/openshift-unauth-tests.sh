#!/bin/bash

LOGFILE="pr-${GIT_PR_NUMBER}-openshift-unauth-tests-${BUILD_NUMBER}"
TEST_NAME="OpenShift Unauthenticated Tests"

source .ibm/pipelines/functions.sh

skip_if_only

ibmcloud login --apikey "${API_KEY_QE}"
ibmcloud target -r eu-de
ibmcloud oc cluster config -c "${CLUSTER_ID}"

(
    set -e
    make install
    make test-integration-openshift-unauth
) |& tee "/tmp/${LOGFILE}"

RESULT=${PIPESTATUS[0]}

save_logs "${LOGFILE}" "${TEST_NAME}" ${RESULT}
save_results "${PWD}/test-integration-unauth.xml" "${LOGFILE}" "${TEST_NAME}" "${BUILD_NUMBER}"

exit ${RESULT}
