#!/bin/bash
set -x

install_etcd_operator(){
  # Create subscription
  oc create -f - <<EOF
  apiVersion: operators.coreos.com/v1alpha1
  kind: Subscription
  metadata:
    name: etcd
    namespace: openshift-operators
  spec:
    channel: clusterwide-alpha
    installPlanApproval: Automatic
    name: etcd
    source: community-operators
    sourceNamespace: openshift-marketplace
    startingCSV: etcdoperator.v0.9.4-clusterwide
EOF
}

install_service_binding_operator(){
  oc create -f - <<EOF
  apiVersion: operators.coreos.com/v1alpha1
  kind: Subscription
  metadata:
    name: service-binding-operator
    namespace: openshift-operators
  spec:
    channel: alpha
    installPlanApproval: Automatic
    name: service-binding-operator
    source: community-operators
    sourceNamespace: openshift-marketplace
    startingCSV: service-binding-operator.v0.1.1-364
EOF
}

# install etcd operator

install_etcd_operator

# install service-binding-operator

install_service_binding_operator
