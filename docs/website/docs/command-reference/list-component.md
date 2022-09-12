---
title: odo list component
---

`odo list component` command is useful for getting information about components running on a specific namespace.

If the command is executed from a directory containing a Devfile, it also displays the component
defined in the Devfile as part of the list, prefixed with a star(*).

For each component, the command displays:
- its name,
- its project type,
- on which mode it is running (None, Dev, Deploy, or both), note that None is only applicable to the component 
defined in the local Devfile,
- by which application the component has been deployed.

### Running the command
```shell
$ odo list component
 ✓  Listing components from namespace 'my-percona-server-mongodb-operator' [292ms]
 NAME              PROJECT TYPE  RUNNING IN  MANAGED                         
 my-nodejs         nodejs        Deploy      odo (v3.0.0-rc1)                
 my-go-app         go            Dev         odo (v3.0.0-rc1)                
 mongodb-instance  Unknown       None        percona-server-mongodb-operator 
```

## Available flags

* `--namespace` - Namespace to list the components from (optional). By default, the current namespace defined in kubeconfig is used
* `-o json` - Outputs the list in JSON format. See [JSON output](json-output.md) for more information

:::tip use of cache

`odo list component` makes use of cache for performance reasons. This is the same cache that is referred by `kubectl` command 
when you do `kubectl api-resources --cached=true`. As a result, if you were to install an Operator/CRD on the 
Kubernetes cluster, and create a resource from it using odo, you might not see it in the `odo list component` output. This 
would be the case for 10 minutes timeframe for which the cache is considered valid. Beyond this 10 minutes, the 
cache is updated anyway.

If you would like to invalidate the cache before the 10 minutes timeframe, you could manually delete it by doing:
```shell
rm -rf ~/.kube/cache/discovery/api.crc.testing_6443/
```
Above example shows how to invalidate the cache for a CRC cluster. Note that you will have to modify the `api.crc.
testing_6443` part based on the cluster you are working against.