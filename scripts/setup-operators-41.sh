#!/bin/bash
set -x

CI_OPERATOR_HUB_PROJECT=ci-operator-hub-project

install_mongo_operator() {
  # Subscription for mongo
  oc create -f -<<EOF
  apiVersion: operators.coreos.com/v1alpha1
  kind: Subscription
  metadata:
    labels:
      csc-owner-name: mongo-csc
      csc-owner-namespace: openshift-marketplace
    name: mongodb-enterprise
    namespace: openshift-operators
  spec:
    channel: stable
    installPlanApproval: Automatic
    name: mongodb-enterprise
    source: mongo-csc
    sourceNamespace: openshift-operators
EOF
}


install_etcd_operator(){
  oc create -f -<<EOF
  apiVersion: operators.coreos.com/v1alpha1
  kind: Subscription
  metadata:
    labels:
      csc-owner-name: etcd-csc
      csc-owner-namespace: openshift-marketplace
    name: etcd
    namespace: ${CI_OPERATOR_HUB_PROJECT}
  spec:
    channel: singlenamespace-alpha
    installPlanApproval: Automatic
    name: etcd
    source: etcd-csc
    sourceNamespace: ${CI_OPERATOR_HUB_PROJECT}
EOF
}

# First, install cluster-wide operator
# CatalogSourceConfig for mongodb
oc create -f -<<EOF
apiVersion: operators.coreos.com/v1
kind: CatalogSourceConfig
metadata:
  name: mongo-csc
  namespace: openshift-marketplace
spec:
  csDisplayName: Certified Operators
  csPublisher: Certified
  packages: mongodb-enterprise
  targetNamespace: openshift-operators
EOF

# install mongo operator
while :
do
    if oc get csv -n openshift-operators | grep mongo; then
        break
    else
        install_mongo_operator
        sleep 15
    fi
done

# Now onto namespace bound operator
# Create OperatorGroup
oc create -f -<<EOF
apiVersion: operators.coreos.com/v1
kind: OperatorGroup
metadata:
  generateName: ${CI_OPERATOR_HUB_PROJECT}-
  generation: 2
  namespace: ${CI_OPERATOR_HUB_PROJECT}
spec:
  serviceAccount:
    metadata:
      creationTimestamp: null
  targetNamespaces:
  - ${CI_OPERATOR_HUB_PROJECT}
EOF
### Create a CatalogSourceConfig for etcd operator
oc create -f -<<EOF
apiVersion: operators.coreos.com/v1
kind: CatalogSourceConfig
metadata:
  finalizers:
  - finalizer.catalogsourceconfigs.operators.coreos.com
  generation: 3
  name: etcd-csc
  namespace: openshift-marketplace
spec:
  csDisplayName: Community Operators
  csPublisher: Community
  packages: etcd
  targetNamespace: ${CI_OPERATOR_HUB_PROJECT}
EOF

# install etcd operator
while :
do
    if oc get csv -n ${CI_OPERATOR_HUB_PROJECT} | grep etcd; then
        break
    else
        install_etcd_operator
        sleep 15
    fi
done
