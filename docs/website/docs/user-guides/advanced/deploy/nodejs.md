---
title: Deploying with Node.JS
sidebar_position: 1
---

## Overview

import Overview from './docs-mdx/prerequisites.mdx';

<Overview/>

## Prerequisites

import PreReq from './docs-mdx/prerequisites.mdx';

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

import EditingDevfile from './docs-mdx/editing_devfile.mdx';

<EditingDevfile name="nodejs" port="3000"/>

import K8sDevfile from './docs-mdx/nodejs/nodejs_final_devfile_kubernetes.mdx';
import OCDevfile from './docs-mdx/nodejs/nodejs_final_devfile_openshift.mdx';
import FinalDevfileDescription from './docs-mdx/final_devfile.mdx';

<FinalDevfileDescription k8sdata=<K8sDevfile /> ocdata=<OCDevfile /> />


## Step 4. Run the `odo deploy` command

import DeployOutput from './docs-mdx/nodejs/nodejs_deploy_output.mdx';

import DeployDescription from './docs-mdx/running_deploy_description.mdx';

<DeployDescription deployout=<DeployOutput /> />


## Step 5. Accessing the application

import AccessingApplicationDescription from './docs-mdx/accessing_application.mdx'
import KubernetesDescribeOutput from './docs-mdx/nodejs/nodejs_describe_component_kubernetes_output.mdx';
import OpenShiftDescribeOutput from './docs-mdx/nodejs/nodejs_describe_component_openshift_output.mdx';

<AccessingApplicationDescription k8sdata=<KubernetesDescribeOutput /> ocdata=<OpenShiftDescribeOutput />/>

## Step 6. Delete the resources

import DeleteOut from './docs-mdx/nodejs/nodejs_delete_component_output.mdx';
import DeleteDescription from './docs-mdx/delete_resources.mdx';

<DeleteDescription deleteout=<DeleteOut /> />