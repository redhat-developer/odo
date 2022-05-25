#!/bin/bash
set -x

install_postgres_operator(){
  $1 create -f - <<EOF
  apiVersion: operators.coreos.com/v1alpha1
  kind: Subscription
  metadata:
    name: my-cloud-native-postgresql
    namespace: $2
  spec:
    channel: stable
    name: cloud-native-postgresql
    source: $3
    sourceNamespace: $4
EOF
}

install_service_binding_operator() {
  $1 create -f - <<EOF
  apiVersion: operators.coreos.com/v1alpha1
  kind: Subscription
  metadata:
    name: my-service-binding-operator
    namespace: $2
  spec:
    channel: stable
    name: $3
    source: $4
    sourceNamespace: $5
EOF
}

if [ "$KUBERNETES" == "true" ]; then
  # install "cloud-native-postgresql" using "kubectl" in "operators" namespace; use "operatorhubio-catalog" catalog source from "olm" namespace
  install_postgres_operator kubectl operators operatorhubio-catalog olm

  # install "service-binding-operator" using "kubectl" in "operators" namespace; use "operatorhubio-catalog" catalog source from "olm" namespace
  install_service_binding_operator kubectl operators service-binding-operator operatorhubio-catalog olm
else
  # install "cloud-native-postgresql" using "oc" in "openshift-operators" namespace; use "certified-operators" catalog source from "openshift-marketplace" namespace
  install_postgres_operator oc openshift-operators certified-operators openshift-marketplace

  # install "rh-service-binding-operator" using "oc" in "openshift-operators" namespace; use "redhat-operators" catalog source from "openshift-marketplace" namespace
  install_service_binding_operator oc openshift-operators rh-service-binding-operator redhat-operators openshift-marketplace
fi
