---
title: odo add binding
sidebar_position: 4
---

The `odo add binding` command can add a link between an operator-backed service and a component. odo uses the Service Binding Operator to create this link. Running this command will make the necessary changes to the Devfile, and once pushed to the cluster, it creates an instance of the `ServiceBinding` resource.

Currently, it only allows connecting to the operator-backed services which support binding via the Service Binding Operator.
To know about the operators supported by the Service Binding Operator, read their [README](https://github.com/redhat-developer/service-binding-operator#known-bindable-operators).

## Pre-requisites
* A directory containing a Devfile; if you don't have one, see [odo init](init.md) on obtaining a devfile.
* A cluster with the Service Binding operator installed, along with the operator whose service you need to bind to

## Interactive Mode
In the interactive mode, you will be guided to choose:
* a service from the list of bindable service instances as supported by the Service Binding Operator,
* option to bind the service as a file,
* a name for the binding.

```shell
# Add binding between service named 'myservice',
# and the component present in the working directory in the interactive mode
odo add binding
```

## Non-interactive mode
In the non-interactive mode, you will have to specify the following required information through the command-line:
* `--service` flag to specify the service you want to bind to,
* `--name` flag to specify a name for the binding,
* `--bind-as-files` flag to specify if the service should be bound as a file; this flag is set to true by default.


```shell
# Add binding between a service named 'myservice',
# and the component present in the working directory in the non-interactive mode
odo add binding --name mybinding --service myRedisService.Redis
```

### Formats supported by the `--service` flag
The `--service` flag supports the following formats to specify the service name:
* `<name>`
* `<name>.<kind>`
* `<name>.<kind>.<apigroup>`
* `<name>/<kind>`
* `<name>/<kind>.<apigroup>`

The above formats are helpful when multiple services with the same name exists on the cluster.

#### Examples - 
```shell
# Add binding between a service named 'myservice',
# and the component present in the working directory
odo add binding --service myservice --name myRedisService

# Add binding between service named 'myservice' of kind 'Redis', and APIGroup 'redis.redis.opstreelab.in',
# and the component present in the working directory 
odo add binding --service myservice/Redis.redis.redis.opstreelab.in --name myRedisService
odo add binding --service myservice.Redis.redis.redis.opstreelab.in --name myRedisService

# Add binding between service named 'myservice' of kind 'Redis',
# and the component present in the working directory
odo add binding --service myservice/Redis --name myRedisService
odo add binding --service myservice.Redis --name myRedisService
```
