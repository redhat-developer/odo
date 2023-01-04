---
title: Developing with Java (Spring Boot)
sidebar_position: 2
---

## Step 0. Creating the initial source code (optional)

import InitialSourceCodeInfo from './docs-mdx/initial_source_code_description.mdx';

<InitialSourceCodeInfo/>

For Java, we will use the [Spring Initializr](https://start.spring.io/) to generate the example source code:

1. Navigate to [start.spring.io](https://start.spring.io/) 
2. Select **11** under **Java**
3. Click on "Add" under "Dependencies"
4. Select "Spring Web"
5. Click "Generate" to generate and download the source code

Finally, open a terminal and navigate to the directory.

Your source code has now been generated and created in the directory.

## Step 1. Connect to your cluster and create a new namespace or project

import ConnectingToCluster from './docs-mdx/connecting_to_the_cluster_description.mdx';

<ConnectingToCluster/>

## Step 2. Initializing your application (`odo init`)


import InitSampleOutput from './docs-mdx/java/java_odo_init_output.mdx';
import InitDescription from './docs-mdx/odo_init_description.mdx';

<InitDescription framework="Java (Spring Boot)" initout=<InitSampleOutput/> />

## Step 3. Developing your application continuously (`odo dev`)

import DevSampleOutput from './docs-mdx/java/java_odo_dev_output.mdx';

import DevDescription from './docs-mdx/odo_dev_description.mdx';

<DevDescription framework="Java (Spring Boot)" devout=<DevSampleOutput/> />

_You can now follow the [advanced guide](../advanced/deploy/java.md) to deploy the application to production._
