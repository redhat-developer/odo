#!/bin/bash

# Script to run validate test on IBM Cloud 
# This script needs update if there is any change in steps of executing validate test

LOGFILE="pr-${GIT_PR_NUMBER}-Validate-tests-${BUILD_NUMBER}"

source .ibm/pipelines/functions.sh

(
    set -e
    make goget-tools
    make validate
) |& tee "/tmp/$LOGFILE"
RESULT=${PIPESTATUS[0]}

ibmcloud login --apikey "${API_KEY}" -r "${IBM_REGION}"
save_logs "${LOGFILE}" "Validate Tests" ${RESULT}

exit ${RESULT}
