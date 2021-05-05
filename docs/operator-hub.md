---
# Page settings
layout: default
keywords:
comments: false

# Hero section
title: Introduction to Operators
description: Deploying an Operator from Operator Hub using odo.

# Micro navigation
micro_nav: true

---
# Creating and linking with Operator backed services

In this document we will go through what Operators are, how to create and delete services from the installed Operators, and how to link an odo component with an Operator backed service. We do this walking through an example Node.js application and linking it with an etcd key-value store.

> **Note**
> 
> We will be updating our documentation with more examples of linking components to different kinds of Operator backed services in future.

## Introduction to Operators

What is an Operator?

An Operator is essentially a [custom controller](https://www.openshift.com/learn/topics/operators). It is a method of packaging, deploying and managing a Kubernetes-native application.

With Operators, odo allows you to create a service as defined by a Custom Resource Definition (CRD).

odo utilizes Operators and [Operator Hub](https://operatorhub.io/) in order to provide a seamless method for custom controller service installation.

> **Warning**
> 
> You cannot install Operators with odo on your cluster.
> 
> To install Operators on a Kubernetes cluster, contact your cluster administrator or see the [Kubernetes documentation](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/).
> 
> To install Operators on an OpenShift cluster, contact your cluster administrator or see the [OpenShift documentation](https://docs.openshift.com/container-platform/4.6/operators/admin/olm-adding-operators-to-cluster.html).

## Deploying your first Operator

### Prerequisites

  - You must have cluster permissions to install an Operator on either [OpenShift](https://docs.openshift.com/container-platform/latest/operators/olm-adding-operators-to-cluster.html) or [Kubernetes](https://github.com/operator-framework/operator-lifecycle-manager/blob/master/doc/install/install.md). If you’re running a [minikube](https://minikube.sigs.k8s.io/docs/) cluster, you can refer to [this guide](https://odo.dev/docs/operators-on-minikube) to install Operators required to run example mentioned in this document.

## Creating a project

Create a project to keep your source code, tests, and libraries organized in a separate single unit.

1.  Log in to your cluster:
    
    ``` sh
    $ odo login -u developer -p developer
    ```

2.  Create a project:
    
    ``` sh
    $ odo project create myproject
     ✓  Project 'myproject' is ready for use
     ✓  New project created and now using project : myproject
    ```

## Installing an Operator

In our examples, we install [etcd](https://etcd.io/), a distributed key-value store from [Operator Hub](https://operatorhub.io/operator/etcd).

> **Important**
> 
> Each Operator we install refers to the built-in `metadata.annotations.alm-examples` annotation in order to correctly deploy. If the Operator does not contain the correct metadata, you will not be able to correctly deploy. For more information, see the the [upstream CRD documentation](https://github.com/operator-framework/operator-lifecycle-manager/blob/master/doc/design/building-your-csv.md#crd-templates).

### Kubernetes installation

For Kubernetes installation, you must need to install the Operator Lifecycle Manager and etcd Operator from the [etcd installation guide on Operator Hub](https://operatorhub.io/operator/etcd). You can refer to [this document](https://odo.dev/docs/operators-on-minikube) for steps setup Operator Lifecycle Manager and etcd Operator on a [minikube](https://minikube.sigs.k8s.io/) cluster.

### OpenShift installation

For OpenShift installation, the etcd Operator can be installed through the [administrative console](https://docs.openshift.com/container-platform/latest/operators/olm-adding-operators-to-cluster.html).

### Listing all available Operators

Before deploying your first Operator, have a look at what is available:

``` sh
$ odo catalog list services
Operators available in the cluster
NAME                          CRDs
etcdoperator.v0.9.4           EtcdCluster, EtcdBackup, EtcdRestore
```

In above output, `etcdoperator.v0.9.4` is the Operator while `EtcdCluster`, `EtcdBackup` and `EtcdRestore` are the CRDs provided by this Operator.

## Creating an Operator backed service

In this example, we will be deploying `EtcdCluster` service from [etcd Operator](https://operatorhub.io/operator/etcd) to an OpenShift / Kubernetes cluster. This service is provided by the Operator `etcdoperator`. Please ensure that this Operator is installed on your OpenShift / Kubernetes cluster before trying to create `EtcdCluster` service from it. If it’s not installed, please install it by logging into your OpenShift / Kubernetes cluster as `kube:admin` user.

1.  Create an `EtcdCluster` service from the `etcdoperator.v0.9.4` Operator:
    
    ``` sh
    $ odo service create etcdoperator.v0.9.4/EtcdCluster
    ```

2.  Confirm the Operator backed service was deployed:
    
    ``` sh
    $ odo service list
    ```

It is important to note that `EtcdBackup` and `EtcdRestore` cannot be deployed the same way as we deployed `EtcdCluster` as they require configuring other parameters in their YAML definition.

## Deploying Operator backed service to a cluster via YAML

In this example, we will be deploying our [installed etcd Operator](https://operatorhub.io/operator/etcd) to an OpenShift / Kubernetes cluster.

However, we will be using the YAML definition where we modify the `metadata.name` and `spec.size`.

> **Important**
> 
> Deploying via YAML is a **temporary** feature as we add support for [passing parameters on the command line](https://github.com/openshift/odo/issues/2785) and [interactive mode](https://github.com/openshift/odo/issues/2799).

1.  Retrieve the YAML output of the operator:
    
    ``` shell
    $ odo service create etcdoperator.v0.9.4/EtcdCluster --dry-run > etcd.yaml
    ```

2.  Modify the YAML file by redefining the name and size:
    
    ``` yaml
    apiVersion: etcd.database.coreos.com/v1beta2
    kind: EtcdCluster
    metadata:
      name: my-etcd-cluster // Change the name
    spec:
      size: 1 // Reduce the size from 3 to 1
      version: 3.2.13
    ```

3.  Create the service from the YAML file:
    
    ``` shell
    $ odo service create --from-file etcd.yaml
    ```

4.  Confirm that the service has been created:
    
    ``` shell
    $ odo service list
    ```

## Linking an odo component with an Operator backed service

Linking a component to a service means, in simplest terms, to make a service usable from the component. odo uses [Service Binding Operator](https://github.com/redhat-developer/service-binding-operator/) to provide the linking feature. Please refer to [this document](https://odo.dev/docs/install-service-binding-operator.adoc) to install it on OpenShift or Kubernetes.

For example, once you link an EtcdCluster service with your Node.js application, you can use (or, interact with) the EtcdCluster from within your node app. The way odo facilitates linking is by making sure that specific environment variables from the pod in which the service is running are configured in the pod of the component as well.

After having created a service using either of the two approaches discussed above, we can now connect an odo component with the service thus created.

1.  Make sure you are executing the command for a component that’s pushed (`odo push`) to the cluster.

2.  Link the component with the service:
    
    ``` shell
    $ odo service list
    NAME                    AGE
    EtcdCluster/example     46m2s
    
    $ odo link EtcdCluster/example
     ✓  Successfully created link between component "node-todo" and service "EtcdCluster/example"
    
    To apply the link, please use `odo push`
    
    $ odo push
    ```

> **Important**
> 
> For the link between a component and Operator Hub backed service to take effect, make sure you do `odo push`. The link won’t be effective otherwise.

## Unlinking an odo component from an Operator backed service

Unlinking unsets the environment variables that were set by linking. This would cause your application to cease being able to communicate with the service linked using `odo link`.

> **Important**
> 
> `odo unlink` doesn’t work on a cluster other than OpenShift (that is, minikube, or vanilla Kubernetes, etc.) because Service Binding Operator cannot be setup the OLM way (that is, we cannot list it by doing `odo catalog list services` or `kubectl get csv` like we can do for etcd Operator in this document). We are [working making this possible](https://github.com/redhat-developer/service-binding-operator/issues/623).

1.  Make sure you are executing the command for a component that’s pushed (`odo push`) to the cluster.

2.  Unlink the component from the service it is connected to:
    
    ``` shell
    $ odo unlink EtcdCluster/example
    ✓  Successfully unlinked component "node-todo" from service "EtcdCluster/example"
    
    To apply the changes, please use `odo push`
    
    $ odo push
    ```

> **Important**
> 
> For unlinking to take effect, make sure you do `odo push`. It won’t be effective otherwise.

## Deleting an Operator backed service

To delete an Operator backed service, provide full name of the service that you see in the output of `odo service list`. For example:

``` shell
$ odo service list
NAME                    AGE
EtcdCluster/example     2s

$ odo service delete EtcdCluster/example
```

To forcefully delete a service without being prompted for confirmation, use the `-f` flag like below:

``` shell
$ odo service delete EtcdCluster/example -f
```
