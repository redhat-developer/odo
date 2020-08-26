---
title: Odo Devfile Reference

language_tabs:
  - yaml

toc_footers:
  - <a href='https://odo.dev'>Odo Homepage</a>
  - <a href='https://devfile.github.io/devfile/index.html'>Devfile GitHub Page</a>

search: true
---

# Introduction

> Download a starter application to get an example devfile.yaml

```sh
$ odo create nodejs --starter
```

**Installation:**

Installing odo can be found at our [installation page](https://odo.dev/docs/installing-odo/).

**Quick Start:**

If you haven't used odo yet, we recommend going through our [Deploying a devfile using odo guide](https://odo.dev/docs/deploying-a-devfile-using-odo/). 


# Devfile Properties

> All definable devfile properties

```yaml
schemaVersion: <string>
metadata: <metadataObject>
starterProjects: 
- <starterProjectObject>
components: 
- <componentObject>
commands: 
- <commandObject>
```

| Key             | Type                                                    | Required | Description                                                             |
|-----------------|---------------------------------------------------------|----------|-------------------------------------------------------------------------|
| schemaVersion   | string                                                  | yes      | Schema version of devfile                                               |
| metadata        | [metadataObject](#metadataobject)                       | no       | Metadata information that describes the project                         |
| starterProjects | array of [starterProjectObject](#starterproject-object) | no       | List of starter projects that can be used to bootstrap new projects     |
| components      | array of [componentObject](#component-object)           | no       | List of components to be used within your development environment       |
| commands        | array of [commandObject](#commandobject)                | no       | List of commands to be executed                                         |                                             

## schemaVersion

```yaml
schemaVersion: 2.0.0
```

| Key           | Type   | Required | Description         |
|---------------|--------|----------|---------------------|
| schemaVersion | string | yes      | 2.0.0 of the schema |


## metadata

```yaml
metadata:
  name: nodejs
  version: 1.0.0
```

| Key            | Type                               | Description         |
|----------------|------------------------------------|---------------------|
| metadataObject | [metadataObject](#metadata-object) | Metadata to be used |

## starter projects

> Example using a [Quarkus quickstart project](https://github.com/quarkusio/quarkus-quickstarts)

```yaml
starterProjects:
  - name: quarkus-quickstart
    description: minimal CRUD app with quarkus
    git:
      location: https://github.com/quarkusio/quarkus-quickstarts
      branch: master
      sparseCheckoutDir: /getting-started/ # optional checkout dir
      startPoint: 1.7.0.Final # tag or commit to start at
```


| Key                  | Type                                           | Description                         |
|----------------------|------------------------------------------------|-------------------------------------|
| starterProjectObject | [starterProjectObject](#starterproject-object) | List of starter projects to be used |


## components

> Example using an [OpenLiberty](https://github.com/odo-devfiles/registry/blob/master/devfiles/openLiberty/devfile.yaml) container

```yaml
components:
  - container:
      name: appsodyrun
      image: 'ajymau/java-openliberty-dev:latest'
      memoryLimit: 1512Mi
      mountSources: true
      endpoints:
        - name: 9080/tcp
          exposure: public 
          public: true
          protocol: http
          targetPort: 9080
      env:
        - name: MODE2
          value: TEST2
        - name: myprop2
          value: myval2
```

| Key             | Type                                 | Description                   |
|-----------------|--------------------------------------|-------------------------------|
| componentObject | [componentObject](#component-object) | List of components to be used |


## commands

> Example using the exec command

```yaml
commands:
  - exec:
      id: devBuild
      commandLine: "/artifacts/bin/build-container-full.sh"
      component: tools
      workingDir: /projects/springbootproject
      group:
        kind: build
        isDefault: true
  - exec:
      id: devRun
      commandLine: "/artifacts/bin/start-server.sh"
      component: runtime
      workingDir: /
      group:
        kind: run
        isDefault: true
```

| Key           | Type                             | Description                                               |
|---------------|----------------------------------|-----------------------------------------------------------|
| commandObject | [commandObject](#command-object) | Command to be executed in an existing component container |


## events

> Example of various events

```yaml
commands:
  - exec:
      id: copy
      commandLine: "cp /tools/myfile.txt tools.txt"
      component: tools
      workingDir: /
  - exec:
      id: initCache
      commandLine: "./init_cache.sh"
      component: tools
      workingDir: /
  - exec:
      id: connectDB
      commandLine: "./connect_db.sh"
      component: runtime
      workingDir: /
  - exec:
      id: disconnectDB
      commandLine: "./disconnect_db.sh"
      component: runtime
      workingDir: /
  - exec:
      id: cleanup
      commandLine: "./cleanup.sh"
      component: tools
      workingDir: /
events:
  preStart:
    - "connectDB"
  postStart:
    - "copy" 
    - "initCache"
  preStop:
    - "disconnectDB"
  postStop:
    - "cleanup"
```

| Key           | Type                             | Description                                      |
|---------------|----------------------------------|--------------------------------------------------|
| eventObject  | [eventObject](#event-object)     | Events to be executed during a project lifecycle |

# Metadata Object

> Example using metadata

```yaml
metadata:
  name: nodejs
  version: 1.0.0
```

| Key     | Type   | Required | Description                                                                       |
|---------|--------|----------|-----------------------------------------------------------------------------------|
| name    | string | yes      | Name of your devfile. This name will propagate to the devfile registry if listed. |
| version | string | yes      | Version of your devfile                                                           |


# StarterProject Object

> Pulling from a git location

```yaml
starterProjects:
  - name: quarkus-quickstart
    description: minimal CRUD app with quarkus
    git:
      location: https://github.com/quarkusio/quarkus-quickstarts
      branch: master
      sparseCheckoutDir: /getting-started/ # optional checkout dir
      startPoint: 1.7.0.Final # tag or commit to start at
```   


> Pulling from a GitHub location

```yaml
starterProjects:
  - name: nodejs-web-app
    description: nodejs app
    github:
      location: https://github.com/odo-devfiles/nodejs-ex.git
      branch: master
      sparseCheckoutDir: /app/ # optional checkout dir
      startPoint: 1.0.0 # tag or commit to start at
```

> Pulling from a zip 

```yaml
starterProjects:
  - name: nodejs-web-app
    description: nodejs web app
    zip:
      location: https://github.com/odo-devfiles/nodejs-ex/archive/0.0.1.zip
      sparseCheckoutDir: /app/ # optional checkout dir
```

Each starter project may contain three different objects, `git`, `github` or `zip`.



| Key       | Type                          | Description                                                                        |
|-----------|-------------------------------|------------------------------------------------------------------------------------|
| name      | string                        | Name of your devfile                                                               |
| description | string                      | Description of a starter project                                                   |
| clonePath | string                        | Path relative the root of your projects to which the project should be cloned into |
| git       | [gitObject](#gitobject)       | Pull from a Git location                                                           |
| github    | [githubObject](#githubobject) | Pull from GitHub                                                                   |
| zip       | [zipObject](#zipobject)       | Get from a zip location                                                            |


## gitObject

> Using the gitObject

```yaml
starterProjects:
  - name: nodejs-web-app
    git:
      location: https://github.com/odo-devfiles/nodejs-ex.git
      branch: master
      sparseCheckoutDir: /app/ # optional checkout dir
      startPoint: 1.0.0 # tag or commit to start at
```

| Key              | Type   | Description                        |
|------------------|--------|------------------------------------|
| location         | string | Location of the git repository     |
| branch           | string | What branch to use                 |
| spareCheckoutDir | string | What directory to use when pulling |
| startPoint       | string | Tag or commit to start from        |

## githubObject

> Using the githubObject

```yaml
starterProjects:
  - name: nodejs-web-app
    github:
      location: https://github.com/odo-devfiles/nodejs-ex.git
      branch: master
      sparseCheckoutDir: /app/ # optional checkout dir
      startPoint: 1.0.0 # tag or commit to start at
```

| Key              | Type   | Description                        |
|------------------|--------|------------------------------------|
| location         | string | Location of the git repository     |
| branch           | string | What branch to use                 |
| spareCheckoutDir | string | What directory to use when pulling |
| startPoint       | string | Tag or commit to start from        |

## zipObject

> Using the zipObject

```yaml
starterProjects:
  - name: nodejs-web-app
    zip:
      location: https://github.com/odo-devfiles/nodejs-ex/archive/0.0.1.zip
      sparseCheckoutDir: /app/ # optional checkout dir
```

| Key              | Type   | Description                        |
|------------------|--------|------------------------------------|
| location         | string | Location of the zip                |
| spareCheckoutDir | string | What directory to use when pulling |


# Component Object

> Example using a container using the minimum amount of keys

```yaml
components:
  - container:
      name: appsodyrun
      image: 'ajymau/java-openliberty-dev:latest'
```

Each component must use the `container` object.


| Key       | Type                                | Description                   |
|-----------|-------------------------------------|-------------------------------|
| container | [containerObject](#containerobject) | List of containers to be used |


## containerObject

> Example using an [OpenLiberty](https://github.com/odo-devfiles/registry/blob/master/devfiles/openLiberty/devfile.yaml) container

```yaml
components:
  - container:
      name: appsodyrun
      image: 'ajymau/java-openliberty-dev:latest'
      memoryLimit: 1512Mi
      mountSources: true
      endpoints:
        - name: 9080/tcp
          exposure: public 
          public: true
          protocol: http
          targetPort: 9080
      env:
        - name: MODE2
          value: TEST2
        - name: myprop2
          value: myval2
```

| Key            | Type                                    | Required | Description                                                                                                 |
|----------------|-----------------------------------------|----------|-------------------------------------------------------------------------------------------------------------|
| name           | string                                  | yes      | Name of your container                                                                                      |
| image          | string                                  | yes      | Image version                                                                                               |
| memoryLimit    | string                                  | no       | Memory limit to be used with your container                                                                 |
| mountSources   | boolean                                 | no       | Mount the source or not                                                                                     |
| sourceMapping  | string                                  | no       | Path in the container where project sources should be transferred / mounted when mountSource is set to true |
| endpoints[]    | [endpointObject](#endpointobject)       | no       | List of endpoints to use                                                                                    |
| volumeMounts[] | [volumeMountsObject](#volumemountsobject) | no       | List of volumes to mount                                                                                    |
| env[]          | [envObject](#envobject)                 | no       | List of environment variables to use                                                                        |



## endpointObject

> Example using an endpoint

```yaml
components:
  - container:
      name: appsodyrun
      image: 'ajymau/java-openliberty-dev:latest'
      endpoints:
        - name: 9080/tcp
          exposure: public 
          public: true
          protocol: http
          targetPort: 9080
```

| Key        | Type    | Required | Description                                                                                                                                    |
|------------|---------|----------|------------------------------------------------------------------------------------------------------------------------------------------------|
| name       | string  | yes      | Name of your endpoint                                                                                                                          |
| targetPort | integer | yes      | Port number that you are targeting                                                                                                             |
| exposure   | string  | no       | `public`, `internal` or `none`. Describes how the endpoint should be exposed on the network                                                    |
| path       | string  | no       | Path to the endpoint URL                                                                                                                       |
| protocol   | string  | no       | `http`, `https`, `ws`, `wss`, `tcp`, `udp`. Describes the application and transport protocols of the traffic that will go through the endpoint |
| secure     | boolean | no       | Whether or not the endpoint is defined as secure                                                                                               |


## volumeMountsObject

> Example using a volume with a container

```yaml
components:
  - container:
      name: tools
      image: maysunfaisal/springbootbuild
      memoryLimit: 768Mi
      command: ['tail']
      args: [ '-f', '/dev/null']
      mountSources: true
      volumeMounts:
        - name: springbootpvc
          path: /data
```

| Key  | Type   | Required | Description                                                                                        |
|------|--------|----------|----------------------------------------------------------------------------------------------------|
| name | string | yes      | Name of the volume                                                                                 |
| path | string | no       | Path in the component container where the volume should be mounted. Defaults to `/<name>` if blank |


# Command Object

> Example using an exec command within a command object

```yaml
commands:
  - exec:
      id: devBuild
      component: tools
      commandLine: "/artifacts/bin/build-container-full.sh"
      workingDir: /projects/springbootproject
      group:
        kind: build
        isDefault: true
```

Each command must use the `exec` object.

| Key  | Type                      | Description               |
|------|---------------------------|---------------------------|
| exec | [execObject](#execobject) | The CLI command to be ran |

## execObject

> Example using exec command within a command object

```yaml
commands:
  - exec:
      id: devBuild
      commandLine: "/artifacts/bin/build-container-full.sh"
      component: tools
      workingDir: /projects/springbootproject
      group:
        kind: build
        isDefault: true
      env:
        - name: FOO
          value: BAR
```

| Key         | Type                        | Required | Description                                            |
|-------------|-----------------------------|----------|--------------------------------------------------------|
| id          | string                      | yes      | ID of the command                                      |
| commandLine | string                      | yes      | Command to be ran                                      |
| component   | string                      | no       | What component that the actions relate to              |
| label       | string                      | no       | Optional label to be used to describe the command      |
| workingDir  | string                      | no       | Working directory where the command should be executed |
| group       | [groupObject](#groupObject) | no       | Group that the command is part of                      |
| env         | [envObject](#envObject)     | no       | List of environment variables used                     |

## groupObject

> Example using the group object within exec 

```yaml
commands:
  - exec:
      id: devBuild
      commandLine: "/artifacts/bin/build-container-full.sh"
      group:
        kind: build
        isDefault: true
```

| Key       | Type    | Required | Description                                                                                                 |
|-----------|---------|----------|-------------------------------------------------------------------------------------------------------------|
| kind      | string  | yes      | Kind of group the command is part of. Options: `build`, `run`, `test`, `debug`                              |
| isDefault | boolean | no       | Identifies that it is the default command to be ran. Only one default command can be defined for each group |

# Event Object

> Example of using life cycle events where each event is an array of devfile commands to be executed of type `exec` or `composite`

```yaml
commands:
  - exec:
      id: copy
      commandLine: "cp /tools/myfile.txt tools.txt"
      component: tools
      workingDir: /
  - exec:
      id: initCache
      commandLine: "./init_cache.sh"
      component: tools
      workingDir: /
  - exec:
      id: connectDB
      commandLine: "./connect_db.sh"
      component: runtime
      workingDir: /
  - exec:
      id: disconnectDB
      commandLine: "./disconnect_db.sh"
      component: runtime
      workingDir: /
  - exec:
      id: cleanup
      commandLine: "./cleanup.sh"
      component: tools
      workingDir: /
events:
  preStart:
    - "connectDB"
  postStart:
    - "copy" 
    - "initCache"
  preStop:
    - "disconnectDB"
  postStop:
    - "cleanup"
```

## preStartObject

PreStart events are executed as init containers for the project pod in the order they are specified. The devfile command's `commandLine` and `workingDir` become the init container's command and as a result the devfile component container's `command` and `args` or the container image's `Command` and `Args` are overwritten. If a composite command with `parallel: true` is used, it will be executed sequentially as Kubernetes init containers only execute in sequence.

## postStartObject

PostStart events are executed when the Kubernetes deployment for the odo component is created. 

## preStopObject

PreStop events are executed before the Kubernetes deployment for the odo component is deleted. 

# Universal objects

List of objects which are found in multiple parts of devfile. For example, the [envObject](#envobject) is found within [containerObject](#containerobject) and [execObject](#execobject).

## envObject

> Example using environment variables with a container

```yaml
components:
  - container:
      name: appsodyrun
      image: 'ajymau/java-openliberty-dev:latest'
          targetPort: 9080
      env:
        - name: MODE2
          value: TEST2
        - name: myprop2
          value: myval2
```

> Example using environment variables within the exec command

```yaml
commands:
  - exec:
      id: devBuild
      commandLine: "/artifacts/bin/build-container-full.sh"
      env:
        - name: FOO
          value: BAR
```

| Key   | Type   | Required | Description                       |
|-------|--------|----------|-----------------------------------|
| name  | string | yes      | Name of the environment variable  |
| value | string | yes      | Value of the environment variable |

# Unsupported features

**PLEASE NOTE:** Devfile v2 is still in its early stages and there are some parts of the schema that have not yet been implemented.

Please refer to the below table for a list of features which are *not yet* implemented:

| Key                                | Key Description                                | Status      | Description                                                                                     |
|------------------------------------|------------------------------------------------|-------------|-------------------------------------------------------------------------------------------------|
| starterProjects[].clonePath        | [starterProjectObject](#starterproject-object) | IN PROGRESS | Refer to issue: https://github.com/openshift/odo/issues/3729                                     |
| projects[]                         | [projectObject](#projectobject)                        | IN PROGRESS | Entire object not yet implemented. Refer to issue: https://github.com/openshift/odo/issues/3798 |
| parent                             | [parentObject](#parent-object)                 | IN PROGRESS | Refer to issue: https://github.com/openshift/odo/issues/2936                                    |
| events                             | [eventObject](#event-object)                   | IN PROGRESS | Refer to postStop issue: https://github.com/openshift/odo/issues/3577                                    |
| component[].kubernetes             | [kubernetesObject](#kubernetesobject)          | IN PROGRESS |                                                                                                 |
| component[].openshift              | [openshiftObject](#openshiftobject)            | IN PROGRESS |                                                                                                 |
| component[].volume                 | [volumeObject](#volumeobject)                  | IN PROGRESS | Refer to: https://github.com/openshift/odo/issues/3407                                          |
| component[].plugin                 | [pluginObject](#pluginobject)                  | IN PROGRESS | Refer to: https://github.com/openshift/odo/issues/3407                                          |
| component[].container.endpoints    | [endpointObject](#endpointobject)              | IN PROGRESS | Refer to: https://github.com/openshift/odo/issues/3544                                          |
| component[].container.dedicatedPod | [containerObject](#containerobject)            | UNKNOWN     | In schema but not yet implemented.                                                              |
| commands[].apply                   | [applyObject](#applyobject)                    | UNKNOWN     | In schema but not yet implemented.                                                              |
| commands[].composite               | [compositeObject](#compositeobject)            | IN PROGRESS | Refer to: https://github.com/openshift/odo/issues/3338                                          |
| commands[].vscodeLaunch            | [vscodeLaunchObject](#vscodeLaunchobject)      | N/A         | Not applicable to odo.                                                                          |
| commands[].vscodeTask              | [vscodeTaskObject](#vscodeTaskobject)          | N/A         | Not applicable to odo.                                                                          |

