#!/usr/bin/env bash

shout() {
  set +x
  echo -e "\n!!!!!!!!!!!!!!!!!!!!\n${1}\n!!!!!!!!!!!!!!!!!!!!\n"
  set -x
}

set -ex

shout "Setting up"

mkdir bin
GOBIN="`pwd`/bin"
KUBECONFIG="`pwd`/config"

shout "Getting oc binary"
if [[ $BASE_OS == "linux"  ]]; then
    set +x
    curl -k ${OCP4X_DOWNLOAD_URL}/${ARCH}/${BASE_OS}/oc.tar -o ./oc.tar
    set -x
    tar -C $GOBIN -xvf ./oc.tar && rm -rf ./oc.tar
else
    set +x
    curl -k ${OCP4X_DOWNLOAD_URL}/${ARCH}/${BASE_OS}/oc.zip -o ./oc.zip
    set -x
    if [[ $BASE_OS == "windows" ]]; then
        GOBIN="$(cygpath -pw $GOBIN)"
        CURRDIR="$(cygpath -pw $WORKDIR)"
        powershell -Command "Expand-Archive -Path $CURRDIR\oc.zip  -DestinationPath $GOBIN"
        chmod +x $GOBIN/*
    fi
    if [[ $BASE_OS == "mac" ]]; then
        unzip ./oc.zip -d $GOBIN && rm -rf ./oc.zip && chmod +x $GOBIN/oc
        PATH="$PATH:/usr/local/bin:/usr/local/go/bin"
    fi
fi

PATH=$PATH:$GOBIN

#-----------------------------------------------------------------------------

shout "Testing"

# Run unit tests
GOFLAGS='-mod=vendor' make test

# Prep for int
shout "Building"
make bin
cp -avrf ./odo $GOBIN/
shout "getting ginkgo"
GOBIN="$GOBIN" make goget-ginkgo

# Integration tests
shout "Testing against 4x cluster"
    
shout "Logging into 4x cluster as developer (logs hidden)"
set +x
oc login -u developer -p password@123 --insecure-skip-tls-verify  ${OCP4X_API_URL}
set -x
    
shout "Running integration/e2e tests"
make test-e2e-all
