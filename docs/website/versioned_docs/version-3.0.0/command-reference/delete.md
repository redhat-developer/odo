---
title: odo delete component
sidebar_position: 4
---

`odo delete component` command is useful for deleting resources that are managed by odo. It deletes the component and its related innerloop and outerloop resources from the cluster.

There are 2 ways to delete a component:
1. [Delete with access to Devfile](#delete-with-access-to-devfile)
2. [Delete without access to Devfile](#delete-without-access-to-devfile)

## Delete with access to Devfile
```shell
odo delete component
```
odo analyzes the Devfile _present in the current directory_ for the component and the outerloop resources.
If the component has been deployed on the cluster, then odo will list all the resources and prompt the user to confirm the deletion.
Otherwise, odo will exit with a message stating that it could not find the resources on the cluster.

:::info
odo does not delete the Devfile, the odo configuration files, or the source code.

## Delete without access to Devfile
```shell
odo delete component --name <component_name> --namespace <namespace>
```

odo builds a label from the component name to search for the component on the cluster in the given namespace, and deletes it along with all related resources.
If odo finds the component, then it will list all the resources and prompt the user to confirm the deletion.
Otherwise, odo will exit with a message stating that it could not find the resources on the cluster.

The `--namespace` is optional, if not provided, odo will use the current active namespace.


:::caution
In both cases, `odo` does not wait for resources to be deleted.


## Available Flags
* `-f`, `--force` - Use this flag to avoid the user prompt.
* `--name` - Name of the component to delete, optional. By default, the component described in the local devfile is deleted
* `--namespace` - Namespace in which to find the component to delete, optional. By default, the current namespace defined in kubeconfig is used
Check the [documentation on flags](flags.md) to see more flags available.