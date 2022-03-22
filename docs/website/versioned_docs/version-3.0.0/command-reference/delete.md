---
title: odo delete component
sidebar_position: 4
---

`odo delete component` command is useful for deleting resources that are managed by odo. It deletes the component and it's related innerloop and outerloop resources from the cluster.

There are 2 ways to deleting a component:
1. [Delete with access to Devfile](/#Delete with access to Devfile)
2. [Delete without access to Devfile](/#Delete without access to Devfile)

## Delete with access to Devfile
```shell
odo delete component
```
odo analyzes the devfile for the devfile component and the outerloop resources.
If the component has been deployed on the cluster, then odo will list all the resources and prompt the user to confirm the deletion.
Otherwise, odo will exit with a message stating that it could not find the resources on the cluster.

_Note:_ odo does not delete the Devfile, and config files.

## Delete without access to Devfile
```shell
odo delete component --name <component_name> --namespace <namespace>
```

odo builds a label from the component name to search for the component on the cluster in the given namespace, and deletes it along with all it's related resources.
If it finds the component, then it will list all the resources and prompt the user to confirm the deletion.
Otherwise, it will exit with a message stating that it could not find the resources on the cluster.


**_Note:_** In the both the cases, odo does not wait for resources to be deleted.


## Available Flags
* `-f`, `--force` - Use this flag to avoid the user prompt.
* `--name` - Name of the component to delete, optional. By default, the component described in the local devfile is deleted
* `--namespace` - Namespace in which to find the component to delete, optional. By default, the current namespace defined in kubeconfig is used
Check the [documentation on flags](flags.md) to see more flags available.