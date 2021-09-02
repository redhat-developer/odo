#!/usr/bin/env bash
# Runs integration tests on K8S cluster hosted in IBM Cloud

shout() {
  set +x
  echo -e "\n.---------------------------------------\n${1}\n'---------------------------------------\n"
  set -x
}

set -ex

# This is one of the variables injected by ci-firewall. Its purpose is to allow scripts to handle uniqueness as needed
SCRIPT_IDENTITY=${SCRIPT_IDENTITY:-"def-id"}

case ${1} in
    k8s)
        ibmcloud login --apikey $IBMC_DEVELOPER_OCLOGIN_APIKEY -a cloud.ibm.com -r eu-de -g "Developer-CI-and-QE"
        ibmcloud ks cluster config --cluster $IBMC_K8S_CLUSTER_ID

        # Integration tests
        shout "| Running integration Tests on Kubernetes cluster in IBM Cloud"
        make test-cmd-project

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
        echo "Need parameter set to k8s"
        exit 1
        ;;
esac
