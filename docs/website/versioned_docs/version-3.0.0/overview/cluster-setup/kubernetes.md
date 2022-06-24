---
title: Kubernetes
sidebar_position: 1
---

# Setting up a Kubernetes cluster

## Introduction

This guide is helpful in setting up a development environment intended to be used with `odo`; this setup is not recommended for a production environment.

`odo` can be used with ANY Kubernetes cluster. However, this development environment will ensure complete coverage of all features of `odo`.

## Requirements

* You have a Kubernetes cluster set up (such as [minikube](https://minikube.sigs.k8s.io/docs/start/))
* You have admin privileges to the cluster

**Important notes:** `odo` will use the __default__  storage provisioning on your cluster. If it has not been set correctly, see our [troubleshooting guide](/docs/getting-started/cluster-setup/kubernetes#troubleshooting) for more details.

## (OPTIONAL) Installing the Service Binding Operator

Service Binding Operator is required to bind an application with microservices.

Visit [Service Binding Operator on OperatorHub.io](https://operatorhub.io/operator/service-binding-operator) or the [official documentation](https://redhat-developer.github.io/service-binding-operator/userguide/getting-started/installing-service-binding.html#installing-on-kubernetes) of Service Binding Operator to see how you can install it on your Kubernetes cluster.

## Troubleshooting

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
