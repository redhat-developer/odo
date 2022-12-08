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

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

<details>
<summary> Your final Devfile should look something like this:</summary>

:::note
Your Devfile might slightly vary from the example above, but the example should give you an idea about the placements of all the components and commands.
:::
<Tabs groupId="quickstart">
  <TabItem value="kubernetes" label="Kubernetes">

```yaml showLineNumbers
commands:
- exec:
    commandLine: kill $(pidof dotnet); dotnet build -c $CONFIGURATION $STARTUP_PROJECT
      /p:UseSharedCompilation=false
    component: dotnet
    group:
      isDefault: true
      kind: build
    workingDir: ${PROJECT_SOURCE}
  id: build
- exec:
    commandLine: dotnet run -c $CONFIGURATION --no-build --project $STARTUP_PROJECT
      --no-launch-profile
    component: dotnet
    group:
      isDefault: true
      kind: run
    workingDir: ${PROJECT_SOURCE}
  id: run
# highlight-start
# This is the main "composite" command that will run all below commands
- id: deploy
  composite:
    commands:
    - build-image
    - k8s-deployment
    - k8s-service
    - k8s-url
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
- id: k8s-url
  apply:
    component: outerloop-url
# highlight-end
components:
- container:
    args:
    - tail
    - -f
    - /dev/null
    endpoints:
    - name: http-dotnet60
      targetPort: 8080
    env:
    - name: CONFIGURATION
      value: Debug
    - name: STARTUP_PROJECT
      value: app.csproj
    - name: ASPNETCORE_ENVIRONMENT
      value: Development
    - name: ASPNETCORE_URLS
      value: http://*:8080
    image: registry.access.redhat.com/ubi8/dotnet-60:6.0
    mountSources: true
  name: dotnet
# highlight-start
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
        type: NodePort
- name: outerloop-url
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
# highlight-end
metadata:
  description: Stack with .NET 6.0
  displayName: .NET 6.0
  icon: https://github.com/dotnet/brand/raw/main/logo/dotnet-logo.png
  language: .NET
  name: my-dotnet-app
  projectType: dotnet
  tags:
  - .NET
  version: 1.0.2
# highlight-next-line
schemaVersion: 2.2.0
starterProjects:
- git:
    checkoutFrom:
      remote: origin
      revision: dotnet-6.0
    remotes:
      origin: https://github.com/redhat-developer/s2i-dotnetcore-ex
  name: dotnet60-example
  subDir: app
# highlight-start
# Add the following variables code anywhere in devfile.yaml
# This MUST be a container registry you are able to access
variables:
  CONTAINER_IMAGE: quay.io/MYUSERNAME/dotnet-odo-example
  RESOURCE_NAME: my-dotnet-app
  CONTAINER_PORT: "8080"
  DOMAIN_NAME: dotnet.example.com
# highlight-end
```
  </TabItem>
  <TabItem value="openshift" label="OpenShift">


```yaml showLineNumbers
commands:
- exec:
    commandLine: kill $(pidof dotnet); dotnet build -c $CONFIGURATION $STARTUP_PROJECT
      /p:UseSharedCompilation=false
    component: dotnet
    group:
      isDefault: true
      kind: build
    workingDir: ${PROJECT_SOURCE}
  id: build
- exec:
    commandLine: dotnet run -c $CONFIGURATION --no-build --project $STARTUP_PROJECT
      --no-launch-profile
    component: dotnet
    group:
      isDefault: true
      kind: run
    workingDir: ${PROJECT_SOURCE}
  id: run
# highlight-start
# This is the main "composite" command that will run all below commands
- id: deploy
  composite:
    commands:
    - build-image
    - k8s-deployment
    - k8s-service
    - k8s-url
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
- id: k8s-url
  apply:
    component: outerloop-url
# highlight-end
components:
- container:
    args:
    - tail
    - -f
    - /dev/null
    endpoints:
    - name: http-dotnet60
      targetPort: 8080
    env:
    - name: CONFIGURATION
      value: Debug
    - name: STARTUP_PROJECT
      value: app.csproj
    - name: ASPNETCORE_ENVIRONMENT
      value: Development
    - name: ASPNETCORE_URLS
      value: http://*:8080
    image: registry.access.redhat.com/ubi8/dotnet-60:6.0
    mountSources: true
  name: dotnet
# highlight-start
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
        type: NodePort
- name: outerloop-url
  kubernetes:
    inlined: |
      apiVersion: route.openshift.io/v1
      kind: Route
      metadata:
        name: {{RESOURCE_NAME}}
      spec:
        path: /
        to:
          kind: Service
          name: {{RESOURCE_NAME}}
        port:
          targetPort: {{CONTAINER_PORT}}
# highlight-end
metadata:
  description: Stack with .NET 6.0
  displayName: .NET 6.0
  icon: https://github.com/dotnet/brand/raw/main/logo/dotnet-logo.png
  language: .NET
  name: my-dotnet-app
  projectType: dotnet
  tags:
  - .NET
  version: 1.0.2
# highlight-next-line
schemaVersion: 2.2.0
starterProjects:
- git:
    checkoutFrom:
      remote: origin
      revision: dotnet-6.0
    remotes:
      origin: https://github.com/redhat-developer/s2i-dotnetcore-ex
  name: dotnet60-example
  subDir: app
# highlight-start
# Add the following variables code anywhere in devfile.yaml
# This MUST be a container registry you are able to access
variables:
  CONTAINER_IMAGE: quay.io/MYUSERNAME/dotnet-odo-example
  RESOURCE_NAME: my-dotnet-app
  CONTAINER_PORT: "8080"
  DOMAIN_NAME: dotnet.example.com
# highlight-end
```
</TabItem>
</Tabs>

</details>


## Step 4. Run the `odo deploy` command

import RunningDeploy from './_running_deploy.mdx';

<RunningDeploy name="dotnet"/>


## Step 5. Accessing the application

import AccessingApplication from './_accessing_application.mdx'

<AccessingApplication name="dotnet" displayName=".NET 6.0" language=".NET" projectType="dotnet" description="Stack with .NET 6.0" tags=".NET" version="1.0.2"/>

## Step 6. Delete the resources

import Delete from './_delete_resources.mdx';

<Delete/>