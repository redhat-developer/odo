---
title: odo catalog
sidebar_position: 1
---
# odo catalog

odo uses different *catalogs* to help you deploy *components* and *services*.

## Components

odo uses the portable *devfile* format to describe the components you want to work on, and can connect to devfile registries to download devfiles for different languages and frameworks. See [`odo registry`](/docs/command-reference/registry) for more information.

### Listing components

You can list the available *devfiles* available on the different registries with the command:

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

The output of the command has been shortened for presentation reasons; you should obtain a longer list when running this command with the default devfile registry.

### Getting information about a component

You can get more information about a specific component with the command:

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

The first information is the registry from which is retrieved the devfile.

The second information is the list of *starter projects* referenced for this devfile. A *starter project* is a simple project in the same language and framework of the devfile, that can help you start a new project. See [`odo create`](/docs/command-reference/create) for more information on creating a project from a start project.

## Services

odo can help you deploy *services* with the help of *operators*.

### Listing services

You can get the list of available operators and their associate services with the command:

```
$ odo catalog list services
Services available through Operators
NAME                                 CRDs
postgresql-operator.v0.1.1           Backup, Database
redis-operator.v0.8.0                RedisCluster, Redis
```

In this example, you can see that two operators are installed in your cluster. The first one to deploy PostgreSQL related services, the second to deploy Redis related services. The PostgreSQL operator offers two services: `Backup` and `Database`.

> Note that only operators deployed with the help of the [*Operator Lifecycle Manager*](https://olm.operatorframework.io/) are supported by odo. See [Installing the Operator Lifecycle Manager (OLM)](/docs/getting-started/cluster-setup/kubernetes#installing-the-operator-lifecycle-manager-olm) for more information.

> Also note that odo fetches the `ClusterServiceVersion` (`CSV`) resources of the current namespace in a *Succeeded* phase to get the list of available operators. When a new namespace is created, these resources are automatically added to the namespace, and some time is necessary for them to be in the *Succeeded* phase. So, please be patient when you try to list the services from a newly created namespace, you could get an empty list for some time.

### Searching services

You can search for specific services by keyword with the command:

```
$ odo catalog search service postgre
Services available through Operators
NAME                           CRDs
postgresql-operator.v0.1.1     Backup, Database
```

A similar list will be displayed, containing only the relevant operators, whose name contains the searched keyword.

### Getting information about a service

You can get more information about a specific service with the command:

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

A service is represented in your cluster by a `CustomResourceDefinition` (commonly named `CRD`). This command will display the details about this CRD: its kind, its version, and the list of fields available to define an instance of this custom resource.

The list of fields is extracted from the *OpenAPI schema* included in the `CRD`. As this information is optional in a `CRD`, if it is not present, the fields information are extracted from the `ClusterServiceVersion` (`CSV`) representing the service instead.
