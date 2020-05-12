#!/usr/bin/env bash

# This script verifies that package trees
# conform to our import restrictions

FORBIDDEN=$(
    go list -f $'{{with $package := .ImportPath}}{{range $.Imports}}{{$package}} imports {{.}}\n{{end}}{{end}}' ./... |
    grep "k8s.io/kubernetes" |
    # the next imports need to disappear to be able to get rid of k/k dependency
    grep -v "k8s.io/kubernetes/pkg/credentialprovider" |
    grep -v "k8s.io/kubernetes/pkg/apis/rbac/v1" |
    # below imports will be automatically gone when kubectl is in staging
    grep -v "k8s.io/kubernetes/pkg/api/legacyscheme" |
    grep -v "k8s.io/kubernetes/pkg/kubectl"
)
if [ -n "${FORBIDDEN}" ]; then
    echo "Forbidden dependencies:"
    echo
    echo "${FORBIDDEN}" | sed 's/^/  /'
    echo
    exit 1
fi

TEST_FORBIDDEN=$(
    go list -f $'{{with $package := .ImportPath}}{{range $.TestImports}}{{$package}} imports {{.}}\n{{end}}{{end}}' ./... |
    grep "k8s.io/kubernetes" |
    # the next imports need to disappear to be able to get rid of k/k dependency
    grep -v "k8s.io/kubernetes/pkg/credentialprovider" |
    grep -v "k8s.io/kubernetes/pkg/apis/rbac/v1" |
    # below imports will be automatically gone when kubectl is in staging
    grep -v "k8s.io/kubernetes/pkg/api/legacyscheme" |
    grep -v "k8s.io/kubernetes/pkg/kubectl"
)
if [ -n "${TEST_FORBIDDEN}" ]; then
    echo "Forbidden dependencies in test code:"
    echo
    echo "${TEST_FORBIDDEN}" | sed 's/^/  /'
    echo
    exit 1
fi
