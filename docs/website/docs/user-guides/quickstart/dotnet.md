---
title: Developing with .NET
sidebar_position: 3
---

## Step 0. Creating the initial source code (optional)

import InitialSourceCodeInfo from './_initial_source_code.mdx';

<InitialSourceCodeInfo/>

For .NET we will use the [ASP.NET Core MVC](https://docs.microsoft.com/en-us/aspnet/core/tutorials/first-mvc-app/start-mvc?view=aspnetcore-6.0&tabs=visual-studio-code) example. 

ASP.NET MVC is a web application framework that implements the model-view-controller (MVC) pattern.

1. Generate an example project:

```console
dotnet new mvc --name app
```
```console
$ dotnet new mvc --name app
Welcome to .NET 6.0!
---------------------
SDK Version: 6.0.104

...

The template "ASP.NET Core Web App (Model-View-Controller)" was created successfully.
This template contains technologies from parties other than Microsoft, see https://aka.ms/aspnetcore/6.0-third-party-notices for details.

Processing post-creation actions...
Running 'dotnet restore' on /Users/user/app/app.csproj...
  Determining projects to restore...
  Restored /Users/user/app/app.csproj (in 84 ms).
Restore succeeded.
```

Your source code has now been generated and created in the directory.


## Step 1. Connect to your cluster and create a new namespace or project

import ConnectingToCluster from './_connecting_to_cluster.mdx';

<ConnectingToCluster/>

## Step 2. Initializing your application (`odo init`)

import CreatingApp from './_creating_app.mdx';

<CreatingApp name="dotnet" port="8080" language="dotnet" framework=".NET"/>

## Step 3. Developing your application continuously (`odo dev`)

import RunningCommand from './_running_command.mdx';

<RunningCommand name="dotnet" port="8080" language="dotnet" framework=".NET"/>