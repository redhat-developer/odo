#!/bin/bash

LOGFILE="pr-${GIT_PR_NUMBER}-openshift-tests-${BUILD_NUMBER}"

source .ibm/pipelines/functions.sh

ibmcloud login --apikey "${API_KEY_QE}"
ibmcloud target -r eu-de
ibmcloud oc cluster config -c "${CLUSTER_ID}"
oc login -u apikey -p "${API_KEY_QE}" "${IBM_OPENSHIFT_ENDPOINT}"

cleanup_namespaces

(
    set -e
    export DEVFILE_PROXY="$(kubectl get svc -n devfile-proxy nginx -o jsonpath='{.status.loadBalancer.ingress[0].hostname}' || true)"
    echo Using Devfile proxy: ${DEVFILE_PROXY}
    make install
    make test-integration-cluster
    make test-e2e
) |& tee "/tmp/${LOGFILE}"

RESULT=${PIPESTATUS[0]}

save_logs "${LOGFILE}" "OpenShift Tests" ${RESULT}

exit ${RESULT}
