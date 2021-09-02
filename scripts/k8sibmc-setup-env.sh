#!/usr/bin/env bash
#Sets up requirements to run tests in K8S cluster hosted in the IBM Cloud

shout() {
  set +x
  echo -e "\n.---------------------------------------\n${1}\n'---------------------------------------\n"
  set -x
}

# Create a bin directory whereever script runs. This will be where all binaries that need to be in PATH will reside.
export HOME="~/"
export GOPATH="~/go"
export GOBIN="$GOPATH/bin"
mkdir -p $GOBIN
# This is one of the variables injected by ci-firewall. Its purpose is to allow scripts to handle uniqueness as needed
SCRIPT_IDENTITY=${SCRIPT_IDENTITY:-"def-id"}

# Add GOBIN which is the bin dir we created earlier to PATH so any binaries there are automatically available in PATH
export PATH=$PATH:$GOBIN

# Prep for integration/e2e
shout "Building odo binaries"
make bin

# copy built odo to GOBIN
cp -avrf ./odo $GOBIN/

setup_kubeconfig() {
    # Login as admin to IBM Cloud and get kubeconfig file for K8S cluster
    ibmcloud login --apikey $IBMC_ADMIN_OCLOGIN_APIKEY -a cloud.ibm.com -r eu-de -g "Developer-CI-and-QE"
    ibmcloud ks cluster config --cluster $IBMC_K8S_CLUSTER_ID

    export KUBECONFIG=$HOME/.kube/config
    if [[ ! -f $KUBECONFIG ]]; then
        echo "Could not find kubeconfig file"
        exit 1
    fi
    if [[ ! -z $KUBECONFIG ]]; then
        # Copy kubeconfig to current directory, to avoid clashes with other test runs
        # Read and Write permission to current kubeconfig file
        cp $KUBECONFIG "`pwd`/config"
        chmod 640 "`pwd`/config"
        export KUBECONFIG="`pwd`/config"
    fi
}

case ${1} in
    k8s)
        setup_kubeconfig
        ;;
    *)
        echo "<<< Need parameter set to K8S >>>"
        exit 1
        ;;
esac

### Applies to both K8S and minikube
# Setup to find nessasary data from cluster setup
## Constants
SETUP_OPERATORS="./scripts/configure-cluster/common/setup-operators.sh"

# The OLM Version
export OLM_VERSION="v0.18.3"
# Enable OLM for running operator tests
curl -sL https://github.com/operator-framework/operator-lifecycle-manager/releases/download/$OLM_VERSION/install.sh | bash -s $OLM_VERSION

set +x
# Get kubectl cluster info
kubectl cluster-info
        
set -x
# Set kubernetes env var as true, to distinguish the platform inside the tests
export KUBERNETES=true

# Create Operators for Operator tests
sh $SETUP_OPERATORS
