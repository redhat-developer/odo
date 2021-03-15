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

This document walks you through the steps to install [Service Binding Operator v0.3.0](https://github.com/redhat-developer/service-binding-operator/tree/v0.3.0) on OpenShift cluster and [Service Binding Operator v0.4.0](https://operatorhub.io/operator/service-binding-operator) on Kubernetes cluster.

## Why do I need the Service Binding Operator?

odo uses Service Binding Operator to provide the `odo link` feature which helps connect an odo component to a service or another component.

## Installing Service Binding Operator on OpenShift

To install Service Binding Operator on OpenShift, refer [this video](https://www.youtube.com/watch?v=8QmewscQwHg).

## Installing Service Binding Operator on Kubernetes

Before installing an Operator, we first need to enable the Operator Lifecycle Manager (OLM).

1.  If you are using [minikube](https://minikube.sigs.k8s.io/), please enable the OLM addon by doing:
    
    ``` sh
    $ minikube addons enable olm
    ```

2.  To install OLM on a Kubernetes cluster setup other than minikube, please refer the [installation instructions on GitHub](https://github.com/operator-framework/operator-lifecycle-manager/#installation).

To install the Operator, execute the following `kubectl` command provided on its [OperatorHub.io page](https://operatorhub.io/operator/service-binding-operator):

``` sh
$ kubectl create -f https://operatorhub.io/install/service-binding-operator.yaml
```

### Making sure that Service Binding Operator installed successfully on Kubernetes

1.  One way to make sure that the Operator installed properly is to verify that its [pod](https://kubernetes.io/docs/concepts/workloads/pods/) started and is in "Running" state (note that you will have to specify the namespace where you installed Service Binding Operator in earlier step, and the pod name will be different in your setup than what’s shown in below output):
    
    ``` sh
    $ kubectl get pods --namespace operators
    NAME                                        READY   STATUS     RESTARTS   AGE
    service-binding-operator-6b7c654c89-rg9gq   1/1     Running    0          15m
    ```

2.  Another aspect to check is output of below command as suggested in the Operator’s installation instruction:
    
    ``` sh
    $ kubectl get csv -n operators
    ```
    
    If you see the value under `PHASE` column to be anything other than `Installing` or `Succeeded`, please take a look at the pods in `olm` namespace and ensure that the pod starting with the name `operatorhubio-catalog` is in `Running` state:
    
    ``` sh
    $ kubectl get pods -n olm
    NAME                                READY   STATUS             RESTARTS   AGE
    operatorhubio-catalog-x24dq         0/1     CrashLoopBackOff   6          9m40s
    ```
    
    If you see output like above where the pod is in `CrashLoopBackOff` state or any other state other than `Running`, delete the pod (note that exact name of the pod will be different on your cluster):
    
    ``` sh
    $ kubectl delete pods -n olm operatorhubio-catalog-x24dq
    ```
