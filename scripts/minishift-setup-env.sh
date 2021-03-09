#!/usr/bin/env bash

shout() {
  set +x
  echo -e "\n!!!!!!!!!!!!!!!!!!!!\n${1}\n!!!!!!!!!!!!!!!!!!!!\n"
  set -x
}

set -ex

shout "| Setting up environment..."

export PATH="$PATH:/usr/local/go/bin/"
export GOPATH=$HOME/go
mkdir -p $GOPATH/bin
export PATH="$PATH:$(pwd):$GOPATH/bin"

export ARTIFACTS_DIR="`pwd`/artifacts"
export CUSTOM_HOMEDIR=$ARTIFACT_DIR
# This si one of the variables injected by ci-firewall. Its purpose is to allow scripts to handle uniqueness as needed
SCRIPT_IDENTITY=${SCRIPT_IDENTITY:-"def-id"}

cd openshift/odo
make goget-ginkgo

executing "| Building ODO..."
make bin
sudo cp odo $GOPATH/bin

export MINISHIFT_ENABLE_EXPERIMENTAL=y 

#Check if minishift is running, OPenStack VM is long running
sh ./minishift-start-if-required.sh
