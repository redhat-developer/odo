#!/usr/bin/sh

# Script to run podman test on IBM Cloud
# This script needs update if there is any change in the podman test make target command


Shout() {
    echo "----------------$1----------------"
}

# Get all env var for tests

BUILD_NUMBER=$1
LOGFILE=$2
REPO=$3
GIT_PR_NUMBER=$4
TEST_EXEC_NODES=$5
Shout "Args Recived"

echo $BUILD_NUMBER $LOGFILE $REPO $GIT_PR_NUMBER $TEST_EXEC_NODES

(
    set -e
    Shout "Cloning Repo"
    git clone --depth 1 $REPO $BUILD_NUMBER
    cd $BUILD_NUMBER

    Shout "Checking out PR #$GIT_PR_NUMBER"
    git fetch --depth 1 origin pull/${GIT_PR_NUMBER}/head:pr${GIT_PR_NUMBER}
    git checkout pr${GIT_PR_NUMBER}

    Shout "Setup ENV variables"
    mkdir bin
    mkdir artifacts

    #set env var
    GOCACHE=$(pwd)/.gocache
    GOBIN=$(pwd)/bin
    export PATH=$GOBIN:$PATH

    Shout "Create Binary"
    make install
    
    Shout "Running test"
    make test-integration-podman

) |& tee "/tmp/${LOGFILE}"
