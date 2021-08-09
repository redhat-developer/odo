---
title: Kubernetes
sidebar_position: 1
---

# Setup a minikube cluster
Please note that this documentation is only useful in setting up a development environment, it is not recommended for a production environment.

This guide assumes that you have [installed minikube](https://minikube.sigs.k8s.io/docs/start/) on your system.

If you are using a Kubernetes cluster other than minikube, this guide assumes that you have admin privileges to the cluster and are logged in with the admin user; operator installation is only possible with an admin user.

**Agenda for this guide:**
* [Install the OLM](#install-the-olm)
* [Install the Service Binding Operator](#install-the-service-binding-operator)
* [Install an operator](#install-an-operator)
* [Verify the operator installation](#verify-the-operator-installation)

## Install the OLM
The Operator Lifecycle Manager(OLM) is a component of the Operator Framework, an open source toolkit to manage Kubernetes native applications, called Operators, in a streamlined and scalable way.[(Source)](https://olm.operatorframework.io/)

To install operators, we will first need to install OLM [(Operator Lifecycle Manager)](https://olm.operatorframework.io/) addon to the minikube cluster.
Running the script below will take some time to install all the necessary namespaces and other resources.
```shell
curl -sL https://github.com/operator-framework/operator-lifecycle-manager/releases/download/v0.18.3/install.sh | bash -s v0.18.3
```

To install OLM on a Kubernetes cluster setup other than minikube, please refer the [installation instructions on GitHub](https://github.com/operator-framework/operator-lifecycle-manager/#installation).

## Install the Service Binding Operator
odo uses [Service Binding Operator](https://operatorhub.io/operator/service-binding-operator) to provide the `odo link` feature which helps connect an odo component to a service or another component.

Operators can be installed in a specific namespace or across the cluster(i.e. in all the namespaces). To install an operator, we need to make sure that the namespace contains `OperatorGroup` resource. Running the command below will create the necessary resource in `operators` namespace.
```shell
kubectl create -f https://operatorhub.io/install/service-binding-operator.yaml
```

If you want to access this resource from other namespaces as well, add your target namespace to `.spec.targetNamespaces` list.

See [Verify the Operator installation](#verify-the-operator-installation) to ensure the operator is installed successfully.

## Install an operator
1. Visit the [OperatorHub](https://operatorhub.io) website.
2. Search for an operator of your choice.
3. Navigate to its detail page.
4. Click on `Install`.
5. Follow the instruction in the installation popup.
6. [Verify the operator was installed successfully](#verify-the-operator-installation).

## Verify the Operator installation
Wait for a few seconds for the operator to install.
To verify that the operator is installed successfully and see the CRDs associated with it, run the following command.
```shell
odo catalog list services
```
The output can look similar to:
```shell
$ odo catalog list services
Services available through Operators
NAME                                CRDs
datadog-operator.v0.6.0             DatadogAgent, DatadogMetric, DatadogMonitor
service-binding-operator.v0.9.1     ServiceBinding, ServiceBinding
```
If you do not see your installed operator in the list, follow the [troubleshooting guide](#troubleshoot-the-operator-installation) to find the issue and debug it.

## Troubleshoot the Operator installation
There are two ways to confirm that the operator has been installed properly.
The examples you may see in this guide uses [Datadog Operator](https://operatorhub.io/operator/datadog-operator) and [Service Binding Operator](https://operatorhub.io/operator/service-binding-operator).
1. Verify that its pod started and is in “Running” state.  
  ```s****************hell
  kubectl get pods -n operators
  ```
The output can look similar to:
  ```shell
  $ kubectl get csv -n operators 
  NAME                                       READY   STATUS    RESTARTS   AGE
  datadog-operator-manager-5db67c7f4-hgb59   1/1     Running   0          2m13s
  service-binding-operator-c8d7587b8-lxztx   1/1     Running   5          6d23h
  ```
2. Verify that the csv is in Succeeded or Installing phase.
  ```shell
  kubectl get csv -n operators
  ```
  The output can look similar to the following:
  ```shell
  $ kubectl get csv -n operators
  NAME                              DISPLAY                    VERSION   REPLACES                          PHASE
  datadog-operator.v0.6.0           Datadog Operator           0.6.0     datadog-operator.v0.5.0           Succeeded
  service-binding-operator.v0.9.1   Service Binding Operator   0.9.1     service-binding-operator.v0.9.0   Succeeded
  ```

  If you see the value under PHASE column to be anything other than _Installing_ or _Succeeded_, please take a look at the pods in `olm` namespace and ensure that the pod starting with name `operatorhubio-catalog` is in Running state:
  ```shell
  $ kubectl get pods -n olm
  NAME                                READY   STATUS             RESTARTS   AGE
  operatorhubio-catalog-x24dq         0/1     CrashLoopBackOff   6          9m40s
  ```
  If you see output like above where the pod is in CrashLoopBackOff state or any other state other than Running, delete the pod:
  ```shell
  kubectl delete pods -n olm <operatorhubio-catalog-name>
  ```