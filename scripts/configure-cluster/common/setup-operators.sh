#!/bin/bash
set -x

install_mongodb_enterprise_operator(){
  # Create subscription
  oc create -f - <<EOF
  apiVersion: operators.coreos.com/v1alpha1
  kind: Subscription
  metadata: 
    labels: 
      operators.coreos.com/mongodb-enterprise-rhmp.openshift-operators: ""
    name: mongodb-enterprise-rhmp
    namespace: openshift-operators
  spec:
    channel: stable
    installPlanApproval: Automatic
    name: mongodb-enterprise-rhmp
    source: redhat-marketplace
    sourceNamespace: openshift-marketplace
    startingCSV: mongodb-enterprise.v1.10.0
EOF
}

install_service_binding_operator(){
  oc create -f - <<EOF
  apiVersion: operators.coreos.com/v1alpha1
  kind: Subscription
  metadata:
    labels:
      operators.coreos.com/rh-service-binding-operator.openshift-operators: ""
    name: rh-service-binding-operator
    namespace: openshift-operators
  spec:
    channel: beta
    installPlanApproval: Automatic
    name: rh-service-binding-operator
    source: redhat-operators
    sourceNamespace: openshift-marketplace
EOF
}


# install mongodb-enterprise operator

install_mongodb_enterprise_operator

# install service-binding-operator

install_service_binding_operator
