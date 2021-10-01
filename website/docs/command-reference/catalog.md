---
title: odo catalog
sidebar_position: 1
---
# odo catalog

odo uses different *catalogs* to deploy *components* and *services*.

## Components

odo uses the portable *devfile* format to describe the components you want to work on. It can connect to various devfile registries to download devfiles for different languages and frameworks. See [`odo registry`](/docs/command-reference/registry) for more information.

### Listing components

You can list all the *devfiles* available on the different registries with the command:

```
odo catalog list components
```

Example:

```
$ odo catalog list components
Odo Devfile Components:
NAME             DESCRIPTION                          REGISTRY
go               Stack with the latest Go version     DefaultDevfileRegistry
java-maven       Upstream Maven and OpenJDK 11        DefaultDevfileRegistry
nodejs           Stack with Node.js 14                DefaultDevfileRegistry
php-laravel      Stack with Laravel 8                 DefaultDevfileRegistry
python           Python Stack with Python 3.7         DefaultDevfileRegistry
[...]
```

### Getting information about a component

You can get more information about a specific component with the command:

```
odo catalog describe component
```

Example:

```
$ odo catalog describe component nodejs
* Registry: DefaultDevfileRegistry

Starter Projects:
---
name: nodejs-starter
attributes: {}
description: ""
subdir: ""
projectsource:
  sourcetype: ""
  git:
    gitlikeprojectsource:
      commonprojectsource: {}
      checkoutfrom: null
      remotes:
        origin: https://github.com/odo-devfiles/nodejs-ex.git
  zip: null
  custom: null
```

`Registry` is the registry from which the devfile is retrieved.

*Starter projects* are sample projects in the same language and framework of the devfile, that can help you start a new project. See [`odo create`](/docs/command-reference/create) for more information on creating a project from a starter project.

## Services

odo can deploy *services* with the help of *operators*.

### Listing services

You can get the list of available operators and their associated services with the command:

```
odo catalog list services
```

Example: 

```
$ odo catalog list services
Services available through Operators
NAME                                 CRDs
postgresql-operator.v0.1.1           Backup, Database
redis-operator.v0.8.0                RedisCluster, Redis
```

In this example, you can see that two operators are installed in the cluster. The `postgresql-operator.v0.1.1` operator can deploy services related to PostgreSQL: `Backup` and `Database`. The `redis-operator.v0.8.0` operator can deploy services related to Redis: `RedisCluster` and `Redis`.

Only operators deployed with the help of the [*Operator Lifecycle Manager*](https://olm.operatorframework.io/) are supported by odo. See [Installing the Operator Lifecycle Manager (OLM)](/docs/getting-started/cluster-setup/kubernetes#installing-the-operator-lifecycle-manager-olm) for more information.

> Note: To get a list of all the available operators, odo fetches the `ClusterServiceVersion` (`CSV`) resources of the current namespace that are in a *Succeeded* phase. For operators that support cluster-wide access, when a new namespace is created, these resources are automatically added to it, but it may take some time before they are in the *Succeeded* phase, and odo may return an empty list until the resources are ready.

### Searching services

You can search for a specific service by a keyword with the command:

```
odo catalog search service
```

Example:

```
$ odo catalog search service postgre
Services available through Operators
NAME                           CRDs
postgresql-operator.v0.1.1     Backup, Database
```

You may see a similar list that contains only the relevant operators, whose name contains the searched keyword.

### Getting information about a service

You can get more information about a specific service with the command:

```
odo catalog describe service
```

Example:

```
$ odo catalog describe service postgresql-operator.v0.1.1/Database
KIND:    Database
VERSION: v1alpha1

DESCRIPTION:
     Database is the Schema for the the Database Database API

FIELDS:
   awsAccessKeyId (string)   
     AWS S3 accessKey/token ID

     Key ID of AWS S3 storage. Default Value: nil Required to create the Secret
     with the data to allow send the backup files to AWS S3 storage.
[...]
```

A service is represented in the cluster by a `CustomResourceDefinition` (commonly named `CRD`). This command will display the details about this CRD such as  `kind`, `version`, and the list of fields available to define an instance of this custom resource.

The list of fields is extracted from the *OpenAPI schema* included in the `CRD`. This information is optional in a `CRD`, and if it is not present, it is extracted from the `ClusterServiceVersion` (`CSV`) representing the service instead.
