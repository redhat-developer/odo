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

cd openshift/odo
make goget-ginkgo

executing "| Building ODO..."
make bin
sudo cp odo /usr/bin

export MINISHIFT_ENABLE_EXPERIMENTAL=y 
