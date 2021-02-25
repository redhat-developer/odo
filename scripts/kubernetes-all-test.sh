#!/usr/bin/env bash

shout() {
  set +x
  echo -e "\n!!!!!!!!!!!!!!!!!!!!\n${1}\n!!!!!!!!!!!!!!!!!!!!\n"
  set -x
}

set -ex

shout "Setting up some stuff"

# sudo -i

# Create a bin directory whereever script runs. This will be where all binaries that need to be in PATH will reside.
mkdir bin artifacts
# Change the default location of go's bin directory (without affecting GOPATH). This is where compiled binaries will end up by default
# for eg go get ginkgo later on will produce ginkgo binary in GOBIN
export GOBIN="`pwd`/bin"

# Set kubeconfig to current dir. This ensures no clashes with other test runs
export KUBECONFIG="`pwd`/config"
export ARTIFACTS_DIR="`pwd`/artifacts"
export CUSTOM_HOMEDIR=$ARTIFACT_DIR

# # Set the default location of go's bin directory. This is where compiled binaries will end up by default
# export GOPATH=$HOME/go

# # Create a bin directory whereever script runs. This will be where all binaries that need to be in PATH will reside.
# mkdir -p $GOPATH/bin
# shout "getting ginkgo"
# make goget-ginkgo
# # Add GOPATH which is the bin dir we created earlier to PATH so any binaries there are automatically available in PATH
# export PATH="$PATH:$(pwd):$GOPATH/bin"

# Add GOBIN which is the bin dir we created earlier to PATH so any binaries there are automatically available in PATH
export PATH=$PATH:$GOBIN

# Prep for integration/e2e
shout "Building odo binaries"
make bin

# copy built odo to GOBIN
cp -avrf ./odo $GOBIN/
shout "getting ginkgo"
make goget-ginkgo

shout "Start minikube"
# Delete minikube instance, if in anycase already exists
minikube delete
minikube start --vm-driver=docker --container-runtime=docker
set +x
# Get kubectl cluster info
kubectl cluster-info

set -x
# # Prep for integration tests
# shout "Building odo binaries"
# make bin
# cp odo /usr/bin
export KUBERNETES=true
make test-cmd-project
make test-integration-devfile
