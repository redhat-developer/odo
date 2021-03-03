#!/usr/bin/env bash

shout() {
   set +x
  echo -e "\n.---------------------------------------\n${1}\n'---------------------------------------\n"
   set -x
}

set -ex

eval $(minishift oc-env)

shout "| Logging in to minishift..."
oc login -u developer -p developer --insecure-skip-tls-verify $(minishift ip):8443

shout "| Executing: generic, login, component command and plugin handler integration tests"
make test-generic
make test-cmd-login-logout
make test-cmd-cmp
make test-plugin-handler

shout "| Executing: preference, config, component sub-commands and debug command integration tests"
make test-cmd-pref-config
make test-cmd-cmp-sub
make test-cmd-debug

shout "| Executing: service, link and component sub-commands command integration tests"
make test-cmd-service
make test-cmd-link-unlink-311-cluster

shout "| Executing: watch, storage, app, project, URL and push command integration tests"
make test-cmd-watch
make test-cmd-storage
make test-cmd-app
make test-cmd-push
make test-cmd-project
make test-cmd-url

shout "| Executing: devfile catalog, create, push, watch, delete, registry, exec, test, env, status, config, debug and log command integration tests"
make test-integration-devfile

shout "| Executing: core beta, java, source e2e tests"
make test-e2e-beta
make test-e2e-java
make test-e2e-source
make test-e2e-images
make test-e2e-devfile

odo logout
