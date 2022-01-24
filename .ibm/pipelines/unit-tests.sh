#!/bin/bash

# Script to run unit test on IBM Cloud 
# This script needs update if there is any change in the unit test make target command

LOGFILE="pr-${GIT_PR_NUMBER}-unit-tests-${BUILD_NUMBER}"

source .ibm/pipelines/functions.sh

(
    set -e
    make test
) |& tee "/tmp/$LOGFILE"
RESULT=${PIPESTATUS[0]}

ibmcloud login --apikey "${API_KEY}" -r "${IBM_REGION}"
save_logs "${LOGFILE}" "Unit Tests" ${RESULT}

exit ${RESULT}
