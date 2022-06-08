---
title: odo remove binding
---

## Description
The `odo remove binding` command removes the link created between the component and a service via Service Binding.

## Running the Command
Running this command removes the reference from the devfile, but does not necessarily remove it from the cluster. To remove the ServiceBinding from the cluster, you must run `odo dev`, or `odo deploy`.

The command takes a required `--name` flag that points to the name of the Service Binding to be removed.
```shell
odo remove binding --name <ServiceBinding_name>
```

## Examples 
```shell
$ odo remove binding --name redis-service-my-nodejs-app
```

There is no interactive mode for this command at the moment.