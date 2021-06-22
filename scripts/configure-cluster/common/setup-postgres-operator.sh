#!/bin/bash

install_postgres_operator(){
  oc create -f - <<EOF
  apiVersion: operators.coreos.com/v1
  kind: OperatorGroup
  metadata:
    generateName: ${1}-
    namespace: ${1}
  spec:
    targetNamespaces:
    - ${1}
EOF

  oc create -f - <<EOF
  apiVersion: operators.coreos.com/v1alpha1
  kind: Subscription
  metadata:
    name: postgresql-operator-dev4devs-com
    namespace: ${1}
  spec:
    channel: alpha
    name: postgresql-operator-dev4devs-com
    source: community-operators
    sourceNamespace: openshift-marketplace
    installPlanApproval: "Automatic"
EOF
}