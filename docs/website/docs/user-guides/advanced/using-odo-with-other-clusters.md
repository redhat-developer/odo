---
title: Using odo on GKE/AKS/EKS
sidebar_position: 9
---
Deploying an application on a GKE/AKS/EKS cluster does not always work out of the box, often due to storage related permission issues.
Depending on the way storage provisioner is set up for the cluster and user set by the container image, it may not always be possible to write anywhere inside the container, hence syncing files or creating new files may not be possible.

However, there are workarounds to fix this storage permission issue.
1. [Use ephemeral volumes.](#use-ephemeral-volumes)
2. [Define a location with access to read/write the files.](#define-a-location-with-access-to-readwrite-the-files)
3. Use a root user for the container image

## Use Ephemeral volumes
This workaround can be useful if you do not need to mount any additional persistent volumes.
To use ephemeral volumes, set `Ephemeral` preference to false:
```shell
odo preference set Ephemeral false -f
```

## Define a location with access to read/write the files
Use `sourceMapping` to mount source to a directory where the user has read/write permissions.
If the user defined by the container image is root, they should not have any problem
By default, if you mention a relative path, for e.g. `sourceMapping: go-app` it mounts the sources to $HOME/go-app; $HOME will be defined by the image. This will also assign $PROJECT_SOURCE to sourceMapping value.
Ensure that any extra volume mount is also done in the location where the non-root user has write permission. Can consider using `${HOME}/${PROJECT_SOURCE}` location for `.m2` for example.

For the majority of the images from registry.access.redhat.com, all of them are configured to use non-root user.
[//]: # (https://catalog.redhat.com/software/containers/ubi8/openjdk-11/5dd6a4b45a13461646f677f4?container-tabs=dockerfile)

Depending on the storage provisioner, these users may of may not have access to write to directories other than their $HOME.

TODO: Compare the dir permissions on GKE and docker desktop for $HOME, /, $PROJECT_SOURCE. See what group/user are set for each.

[//]: # (Test odo on AKS with low cpu/memory B2ms or the lowest one)

## Example: Deploying a Go application on a GKE/AKS/EKS cluster

### Pre-requisite:
1. Login to your GKE/AKS/EKS cluster.
2. [Initialize the application with `odo init`](../quickstart/go#step-2-initializing-your-application-odo-init).

### Modify the Devfile
```yaml showLineNumbers
commands:
- exec:
    commandLine: go build main.go
    component: runtime
    env:
    - name: GOPATH
      #      highlight-next-line
      value: ${HOME}/${PROJECT_SOURCE}/.go
    - name: GOCACHE
      #      highlight-next-line
      value: ${HOME}/${PROJECT_SOURCE}/.cache
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
    #      highlight-next-line
    sourceMapping: go-app
  name: runtime
metadata:
  description: Go is an open source programming language that makes it easy to build
    simple, reliable, and efficient software.
  displayName: Go Runtime
  icon: https://raw.githubusercontent.com/devfile-samples/devfile-stack-icons/main/golang.svg
  language: Go
  name: places
  projectType: Go
  provider: Red Hat
  tags:
  - Go
  version: 1.0.2
schemaVersion: 2.1.0
starterProjects:
- description: A Go project with a simple HTTP server
  git:
    checkoutFrom:
      revision: main
    remotes:
      origin: https://github.com/devfile-samples/devfile-stack-go.git
  name: go-starter
```
We add `sourceMapping` to container component "runtime". This will mount the source files on $HOME/go-app directory where the non-root user set by container image has RW permission.
If the `sourceMapping` is not defined, `odo` will attempt to mount the source files in `/projects` directory.

Read more about Project Sources in [How odo works](/docs/development/architecture/how-odo-works#project-sources).