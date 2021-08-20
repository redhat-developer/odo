---
title: Kubernetes
sidebar_position: 1
---

# Setting up a Kubernetes cluster
*Note that this guide is only helpful in setting up a development environment; this setup is not recommended for a production environment.*

## Prerequisites
* You have a Kubernetes cluster setup, this could for example be a [minikube](https://minikube.sigs.k8s.io/docs/start/) cluster.
* You have admin privileges to the cluster, since the Operator installation is only possible with an admin user.

## Enabling Ingress
To access an application externally, you will create _URLs_ using odo, which are implemented on a Kubernetes cluster by Ingress resources; installing an Ingress controller helps in using this feature on a Kubernetes cluster.

**Minikube:** To install an Ingress controller on a minikube cluster, enable the **ingress** addon with the following command:
```shell
minikube addons enable ingress
````
To learn more about ingress addon, see [the documentation on Kubernetes website](https://kubernetes.io/docs/tasks/access-application-cluster/ingress-minikube/).

**Other Kubernetes Cluster**: To enable the Ingress feature on a Kubernetes cluster _other than minikube_, using the NGINX Ingress controller see [the official NGINX Ingress controller installation documentation](https://kubernetes.github.io/ingress-nginx/deploy/).

To use a different controller, see [the Ingress controller documentation](https://kubernetes.io/docs/concepts/services-networking/ingress-controllers/).

To learn more about enabling this feature on your cluster, see the [Ingress prerequisites](https://kubernetes.io/docs/concepts/services-networking/ingress/#prerequisites) on the official kubernetes documentation.

## Installing the Operator Lifecycle Manager (OLM)
The Operator Lifecycle Manager(OLM) is a component of the Operator Framework, an open source toolkit to manage Kubernetes native applications, called Operators, in a streamlined and scalable way. [(Source)](https://olm.operatorframework.io/)

[//]: # (Move this section to Architecture > Service Binding or create a new Operators doc)
What are Operators?
>The Operator pattern aims to capture the key aim of a human operator who is managing a service or set of services. Human operators who look after specific applications and services have deep knowledge of how the system ought to behave, how to deploy it, and how to react if there are problems.
>
>People who run workloads on Kubernetes often like to use automation to take care of repeatable tasks. The Operator pattern captures how you can write code to automate a task beyond what Kubernetes itself provides.
> [(Source)](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/#motivation)

[//]: # (Move until here)

To install an Operator, we will first need to install OLM [(Operator Lifecycle Manager)](https://olm.operatorframework.io/) on the cluster.
```shell
curl -sL https://github.com/operator-framework/operator-lifecycle-manager/releases/download/v0.18.3/install.sh | bash -s v0.18.3
```
Running the script will take some time to install all the necessary resources in the Kubernetes cluster including the `OperatorGroup` resource.

Note: Check the OLM [release page](https://github.com/operator-framework/operator-lifecycle-manager/releases/) to use the latest version.

## Installing the Service Binding Operator
odo uses [Service Binding Operator](https://operatorhub.io/operator/service-binding-operator) to provide the `odo link` feature which helps to connect an odo component to a service or another component.

Operators can be installed in a specific namespace or across the cluster(i.e. in all the namespaces).
```shell
kubectl create -f https://operatorhub.io/install/service-binding-operator.yaml
```
Running the command will create the necessary resource in the `operators` namespace.

If you want to access this resource from other namespaces as well, add your target namespace to `.spec.targetNamespaces` list in the `service-binding-operator.yaml` file before running `kubectl create`.

See [Verifying the Operator installation](#verifying-the-operator-installation) to ensure that the Operator was installed successfully.

## Installing an Operator
To install an operator from the OperatorHub website:
1. Visit the [OperatorHub](https://operatorhub.io) website.
2. Search for an Operator of your choice.
3. Navigate to its detail page.
4. Click on **Install**.
5. Follow the instruction in the installation popup. Please make sure to install the Operator in your desired namespace or cluster-wide, depending on your choice and the Operator capability.
6. [Verify the Operator installation](#verifying-the-operator-installation).

## Verifying the Operator installation
Wait for a few seconds for the Operator to install.

Once the Operator is successfully installed on the cluster, you can use `odo` to verify the Operator installation and see the CRDs associated with it; run the following command:
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
If you do not see your installed Operator in the list, follow the [troubleshooting guide](#troubleshoot-the-operator-installation) to find the issue and debug it.

## Troubleshooting the Operator installation
There are two ways to confirm that the Operator has been installed properly.
The examples you may see in this guide use [Datadog Operator](https://operatorhub.io/operator/datadog-operator) and [Service Binding Operator](https://operatorhub.io/operator/service-binding-operator).
1. Verify that its pod started and is in “Running” state.
  ```shell
  kubectl get pods -n operators
  ```
The output can look similar to:
  ```shell
  $ kubectl get pods -n operators
  NAME                                       READY   STATUS    RESTARTS   AGE
  datadog-operator-manager-5db67c7f4-hgb59   1/1     Running   0          2m13s
  service-binding-operator-c8d7587b8-lxztx   1/1     Running   5          6d23h
  ```
2. Verify that the ClusterServiceVersion (csv) resource is in Succeeded or Installing phase.
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
  kubectl delete pods/<operatorhubio-catalog-name> -n olm
  ```
