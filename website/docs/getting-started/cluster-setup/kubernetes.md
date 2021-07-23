---
title: Kubernetes
sidebar_position: 2
---

# Setup a minikube cluster

Please note that this documentation only helps setup a development environment.

This guide assumes that you have installed minikube on your system. 
If you do not have it installed, follow the official [minikube installation guide](https://minikube.sigs.k8s.io/docs/start/) to get started.

Next step is installing operators. To learn about what operators, see [Getting Started > Cluster Setup > Operators](operators.md). We will be using the [Service Binding Operator](https://operatorhub.io/operator/service-binding-operator) operator as an example for this guide.

### Installing OLM on the cluster
1. To install operators, we will first need to install OLM [(Operator Lifecycle Manager)](https://olm.operatorframework.io/) to the minikube cluster.
   Running the script below will take some time to install all the necessary namespaces and other resources.
    ```shell
    curl -sL https://github.com/operator-framework/operator-lifecycle-manager/releases/download/v0.18.1/install.sh | bash -s v0.18.1
    ```
  Note: We are not using the latest v0.18.2 because it is buggy.

2. Operators can be installed in a specific namespace or across the cluster(i.e. in all the namespaces). To install an operator, we need to make sure that the namespace contains `OperatorGroup` resource. Running the command below will create the necessary resource in `operators` namespace.
  ```shell
  kubectl create -f https://operatorhub.io/install/service-binding-operator.yaml
  ```
  If you want to access this resource from other namespaces as well, add your target namespace to `.spec.targetNamespaces` list.
  Wait for a few seconds for the operator to install. Confirm that it has been installed by running the below command:
  ```shell
  kubectl get csv -n operators
  ```