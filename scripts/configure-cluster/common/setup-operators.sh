#!/bin/bash
set -x

export SBO_SOURCE="redhat-operators"
export SBO_NAME="rh-service-binding-operator"

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

install_service_binding_operator(){
  $1 create -f - <<EOF
  apiVersion: operators.coreos.com/v1alpha1
  kind: Subscription
  metadata:
    name: my-service-binding-operator
    namespace: $2
  spec:
    channel: beta
    name: $3
    source: $4
    sourceNamespace: $5
EOF
}

deploy_service_binding_operator_master(){
  oc create -f - <<EOF
  apiVersion: operators.coreos.com/v1alpha1
  kind: CatalogSource
  metadata:
    name: service-binding-master
    namespace: openshift-marketplace
  spec:
    displayName: Service Binding Operator build from master
    image: quay.io/redhat-developer/servicebinding-operator:index
    priority: 500
    publisher: Red Hat
    sourceType: grpc
    updateStrategy:
      registryPoll:
        interval: 10m0s
EOF
}

if [ $KUBERNETES == "true" ]; then
  # install "redis-oprator" using "kubectl" in "operators" namespace; use "operatorhubio-catalog" catalog source from "olm" namespace
  install_redis_operator kubectl operators operatorhubio-catalog olm
else
  if [$NIGHTLY == "true"]; then
    # Deploy SBO master catalog source on OCP Nightly test run
    deploy_service_binding_operator_master

    SBO_SOURCE="service-binding-master"
    SBO_NAME="service-binding-operator"
  fi

  # install "redis-oprator" using "oc" in "openshift-operators" namespace; use "community-operators" catalog source from "openshift-marketplace" namespace
  install_redis_operator oc openshift-operators community-operators openshift-marketplace

  # install "service-binding-operator" using "oc" in "openshift-operators" namespace; use SBO_SOURCE env var catalog source from "openshift-marketplace" namespace
  install_service_binding_operator oc openshift-operators $SBO_NAME $SBO_SOURCE openshift-marketplace
fi