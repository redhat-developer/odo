#!/usr/bin/env bash

shout() {
  set +x
  echo -e "\n!!!!!!!!!!!!!!!!!!!!\n${1}\n!!!!!!!!!!!!!!!!!!!!\n"
  set -x
}

set -ex

shout "Setting up some stuff"

# Create a bin directory whereever script runs. This will be where all binaries that need to be in PATH will reside.
mkdir bin artifacts
# Change the default location of go's bin directory (without affecting GOPATH). This is where compiled binaries will end up by default
# for eg go get ginkgo later on will produce ginkgo binary in GOBIN
export GOBIN="`pwd`/bin"
# Set kubeconfig to current dir. This ensures no clashes with other test runs
export KUBECONFIG="`pwd`/config"
export ARTIFACT_DIR="`pwd`/artifacts"
export CUSTOM_HOMEDIR=$ARTIFACT_DIR

# This si one of the variables injected by ci-firewall. Its purpose is to allow scripts to handle uniqueness as needed
SCRIPT_IDENTITY=${SCRIPT_IDENTITY:-"def-id"}

if [[ $BASE_OS == "windows" ]]; then
    shout "Setting GOBIN for windows"
    GOBIN="$(cygpath -pw $GOBIN)"
elif [[ $BASE_OS == "mac" ]]; then
    PATH="$PATH:/usr/local/bin:/usr/local/go/bin"                           #Path to `go` command as `/usr/local/go/bin:/usr/local/bin` is not included in $PATH while running test
fi
    
shout "Setting PATH"
# Add GOBIN which is the bin dir we created earlier to PATH so any binaries there are automatically available in PATH
export PATH=$PATH:$GOBIN

#-----------------------------------------------------------------------------