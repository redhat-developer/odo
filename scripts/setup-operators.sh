#!/bin/bash
set -x

CI_OPERATOR_HUB_PROJECT=ci-operator-hub-project
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
