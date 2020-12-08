---
# Page settings
layout: default
keywords:
comments: false

# Hero section
title: Installing Service Binding Operator
description: Installing Service Binding Operator on OpenShift and Kubernetes

# Micro navigation
micro_nav: true

---
# Installing Service Binding Operator

This document walks you through the steps to install [Service Binding Operator v0.3.0](https://github.com/redhat-developer/service-binding-operator/tree/v0.3.0) on OpenShift and Kubernetes clusters.

## Why do I need the Service Binding Operator?

odo uses Service Binding Operator to provide the `odo link` feature which helps connect an odo component to a service or another component.

## Installing Service Binding Operator on OpenShift

To install Service Binding Operator on OpenShift, refer [this video](https://www.youtube.com/watch?v=8QmewscQwHg).

## Installing Service Binding Operator on Kubernetes

Steps mentioned in this section were tested on [minikube](https://minikube.sigs.k8s.io/). We tested this on minikube v1.15.1 but it should work on v1.11.0 and above.

For Kubernetes, Service Binding Operator is not yet available via the OLM. The team is [working on making it available](https://github.com/redhat-developer/service-binding-operator/issues/727).

To install the Operator, execute the following `kubectl` command:

``` sh
$ kubectl apply -f https://gist.githubusercontent.com/dharmit/0e05be20e98c9271b2117acea7908cc2/raw/1e45fc89fc576e184e41fcc23e88d35f0e08a7e9/install.yaml
```

You should now see a `Deployment` for Service Binding Operator in the `default` namespace:

``` sh
$ kubectl get deploy -n default
```

If you would like to install the Service Binding Operator in a different namespace, edit [this line](https://gist.github.com/dharmit/0e05be20e98c9271b2117acea7908cc2#file-install-yaml-L464) by downloading the YAML mentioned in previous set to the namespace of your choice.

### Making sure that Service Binding Operator got correctly installed

One way to make sure that the Operator installed properly is to verify that its [pod](https://kubernetes.io/docs/concepts/workloads/pods/) started and is in "Running" state (note that you will have to specify the namespace where you installed Service Binding Operator in earlier step, and the pod name will be different in your setup than what’s shown in below output):

``` sh
$ kubectl get pods --namespace default
NAME                                        READY   STATUS     RESTARTS   AGE
service-binding-operator-6b7c654c89-rg9gq   1/1     Running    0          15m
```
