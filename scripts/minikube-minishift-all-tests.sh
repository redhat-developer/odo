#!/usr/bin/env bash

shout() {
    set +x
    echo -e "\n.---------------------------------------\n${1}\n'---------------------------------------\n"
    set -x
}

set -ex

# This is one of the variables injected by ci-firewall. Its purpose is to allow scripts to handle uniqueness as needed
SCRIPT_IDENTITY=${SCRIPT_IDENTITY:-"def-id"}

# Integration tests
shout "| Running integration Tests on MiniKube"
make test-cmd-project
make test-integration-devfile

shout "Cleaning up some leftover namespaces"

set +x
for i in $(kubectl get namespace -o name); do
    if [[ $i == "namespace/${SCRIPT_IDENTITY}"* ]]; then
        kubectl delete $i
    fi
done
set -x

odo logout
