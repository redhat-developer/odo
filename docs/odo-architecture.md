---
# Page settings
layout: default
keywords:
comments: false

# Hero section
title: Architecture of odo
description: A general overview of the odo architecture

# Micro navigation
micro_nav: true
---
This section describes `odo` architecture and how `odo` manages resources on a cluster.

# Developer setup

With odo you can create and deploy application on OpenShift clusters from a terminal. Code editor plug-ins use odo which allows users to interact with OpenShift clusters from their IDE terminals. Examples of plug-ins that use odo: VS Code OpenShift Connector, OpenShift Connector for Intellij, Codewind for Eclipse Che.

odo works on Windows, macOS, and Linux operating systems and from any terminal. odo provides autocompletion for bash and zsh command line shells.

odo supports Node.js and Java components.

# OpenShift source-to-image

OpenShift Source-to-Image (S2I) is an open-source project which helps in building artifacts from source code and injecting these into container images. S2I produces ready-to-run images by building source code without the need of a Dockerfile. odo uses S2I builder image for executing developer source code inside a container.

# OpenShift cluster objects

## Init Containers

Init containers are specialized containers that run before the application container starts and configure the necessary environment for the application containers to run. Init containers can have files that application images do not have, for example setup scripts. Init containers always run to completion and the application container does not start if any of the init containers fails.

The pod created by odo executes two Init Containers:

  - The `copy-supervisord` Init container.

  - The `copy-files-to-volume` Init container.

### `copy-supervisord`

The `copy-supervisord` Init container copies necessary files onto an `emptyDir` volume. The main application container utilizes these files from the `emptyDir` volume.

  - Binaries:
    
      - `go-init` is a minimal init system. It runs as the first process (PID 1) inside the application container. go-init starts the `SupervisorD` daemon which runs the developer code. go-init is required to handle orphaned processes.
    
      - `SupervisorD` is a process control system. It watches over configured processes and ensures that they are running. It also restarts services when necessary. For odo, `SupervisorD` executes and monitors the developer code.

  - Configuration files:
    
      - `supervisor.conf` is the configuration file necessary for the SupervisorD daemon to start.

  - Scripts:
    
      - `assemble-and-restart` is an OpenShift S2I concept to build and deploy user-source code. The assemble-and-restart script first assembles the user source code inside the application container and then restarts SupervisorD for user changes to take effect.
    
      - `Run` is an OpenShift S2I concept of executing the assembled source code. The `run` script executes the assembled code created by the `assemble-and-restart` script.
    
      - `s2i-setup` is a script that creates files and directories which are necessary for the `assemble-and-restart` and run scripts to execute successfully. The script is executed whenever the application container starts.

  - Directories:
    
      - `language-scripts`: OpenShift S2I allows custom `assemble` and `run` scripts. A few language specific custom scripts are present in the `language-scripts` directory. The custom scripts provide additional configuration to make odo debug work.

The `emptyDir` volume is mounted at the `/opt/odo` mount point for both the Init container and the application container.

### `copy-files-to-volume`

The `copy-files-to-volume` Init container copies files that are in `/opt/app-root` in the S2I builder image onto the persistent volume. The volume is then mounted at the same location (`/opt/app-root`) in an application container.

Without the persistent volume on `/opt/app-root` the data in this directory is lost when the persistent volume claim is mounted at the same location.

The PVC is mounted at the `/mnt` mount point inside the Init container.

## Application container

Application container is the main container inside of which the user-source code executes.

Application container is mounted with two volumes:

  - `emptyDir` volume mounted at `/opt/odo`

  - The persistent volume mounted at `/opt/app-root`

`go-init` is executed as the first process inside the application container. The `go-init` process then starts the `SupervisorD` daemon.

`SupervisorD` executes and monitors the user assembled source code. If the user process crashes, `SupervisorD` restarts it.

## Persistent volumes and persistent volume claims

A persistent volume claim (PVC) is a volume type in Kubernetes which provisions a persistent volume. The life of a persistent volume is independent of a pod lifecycle. The data on the persistent volume persists across pod restarts.

The `copy-files-to-volume` Init container copies necessary files onto the persistent volume. The main application container utilizes these files at runtime for execution.

The naming convention of the persistent volume is \<component\_name\>-s2idata.

| Container              | PVC mounted at  |
| ---------------------- | --------------- |
| `copy-files-to-volume` | `/mnt`          |
| Application container  | `/opt/app-root` |

## `emptyDir` volume

An `emptyDir` volume is created when a pod is assigned to a node, and exists as long as that pod is running on the node. If the container is restarted or moved, the content of the `emptyDir` is removed, Init container restores the data back to the `emptyDir`. `emptyDir` is initially empty.

The `copy-supervisord` Init container copies necessary files onto the `emptyDir` volume. These files are then utilized by the main application container at runtime for execution.

| Container             | `emptyDir volume` mounted at |
| --------------------- | ---------------------------- |
| `copy-supervisord`    | `/opt/odo`                   |
| Application container | `/opt/odo`                   |

## Service

A service is a Kubernetes concept of abstracting the way of communicating with a set of pods.

odo creates a service for every application pod to make it accessible for communication.

# `odo push` workflow

This section describes `odo push` workflow. odo push deploys user code on an OpenShift cluster with all the necessary OpenShift resources.

1.  Creating resources
    
    If not already created, `odo push` creates the following OpenShift resources:
    
      - `DeploymentConfig` object:
        
          - Two init containers are executed: `copy-supervisord` and `copy-files-to-volume`. The init containers copy files onto the `emptyDir` and the `PersistentVolume` type of volumes respectively.
        
          - The application container starts. The first process in the application container is the `go-init` process with PID=1.
        
          - `go-init` process starts the SupervisorD daemon.
            
            > **Note**
            > 
            > The user application code has not been copied into the application container yet, so the `SupervisorD` daemon does not execute the `run` script.
    
      - `Service` object
    
      - `Secret` objects
    
      - `PersistentVolumeClaim` object

2.  Indexing files
    
      - A file indexer indexes the files in the source code directory. The indexer traverses through the source code directories recursively and finds files which have been created, deleted, or renamed.
    
      - A file indexer maintains the indexed information in an odo index file inside the `.odo` directory.
    
      - If the odo index file is not present, it means that the file indexer is being executed for the first time, and creates a new odo index JSON file. The odo index JSON file contains a file map - the relative file paths of the traversed files and the absolute paths of the changed and deleted files.

3.  Pushing code
    
    Local code is copied into the application container, usually under `/tmp/src`.

4.  Executing `assemble-and-restart`
    
    On a successful copy of the source code, the `assemble-and-restart` script is executed inside the running application container.
