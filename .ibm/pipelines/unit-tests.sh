#!/bin/bash

LOGFILE="pr-${GIT_PR_NUMBER}-unit-tests-${BUILD_NUMBER}"

source .ibm/pipelines/functions.sh

(
    set -e
    make goget-tools
    make validate
    make test
) |& tee "/tmp/$LOGFILE"
RESULT=${PIPESTATUS[0]}

install_ibmcloud cloud-object-storage
install_gh
save_logs "${LOGFILE}" "Unit Tests"

exit ${RESULT}
