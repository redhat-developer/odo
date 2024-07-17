---
title: odo list
---

`odo list` command combines the [`odo list binding`](./list-binding.md) and [`odo list component`](./list-component.md) commands.

## Running the command

```shell
odo list
```

<details>
<summary>Example</summary>

```shell
$ odo list
 ✓  Listing components from namespace 'my-percona-server-mongodb-operator' [292ms]
 NAME              PROJECT TYPE  RUNNING IN  MANAGED                          PLATFORM
 * my-nodejs       nodejs        Deploy      odo (v3.7)                       cluster
 my-go-app         go            Dev         odo (v3.7)                       podman
 mongodb-instance  Unknown       None        percona-server-mongodb-operator  cluster

Bindings:
 NAME                        APPLICATION                 SERVICES                                                   RUNNING IN 
 my-go-app-mongodb-instance  my-go-app-app (Deployment)  mongodb-instance (PerconaServerMongoDB.psmdb.percona.com)  Dev
```
</details>
