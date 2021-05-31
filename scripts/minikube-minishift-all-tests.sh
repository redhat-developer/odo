#!/usr/bin/env bash

shout() {
  set +x
  echo -e "\n.---------------------------------------\n${1}\n'---------------------------------------\n"
  set -x
}

set -ex

# This is one of the variables injected by ci-firewall. Its purpose is to allow scripts to handle uniqueness as needed
SCRIPT_IDENTITY=${SCRIPT_IDENTITY:-"def-id"}

case ${1} in
    minikube)
        # Integration tests
        shout "| Running integration Tests on MiniKube"
        make test-operator-hub
        # make test-cmd-project
        # make test-integration-devfile

        shout "Cleaning up some leftover namespaces"

        set +x
        for i in $(kubectl get namespace -o name); do
	        if [[ $i == "namespace/${SCRIPT_IDENTITY}"* ]]; then
	            kubectl delete $i
            fi
        done
        set -x

        odo logout
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
