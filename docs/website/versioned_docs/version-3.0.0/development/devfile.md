---
title: Devfile Reference
sidebar_position: 30
---

## Description

### Overview

We implement the latest [Devfile 2.2.0 reference](https://devfile.io/docs/devfile/2.2.0/user-guide/api-reference/) in `odo`.

Below are `odo`-specific examples regarding the Devfile reference. Everything we have listed below is implemented in the current version of `odo`.

This API Reference uses examples and snippets to help you create your own `devfile.yaml` file for your project.

For this reference, we use one singular (working) Node.js example which can be ran from your own environment.

### How `odo` Creates Resources from Devfile

When `odo dev` is ran, `odo` creates the following Kubernetes resources from `container` (component) and `exec` (command) sections of the Devfile:
* Deployment for running the containers
* Service for accessibility 

### File Synchronization

We use the "inner-loop" process intensively in order to propagate and see changes immediately so you spend less time deploying, and more time writing code. 

Below is the loop that `odo` uses for file synchronization:

1. Deployment and Service resources are created (or updated) on the Kubernetes cluster
2. Files are transfered over to the running container using the Kubernetes API 
3. `odo` watches for file changes locally and repeats Step 1 and 2.

The "loop" is cancelled once the user inputs Ctrl+C.

#### Hot Reloading

`hotReloadCapable` is a special boolean within `exec` that allows you to specify if a framework is "hot reloadable".

If set to `true`, the container won't be restarted as the framework will handle file changes on its own.

#### Full Example

```yaml
commands:
  - exec:
      commandLine: yarn dev
      component: runtime
      group:
        isDefault: true
        kind: run
      hotReloadCapable: true
      workingDir: ${PROJECT_SOURCE}
    id: build
```


### What Commands are Executed in Dev or Deploy

Each command has a group `kind`:

```yaml
- exec:
    commandLine: npm start
    component: runtime
    group:
      isDefault: true
      kind: run
    workingDir: ${PROJECT_SOURCE}
  id: run
```

When `odo dev` is executed, we use: build, run, test and debug.

These commands are typically ran within the `container` component that has been defined.

The order in which the commands are ran for `odo dev` are:

0. the sources are synchronized to the container
1. `build`: We build the program from the sources
2. `test`: NOT YET IMPLEMENTED?
3. `run`: The application is ran within the container
4. `debug`: This is ran when `odo dev --debug` is executed

When `odo deploy` is executed, we use: deploy.

These commands are typically tied to Kubernetes or OpenShift inline resources. They are defined as a component. However, you can use `container` or `image` components as well under the deploy group.

## File Reference

This file reference outlines the **major** components of the Devfile API Reference using *snippets* and *examples*.

We use practical approach outlining the details of Devfile. The example is a modified version from [registry/nodejs](https://github.com/devfile/registry/blob/main/stacks/nodejs/devfile.yaml) on GitHub for demonstration purposes.

In this example, we are deploying a full Node.js application that is available through both `odo dev` and `odo deploy`. 

For a more in-depth detailed outline of Devfile, please refer to the official [Devfile 2.2.0 reference](https://devfile.io/docs/devfile/2.2.0/user-guide/api-reference/).

**NOTE:** Some portions of the Devfile examples are commented out to show what is *available* but it does not apply to the practical example.

### Special Variables

There are two special variables available to use in `devfile.yaml`:
* `$PROJECTS_ROOT`: A path where projects sources are mounted as defined by container component's sourceMapping.
* `$PROJECT_SOURCE`: A path to a project source (`$PROJECTS_ROOT/`). If there are multiple projects, this will point to the directory of the first one.

These are helpful when defining or using multiple sources during development or deployment.

### `schemaVersion`

> `type: string` | **required**

Devfile schema version. This outlines what version of the schema `odo` will use and validate against.

#### Full Example

```yaml
schemaVersion: 2.2.0
```

### `metadata`

> `type: object` | **required**

The metadata of the Devfile describes what the following Devfile is about. Information displayed here is helpful to figure out what the project does.

#### Full Example

We are describing a simple Node.js app. This information is displayed when running `odo registry`.

```yaml
metadata:
  description: Stack with Node.js 14
  name: my-nodejs-app
  displayName: Node.js Runtime
  icon: https://nodejs.org/static/images/logos/nodejs-new-pantone-black.svg
  language: javascript
  projectType: nodejs
  tags:
  - NodeJS
  - Express
  - ubi8
  version: 1.0.1
```

### `starterProjects`

> `type: object` | **optional**

StarterProjects is a project that can be used as a starting point when bootstrapping new projects.

When running `odo init`, `odo` looks at the `starterProjects` object to see if there are any available to initially generate.

We support using both `git` and `zip` file formats.

#### Full Example

We will use the starter project from [odo-devfiles/nodejs-ex](https://github.com/odo-devfiles/nodejs-ex) on GitHub.

```yaml
starterProjects:
- git:
    remotes:
      origin: https://github.com/odo-devfiles/nodejs-ex.git
    # You can also checkout from a specific remote as well as revision / tag
    # checkoutFrom:
    #  remote: subproject
    #  revision: 1.1.0Final
    
    # As well as a subdirectory
    # subDir: demo
  name: nodejs-starter

# Alternatively, you can also provide a zipfile location
- zip:
    location: https://github.com/odo-devfiles/nodejs-ex/archive/refs/tags/0.0.2.zip
  name: nodejs-zip-starter
```

### `variables`

> `type: object` | **optional**

Map of key-value variables used for string replacement in the Devfile.

This is important so that you can easily replace variables and make Devfile more *portable* between different environments.

#### Full Example

**NOTE:** For the below example to run correctly, change the `CONTAINER_IMAGE` to an accessible container registry.

```yaml
variables:
  CONTAINER_IMAGE: quay.io/MYUSERNAME/nodejs-odo-example
  RESOURCE_NAME: my-nodejs-app
  CONTAINER_PORT: "3000"
  DOMAIN_NAME: nodejs.example.com
```


### `commands`
> `type: object` | **required**

Predefined, ready-to-use, devworkspace-related commands

Each command is categorized into five groups: build, run, test, debug and deploy.

For `odo dev`, we use: build, run, test and debug.

For `odo deploy`, we use: deploy.

For each command categorized, there are three different "properties" you can use, either: `exec`, `apply` or `composite`:
  * `exec`: A CLI command is executed inside an existing component container
  * `apply`: A command that consists of "applying" a component definition. Typically Kubernetes inline code.
  * `composite`: A composite command that allows executing several sub-commands either sequentially or concurrently

Each command is ran at different points of `odo`'s execution and is further explained [here](#what-commands-are-executed-in-dev-or-deploy).

#### Full Example using `odo dev` capabilities

In the below example, we showcase what commands we would run for `odo dev`:

```yaml

# We will go into detail on the "build" part
# 
# All of these commands are ran in a *container* and at different stages
# of `odo dev`
commands:
- exec:
    # The command you want to run
    commandLine: npm install

    # Which component we are using to run the above command
    # The component contains the container image we are using
    component: runtime

    # What group we are running this in, in our case it's "build"
    group:

      # Identifies the default command for a given group kind as there can be multiple
      # exec's for one kind
      isDefault: true
      # Executes in the "build" stage of dev
      kind: build

    # The working directory being used (see "Special Variables")
    workingDir: ${PROJECT_SOURCE}

  # The to refer to
  id: install

# Executes in the "run" stage of dev
- exec:
    commandLine: npm start
    component: runtime
    group:
      isDefault: true
      kind: run
    workingDir: ${PROJECT_SOURCE}
  id: run

# Executes in the "debug" stage of dev
- exec:
    commandLine: npm run debug
    component: runtime
    group:
      isDefault: true
      kind: debug
    workingDir: ${PROJECT_SOURCE}
  id: debug

# Executes in the "test" stage of dev
- exec:
    commandLine: npm test
    component: runtime
    group:
      isDefault: true
      kind: test
    workingDir: ${PROJECT_SOURCE}
  id: test
```

#### Full Example using `odo deploy` capabilities

In the below example, we showcase what commands we would run for `odo deploy`:

```yaml
# This is the main "composite" command that will run all below commands
commands:
- id: deploy
  composite:
    # In this composite, we will run the following commands:
    commands:
    - build-image
    - k8s-deployment
    - k8s-service
    - k8s-ingress
    
    # This is part of the deploy group and will be executed when running `odo deploy`
    group:
      isDefault: true
      kind: deploy

# In order to run all the commands under composite, we must also define each command:
# Below are the commands and their respective components that they are "linked" to deploy
- id: build-image
  # In this case, we will apply the following outerloop-build component 
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

### `components`

> `type: object` | **required**

List of the devworkspace components, such as editor and plugins, user-provided containers, or other types of components.

Each component can be as simple as container to a full Kubernetes inline object.

There are four different kinds of components:
* `container`: Allows adding and configuring devworkspace-related containers
* `image`: Allows specifying the definition of an image for outer loop builds
* `kubernetes`: Allows importing into the devworkspace the Kubernetes resources defined in a given manifest
* `openshift`: Allows importing into the devworkspace the OpenShift resources defined in a given manifest 

#### Full Example using `odo dev` capabilities

In the below example, we utilize one container component for all `odo dev` commands (build, run, test and debug):

```yaml
components:
- container:
    # A list of all endpoints to be used.
    endpoints:
    - name: http-3000
      targetPort: 3000

    # The container image
    image: registry.access.redhat.com/ubi8/nodejs-14:latest

    # Arguments to be passed into the container by default
    # args: "foo", "bar"

    # The default command to be ran
    # command: "foo", "bar"

    # Environment variables to be passed into the container
    # env:
    #  HELLO: WORLD

    # Limits that can be set
    memoryLimit: 1024Mi
    # memoryRequest: 2048Mi
    # cpuRequest: 250m
    # cpuLimit: 500m

    # Specify if a container should run in its own separated pod, instead of running as part of the main development environment pod.
    # By default, this is set to false
    dedicatedPod: false

    # Annotations can also be passed in too to the Deployment and Service created by this container
    # annotation:
    #   service:
    #     FOO: BAR
    #   deployment:
    #     FOO: BAR

    # You can also mount for persistent storage. In our case, we'll create "myvol" which mounts to /data
    # Creating a mount, you will also need to create a "myvol" volume component
    # volumeMounts:
    #  - name: myvol
    #    path: /data

    # Toggles whether or not the project source code should be mounted in the component
    # By default, this is set to true
    mountSources: true

    # Optional specification of the path in the container where project sources should be transferred/mounted when mountSources is true. When omitted, the default value of /projects is used.
    # sourceMapping: /projects

  # The name of the component
  name: runtime

# In our commented example, if we were to use a volume, we would add the following to our component list:
# - name: myvol
#   volume:
#     size: 3Gi
```

#### Full Example using `odo deploy` capabilities

In the below example, we implement a image component as well as multiple Kubernetes inline resources for our `odo deploy` scenario:

```yaml
components:
# This will build AND push the container image locally before deployment
- name: outerloop-build
  image:

    # Here we provide the Dockerfile that we are going to be using
    dockerfile:

      # Where the Dockerfile will be built
      buildContext: ${PROJECT_SOURCE}

      # If root is going to be required or not
      rootRequired: false

      # Now we point to where the Dockerfile is located
      # Here we can either provide: URI, Git or a Devfile Registry
      uri: ./Dockerfile

    # Provide the image name to be used, in our case, we use the image name from `variables:`
    imageName: "{{CONTAINER_IMAGE}}"
    
    # For auto build, you can define if an image should be built during startup
    # autoBuild: false

# This will create a Deployment in order to run your container image across
# the cluster. We provide raw Kubernetes yaml inlined into our devfile.yaml
- name: outerloop-deployment
  kubernetes:

    # Below we provide the inline code for our Deployment
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
              - name: nodejs
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

# You can also define OpenShift inline resources:
# - name: outlerloop-openshift-route
#   openshift:
#     inlined: |
#       apiVersion: v1
#       kind: Route
#       ...
```

### Full Example

Below is the full example which you can run locally:

```yaml
schemaVersion: 2.2.0

metadata:
  description: Stack with Node.js 14
  displayName: Node.js Runtime
  icon: https://nodejs.org/static/images/logos/nodejs-new-pantone-black.svg
  language: javascript
  name: my-nodejs-app
  projectType: nodejs
  tags:
  - NodeJS
  - Express
  - ubi8
  version: 1.0.1

starterProjects:
- git:
    remotes:
      origin: https://github.com/odo-devfiles/nodejs-ex.git
  name: nodejs-starter

variables:
  CONTAINER_IMAGE: quay.io/MYUSERNAME/nodejs-odo-example
  RESOURCE_NAME: my-nodejs-app
  CONTAINER_PORT: "3000"
  DOMAIN_NAME: nodejs.example.com

commands:
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
- exec:
    commandLine: npm install
    component: runtime
    group:
      isDefault: true
      kind: build
    workingDir: ${PROJECT_SOURCE}
  id: install
- exec:
    commandLine: npm start
    component: runtime
    group:
      isDefault: true
      kind: run
    workingDir: ${PROJECT_SOURCE}
  id: run
- exec:
    commandLine: npm run debug
    component: runtime
    group:
      isDefault: true
      kind: debug
    workingDir: ${PROJECT_SOURCE}
  id: debug
- exec:
    commandLine: npm test
    component: runtime
    group:
      isDefault: true
      kind: test
    workingDir: ${PROJECT_SOURCE}
  id: test

components:
- name: outerloop-build
  image:
    dockerfile:
      buildContext: ${PROJECT_SOURCE}
      rootRequired: false
      uri: ./Dockerfile
    imageName: "{{CONTAINER_IMAGE}}"
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
- container:
    endpoints:
    - name: http-3000
      targetPort: 3000
    image: registry.access.redhat.com/ubi8/nodejs-14:latest
    memoryLimit: 1024Mi
    mountSources: true
  name: runtime
```

### Not Yet Implemented


All full descriptions of missing specification features can be found on the [2.2.0 API Specification](https://devfile.io/docs/devfile/2.2.0/user-guide/api-reference/).

List of Devfile spec features not yet implemented in `odo`:

#### Components: Git Checkout

```yaml
components:
- name: outerloop-build
  image:
    imageName: "{{CONTAINER_IMAGE}}"
    dockerfile:
      buildContext: ${PROJECT_SOURCE}
      rootRequired: false
      uri: ./Dockerfile

      # NOT YET IMPLEMENTED
      git:
        checkoutFrom:
          remote: subproject
          revision: 1.0.0Final
        fileLocation: ./Dockerfile
        remotes:
          origin: https://github.com/myusername/exampleproject.git
```

#### Components: Devfile Checkout

```yaml
components:
- name: outerloop-build
  image:
    imageName: "{{CONTAINER_IMAGE}}"
    dockerfile:
      buildContext: ${PROJECT_SOURCE}
      rootRequired: false
      uri: ./Dockerfile

      # NOT YET IMPLEMENTED
      devfileRegistry:
        id: myregistry
        registryUrl: https://registry.devfile.io
```

#### Components: Kubernetes Endpoints and DeployByDefault

```yaml
components:
- name: outerloop-deployment
  kubernetes:
    # NOT YET IMPLEMENTED
    # See: https://devfile.io/docs/devfile/2.2.0/user-guide/api-reference/
    # for full details
    # endpoints:

    # NOT YET IMPLEMENTED
    # Defines if the component should be deployed during startup.
    deployByDefault: false

    # Below we provide the inline code for our Deployment
    inlined: |
      kind: Deployment
      ...
```
