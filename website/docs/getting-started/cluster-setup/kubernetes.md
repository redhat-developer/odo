---
title: Kubernetes
sidebar_position: 2
---

# Setup a minikube cluster

Please note that this documentation only helps setup a development environment.

This guide assumes that you have installed minikube on your system. 
If you do not have it installed, follow the official [minikube installation guide](https://minikube.sigs.k8s.io/docs/start/) to get started.

Next step is installing operators. To learn about what operators, see [Getting Started > Cluster Setup > Operators](operators.md). We will be using the [Redis Enterprise](https://operatorhub.io/operator/redis-enterprise) operator as an example for this guide.

### Installing operators on a minikube cluster
1. To install operators, we will first need to install OLM [(Operator Lifecycle Manager)](https://olm.operatorframework.io/) to the minikube cluster.
    ```shell
    minikube addons enable olm
    ```
   
  Operators can be installed in a specific namespace or across the cluster(i.e. in all the namespaces). To install an operator, we need to make sure of two things: 
    1. The namespace in which we are installing has an `OperatorGroup` resource.
    2. An operator installed in one namespace should also be accessible in all the other namespaces. 

  By default, enabling `olm` addon takes care of creating an `OperatorGroup` resource in its default `operators` namespace; check that they were created with the help of the command below:
  ```shell
  kubectl get og -n operators
  ```

  **Note**: Not all operators support installation in all the namespaces, some restrict access to a single namespace.

  If you do not see the `OperatorGroup` resource, you can manually create one with the help of the command below:
  ```shell
  kubectl create -f - << EOF
  apiVersion: operators.coreos.com/v1
  kind: OperatorGroup
  metadata:
    name: global-operators
    namespace: operators
  spec:
    targetNamespaces:
    - operators
  EOF
  ```
  If you want to access this resource from other namespaces as well, add your target namespace to `.spec.targetNamespaces` list.

2. Now we install the postgresql operator.
  ```shell
  kubectl create -f - << EOF
  apiVersion: operators.coreos.com/v1alpha1
  kind: Subscription
  metadata:
    name: my-redis-enterprise
    namespace: operators
  spec:
    channel: alpha
    name: redis-enterprise
    source: operatorhubio-catalog
    sourceNamespace: olm
  EOF
  ```
  Wait for a few seconds for the operator to install. Confirm that it has been installed by running the below command:
  ```shell
  kubectl get csv
  ```