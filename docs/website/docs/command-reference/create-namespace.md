---
title: odo create namespace
---

`odo create namespace` lets you create a namespace/project on your cluster. If you are on a Kubernetes cluster, running the command will create a Namespace resource for you, and for an OpenShift cluster, it will create a Project resource.

Any new namespace created with this command will also be set as the current active namespace, this applies to project as well.

## Running the command
To create a namespace you can run `odo create namespace <name>`:
```shell
odo create namespace mynamespace
```
```shell
$ odo create namespace mynamespace
 ✓  Namespace "mynamespace" is ready for use
 ✓  New namespace created and now using namespace: mynamespace
```

Optionally, you can also use `project` as an alias to `namespace`.

To create a project you can run `odo create project <name>`:
```shell
odo create project myproject
```
```shell
$ odo create project myproject
 ✓  Project "myproject" is ready for use
 ✓  New project created and now using project: myproject
```

:::note
Using either of the aliases will not make any change to the resource created on the cluster. This command is smart enough to detect the resources supported by your cluster and make an informed decision on the type of resource that should be created.
So you can run `odo create project` on a Kubernetes cluster, and it will create a Namespace resource, and you can run `odo create namespace` on an OpenShift cluster, it will create a Project resource.
:::
