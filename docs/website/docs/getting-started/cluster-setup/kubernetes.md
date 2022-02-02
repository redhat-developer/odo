---
title: Kubernetes
sidebar_position: 1
---

# Setting up a Kubernetes cluster

## Introduction
This guide is helpful in setting up a development environment intended to be used with `odo`; this setup is not recommended for a production environment.

`odo` can be used with ANY Kubernetes cluster. However, this development environment will ensure complete coverage of all features of `odo`.

## Prerequisites
* You have a Kubernetes cluster set up (such as [minikube](https://minikube.sigs.k8s.io/docs/start/))
* You have admin privileges to the cluster

**Important notes:** `odo` will use the __default__ ingress and storage provisioning on your cluster. If they have not been set correctly, see our [troubleshooting guide](/docs/getting-started/cluster-setup/kubernetes#troubleshooting) for more details.

## Summary
* An Ingress controller in order to use `odo url create`
* Operator Lifecycle Manager in order to use `odo service create`
* (Optional) Service Binding Operator in order to use `odo link`

## Installing an Ingress controller

Creating an Ingress controller is required to use the `odo url create` feature.

This can be enabled by installing [an Ingress addon as per the Kubernetes documentation](https://kubernetes.io/docs/tasks/access-application-cluster/ingress-minikube/) such as: the built-in one on [minikube](https://minikube.sigs.k8s.io/) or [NGINX Ingress](https://kubernetes.github.io/ingress-nginx/).


**IMPORTANT:** `odo` cannot specify an Ingress controller and will use the *default* Ingress controller. 


If you are unable to access your components, check that your [default Ingress controller](https://kubernetes.github.io/ingress-nginx/#i-have-only-one-ingress-controller-in-my-cluster-what-should-i-do) has been set correctly.

### Minikube

To install an Ingress controller on a minikube cluster, enable the **ingress** addon with the following command:
```shell
minikube addons enable ingress
````

### NGINX Ingress

To enable the Ingress feature on a Kubernetes cluster _other than minikube_, we reccomend to use the [NGINX Ingress controller](https://kubernetes.github.io/ingress-nginx/deploy/).

On the default installation method, you will need to set NGINX Ingress as your [default Ingress controller](https://kubernetes.github.io/ingress-nginx/#i-have-only-one-ingress-controller-in-my-cluster-what-should-i-do), so `odo` may deploy URLs correctly.

### Other Ingress controllers

For a list of all available Ingress controllers see the [the Ingress controller documentation](https://kubernetes.io/docs/concepts/services-networking/ingress-controllers/).

To learn more about enabling this feature on your cluster, see the [Ingress prerequisites](https://kubernetes.io/docs/concepts/services-networking/ingress/#prerequisites) on the official kubernetes documentation.


## Installing the Operator Lifecycle Manager (OLM)

Installing the Operator Lifecycle Manager (OLM) is required to use the `odo service create` feature.

The [Operator Lifecycle Manager (OLM)](https://olm.operatorframework.io/) is an open source toolkit to manage Kubernetes native applications, called Operators, in a streamlined and scalable way.

`odo` utilizes Operators in order to create and link services to applications.

The following command will install OLM cluster-wide as well as create two new namespaces: `olm` and `operators`.

```shell
curl -sL https://github.com/operator-framework/operator-lifecycle-manager/releases/download/v0.20.0/install.sh | bash -s v0.20.0
```
Running the script will take some time to install all the necessary resources in the Kubernetes cluster including the `OperatorGroup` resource.

Note: Check the OLM [release page](https://github.com/operator-framework/operator-lifecycle-manager/releases/) for the latest release.


### Installing an Operator

Installing an Operator allows you to install a service such as Postgres, Redis or DataDog.

To install an operator from the OperatorHub website:
1. Visit the [OperatorHub](https://operatorhub.io) website.
2. Search for an Operator of your choice.
3. Navigate to its detail page.
4. Click on **Install**.
5. Follow the instruction in the installation popup. Please make sure to install the Operator in your desired namespace or cluster-wide, depending on your choice and the Operator capability.
6. [Verify the Operator installation](#verifying-the-operator-installation).

### Verifying the Operator installation

Once the Operator is successfully installed on the cluster, you can use `odo` to verify the Operator installation and see the CRDs associated with it; run the following command:

```shell
odo catalog list services
```

The output may look similar to:

```shell
odo catalog list services
Services available through Operators
NAME                                CRDs
datadog-operator.v0.6.0             DatadogAgent, DatadogMetric, DatadogMonitor
service-binding-operator.v0.9.1     ServiceBinding, ServiceBinding
```

If you do not see your installed Operator in the list, follow the [troubleshooting guide](#troubleshoot-the-operator-installation) to find the issue and debug it.

### Troubleshooting the Operator installation

There are two ways to confirm that the Operator has been installed properly.

The examples you may see in this guide use [Datadog Operator](https://operatorhub.io/operator/datadog-operator) and [Service Binding Operator](https://operatorhub.io/operator/service-binding-operator).

1. Verify that its pod started and is in “Running” state.
  ```shell
  kubectl get pods -n operators
  ```
The output may look similar to:
  ```shell
  kubectl get pods -n operators
  NAME                                       READY   STATUS    RESTARTS   AGE
  datadog-operator-manager-5db67c7f4-hgb59   1/1     Running   0          2m13s
  service-binding-operator-c8d7587b8-lxztx   1/1     Running   5          6d23h
  ```
2. Verify that the ClusterServiceVersion (csv) resource is in Succeeded or Installing phase.
  ```shell
  kubectl get csv -n operators
  ```
  The output may look similar to:
  ```shell
  kubectl get csv -n operators
  NAME                              DISPLAY                    VERSION   REPLACES                          PHASE
  datadog-operator.v0.6.0           Datadog Operator           0.6.0     datadog-operator.v0.5.0           Succeeded
  service-binding-operator.v0.9.1   Service Binding Operator   0.9.1     service-binding-operator.v0.9.0   Succeeded
  ```

  If you see the value under PHASE column to be anything other than _Installing_ or _Succeeded_, please take a look at the pods in `olm` namespace and ensure that the pod starting with name `operatorhubio-catalog` is in Running state:
  ```shell
  kubectl get pods -n olm
  NAME                                READY   STATUS             RESTARTS   AGE
  operatorhubio-catalog-x24dq         0/1     CrashLoopBackOff   6          9m40s
  ```
  If you see output like above where the pod is in CrashLoopBackOff state or any other state other than Running, delete the pod:
  ```shell
  kubectl delete pods/<operatorhubio-catalog-name> -n olm
  ```

### Checking to see if an Operator has been installed

For this example, we will check the [PostgreSQL Operator](https://operatorhub.io/operator/postgresql) installation.

Check `kubectl get csv` to see if your Operator exists:
```shell
$ kubectl get csv                         
NAME                      DISPLAY                           VERSION   REPLACES                  PHASE
postgresoperator.v5.0.3   Crunchy Postgres for Kubernetes   5.0.3     postgresoperator.v5.0.2   Succeeded
```

If the `PHASE` is something other than `Succeeded`, you won't see it in `odo catalog list services` output, and you won't be able to create a working Operator backed service out of it either. You will have to wait patiently until `PHASE` says `Suceeded`.


## (Optional) Installing the Service Binding Operator

`odo` uses [Service Binding Operator](https://operatorhub.io/operator/service-binding-operator) to provide the `odo link` feature which helps to connect an odo component to a service or another component.

The Service Binding Operator is _optional_ and is used to provide extra metadata support for `odo` deployments.

Operators can be installed in a specific namespace or across the cluster-wide.

```shell
kubectl create -f https://operatorhub.io/install/service-binding-operator.yaml
```
Running the command will create the necessary resource in the `operators` namespace.

If you want to access this resource from other namespaces as well, add your target namespace to `.spec.targetNamespaces` list in the `service-binding-operator.yaml` file before running `kubectl create`.

See [Verifying the Operator installation](#verifying-the-operator-installation) to ensure that the Operator was installed successfully.

## Troubleshooting

### Confirming your Ingress Controller functionality

`odo` will use the *default* Ingress Controller. By default, when you install an Ingress Controller such as [NGINX Ingress](https://kubernetes.github.io/ingress-nginx/), it will *not* be set as the default.

You must set it as the default Ingress Controller by modifying the annotation your IngressClass:
```sh
kubectl get IngressClass -A
kubectl edit IngressClass/YOUR-INGRESS -n YOUR-NAMESPACE
```

And add the following annotation:
```yaml
annotation:
  ingressclass.kubernetes.io/is-default-class: "true"
```

### Confirming your Storage Provisioning functionality

`odo` deploys with [Persistent Volume Claims](https://kubernetes.io/docs/concepts/storage/persistent-volumes/). By default, when you install a [StorageClass](https://kubernetes.io/docs/concepts/storage/storage-classes/) such as [GlusterFS](https://kubernetes.io/docs/concepts/storage/storage-classes/#glusterfs), it will *not* be set as the default.

You must set it as the default storage provisioner by modifying the annotation your StorageClass:
```sh
kubectl get StorageClass -A
kubectl edit StorageClass/YOUR-STORAGE-CLASS -n YOUR-NAMESPACE
```

And add the following annotation:
```yaml
annotation:
  storageclass.kubernetes.io/is-default-class: "true"
```
