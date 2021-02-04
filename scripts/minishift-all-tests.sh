#!/usr/bin/env bash

executing() {
   set +x
  echo -e "\n------------------------------\n${1}\n------------------------------\n"
   set -x
}

set -ex

#Export GitHub token to avoid 
executing "Setting up environment..."

export PATH="$PATH:/usr/local/go/bin/"
export GOPATH=$HOME/go
git clone https://github.com/openshift/odo.git openshift/odo
cd openshift/odo

mkdir -p $GOPATH/bin
make goget-ginkgo
export PATH="$PATH:$(pwd):$GOPATH/bin"

executing "Building ODO..."
make bin
sudo cp odo /usr/bin

executing "Stopping minishift..."
minishift stop
yes | minishift delete
MINISHIFT_ENABLE_EXPERIMENTAL=y 

executing "Starting minishift..."
minishift start 
executing "Adding components: service-catalog, automation-service-broker, and template-service-broker ..."
minishift openshift component add service-catalog
minishift openshift component add automation-service-broker
minishift openshift component add template-service-broker
sleep 3m
eval $(minishift oc-env)

executing "Logging in to minishift..."
oc login -u developer -p developer --insecure-skip-tls-verify $(minishift ip):8443

executing "Executing tests..."
make test-cmd-project
make test-cmd-service

executing "Removing cloned repo..."
rm -Rf openshift
