#!/bin/bash

SCRIPT_ROOT=$(dirname ${BASH_SOURCE})/..
CODEGEN_PKG=${CODEGEN_PKG:-$(cd ${SCRIPT_ROOT}; ls -d -1 ./vendor/k8s.io/code-generator 2>/dev/null || echo ../../../k8s.io/code-generator)}

verify="${VERIFY:-}"

bash ${CODEGEN_PKG}/generate-groups.sh "deepcopy" \
	github.com/openshift/custom-resource-status/generated \
	github.com/openshift/custom-resource-status \
	"conditions:v1" \
	"objectreferences:v1" \
	--go-header-file ${SCRIPT_ROOT}/tools/empty.txt \
	${verify}
