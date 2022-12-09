#!/bin/bash

LOGFILE="pr-${GIT_PR_NUMBER}-openshift-tests-${BUILD_NUMBER}"

source .ibm/pipelines/functions.sh

ibmcloud login --apikey "${API_KEY_QE}"
ibmcloud target -r eu-de
ibmcloud oc cluster config -c "${CLUSTER_ID}"

(
    set -e
    make install
    make test-integration-openshift-unauth
) |& tee "/tmp/${LOGFILE}"

RESULT=${PIPESTATUS[0]}

save_logs "${LOGFILE}" "OpenShift Unauthenticated Tests" ${RESULT}

exit ${RESULT}
