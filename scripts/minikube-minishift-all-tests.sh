#!/usr/bin/env bash

shout() {
  set +x
  echo -e "\n.---------------------------------------\n${1}\n'---------------------------------------\n"
  set -x
}

set -ex

# Just checking the ${CI_INFO} value
echo ${CI_INFO}

case ${CI_INFO} in
    minikube)
        # Integration tests
        shout "| Running integration Tests on MiniKube"
        make test-cmd-project
        make test-integration-devfile
        ;;
    minishift)
        cd $HOME/openshift/odo
        eval $(minishift oc-env)

        shout "| Logging in to minishift..."
        oc login -u developer -p developer --insecure-skip-tls-verify $(minishift ip):8443

        shout "| Executing on minishift: generic, login, component command and plugin handler integration tests"
        make test-integration


        shout "| Executing on minishift: devfile catalog, create, push, watch, delete, registry, exec, test, env, status, config, debug and log command integration tests"
        make test-integration-devfile

        shout "| Executing on minishift: core beta, java, source e2e tests"
        make test-e2e-beta
        make test-e2e-java
        make test-e2e-source
        make test-e2e-images
        make test-e2e-devfile

        odo logout
        ;;
    *)
        echo "Need parameter set to minikube or minishift"
        exit 1
        ;;
esac
