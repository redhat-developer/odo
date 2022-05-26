---
title: odo delete namespace
---

`odo delete namespace` lets you delete a namespace/project on your cluster. If you are on a Kubernetes cluster, running the command will delete a Namespace resource for you, and for an OpenShift cluster, it will delete a Project resource.

To delete a namespace you can run `odo delete namespace <name>`:
```shell
odo delete namespace mynamespace
```

Example:
```shell
odo delete namespace mynamespace
 ✓  Namespace "mynamespace" deleted
```

Optionally, you can also use `project` as an alias to `namespace`.

To delete a project you can run `odo delete project <name>`:
```shell
odo delete project myproject
```

Example:
```shell
odo delete project myproject    
  ✓  Project "myproject" deleted
```

:::note
This command is smart enough to detect the resources supported by your cluster and make an informed decision on the type of resource that should be deleted, using either of the aliases.
So you can run `odo delete project` on a Kubernetes cluster, and it will delete a Namespace resource, and you can run `odo delete namespace` on an OpenShift cluster, it will delete a Project resource.
:::

## Available Flags
* `-f`, `--force` - Use this flag to avoid being prompted for confirmation.
* `--wait` - Use this flag to wait until the namespace no longer exists
