#!/usr/bin/env bash

shout() {
  set +x
  echo -e "\n.---------------------------------------\n${1}\n'---------------------------------------\n"
  set -x
}

# Create a bin directory whereever script runs. This will be where all binaries that need to be in PATH will reside.
mkdir bin artifacts
# Change the default location of go's bin directory (without affecting GOPATH). This is where compiled binaries will end up by default
# for eg go get ginkgo later on will produce ginkgo binary in GOBIN
export GOBIN="`pwd`/bin"

# Set kubeconfig to current dir. This ensures no clashes with other test runs
export KUBECONFIG="`pwd`/config"
export ARTIFACTS_DIR="`pwd`/artifacts"
export CUSTOM_HOMEDIR=$ARTIFACT_DIR

# This si one of the variables injected by ci-firewall. Its purpose is to allow scripts to handle uniqueness as needed
SCRIPT_IDENTITY=${SCRIPT_IDENTITY:-"def-id"}

# Add GOBIN which is the bin dir we created earlier to PATH so any binaries there are automatically available in PATH
export PATH=$PATH:$GOBIN

# Prep for integration/e2e
shout "Building odo binaries"
make bin

# copy built odo to GOBIN
cp -avrf ./odo $GOBIN/
shout "| Getting ginkgo"
make goget-ginkgo

#Workaround for https://github.com/openshift/odo/issues/4523 use env varibale CLUSTER instead of parameter
case $CLUSTER in
    minishift)
        export MINISHIFT_ENABLE_EXPERIMENTAL=y 
        export PATH="$PATH:/usr/local/go/bin/"
        export GOPATH=$HOME/go
        mkdir -p $GOPATH/bin
        export PATH="$PATH:$(pwd):$GOPATH/bin"
        curl -kJLO https://github.com/openshift/odo/blob/master/scripts/minishift-start-if-required.sh
        chmod +x minishift-start-if-required.sh
        sh ./minishift-start-if-required.sh
        ;;
    minikube)
        shout "| Start minikube"
        # Delete minikube instance, if in anycase already exists
        minikube delete
        minikube start --vm-driver=docker --container-runtime=docker
        set +x
        # Get kubectl cluster info
        kubectl cluster-info

        set -x
        # Set kubernetes env var as true, to distinguish the platform inside the tests
        export KUBERNETES=true
        ;;
    *)
        echo "<<< Need (parameter) CLUSTER env. variable set to minikube or minishift >>>"
        exit 1
        ;;
esac
