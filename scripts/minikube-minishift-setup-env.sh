#!/usr/bin/env bash

shout() {
  set +x
  echo -e "\n.---------------------------------------\n${1}\n'---------------------------------------\n"
  set -x
}

# Create a bin directory whereever script runs. This will be where all binaries that need to be in PATH will reside.
export HOME=`pwd`/home
export GOPATH="`pwd`/home/go"
export GOBIN="$GOPATH/bin"
mkdir -p $GOBIN
# This si one of the variables injected by ci-firewall. Its purpose is to allow scripts to handle uniqueness as needed
SCRIPT_IDENTITY=${SCRIPT_IDENTITY:-"def-id"}

# Add GOBIN which is the bin dir we created earlier to PATH so any binaries there are automatically available in PATH
export PATH=$PATH:$GOBIN

# Prep for integration/e2e
shout "Building odo binaries"
make bin

# copy built odo to GOBIN
cp -avrf ./odo $GOBIN/

setup_kubeconfig() {
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
    minishift)
        export MINISHIFT_ENABLE_EXPERIMENTAL=y 
        export PATH="$PATH:/usr/local/go/bin/"
        sh .scripts/minishift-start-if-required.sh
        ;;
    minikube)
        mkStatus=$(minikube status)
        shout "| Checking if Minikube needs to be started..."
        if [[ "$mkStatus" == *"host: Running"* ]] && [[ "$mkStatus" == *"kubelet: Running"* ]]; then 
            if [[ "$mkStatus" == *"kubeconfig: Misconfigured"* ]]; then
                minikube update-context
            fi
            setup_kubeconfig
            kubectl config use-context minikube
        else
            minikube delete
            shout "| Start minikube"
            minikube start --vm-driver=docker --container-runtime=docker
            setup_kubeconfig
        fi
        
        minikube version
        # Setup to find nessasary data from cluster setup
        ## Constants
        SETUP_OPERATORS="./scripts/configure-cluster/common/setup-kube-operators.sh"

        # Enable OLM for running operator tests
        curl -sL https://github.com/operator-framework/operator-lifecycle-manager/releases/download/v0.17.0/install.sh | bash -s v0.17.0

        set +x
        # Get kubectl cluster info
        kubectl cluster-info

        set -x
        # Set kubernetes env var as true, to distinguish the platform inside the tests
        export KUBERNETES=true

        # Create Operators for Operator tests
        sh $SETUP_OPERATORS
        ;;
    *)
        echo "<<< Need parameter set to minikube or minishift >>>"
        exit 1
        ;;
esac
