#!/usr/bin/env bash

shout() {
  set +x
  echo -e "\n.---------------------------------------\n${1}\n'---------------------------------------\n"
  set -x
}

set -ex

# This is one of the variables injected by ci-firewall. Its purpose is to allow scripts to handle uniqueness as needed
SCRIPT_IDENTITY=${SCRIPT_IDENTITY:-"def-id"}

case ${1} in
    minikube)
        # Integration tests
        shout "| Running integration Tests on MiniKube"
        make test-operator-hub
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
        ;;
    k8s)
        ibmcloud login --apikey $IBMC_developer_APIKEY -a cloud.ibm.com -r eu-de -g "Developer-CI-and-QE"
        ibmcloud ks cluster config --cluster $IBMC_K8S_CLUSTER_ID

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
        ;;
    *)
        echo "Need parameter set to minikube or minishift"
        exit 1
        ;;
esac
