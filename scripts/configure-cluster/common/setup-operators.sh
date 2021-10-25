#!/bin/bash
set -x

install_redis_operator(){
  $1 create -f - <<EOF
  apiVersion: operators.coreos.com/v1alpha1
  kind: Subscription
  metadata:
    name: my-redis-operator
    namespace: $2
  spec:
    channel: stable
    name: redis-operator
    source: $3
    sourceNamespace: $4
EOF
}

if [ $KUBERNETES == "true" ]; then
  # install "redis-oprator" using "kubectl" in "operators" namespace; use "operatorhubio-catalog" catalog source from "olm" namespace
  install_redis_operator kubectl operators operatorhubio-catalog olm
elif [ $(uname -m) == "s390x" ]; then
  # create "operator-ibm-catalog" CatalogSource for s390x
  oc apply -f https://raw.githubusercontent.com/openshift/odo/main/website/manifests/catalog-source-$(uname -m).yaml
  # install "redis-oprator" using "oc" in "openshift-operators" namespace; use "operator-ibm-catalog" catalog source from "openshift-marketplace" namespace
  install_redis_operator oc openshift-operators operator-ibm-catalog openshift-marketplace
else
  # install "redis-oprator" using "oc" in "openshift-operators" namespace; use "community-operators" catalog source from "openshift-marketplace" namespace
  install_redis_operator oc openshift-operators community-operators openshift-marketplace
fi
