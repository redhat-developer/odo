---
title: odo list
---

`odo list` command combines the `odo list binding` and `odo list component` commands.

## Running the command

```shell
odo list
```
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
