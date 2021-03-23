#!/usr/bin/env bash

shout() {
  set +x
  echo -e "\n.---------------------------------------\n${1}\n'---------------------------------------\n"
  set -x
}

set -ex

#Workaround for https://github.com/openshift/odo/issues/4523 use env varibale CLUSTER instead of parameter


eval $(minishift oc-env)

shout "| Logging in to minishift..."
oc login -u developer -p developer --insecure-skip-tls-verify $(minishift ip):8443

shout "| Executing test-integration"
make test-integration

shout "| Executing test-integration-devfile"
make test-integration-devfile

shout "| Executing e2e tests"
make test-e2e-beta
make test-e2e-java
make test-e2e-source
make test-e2e-images
make test-e2e-devfile

odo logout

