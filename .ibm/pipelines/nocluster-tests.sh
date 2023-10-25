#!/bin/bash

LOGFILE="pr-${GIT_PR_NUMBER}-nocluster-tests-${BUILD_NUMBER}"
TEST_NAME="NoCluster Tests"

source .ibm/pipelines/functions.sh

skip_if_only

ibmcloud login --apikey "${API_KEY_QE}"
ibmcloud target -r "${IBM_REGION}"

(
    set -e
    make install
    make test-integration-no-cluster
) |& tee "/tmp/${LOGFILE}"

RESULT=${PIPESTATUS[0]}

save_logs "${LOGFILE}" "${TEST_NAME}" ${RESULT}
save_results "${PWD}/test-integration-nc.xml" "${LOGFILE}" "${TEST_NAME}" "${BUILD_NUMBER}"

exit ${RESULT}
