#!/bin/bash
set -x

install_etcd_operator() {
    # Create subscription
    kubectl create -f - << EOF
    apiVersion: operators.coreos.com/v1alpha1
    kind: Subscription
    metadata:
        name: etcd
        namespace: operators
    spec:
        channel: clusterwide-alpha
        name: etcd
        source: operatorhubio-catalog
        sourceNamespace: olm
        startingCSV: etcdoperator.v0.9.4-clusterwide
        installPlanApproval: Automatic
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

# install etcd operator

install_etcd_operator

# install service-binding-operator

install_service_binding_operator
