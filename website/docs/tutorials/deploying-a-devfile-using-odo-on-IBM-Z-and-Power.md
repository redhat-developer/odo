---
title: Using odo on IBM-Z and Power
sidebar_position: 2
---
### Pre-requisites
1. Read [Architecture > Devfile](../architecture/devfile.md).

### Deploying your first devfile on IBM Z & Power
Since the [DefaultDevfileRegistry](https://github.com/odo-devfiles/registry) doesn't support IBM Z & Power now, you will need to create a secure private DevfileRegistry first. To create a new secure private DevfileRegistry, please check the doc [secure registry](../architecture/secure-registry.md).

The images can be used for devfiles on IBM Z & Power

|Language   | Devfile Name  | Description   | Image Source  | Supported Platform    |
| ----------- | ----------- | ----------- | ----------- | ----------- |
| Java      | java-maven    | Upstream Maven and OpenJDK 11 | registry.redhat.io/codeready-workspaces/plugin-java11-openj9-rhel8 | s390x, ppc64le |
| Java      | java-openliberty | Open Liberty microservice in Java | registry.redhat.io/codeready-workspaces/plugin-java11-openj9-rhel8 | s390x, ppc64le |
| Java | java-quarkus | Upstream Quarkus with Java+GraalVM | registry.redhat.io/codeready-workspaces/plugin-java8-openj9-rhel8 | s390x, ppc64le|
| Java | java-springboot | Spring Boot® using Java| registry.redhat.io/codeready-workspaces/plugin-java11-openj9-rhel8 | s390x, ppc64le|
| Vert.x Java| java-vertx | Upstream Vert.x using Java | registry.redhat.io/codeready-workspaces/plugin-java11-openj9-rhel8 | s390x, ppc64le|
| Node.JS | nodejs | Stack with NodeJS 12 | registry.redhat.io/codeready-workspaces/plugin-java8-openj9-rhel8 | s390x, ppc64le|
| Python| python | Python Stack with Python 3.7 | registry.redhat.io/codeready-workspaces/plugin-java8-openj9-rhel8 | s390x, ppc64le|
| Django| python-django| Python3.7 with Django| registry.redhat.io/codeready-workspaces/plugin-java8-openj9-rhel8| s390x, ppc64le|

**Note**: Access to the Red Hat registry is required to use these images on IBM Power Systems & IBM Z.

Steps to use devfiles can be found in [Deploying your first devfile](deploying-your-first-devfile.md).
