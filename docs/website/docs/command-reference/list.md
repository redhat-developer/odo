---
title: odo list
---

`odo list` command combines the `odo list binding` and `odo list component` commands.

```shell
$ odo list
 ✓  Listing resources from the namespace "my-percona-server-mongodb-operator" [302ms]
 NAME              PROJECT TYPE  RUNNING IN  MANAGED
 my-nodejs         nodejs        Deploy      odo (v3.0.0-rc1)
 my-go-app         go            Dev         odo (v3.0.0-rc1)
 mongodb-instance  Unknown       None        percona-server-mongodb-operator 

Bindings:
 NAME                        APPLICATION                 SERVICES                                                   RUNNING IN 
 my-go-app-mongodb-instance  my-go-app-app (Deployment)  mongodb-instance (PerconaServerMongoDB.psmdb.percona.com)  Dev
```
## Available flags

* `--namespace` - Namespace to list the resources from (optional). By default, the current namespace defined in kubeconfig is used
* `-o json` - Outputs the list in JSON format. See [JSON output](json-output.md) for more information
