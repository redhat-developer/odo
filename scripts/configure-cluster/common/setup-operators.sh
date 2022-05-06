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
elif [ $(uname -m) == "s390x" ]; then
  # create "operator-ibm-catalog" CatalogSource for s390x
  oc apply -f https://raw.githubusercontent.com/redhat-developer/odo/main/docs/website/manifests/catalog-source-$(uname -m).yaml

  # install "cloud-native-postgresql" using "oc" in "openshift-operators" namespace; use "operator-ibm-catalog" catalog source from "openshift-marketplace" namespace
  install_postgres_operator oc openshift-operators operator-ibm-catalog openshift-marketplace

  # install "service-binding-operator" using "oc" in "openshift-operators" namespace; use "operator-ibm-catalog" catalog source from "openshift-marketplace" namespace
  install_service_binding_operator oc openshift-operators service-binding-operator operator-ibm-catalog openshift-marketplace
else
  # install "cloud-native-postgresql" using "oc" in "openshift-operators" namespace; use "certified-operators" catalog source from "openshift-marketplace" namespace
  install_postgres_operator oc openshift-operators certified-operators openshift-marketplace

  # install "rh-service-binding-operator" using "oc" in "openshift-operators" namespace; use "redhat-operators" catalog source from "openshift-marketplace" namespace
  install_service_binding_operator oc openshift-operators rh-service-binding-operator redhat-operators openshift-marketplace
fi
