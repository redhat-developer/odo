---
title: Deploying with Java (Spring Boot)
sidebar_position: 2
---

## Overview

import Overview from './_overview.mdx';

<Overview/>

## Prerequisites

import PreReq from './_prerequisites.mdx';

<PreReq/>

## Step 1. Create the initial development application

Complete the [Developing with Java (Spring Boot)](/docs/user-guides/quickstart/java) guide before continuing.

## Step 2. Containerize the application

In order to deploy our application, we must containerize it in order to build and push to a registry. Create the following `Dockerfile` in the same directory:

```dockerfile
FROM registry.access.redhat.com/ubi8/openjdk-11 as builder

USER jboss
WORKDIR /tmp/src
COPY --chown=jboss . /tmp/src
RUN mvn package

FROM registry.access.redhat.com/ubi8/openjdk-11
COPY --from=builder /tmp/src/target/*.jar /deployments/app.jar
```

## Step 3. Modify the Devfile

import EditingDevfile from './_editing_devfile.mdx';

<EditingDevfile name="java" port="8080"/>


## Step 4. Run the `odo deploy` command

import RunningDeploy from './_running_deploy.mdx';

<RunningDeploy name="java"/>

## Step 5. Accessing the application

import AccessingApplication from './_accessing_application.mdx'

<AccessingApplication name="java" displayName="Spring Boot®" language="Java" projectType="springboot" description="Spring Boot® using Java" tags="Java, Spring Boot" version="1.2.0"/>

## Step 6. Delete the resources

import Delete from './_delete_resources.mdx';

<Delete/>