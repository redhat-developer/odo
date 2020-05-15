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
        content: Creating a single-component application with odo
        url: '/docs/creating-a-single-component-application-with-odo'
---
`odo` is a CLI tool for creating applications on OpenShift and
Kubernetes. `odo` allows developers to concentrate on creating
applications without the need to administer a cluster itself. Creating
deployment configurations, build configurations, service routes and
other OpenShift or Kubernetes elements are all automated by `odo`.

Existing tools such as `oc` are more operations-focused and require a
deep understanding of Kubernetes and OpenShift concepts. `odo` abstracts
away complex Kubernetes and OpenShift concepts allowing developers to
focus on what is most important to them: code.

# Key features

`odo` is designed to be simple and concise with the following key
features:

  - Simple syntax and design centered around concepts familiar to
    developers, such as projects, applications, and components.

  - Completely client based. No server is required within the cluster
    for deployment.

  - Official support for Node.js and Java components.

  - Partial compatibility with languages and frameworks such as Ruby,
    Perl, PHP, and Python.

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
    applications: a video game, a media player, a web browser.

  - Component  
    A component is a set of Kubernetes resources which host code or
    data. Each component can be run and deployed separately. Examples of
    components: Node.js, Perl, PHP, Python, Ruby.

  - Service  
    A service is software that your component links to or depends on.
    Examples of services: MariaDB, Jenkins, MySQL. In `odo`, services
    are provisioned from the OpenShift Service Catalog and must be
    enabled within your
cluster.

## Officially supported languages and corresponding container images

| Language    | Container image                                                                                                                                  | Package manager |
| ----------- | ------------------------------------------------------------------------------------------------------------------------------------------------ | --------------- |
| **Node.js** | [centos/nodejs-8-centos7](https://github.com/sclorg/s2i-nodejs-container)                                                                        | NPM             |
|             | [rhoar-nodejs/nodejs-8](https://access.redhat.com/articles/3376841)                                                                              | NPM             |
|             | [bucharestgold/centos7-s2i-nodejs](https://www.github.com/bucharest-gold/centos7-s2i-nodejs)                                                     | NPM             |
|             | [rhscl/nodejs-8-rhel7](https://access.redhat.com/containers/#/registry.access.redhat.com/rhscl/nodejs-8-rhel7)                                   | NPM             |
|             | [rhscl/nodejs-10-rhel7](https://access.redhat.com/containers/#/registry.access.redhat.com/rhscl/nodejs-10-rhel7)                                 | NPM             |
| **Java**    | [redhat-openjdk-18/openjdk18-openshift](https://access.redhat.com/containers/#/registry.access.redhat.com/redhat-openjdk-18/openjdk18-openshift) | Maven, Gradle   |
|             | [openjdk/openjdk-11-rhel8](https://access.redhat.com/containers/#/registry.access.redhat.com/openjdk/openjdk-11-rhel8)                           | Maven, Gradle   |
|             | [openjdk/openjdk-11-rhel7](https://access.redhat.com/containers/#/registry.access.redhat.com/openjdk/openjdk-11-rhel7)                           | Maven, Gradle   |

Supported languages, container images, and package managers

### Listing available container images

> **Note**
> 
> The list of available container images is sourced from the cluster’s
> internal container registry and external registries associated with
> the cluster.

To list the available components and associated container images for
your cluster:

1.  Log in to the cluster with `odo`:
    
        $ odo login -u developer -p developer

2.  List the available `odo` supported and unsupported components and
    corresponding container images:
    
        $ odo catalog list components
        Odo Supported OpenShift Components:
        NAME        PROJECT      TAGS
        java       openshift     8,latest
        nodejs     openshift     10,8,8-RHOAR,latest
        
        Odo Unsupported OpenShift Components:
        NAME                      PROJECT       TAGS
        dotnet                    openshift     1.0,1.1,2.1,2.2,latest
        fuse7-eap-openshift       openshift     1.3
    
    The `TAGS` column represents the available image versions, for
    example, `10` represents the `rhoar-nodejs/nodejs-10` container
    image.
