---
title: Deploying with .NET
sidebar_position: 3
---

## Overview

import Overview from './docs-mdx/overview.mdx';

<Overview/>

## Prerequisites

import PreReq from './docs-mdx/prerequisites.mdx';

<PreReq/>

## Step 1. Create the initial development application

Complete the [Developing with .Net](/docs/user-guides/quickstart/dotnet) guide before continuing.

## Step 2. Containerize the application

In order to deploy our application, we must containerize it in order to build and push to a registry. Create the following `Dockerfile` in the same directory:
import Dockerfile from './docs-mdx/dotnet/dotnet_Dockerfile.mdx';

<Dockerfile />

## Step 3. Modify the Devfile

import EditingDevfile from './docs-mdx/editing_devfile.mdx';

<EditingDevfile name="dotnet" port="8080"/>

import K8sDevfile from './docs-mdx/dotnet/dotnet_final_devfile_kubernetes.mdx';
import OCDevfile from './docs-mdx/dotnet/dotnet_final_devfile_openshift.mdx';
import FinalDevfileDescription from './docs-mdx/final_devfile.mdx';

<FinalDevfileDescription k8sdata=<K8sDevfile /> ocdata=<OCDevfile /> />


## Step 4. Run the `odo deploy` command

import DeployOutput from './docs-mdx/dotnet/dotnet_deploy_output.mdx';

import DeployDescription from './docs-mdx/running_deploy_description.mdx';

<DeployDescription deployout=<DeployOutput /> />


## Step 5. Accessing the application

import AccessingApplicationDescription from './docs-mdx/accessing_application.mdx'
import KubernetesDescribeOutput from './docs-mdx/dotnet/dotnet_describe_component_kubernetes_output.mdx';
import OpenShiftDescribeOutput from './docs-mdx/dotnet/dotnet_describe_component_openshift_output.mdx';

<AccessingApplicationDescription k8sdata=<KubernetesDescribeOutput /> ocdata=<OpenShiftDescribeOutput />/>

## Step 6. Delete the resources

import DeleteOut from './docs-mdx/dotnet/dotnet_delete_component_output.mdx';
import DeleteDescription from './docs-mdx/delete_resources.mdx';

<DeleteDescription deleteout=<DeleteOut /> />