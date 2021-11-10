#!/bin/bash

LOGFILE="pr-${GIT_PR_NUMBER}-kubernetes-tests-${BUILD_NUMBER}"

source .ibm/pipelines/functions.sh

install_ibmcloud cloud-object-storage kubernetes-service

ibmcloud ks cluster config --cluster "${IBM_KUBERNETES_ID}" --admin

install_kubectl

(
    set -e
    make install
    make test-integration-devfile
    make test-e2e-devfile
    make test-cmd-project
) |& tee "/tmp/${LOGFILE}"
RESULT=${PIPESTATUS[0]}

install_gh
save_logs "${LOGFILE}" "Kubernetes Tests"

exit ${RESULT}
