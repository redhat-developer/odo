---
# Page settings
layout: default
keywords:
comments: false

# Hero section
title: Installing Operators on minikube
description: Installing etcd Operator and Service Binding Operator on minikube

# Micro navigation
micro_nav: true

---
# Installing Operators on minikube

This guide assumes that you are using [minikube](https://minikube.sigs.k8s.io/docs/) v1.11.0 or newer.

In this guide, we will discuss installing two Operators on a minikube environment:

1.  etcd Operator

2.  Service Binding Operator

## Prerequisites

You must enable the `olm` addon for your minikube cluster by doing:

``` sh
$ minikube addons enable olm
```

## Installing etcd Operator

Operators can be installed in a specific namepsace or across the cluster (that is, for all the namespaces). We will install etcd Operator across the cluster such that if you create a new namespace, the etcd Operator will be automatically available for use.

To install an Operator, we need to make sure that the namespace in which we’re installing it has an `OperatorGroup`. Since we want to install etcd Operator across all the namespaces, we will install it in `operators` namespace and `olm` takes care of making it available across all the namespace.

> **Note**
> 
> You can’t always install an Operator in the `operators` namespace and expect it to be available across all namespaces. The Operator you’re trying to installing needs to be designed to be available in this way as well. Certain Operators only support installation in a single namespace.
> 
> Discussing this topic is out of scope of this guide so we have stated it as a note.

Enabling the `olm` addon will, among other things, create an `OperatorGroup` in the `operators` namepsace. Make sure that it’s there:

``` sh
$ kubectl get og -n operators
NAME               AGE
global-operators   3m37s
```

If you don’t see one, create it using below command:

``` sh
$ kubectl create -f - << EOF
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

Now, install the etcd Operator using below command:

``` sh
$ kubectl create -f - << EOF
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
```

Give it a few seconds before checking the availability of the etcd Operator:

``` sh
$ odo catalog list services
Operators available in the cluster
NAME                                CRDs
etcdoperator.v0.9.4-clusterwide     EtcdCluster, EtcdBackup, EtcdRestore
```

### Troubleshooting

If you don’t see etcd Operator using above command or by doing `kubectl get csv -n operators`, make sure that pod belonging to the `CatalogSource` named `operatorhubio-catalog` is running:

``` sh
$ kubectl get po -n olm | grep operatorhubio-catalog
```

If the state of this pod is `CrashLoopBackOff`, delete it so that Kubernetes will automatically spin up a new pod for the `CatalogSource`:

``` sh
$ kubectl delete po -n olm <name-of-operatorhubio-catalog-pod>
```

Once the pod for this `CatalogSource` is up, wait a few seconds before trying to find the etcd Operator when you do `odo catalog list services`.

## Installing the Service Binding Operator

odo uses the [Service Binding Operator](https://github.com/redhat-developer/service-binding-operator/) to provide `odo link` feature which links an odo component with an Operator backed service. Thus, to be able to use this feature, it’s essential that we install the Operator first.

Service Binding Operator is not yet available via the OLM. The team is [working on making it available](https://github.com/redhat-developer/service-binding-operator/issues/727) through OLM.

To install the Operator, execute the following `kubectl` command:

``` sh
$ kubectl apply -f https://github.com/redhat-developer/service-binding-operator/releases/download/v0.1.1-364/install-v0.1.1-364.yaml
```

You should now see a `Deployment` for Service Binding Operator in the namespace where you installed it:

``` sh
$ kubectl get deploy -n <replace-namespace-value>
```

### Troubleshooting

If linking doesn’t work after installing the Service Binding Operator as described above, please open an issue on the odo repository describing the issue. On the issue, please provide the logs of the `Pod` created by the `Deployment` of Service Binding Operator. Execute below command in the namespace where you installed Service Binding Operator (note that `Pod` name will be different in your environment than what’s shown below):

``` sh
$ kubectl get pods
NAME                                        READY   STATUS    RESTARTS   AGE
service-binding-operator-65745c4bdc-gc6km   1/1     Running   0          34m

$ kubectl logs service-binding-operator-65745c4bdc-gc6km --follow
```
