#!/usr/bin/env bash

shout() {
   set +x
  echo -e "\n.---------------------------------------\n${1}\n'---------------------------------------\n"
   set -x
}

set -ex

shout "| Cleaning up projects from cluster ..."

shout "| Logging in to minishift..."

oc login -u developer -p developer --insecure-skip-tls-verify $(minishift ip):8443
PROJECT_LIST=$(oc get projects)

for i in $PROJECT_LIST; do
    # delete existing ns, if any
    NAMESPACE=${i:0:10}
     if [[ $NAMESPACE != *"Active"* ]] && [[ $NAMESPACE != *"NAME"* ]] && [[ $NAMESPACE != *"STATUS"* ]] && [[ $NAMESPACE != *"DISPLAY"* ]] && [[ $NAMESPACE != *"My"* ]] && [[ $NAMESPACE != *"Project"* ]] && [[ $NAMESPACE != *"myproject"* ]]; then
       echo "Namespace is: $NAMESPACE"
       oc delete project $i
    fi
done
