#!/bin/bash

set -e

ibmcloud login --apikey "${IBM_API_KEY}" -r "${IBM_REGION}"
ibmcloud ks cluster config --cluster "${IBM_KUBERNETES_ID}" --admin

export DEVFILE_PROXY="$(kubectl get svc -n devfile-proxy nginx -o jsonpath='{.status.loadBalancer.ingress[0].hostname}' || true)"
echo Using Devfile proxy: ${DEVFILE_PROXY}

PATH=${PATH}:${GOPATH}/bin
make install

export KUBERNETES=true
export TEST_EXEC_NODES=24
make test-integration
