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
| .NET | dotnet60 | Stack with .NET 6.0 | registry.access.redhat.com/ubi8/dotnet-60:6.0 | s390x |
| Go   | go | Stack with the latest Go version | golang:latest | s390x |
| Java      | java-maven    | Upstream Maven and OpenJDK 11 | rregistry.redhat.io/codeready-workspaces/plugin-java11-rhel8 | s390x, ppc64le |
| Java      | java-openliberty | Java application Maven-built stack using the Open Liberty runtime | icr.io/appcafe/open-liberty-devfile-stack:22.0.0.1 | s390x, ppc64le |
| Java      | java-openliberty-gradle | Java application Gradle-built stack using the Open Liberty runtime | icr.io/appcafe/open-liberty-devfile-stack:22.0.0.1-gradle | s390x |
| Java | java-quarkus | Quarkus with Java | registry.access.redhat.com/ubi8/openjdk-11 | s390x, ppc64le|
| Java | java-springboot | Spring Boot® using Java| registry.redhat.io/codeready-workspaces/plugin-java11-rhel8 | s390x, ppc64le|
| Java | java-vertx | Upstream Vert.x using Java | registry.redhat.io/codeready-workspaces/plugin-java11-rhel8 | s390x, ppc64le|
| Java | java-websphereliberty | Java application Maven-built stack using the WebSphere Liberty runtime | icr.io/appcafe/websphere-liberty-devfile-stack:22.0.0.1 | s390x |
| Java | java-websphereliberty-gradle | Java application Gradle-built stack using the WebSphere Liberty runtime | icr.io/appcafe/websphere-liberty-devfile-stack:22.0.0.1-gradle | s390x |
| Java | java-wildfly-bootable-jar | Java stack with WildFly in bootable Jar mode, OpenJDK 11 and Maven 3.5 | registry.access.redhat.com/ubi8/openjdk-11 | s390x |
| JavaScript | nodejs | Stack with Node.js 14 | registry.access.redhat.com/ubi8/nodejs-14:latest | s390x, ppc64le|
| TypeScript | nodejs-angular | Stack with Angular 12 | node:lts-slim | s390x |
| JavaScript | nodejs-nextjs | Stack with Next.js 11 | node:lts-slim | s390x |
| JavaScript | nodejs-nuxtjs | Stack with Nuxt.js 2 | node:lts | s390x |
| JavaScript | nodejs-react | Stack with React 17 | node:lts-slim | s390x |
| JavaScript | nodejs-svelte | Stack with Svelte 3 | node:lts-slim | s390x |
| JavaScript | nodejs-vue | Stack with Vue 3 | node:lts-slim | s390x |
| PHP | php-laravel | Stack with Laravel 8 | composer:2.1.11 | s390x |
| Python | python | Python Stack with Python 3.7 | registry.redhat.io/codeready-workspaces/plugin-java8-rhel8 | s390x, ppc64le|
| Python | python-django| Python3.7 with Django| registry.redhat.io/codeready-workspaces/plugin-java8-rhel8 | s390x, ppc64le|

**Note**: Access to the Red Hat registry is required to use these images on IBM Power Systems & IBM Z.

[//]: # (Steps to use devfiles can be found in Deploying your first devfile)