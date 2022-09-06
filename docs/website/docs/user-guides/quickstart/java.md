---
title: Developing with Java (Spring Boot)
sidebar_position: 2
---

## Step 0. Creating the initial source code (optional)

import InitialSourceCodeInfo from './_initial_source_code.mdx';

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

import ConnectingToCluster from './_connecting_to_cluster.mdx';

<ConnectingToCluster/>

## Step 2. Initializing your application (`odo init`)

import CreatingApp from './_creating_app.mdx';

<CreatingApp name="java" port="8080" language="java" framework="Java (Spring Boot)"/>

## Step 3. Developing your application continuously (`odo dev`)

import RunningCommand from './_running_command.mdx';

<RunningCommand name="java" port="8080" language="java" framework="Java (Spring Boot)"/>