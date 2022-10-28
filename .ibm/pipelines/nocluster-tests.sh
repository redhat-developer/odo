#!/bin/bash

LOGFILE="pr-${GIT_PR_NUMBER}-nocluster-tests-${BUILD_NUMBER}"

source .ibm/pipelines/functions.sh

ibmcloud login --apikey "${API_KEY_QE}"
ibmcloud target -r "${IBM_REGION}"

(
    set -e
    make install
    make test-integration-no-cluster
) |& tee "/tmp/${LOGFILE}"

RESULT=${PIPESTATUS[0]}

save_logs "${LOGFILE}" "NoCluster Tests" ${RESULT}

exit ${RESULT}
