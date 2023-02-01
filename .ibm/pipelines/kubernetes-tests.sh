#!/bin/bash

LOGFILE="pr-${GIT_PR_NUMBER}-kubernetes-tests-${BUILD_NUMBER}"

source .ibm/pipelines/functions.sh

ibmcloud login --apikey "${API_KEY_QE}"
ibmcloud target -r "${IBM_REGION}"
ibmcloud ks cluster config --cluster "${IBM_KUBERNETES_ID}" --admin

cleanup_namespaces
export SKIP_USER_LOGIN_TESTS=true
(
    set -e
    export DEVFILE_PROXY="$(kubectl get svc -n devfile-proxy nginx -o jsonpath='{.status.loadBalancer.ingress[0].hostname}' || true)"
    echo Using Devfile proxy: ${DEVFILE_PROXY}
    make install
    RES=0
    make test-integration-cluster || RES=1
    make test-doc-automation || RES=1
    exit $RES
) |& tee "/tmp/${LOGFILE}"

RESULT=${PIPESTATUS[0]}

save_logs "${LOGFILE}" "Kubernetes Tests" ${RESULT}

exit ${RESULT}
