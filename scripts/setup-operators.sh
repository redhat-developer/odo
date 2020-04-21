#!/bin/bash
set -x

CI_OPERATOR_HUB_PROJECT=ci-operator-hub-project

install_mongo_operator() {
  # First, enable a cluster-wide mongo operator
  oc create -f - <<EOF
  apiVersion: operators.coreos.com/v1alpha1
  kind: Subscription
  metadata:
    generation: 1
    name: mongodb-enterprise
    namespace: openshift-operators
  spec:
    channel: stable
    installPlanApproval: Automatic
    name: mongodb-enterprise
    source: certified-operators
    sourceNamespace: openshift-marketplace
EOF
}

install_etcd_operator(){
  # Create subscription
  oc create -f - <<EOF
  apiVersion: operators.coreos.com/v1alpha1
  kind: Subscription
  metadata:
    name: etcd
    namespace: ${OPERATOR_HUB_PROJECT}
  spec:
    channel: singlenamespace-alpha
    installPlanApproval: Automatic
    name: etcd
    source: community-operators
    sourceNamespace: openshift-marketplace
EOF
}

# install mongo operator
count=0
while [ "$count" -lt "5" ];
do
    if oc get csv -n openshift-operators | grep mongo; then
        break
    else
        install_mongo_operator
        count=`expr $count + 1`
        sleep 15
    fi
done

# Now onto namespace bound operator
# Create OperatorGroup
oc create -f - <<EOF
apiVersion: operators.coreos.com/v1
kind: OperatorGroup
metadata:
  generateName: ${CI_OPERATOR_HUB_PROJECT}-
  generation: 1
  namespace: ${CI_OPERATOR_HUB_PROJECT}
spec:
  targetNamespaces:
  - ${CI_OPERATOR_HUB_PROJECT}
EOF

# install etcd operator
count=0
while [ "$count" -lt "5" ];
do
    if oc get csv -n ${CI_OPERATOR_HUB_PROJECT} | grep etcd; then
        break
    else
        install_etcd_operator
        count=`expr $count + 1`
        sleep 15
    fi
done
