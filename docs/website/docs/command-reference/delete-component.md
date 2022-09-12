---
title: odo delete component
---

`odo delete component` command is useful for deleting resources that are managed by `odo`. It deletes the component and its related innerloop, and outerloop resources from the cluster.

There are 2 ways to delete a component:
- [Delete with access to Devfile](#delete-with-access-to-devfile)
- [Delete without access to Devfile](#delete-without-access-to-devfile)
- [Available Flags](#available-flags)

## Delete with access to Devfile
```shell
odo delete component
```
```shell
$ odo delete component
Searching resources to delete, please wait...
This will delete "my-nodejs" from the namespace "my-project".
 •  The component contains the following resources that will get deleted:
        - Deployment: my-component
? Are you sure you want to delete "my-nodejs" and all its resources? Yes
The component "my-nodejs" is successfully deleted from namespace "my-project"
```
`odo` looks into the Devfile _present in the current directory_ for the component resources for the innerloop, and outerloop.
If these resources have been deployed on the cluster, then `odo` will delete them after user confirmation.
Otherwise, `odo` will exit with a message stating that it could not find the resources on the cluster.

:::note
If some resources attached to the component are present on the clutser, but not in the Devfile, then they will not be deleted.
You can delete these resources by running the command in the [next section](#delete-without-access-to-devfile).
:::

`odo` does not delete the Devfile, the `odo` configuration files, or the source code.

## Delete without access to Devfile
```shell
odo delete component --name <component_name> [--namespace <namespace>]
```
```shell
$ odo delete component --name my-nodejs
Searching resources to delete, please wait...
This will delete "my-nodejs" from the namespace "my-project".
 •  The component contains the following resources that will get deleted:
        - Deployment: my-component
? Are you sure you want to delete these resources? Yes
The component "my-nodejs" is successfully deleted from namespace "my-project"
```

`odo` searches for resources attached to the given component in the given namespace on the cluster.
If `odo` finds the resources, it will delete them after user confirmation.
Otherwise, `odo` will exit with a message stating that it could not find the resources on the cluster.

`--namespace` is optional, if not provided, `odo` will use the current active namespace.

:::info
In both cases, `odo` does not wait for resources to be deleted.
:::

## Available Flags
* `-f`, `--force` - Use this flag to avoid being prompted for confirmation.
* `--name` - Name of the component to delete (optional). By default, the component described in the local devfile is deleted
* `--namespace` - Namespace to find the component to delete (optional). By default, the current namespace defined in kubeconfig is used
