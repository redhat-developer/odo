---
# Page settings
layout: default
keywords:
comments: false

# Hero section
title: Understanding odo
description: Understanding the concepts of odo

# Micro navigation
micro_nav: true

# Page navigation
page_nav:
    prev:
        content: Installing odo
        url: '/docs/installing-odo'
    next:
        content: Deploying a devfile using odo
        url: '/docs/deploying-a-devfile-using-odo/'
---
`odo` is a CLI tool for creating applications on Kubernetes and
OpenShift . `odo` allows developers to concentrate on creating
applications without the need to administer a cluster itself. Creating
deployment configurations, build configurations, service routes and
other Kubernetes or OpenShift elements are all automated by `odo`.

Existing tools such as `oc` are more operations-focused and require a
deep understanding of Kubernetes and OpenShift concepts. `odo` abstracts
away complex Kubernetes and OpenShift concepts allowing developers to
focus on what is most important to them: code.

# Key features

`odo` is designed to be simple and concise with the following key
features:

  - Simple syntax and design centered around concepts familiar to
    developers, such as projects, applications, and components.

  - Completely client based. No additional server other than Kubernetes
    or OpenShift is required for deployment.

  - Official support for Node.js and Java components.

  - Detects changes to local code and deploys it to the cluster
    automatically, giving instant feedback to validate changes in real
    time.

  - Lists all the available components and services from the cluster.

# Core concepts

  - Project  
    A project is your source code, tests, and libraries organized in a
    separate single unit.

  - Application  
    An application is a program designed for end users. An application
    consists of multiple microservices or components that work
    individually to build the entire application. Examples of
    applications: e-Shop, Hotel Reservation System, Online Booking

  - Component  
    A component is a set of Kubernetes resources which host code or
    data. Each component can be run and deployed separately. Examples of
    components: Warehouse API Backend, Inventory API, Web Frontend,
    Payment Backend

  - Service  
    A service is software that your component links to or depends on.
    Examples of services: MariaDB, MySQL.

  - Devfile  
    A portable file responsible for your entire reproducable development
    environment.

## Official Devfiles

Devfiles describe your development environment link. [Click here for
more information on
Devfile.](https://odo.dev/docs/deploying-a-devfile-using-odo/)

| Language | Devfile Name     | Description                        | Devfile Source                                                                                                               | Supported Platform    |
| -------- | ---------------- | ---------------------------------- | ---------------------------------------------------------------------------------------------------------------------------- | --------------------- |
| Java     | java-maven       | Upstream Maven and OpenJDK 11      | [java-maven/devfile.yaml](https://github.com/odo-devfiles/registry/blob/master/devfiles/java-maven/devfile.yaml)             | amd64                 |
| Java     | java-openliberty | Open Liberty microservice in Java  | [java-openliberty/devfile.yaml](https://github.com/odo-devfiles/registry/blob/master/devfiles/java-openliberty/devfile.yaml) | amd64                 |
| Java     | java-quarkus     | Upstream Quarkus with Java+GraalVM | [java-quarkus/devfile.yaml](https://github.com/odo-devfiles/registry/blob/master/devfiles/java-quarkus/devfile.yaml)         | amd64                 |
| Java     | java-springboot  | Spring Boot® using Java            | [java-springboot/devfile.yaml](https://github.com/odo-devfiles/registry/blob/master/devfiles/java-springboot/devfile.yaml)   | amd64                 |
| Node.JS  | nodejs           | Stack with NodeJS 12               | [nodejs/devfile.yaml](https://github.com/odo-devfiles/registry/blob/master/devfiles/nodejs/devfile.yaml)                     | amd64, s390x, ppc64le |

List of Devfiles which are officially supported by odo

### Listing available Devfiles

> **Note**
> 
> The list of available Devfiles is sourced from the official [odo
> registry](https://github.com/odo-devfiles/registry) as well as any
> other registies added via `odo registry add`.

To list the available Devfiles:

    $ odo catalog list components
    Odo Devfile Components:
    NAME                 DESCRIPTION                            REGISTRY
    java-maven           Upstream Maven and OpenJDK 11          DefaultDevfileRegistry
    java-openliberty     Open Liberty microservice in Java      DefaultDevfileRegistry
    java-quarkus         Upstream Quarkus with Java+GraalVM     DefaultDevfileRegistry
    java-springboot      Spring Boot® using Java                DefaultDevfileRegistry
    nodejs               Stack with NodeJS 12                   DefaultDevfileRegistry
