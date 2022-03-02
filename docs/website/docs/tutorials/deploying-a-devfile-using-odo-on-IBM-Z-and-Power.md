---
title: Using odo on IBM-Z and Power
sidebar_position: 3
---
[//]: # (Add prerequisite section)

### Deploying your first devfile on IBM Z & Power
Since the [DefaultDevfileRegistry](https://registry.devfile.io/viewer) doesn't support IBM Z & Power now, you will need to create a secure private DevfileRegistry first. To create a new secure private DevfileRegistry, please check the doc [secure registry](../architecture/secure-registry.md).

The images can be used for devfiles on IBM Z & Power

|Language   | Devfile Name  | Description   | Image Source  | Supported Platform    |
| ----------- | ----------- | ----------- | ----------- | ----------- |
| dotnet | dotnet60 | Stack with .NET 6.0 | registry.access.redhat.com/ubi8/dotnet-60:6.0 | s390x |
| Go   | go | Stack with the latest Go version | golang:latest | s390x |
| Java      | java-maven    | Upstream Maven and OpenJDK 11 | registry.redhat.io/codeready-workspaces/plugin-java11-openj9-rhel8 | s390x, ppc64le |
| Java      | java-openliberty | Open Liberty microservice in Java | registry.redhat.io/codeready-workspaces/plugin-java11-openj9-rhel8 | s390x, ppc64le |
| Java      | java-openliberty-gradle | Java application Gradle-built stack using the Open Liberty runtime | openliberty/application-stack:gradle-0.2 | s390x |
| Java | java-quarkus | Upstream Quarkus with Java+GraalVM | registry.redhat.io/codeready-workspaces/plugin-java8-openj9-rhel8 | s390x, ppc64le|
| Java | java-springboot | Spring Boot® using Java| registry.redhat.io/codeready-workspaces/plugin-java11-openj9-rhel8 | s390x, ppc64le|
| Vert.x Java| java-vertx | Upstream Vert.x using Java | registry.redhat.io/codeready-workspaces/plugin-java11-openj9-rhel8 | s390x, ppc64le|
| Java | java-wildfly-bootable-jar | Java stack with WildFly in bootable Jar mode, OpenJDK 11 and Maven 3.5 | registry.access.redhat.com/ubi8/openjdk-11 | s390x |
| Node.JS | nodejs | Stack with NodeJS 12 | registry.redhat.io/codeready-workspaces/plugin-java8-openj9-rhel8 | s390x, ppc64le|
| Node.JS | nodejs-angular | Stack with Angular 12 | node:lts-slim | s390x |
| Node.JS | nodejs-nextjs | Stack with Next.js 11 | node:lts-slim | s390x |
| Node.JS | nodejs-nuxtjs | Stack with Nuxt.js 2 | node:lts | s390x |
| Node.JS | nodejs-react | Stack with React 17 | node:lts-slim | s390x |
| Node.JS | nodejs-svelte | Stack with Svelte 3 | node:lts-slim | s390x |
| Node.JS | nodejs-vue | Stack with Vue 3 | node:lts-slim | s390x |
| PHP | php-laravel | Stack with Laravel 8 | composer:latest | s390x |
| Python| python | Python Stack with Python 3.7 | registry.redhat.io/codeready-workspaces/plugin-java8-openj9-rhel8 | s390x, ppc64le|
| Django| python-django| Python3.7 with Django| registry.redhat.io/codeready-workspaces/plugin-java8-openj9-rhel8| s390x, ppc64le|

**Note**: Access to the Red Hat registry is required to use these images on IBM Power Systems & IBM Z.

[//]: # (Steps to use devfiles can be found in Deploying your first devfile)