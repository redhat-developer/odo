---
title: odo list binding
---

## Description

You can use `odo list binding` to list all the Service Bindings declared in the current namespace and, if present, 
in the Devfile of the current directory.

This command supports the service bindings added with the command `odo add binding`, and bindings added manually
to the Devfile, using a `ServiceBinding` resource from one of these apiVersion:
- `binding.operators.coreos.com/v1alpha1`
- `servicebinding.io/v1alpha3`

The name of the service binding is prefixed with `*` when the service binding is declared in the Devfile present in the current directory.

To get more information about a specific service binding, you can run the command `odo describe binding --name <name>` (see [`odo describe binding` command reference](./describe-binding.md)).

## Running the Command

To list all the service bindings, you can run `odo list binding`:
```shell
odo list binding
```

Example:

```sh
$ odo list binding
 NAME                              APPLICATION                     SERVICES                                                   RUNNING IN 
 binding-to-redis                  my-nodejs-app-app (Deployment)  redis (Service)                                            Dev
 * my-nodejs-app-cluster-sample    my-nodejs-app-app (Deployment)  cluster-sample (Cluster.postgresql.k8s.enterprisedb.io)    Dev       
 * my-nodejs-app-cluster-sample-2  my-nodejs-app-app (Deployment)  cluster-sample-2 (Cluster.postgresql.k8s.enterprisedb.io)  Dev       
```
