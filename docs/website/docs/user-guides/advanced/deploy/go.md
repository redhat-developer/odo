---
title: Deploying with Go
sidebar_position: 4
---

## Overview

import Overview from './_overview.mdx';

<Overview/>

## Prerequisites

import PreReq from './_prerequisites.mdx';

<PreReq/>

## Step 1. Create the initial development application

Complete the [Developing with Go](/docs/user-guides/quickstart/go) guide before continuing.

## Step 2. Containerize the application

In order to deploy our application, we must containerize it in order to build and push to a registry. Create the following `Dockerfile` in the same directory:

```dockerfile
# This Dockerfile is referenced from:
# https://github.com/GoogleCloudPlatform/golang-samples/blob/main/run/helloworld/Dockerfile

# Build
FROM golang:1.17-buster as builder
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . ./
RUN go build -v -o server

# Create a "lean" container
FROM debian:buster-slim
RUN set -x && apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
    ca-certificates && \
    rm -rf /var/lib/apt/lists/*
COPY --from=builder /app/server /app/server

# Run
CMD ["/app/server"]
```

## Step 3. Modify the Devfile

import EditingDevfile from './_editing_devfile.mdx';

<EditingDevfile name="go" port="8080"/>

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
    commandLine: go build main.go
    component: runtime
    env:
    - name: GOPATH
      value: ${PROJECT_SOURCE}/.go
    - name: GOCACHE
      value: ${PROJECT_SOURCE}/.cache
    group:
      isDefault: true
      kind: build
    workingDir: ${PROJECT_SOURCE}
  id: build
- exec:
    commandLine: ./main
    component: runtime
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
    - name: http-go
      targetPort: 8080
    image: registry.access.redhat.com/ubi9/go-toolset:latest
    memoryLimit: 1024Mi
    mountSources: true
  name: runtime
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
  description: Go is an open source programming language that makes it easy to build
    simple, reliable, and efficient software.
  displayName: Go Runtime
  icon: https://raw.githubusercontent.com/devfile-samples/devfile-stack-icons/main/golang.svg
  language: Go
  name: my-go-app
  projectType: Go
  provider: Red Hat
  tags:
  - Go
  version: 1.0.2
# highlight-start
# Deploy "kind" ID's use schema 2.2.0+
schemaVersion: 2.2.0
# highlight-end
starterProjects:
- description: A Go project with a simple HTTP server
  git:
    checkoutFrom:
      revision: main
    remotes:
      origin: https://github.com/devfile-samples/devfile-stack-go.git
  name: go-starter
# highlight-start
# Add the following variables code anywhere in devfile.yaml
# This MUST be a container registry you are able to access
variables:
  CONTAINER_IMAGE: quay.io/MYUSERNAME/go-odo-example
  RESOURCE_NAME: my-go-app
  CONTAINER_PORT: "8080"
  DOMAIN_NAME: go.example.com
# highlight-end
```
  </TabItem>
  <TabItem value="openshift" label="OpenShift">


```yaml showLineNumbers
commands:
- exec:
    commandLine: go build main.go
    component: runtime
    env:
    - name: GOPATH
      value: ${PROJECT_SOURCE}/.go
    - name: GOCACHE
      value: ${PROJECT_SOURCE}/.cache
    group:
      isDefault: true
      kind: build
    workingDir: ${PROJECT_SOURCE}
  id: build
- exec:
    commandLine: ./main
    component: runtime
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
    - name: http-go
      targetPort: 8080
    image: registry.access.redhat.com/ubi9/go-toolset:latest
    memoryLimit: 1024Mi
    mountSources: true
  name: runtime
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
  description: Go is an open source programming language that makes it easy to build
    simple, reliable, and efficient software.
  displayName: Go Runtime
  icon: https://raw.githubusercontent.com/devfile-samples/devfile-stack-icons/main/golang.svg
  language: Go
  name: my-go-app
  projectType: Go
  provider: Red Hat
  tags:
  - Go
  version: 1.0.2
# highlight-start
# Deploy "kind" ID's use schema 2.2.0+
schemaVersion: 2.2.0
# highlight-end
starterProjects:
- description: A Go project with a simple HTTP server
  git:
    checkoutFrom:
      revision: main
    remotes:
      origin: https://github.com/devfile-samples/devfile-stack-go.git
  name: go-starter
# highlight-start
# Add the following variables code anywhere in devfile.yaml
# This MUST be a container registry you are able to access
variables:
  CONTAINER_IMAGE: quay.io/MYUSERNAME/go-odo-example
  RESOURCE_NAME: my-go-app
  CONTAINER_PORT: "8080"
  DOMAIN_NAME: go.example.com
# highlight-end
```
</TabItem>
</Tabs>

</details>

## Step 4. Run the `odo deploy` command

import RunningDeploy from './_running_deploy.mdx';

<RunningDeploy name="go"/>

## Step 5. Accessing the application

import AccessingApplication from './_accessing_application.mdx'

<AccessingApplication name="go" displayName="Go Runtime" language="Go" projectType="Go" description="Go is an open source programming language that makes it easy to build simple, reliable, and efficient software." tags="Go"/>

## Step 6. Delete the resources

import Delete from './_delete_resources.mdx';

<Delete/>