---
title: Creating Kubernetes resources
sidebar_position: 2
---
# Creating Kubernetes resources using odo

While odo is mainly focused on application developers who would like to care less about Kubernetes and more about getting their application running on top of it, it also tries to make things simple for application architects or devfile stack authors who are comfortable with Kubernetes. One such feature of odo that we will discuss in this guide is creation of Kubernetes resources like Pods, Deployments, and such using odo. Using this, if an advanced user would like to create some Kubernetes resources, they could edit the `devfile.yaml` and add it there. An `odo push` after the edit would create the resource on the cluster.

In this guide, we will create an nginx Deployment using its Kubernetes manifest. We will write this manifest in the `devfile.yaml`. Upon doing `odo push`, you will be able to see a Deployment and its Pods on the Kubernetes cluster using `kubectl` commands. We will also see a "hack" to add this manifest into the `devfile.yaml` using an odo command.

## Create an odo component

As with other resources like URL, Storage and Services, to create a Kubernetes resource, we first need to have an odo component. We will keep it simple by using a nodejs starter project. In an empty directory, create a component using:
```shell
odo create nodejs --starter
```
Example:
```shell
$ odo create nodejs --starter
Devfile Object Validation
 ✓  Checking devfile existence [52416ns]
 ✓  Creating a devfile component from registry: stage [95517ns]
Validation
 ✓  Validating if devfile name is correct [97488ns]

Starter Project
 ✓  Downloading starter project nodejs-starter from https://github.com/odo-devfiles/nodejs-ex.git [593ms]

Please use `odo push` command to create the component with source deployed
```

## Add Deployment manifest to `devfile.yaml`

Open the `devfile.yaml` created by above command and add below in the `components` section. Note that there is already a `runtime` component in this section.:
```yaml
- name: nginx-deploy
  kubernetes:
    inlined: |
      apiVersion: apps/v1
      kind: Deployment
      metadata:
        name: nginx-deployment
        labels:
          app: nginx
      spec:
        replicas: 3
        selector:
          matchLabels:
            app: nginx
        template:
          metadata:
            labels:
              app: nginx
          spec:
            containers:
              - name: nginx
                image: quay.io/bitnami/nginx
                ports:
                  - containerPort: 80
```

### The "hack"

When you do `odo service create`, odo adds the information of the service as a Kubernetes inlined component to your `devfile.yaml`. So, if you don't want to edit the `devfile.yaml` yourself and use odo to add this information to it, copy below YAML into a file on your system:
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  labels:
    app: nginx
spec:
  replicas: 3
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
        - name: nginx
          image: quay.io/bitnami/nginx
          ports:
            - containerPort: 80
```
Let's say you copy it to `deployment.yaml` in your current working directory. Now, run the following command:
```shell
odo service create --from-file deployment.yaml
```
Example:
```shell
$ odo service create --from-file deployment.yaml 
Successfully added service to the configuration; do 'odo push' to create service on the cluster
```
Voilà! The Deployment manifest has been added to the `devfile.yaml`.

## Push to the cluster

Now all you need to do is `odo push`. Note that, unlike for Operator backed service, odo won't show you any message indicating that a service or some resource was created on the cluster. This is by design because the feature is meant for advanced users who can play with resources created in this way through `kubectl` CLI. However, if odo fails to create a resource, it will error out and let you know about it.

See if the Deployment and its Pods were created on the cluster using:
```shell
kubectl get deploy
kubectl get pods
```

## Good to know

odo adds a Kubernetes label to the resources created in this way. It is `app.kubernetes.io/managed-by: odo`. odo also sets the `ownerReferences` for such objects to the underlying odo component so that when you do `odo delete`, such resources are deleted from the cluster. Other than that, odo doesn't help in managing such resources and users are expected to know how to do so.