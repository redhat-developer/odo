---
title: Deploying with Node.JS
sidebar_position: 1
---

## Overview

import Overview from './_overview.mdx';

<Overview/>

## Prerequisites

import PreReq from './_prerequisites.mdx';

<PreReq/>

## Step 1. Create the initial development application

Complete the [Developing with Node.JS](/docs/user-guides/quickstart/nodejs) guide before continuing.

## Step 2. Containerize the application

In order to deploy our application, we must containerize it in order to build and push to a registry. Create the following `Dockerfile` in the same directory:

```dockerfile
# Sample copied from https://github.com/nodeshift-starters/devfile-sample/blob/main/Dockerfile

# Install the app dependencies in a full Node docker image
FROM registry.access.redhat.com/ubi8/nodejs-14:latest

# Copy package.json and package-lock.json
COPY package*.json ./

# Install app dependencies
RUN npm install --production

# Copy the dependencies into a Slim Node docker image
FROM registry.access.redhat.com/ubi8/nodejs-14-minimal:latest

# Install app dependencies
COPY --from=0 /opt/app-root/src/node_modules /opt/app-root/src/node_modules
COPY . /opt/app-root/src

ENV NODE_ENV production
ENV PORT 3000

CMD ["npm", "start"]
```

## Step 3. Modify the Devfile

import EditingDevfile from './_editing_devfile.mdx';

<EditingDevfile name="nodejs" port="3000"/>


## Step 4. Run the `odo deploy` command

import RunningDeploy from './_running_deploy.mdx';

<RunningDeploy name="nodejs"/>

## Step 5. Accessing the application

import AccessingApplication from './_accessing_application.mdx'

<AccessingApplication name="node" displayName="Node.js Runtime" language="JavaScript" projectType="Node.js" description="Stack with Node.js 16" tags="Node.js, Express, ubi8" version="2.1.1"/>

## Step 6. Delete the resources

import Delete from './_delete_resources.mdx';

<Delete/>