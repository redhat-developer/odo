---
title: odo list namespace
sidebar_position: 2
---

## Description

You can use `odo list namespace` to list all the available namespaces within the cluster. 
If you are on a Kubernetes cluster, this command will return a list of Namespace resources, but on an OpenShift cluster, 
it will return a list of Project resources.

## Running the Command

To list all the namespaces, you can run `odo list namespace`:
```shell
odo list namespace
```

Example - 
```sh
$ odo list namespace
ACTIVE     NAME
*          default
           kube-node-lease
           kube-public
           kube-system
           mynamespace
           myproject
           olm
           operators
```

 Optionally, you can also use `project` as an alias to `namespace`.
 
 To list all the projects, you can run `odo list project`:
 ```shell
 odo list project
 ```

 Example -
 ```sh
 $ odo list project
 ACTIVE     NAME
*          default
           kube-node-lease
           kube-public
           kube-system
           mynamespace
           myproject
           olm
           operators
```

:::note
Using either of the aliases will not affect the resources returned by the cluster. This command is smart enough to detect the resources supported by your cluster and make an informed decision on the type of resource that should be listed.
So you can run `odo list project` on a Kubernetes cluster, and it will list Namespace resources, and you can run `odo list namespace` on an OpenShift cluster, it will list Project resources.
:::