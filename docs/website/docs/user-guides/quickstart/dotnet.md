---
title: Developing with .NET
sidebar_position: 3
---

## Step 0. Creating the initial source code (optional)

import InitialSourceCodeInfo from './docs-mdx/initial_source_code_description.mdx';

<InitialSourceCodeInfo/>

For .NET we will use the [ASP.NET Core MVC](https://docs.microsoft.com/en-us/aspnet/core/tutorials/first-mvc-app/start-mvc?view=aspnetcore-6.0&tabs=visual-studio-code) example. 

ASP.NET MVC is a web application framework that implements the model-view-controller (MVC) pattern.

1. Generate an example project:

```console
dotnet new mvc --name app --output .
```
<details>
<summary>Example</summary>

```shell
$ dotnet new mvc --name app --output .
Welcome to .NET 6.0!
---------------------
SDK Version: 6.0.104

...

The template "ASP.NET Core Web App (Model-View-Controller)" was created successfully.
This template contains technologies from parties other than Microsoft, see https://aka.ms/aspnetcore/6.0-third-party-notices for details.

Processing post-creation actions...
Running 'dotnet restore' on /home/user/quickstart-demo/app.csproj...
  Determining projects to restore...
  Restored /home/user/quickstart-demo/app.csproj (in 96 ms).
Restore succeeded.

```
</details>

Your source code has now been generated and created in the directory.


## Step 1. Connect to your cluster and create a new namespace or project

import ConnectingToCluster from './docs-mdx/connecting_to_the_cluster_description.mdx';

<ConnectingToCluster/>

## Step 2. Initializing your application (`odo init`)

import InitSampleOutput from './docs-mdx/dotnet/dotnet_odo_init_output.mdx';
import InitDescription from './docs-mdx/odo_init_description.mdx';

<InitDescription framework=".NET" initout=<InitSampleOutput/> />

:::note
When you first run `odo init`, it will detect the required devfile to be 'dotnet50', if this happens to you, please select <b>No</b> when asked <b>Is this correct?</b> and then select <b>.NET 6.0</b> when asked for <b>Select project type:</b>. Take a look at the sample output for a reference.
:::

## Step 3. Developing your application continuously (`odo dev`)

import DevSampleOutput from './docs-mdx/dotnet/dotnet_odo_dev_output.mdx';

import DevDescription from './docs-mdx/odo_dev_description.mdx';

<DevDescription framework=".NET" devout=<DevSampleOutput/> />


_You can now follow the [advanced guide](../advanced/deploy/dotnet.md) to deploy the application to production._
