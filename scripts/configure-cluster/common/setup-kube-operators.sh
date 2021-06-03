#!/bin/bash
set -x

install_mongodb_enterprise_operator(){
  # Create subscription
  oc create -f - <<EOF
  apiVersion: operators.coreos.com/v1alpha1
  kind: Subscription
  metadata:
    name: my-mongodb-enterprise
    namespace: operators
  spec:
    channel: stable
    name: mongodb-enterprise
    source: operatorhubio-catalog
    sourceNamespace: olm
EOF
}

install_service_binding_operator(){
kubectl create -f - << EOF
    apiVersion: operators.coreos.com/v1alpha1 
    kind: Subscription 
    metadata: 
      name: my-service-binding-operator 
      namespace: operators 
    spec: 
      channel: beta 
      name: service-binding-operator 
      source: operatorhubio-catalog 
      sourceNamespace: olm
EOF
}

# install mongodb-enterprise operator

install_mongodb_enterprise_operator

# install service-binding-operator

install_service_binding_operator
