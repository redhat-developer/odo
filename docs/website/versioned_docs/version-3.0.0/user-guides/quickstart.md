---
title: Quickstart Guide
sidebar_position: 5
---

# Quickstart Guide

In this guide, we will be using odo to create a "Hello World" application.

You have the option of choosing from the following frameworks for the quickstart guide:
* Node.js
* .NET

A full list of example applications can be viewed with the `odo registry` command.

## Prerequisites

* Have the odo binary [installed](../overview/installation.md).
* A [Kubernetes](../overview/cluster-setup/kubernetes) or [OpenShift cluster](../overview/cluster-setup/openshift) 

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

## Step 0. Creating the initial source code (optional)

We will create the example source code by using some popular frameworks.

This is *optional* and you may use an existing project instead or a starter project from `odo init`.

<Tabs groupId="quickstart">
  <TabItem value="nodejs" label="Node.js">

For Node.JS we will use the [Express](https://expressjs.com/) framework for our example.

1. Install Express:
```console
$ npm install express --save
```

2. Generate an example project:
```console
$ npx express-generator

  warning: the default view engine will not be jade in future releases
  warning: use `--view=jade' or `--help' for additional options


   create : public/
   create : public/javascripts/
   create : public/images/
   create : public/stylesheets/
   create : public/stylesheets/style.css
   create : routes/
   create : routes/index.js
   create : routes/users.js
   create : views/
   create : views/error.jade
   create : views/index.jade
   create : views/layout.jade
   create : app.js
   create : package.json
   create : bin/
   create : bin/www

   install dependencies:
     $ npm install

   run the app:
     $ DEBUG=express:* npm start
```


</TabItem>
  <TabItem value="dotnet" label=".NET">

  For .NET we will use the [ASP.NET Core MVC](https://docs.microsoft.com/en-us/aspnet/core/tutorials/first-mvc-app/start-mvc?view=aspnetcore-6.0&tabs=visual-studio-code) example. 

  ASP.NET MVC is a web application framework that implements the model-view-controller (MVC) pattern.

  1. Generate an example project:

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

  </TabItem>
</Tabs>

Your source code has now been generated and created in the directory.

## Step 1. Creating your application (`odo init`)

Now we'll initialize your application by creating a `devfile.yaml` to be deployed.

`odo` handles this automatically with the `odo init` command by autodetecting your source code and downloading the appropriate Devfile.

**Note:** If you skipped *Step 0*, select a "starter project" when running `odo init`.

<Tabs groupId="quickstart">
  <TabItem value="nodejs" label="Node.js">

Let's run `odo init` and select Node.js:

```console
$ odo init
  __
 /  \__     Initializing new component
 \__/  \    Files: Source code detected, a Devfile will be determined based upon source code autodetection
 /  \__/    odo version: v3.0.0-alpha2
 \__/

Interactive mode enabled, please answer the following questions:
Based on the files in the current directory odo detected
Language: javascript
Project type: nodejs
The devfile "nodejs" from the registry "DefaultDevfileRegistry" will be downloaded.
? Is this correct? Yes
 ✓  Downloading devfile "nodejs" from registry "DefaultDevfileRegistry" [501ms]
Current component configuration:
Container "runtime":
  Opened ports:
   - 3000
  Environment variables:
? Select container for which you want to change configuration? NONE - configuration is correct
? Enter component name: my-nodejs-app

Your new component 'my-nodejs-app' is ready in the current directory.
To start editing your component, use 'odo dev' and open this folder in your favorite IDE.
Changes will be directly reflected on the cluster.
```

A `devfile.yaml` has now been added to your directory and now you're ready to start development.

  </TabItem>
  <TabItem value="dotnet" label=".NET">

Let's run `odo init` and select .NET 6.0:

```console
$ odo init
  __
 /  \__     Initializing new component
 \__/  \    Files: Source code detected, a Devfile will be determined based upon source code autodetection
 /  \__/    odo version: v3.0.0-alpha2
 \__/

Interactive mode enabled, please answer the following questions:
Based on the files in the current directory odo detected
Language: dotnet
Project type: dotnet
The devfile "dotnet50" from the registry "DefaultDevfileRegistry" will be downloaded.
? Is this correct? No
? Select language: dotnet
? Select project type: .NET 6.0
 ✓  Downloading devfile "dotnet60" from registry "DefaultDevfileRegistry" [596ms]
Current component configuration:
Container "dotnet":
  Opened ports:
   - 8080
  Environment variables:
   - STARTUP_PROJECT = app.csproj
   - ASPNETCORE_ENVIRONMENT = Development
   - ASPNETCORE_URLS = http://*:8080
   - CONFIGURATION = Debug
? Select container for which you want to change configuration? NONE - configuration is correct
? Enter component name: my-dotnet60-app

Your new component 'my-dotnet60-app' is ready in the current directory.
To start editing your component, use 'odo dev' and open this folder in your favorite IDE.
Changes will be directly reflected on the cluster.
```

A `devfile.yaml` has now been added to your directory and now you're ready to start development.
  </TabItem>
</Tabs>

## Step 2. Developing your application continuously (`odo dev`)

Now that we've generated our code as well as our Devfile, let's start on development.

`odo` uses inner loop development and allows you to code, build, run and test the application in a continuous workflow.

Once you run `odo dev`, you can freely edit code in your favourite IDE and watch as `odo` rebuilds and redeploys it.

<Tabs groupId="quickstart">
  <TabItem value="nodejs" label="Node.js">

Let's run `odo dev` to start development on your Node.JS application:
```console
$ odo dev
  __
 /  \__     Developing using the my-nodejs-app Devfile
 \__/  \    Namespace: default
 /  \__/    odo version: v3.0.0-alpha2
 \__/

↪ Deploying to the cluster in developer mode
 ✓  Waiting for Kubernetes resources [3s]
 ✓  Syncing files into the container [330ms]
 ✓  Building your application in container on cluster [4s]
 ✓  Executing the application [1s]

Your application is now running on the cluster
 - Forwarding from 127.0.0.1:40001 -> 3000

Watching for changes in the current directory /Users/user/express
Press Ctrl+c to exit `odo dev` and delete resources from the cluster
```
  </TabItem>

  <TabItem value="dotnet" label=".NET">

Let's run `odo dev` to start development on your .NET application:

```console
$ odo dev
  __
 /  \__     Developing using the my-dotnet60-app Devfile
 \__/  \    Namespace: default
 /  \__/    odo version: v3.0.0-alpha2
 \__/

↪ Deploying to the cluster in developer mode
 ✓  Waiting for Kubernetes resources [3s]
 ✓  Syncing files into the container [2s]
 ✓  Building your application in container on cluster [5s]
 ✓  Executing the application [1s]

Your application is now running on the cluster
 - Forwarding from 127.0.0.1:40001 -> 8080

Watching for changes in the current directory /Users/user/dotnet
Press Ctrl+c to exit `odo dev` and delete resources from the cluster
```


  </TabItem>
</Tabs>

You can now access the application at [127.0.0.1:40001](http://127.0.0.1:40001) in your local browser and start your development loop. `odo` will watch for changes and push the code for real-time updates.


## Step 3. Deploying your application to the world (`odo deploy`)

**Prerequisites:**

Before we begin, you must login to a container registry that we will be pushing our application to.

Login to your container registry with either `podman` or `docker`:

```console
$ podman login
# or
$ docker login
```

**NOTE:** If you are running Apple Silicon (M1), you must set your Docker build platform to the cluster you are deploying to.

For example, if you are deploying to `linux/amd64`:

```console
export DOCKER_DEFAULT_PLATFORM=linux/amd64  
```

**Overview:**

There are three steps to deploy your application:

1. Containerize your application by creating a `Dockerfile`
2. Modify `devfile.yaml` to add your Kubernetes code
3. Run `odo deploy`

<Tabs groupId="quickstart">
  <TabItem value="nodejs" label="Node.js">

#### 1. Containerize the application

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

#### 2. Modify the Devfile

Let's modify the `devfile.yaml` and add the respective deployment code.

`odo deploy` uses Devfile schema **2.2.0**. Change the schema to reflect the change:

```yaml
# Deploy "kind" ID's use schema 2.2.0+
schemaVersion: 2.2.0
```

Add the `variables` section:

```yaml
# Add the following variables code anywhere in devfile.yaml
# This MUST be a container registry you are able to access
variables:
  CONTAINER_IMAGE: quay.io/MYUSERNAME/nodejs-odo-example
  RESOURCE_NAME: my-nodejs-app
  CONTAINER_PORT: "3000"
  DOMAIN_NAME: nodejs.example.com
```

Add the commands used to deploy:

```yaml
# This is the main "composite" command that will run all below commands
- id: deploy
  composite:
    commands:
    - build-image
    - k8s-deployment
    - k8s-service
    - k8s-ingress
    group:
      isDefault: true
      kind: deploy

# Below are the commands and their respective components that they are "linked" to deploy
- id: build-image
  apply:
    component: outerloop-build
- id: k8s-deployment
  apply:
    component: outerloop-deployment
- id: k8s-service
  apply:
    component: outerloop-service
- id: k8s-ingress
  apply:
    component: outerloop-ingress
```

Add the Kubernetes Service and Ingress inline code to `components`:
```yaml
components:

# This will build the container image before deployment
- name: outerloop-build
  image:
    dockerfile:
      buildContext: ${PROJECT_SOURCE}
      rootRequired: false
      uri: ./Dockerfile
    imageName: "{{CONTAINER_IMAGE}}"

# This will create a Deployment in order to run your container image across
# the cluster.
- name: outerloop-deployment
  kubernetes:
    inlined: |
      kind: Deployment
      apiVersion: apps/v1
      metadata:
        name: {{RESOURCE_NAME}}
      spec:
        replicas: 1
        selector:
          matchLabels:
            app: {{RESOURCE_NAME}}
        template:
          metadata:
            labels:
              app: {{RESOURCE_NAME}}
          spec:
            containers:
              - name: {{RESOURCE_NAME}}
                image: {{CONTAINER_IMAGE}}
                ports:
                  - name: http
                    containerPort: {{CONTAINER_PORT}}
                    protocol: TCP
                resources:
                  limits:
                    memory: "1024Mi"
                    cpu: "500m"

# This will create a Service so your Deployment is accessible.
# Depending on your cluster, you may modify this code so it's a
# NodePort, ClusterIP or a LoadBalancer service.
- name: outerloop-service
  kubernetes:
    inlined: |
      apiVersion: v1
      kind: Service
      metadata:
        name: {{RESOURCE_NAME}}
      spec:
        ports:
        - name: "{{CONTAINER_PORT}}"
          port: {{CONTAINER_PORT}}
          protocol: TCP
          targetPort: {{CONTAINER_PORT}}
        selector:
          app: {{RESOURCE_NAME}}
        type: ClusterIP

# Let's create an Ingress so we can access the application via a domain name
- name: outerloop-ingress
  kubernetes:
    inlined: |
      apiVersion: networking.k8s.io/v1
      kind: Ingress
      metadata:
        name: {{RESOURCE_NAME}}
      spec:
        rules:
          - host: "{{DOMAIN_NAME}}"
            http:
              paths:
                - path: "/"
                  pathType: Prefix
                  backend:
                    service:
                      name: {{RESOURCE_NAME}} 
                      port:
                        number: {{CONTAINER_PORT}}
```


#### 3. Run the `odo deploy` command

Now we're ready to run `odo deploy`:

```console
$ odo deploy
  __
 /  \__     Deploying the application using my-nodejs-app Devfile
 \__/  \    Namespace: default
 /  \__/    odo version: v3.0.0-alpha2
 \__/

↪ Building & Pushing Container: MYUSERNAME/test
 •  Building image locally  ...
 ✓  Building image locally [880ms]
 •  Pushing image to container registry  ...
 ✓  Pushing image to container registry [5s]

↪ Deploying Kubernetes Component: nodejs-example
 ✓  Searching resource in cluster
 ✓  Creating kind Deployment [48ms]

↪ Deploying Kubernetes Component: nodejs-example
 ✓  Searching resource in cluster
 ✓  Creating kind Service [51ms]

↪ Deploying Kubernetes Component: nodejs-example
 ✓  Searching resource in cluster
 ✓  Creating kind Ingress [49ms]

Your Devfile has been successfully deployed
```

Your application has now been deployed to the Kubernetes cluster with Deployment, Service, and Ingress resources.

Test your application by visiting the `DOMAIN_NAME` variable that you had set in the `devfile.yaml`.

  </TabItem>
  <TabItem value="dotnet" label=".NET">

#### 1. Containerize the application

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

#### 2. Modify the Devfile

Let's modify the `devfile.yaml` and add the respective deployment code.

`odo deploy` uses Devfile schema **2.2.0**. Change the schema to reflect the change:

```yaml
# Deploy "kind" ID's use schema 2.2.0+
schemaVersion: 2.2.0
```

Add the `variables` section:

```yaml
# Add the following variables code anywhere in devfile.yaml
# This MUST be a container registry you are able to access
variables:
  CONTAINER_IMAGE: quay.io/MYUSERNAME/dotnet-odo-example
  RESOURCE_NAME: my-dotnet-app
  CONTAINER_PORT: "8080"
  DOMAIN_NAME: dotnet.example.com
```

Add the commands used to deploy:

```yaml
# This is the main "composite" command that will run all below commands
- id: deploy
  composite:
    commands:
    - build-image
    - k8s-deployment
    - k8s-service
    - k8s-ingress
    group:
      isDefault: true
      kind: deploy

# Below are the commands and their respective components that they are "linked" to deploy
- id: build-image
  apply:
    component: outerloop-build
- id: k8s-deployment
  apply:
    component: outerloop-deployment
- id: k8s-service
  apply:
    component: outerloop-service
- id: k8s-ingress
  apply:
    component: outerloop-ingress
```

Add the Kubernetes Service and Ingress inline code to `components`:
```yaml
components:

# This will build the container image before deployment
- name: outerloop-build
  image:
    dockerfile:
      buildContext: ${PROJECT_SOURCE}
      rootRequired: false
      uri: ./Dockerfile
    imageName: "{{CONTAINER_IMAGE}}"

# This will create a Deployment in order to run your container image across
# the cluster.
- name: outerloop-deployment
  kubernetes:
    inlined: |
      kind: Deployment
      apiVersion: apps/v1
      metadata:
        name: {{RESOURCE_NAME}}
      spec:
        replicas: 1
        selector:
          matchLabels:
            app: {{RESOURCE_NAME}}
        template:
          metadata:
            labels:
              app: {{RESOURCE_NAME}}
          spec:
            containers:
              - name: {{RESOURCE_NAME}}
                image: {{CONTAINER_IMAGE}}
                ports:
                  - name: http
                    containerPort: {{CONTAINER_PORT}}
                    protocol: TCP
                resources:
                  limits:
                    memory: "1024Mi"
                    cpu: "500m"

# This will create a Service so your Deployment is accessible.
# Depending on your cluster, you may modify this code so it's a
# NodePort, ClusterIP or a LoadBalancer service.
- name: outerloop-service
  kubernetes:
    inlined: |
      apiVersion: v1
      kind: Service
      metadata:
        name: {{RESOURCE_NAME}}
      spec:
        ports:
        - name: "{{CONTAINER_PORT}}"
          port: {{CONTAINER_PORT}}
          protocol: TCP
          targetPort: {{CONTAINER_PORT}}
        selector:
          app: {{RESOURCE_NAME}}
        type: ClusterIP

# Let's create an Ingress so we can access the application via a domain name
- name: outerloop-ingress
  kubernetes:
    inlined: |
      apiVersion: networking.k8s.io/v1
      kind: Ingress
      metadata:
        name: {{RESOURCE_NAME}}
      spec:
        rules:
          - host: "{{DOMAIN_NAME}}"
            http:
              paths:
                - path: "/"
                  pathType: Prefix
                  backend:
                    service:
                      name: {{RESOURCE_NAME}} 
                      port:
                        number: {{CONTAINER_PORT}}
```


#### 3. Run the `odo deploy` command

Now we're ready to run `odo deploy`:

```console
$ odo deploy
  __
 /  \__     Deploying the application using my-dotnet-app Devfile
 \__/  \    Namespace: default
 /  \__/    odo version: v3.0.0-alpha2
 \__/

↪ Building & Pushing Container: MYUSERNAME/test
 •  Building image locally  ...
 ✓  Building image locally [880ms]
 •  Pushing image to container registry  ...
 ✓  Pushing image to container registry [5s]

↪ Deploying Kubernetes Component: dotnet-example
 ✓  Searching resource in cluster
 ✓  Creating kind Deployment [48ms]

↪ Deploying Kubernetes Component: dotnet-example
 ✓  Searching resource in cluster
 ✓  Creating kind Service [51ms]

↪ Deploying Kubernetes Component: dotnet-example
 ✓  Searching resource in cluster
 ✓  Creating kind Ingress [49ms]

Your Devfile has been successfully deployed
```

Your application has now been deployed to the Kubernetes cluster with Deployment, Service, and Ingress resources.

Test your application by visiting the `DOMAIN_NAME` variable that you had set in the `devfile.yaml`.

  </TabItem>
</Tabs>
