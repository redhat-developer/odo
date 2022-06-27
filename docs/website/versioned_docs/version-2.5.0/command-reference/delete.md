---
title: odo delete
sidebar_position: 4
---

`odo delete` command is useful for deleting resources that are managed by odo.

## Deleting a component

To delete a _devfile_ component, you can execute `odo delete`.

```shell
odo delete
```

If the component is pushed to the cluster, running the above command will delete the component from the cluster, and it's dependant storage, url, secrets, and other resources.
If it is not pushed, the command would exit with an error stating that it could not find the resources on the cluster.

Use `-f` or `--force` flag to avoid the confirmation questions. 

## Un-deploying Devfile Kubernetes components

To undeploy the Devfile Kubernetes components deployed with `odo deploy` from the cluster, you can execute the `odo delete` command with `--deploy` flag:
```shell
odo delete --deploy
```

Use `-f` or `--force` flag to avoid the confirmation questions.

## Delete Everything

To delete a _devfile_ component, the Devfile Kubernetes component(deployed via `odo deploy`), Devfile, and the local configuration, you can execute the `odo delete` command with `--all` flag:
```shell
odo delete --all
```

## Available Flags
* `-f`, `--force` - Use this flag to avoid the confirmation questions.
* `-w`, `--wait` - Use this flag to wait for component deletion, and it's dependant; this does not work with the un-deployment.
Check the [documentation on flags](flags.md) to see more flags available.