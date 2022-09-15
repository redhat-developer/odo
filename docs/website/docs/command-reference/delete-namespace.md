---
title: odo delete namespace
---

`odo delete namespace` lets you delete a namespace/project on your cluster. If you are on a Kubernetes cluster, running the command will delete a Namespace resource for you, and for an OpenShift cluster, it will delete a Project resource.

## Running the command
To delete a namespace, run the following command:
```shell
odo delete namespace <name> [--wait] [--force]
```
<details>
<summary>Example</summary>

```shell
$ odo delete namespace mynamespace
 ✓  Namespace "mynamespace" deleted
```
</details>

Optionally, you can also use `project` as an alias to `namespace`.

To delete a project, run the following command:
```shell
odo delete project <name> [--wait] [--force]
```
<details>
<summary>Example</summary>

```shell
$ odo delete project myproject
✓  Project "myproject" deleted
```
</details>


:::tip
This command is smart enough to detect the resources supported by your cluster and make an informed decision on the type of resource that should be deleted, using either of the aliases.

So you can run `odo delete project` on a Kubernetes cluster, and it will delete a `Namespace` resource, or you can run `odo delete namespace` on an OpenShift cluster, it will delete a `Project` resource.
:::
