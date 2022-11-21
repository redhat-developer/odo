---
title: Deploying with .NET
sidebar_position: 3
---

## Overview

import Overview from './_overview.mdx';

<Overview/>

## Prerequisites

import PreReq from './_prerequisites.mdx';

<PreReq/>

## Step 1. Create the initial development application

Complete the [Developing with .Net](/docs/user-guides/quickstart/dotnet) guide before continuing.

## Step 2. Containerize the application

In order to deploy our application, we must containerize it in order to build and push to a registry. Create the following `Dockerfile` in the same directory:

```dockerfile
FROM registry.access.redhat.com/ubi8/dotnet-60:6.0 as builder
WORKDIR /opt/app-root/src
COPY --chown=1001 . .
RUN dotnet publish -c Release


FROM registry.access.redhat.com/ubi8/dotnet-60:6.0
EXPOSE 8080
COPY --from=builder /opt/app-root/src/bin /opt/app-root/src/bin
WORKDIR /opt/app-root/src/bin/Release/net6.0/publish
CMD ["dotnet", "app.dll"]
```

## Step 3. Modify the Devfile

import EditingDevfile from './_editing_devfile.mdx';

<EditingDevfile name="dotnet" port="8080"/>


## Step 4. Run the `odo deploy` command

import RunningDeploy from './_running_deploy.mdx';

<RunningDeploy name="dotnet"/>


## Step 5. Accessing the application

import AccessingApplication from './_accessing_application.mdx'

<AccessingApplication name="dotnet" displayName=".NET 6.0" language=".NET" projectType="dotnet" description="Stack with .NET 6.0" tags=".NET" version="1.0.2"/>

## Step 6. Delete the resources

import Delete from './_delete_resources.mdx';

<Delete/>