#!/usr/bin/env bash

shout() {
  set +x
  echo -e "\n!!!!!!!!!!!!!!!!!!!!\n${1}\n!!!!!!!!!!!!!!!!!!!!\n"
  set -x
}

set -ex

shout "Setting up some stuff"

sudo -i

# Set the default location of go's bin directory. This is where compiled binaries will end up by default
export GOPATH=$HOME/go

# Create a bin directory whereever script runs. This will be where all binaries that need to be in PATH will reside.
mkdir -p $GOPATH/bin
shout "getting ginkgo"
make goget-ginkgo
# Add GOPATH which is the bin dir we created earlier to PATH so any binaries there are automatically available in PATH
export PATH="$PATH:$(pwd):$GOPATH/bin"

shout "Start minikube"
# Delete minikube instance, if in anycase already exists
minikube delete
minikube start --vm-driver=none --container-runtime=docker
set +x
# Get kubectl cluster info
kubectl cluster-info

set -x
# Prep for integration tests
shout "Building odo binaries"
make bin
cp odo /usr/bin
export KUBERNETES=true
make test-cmd-project
make test-integration-devfile
