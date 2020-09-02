#!/bin/bash
set -x

CI_OPERATOR_HUB_PROJECT=ci-operator-hub-project

install_mongo_operator() {
  # First, enable a cluster-wide mongo operator
  oc create -f - <<EOF
  apiVersion: operators.coreos.com/v1alpha1
  kind: Subscription
  metadata:
    name: percona-server-mongodb-operator-certified
    namespace: openshift-operators
  spec:
    channel: stable
    installPlanApproval: Automatic
    name: percona-server-mongodb-operator-certified
    source: certified-operators
    sourceNamespace: openshift-marketplace
    startingCSV: percona-server-mongodb-operator.v1.4.0
EOF
}

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

# install etcd operator
count=0
while [ "$count" -lt "5" ];
do
    if oc get csv -n openshift-operators | grep etcd; then
        break
    else
        install_etcd_operator
        count=`expr $count + 1`
        sleep 15
    fi
done

# install service-binding-operator
count=0
while [ "$count" -lt "5" ];
do
    if oc get csv -n openshift-operators | grep service-binding-operator; then
        break
    else
        install_service_binding_operator
        count=`expr $count + 1`
        sleep 15
    fi
done
