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

# Page navigation
page_nav:
    prev:
        content: Configure your terminal for autocompletion
        url: '/docs/configuring-the-odo-cli'
    next:
        content: odo CLI reference
        url: '/docs/odo-cli-reference'
---
This section describes `odo` architecture and how `odo` manages
OpenShift resources on a cluster.

# Developer setup

With odo you can create and deploy application on OpenShift clusters
from a terminal. Code editor plug-ins use odo which allows users to
interact with OpenShift clusters from their IDE terminals. Examples of
plug-ins that use odo: VS Code Openshift Connector, Openshift Connector
for Intellij, Codewind for Eclipse Che.

odo works on Windows, macOS, and Linux operating systems and from any
terminal. odo provides autocompletion for bash and zsh command line
shells.

odo 1.1.2 supports Node.js and Java components.

# OpenShift source-to-image

Openshift Source-to-Image (S2I) is an open-source project which helps in
building artifacts from source code and injecting these into container
images. S2I produces ready-to-run images by building source code without
the need of a Dockerfile. odo uses S2I builder image for executing
developer source code inside a container.

# OpenShift cluster objects

## Init Containers

Init containers are specialized containers that run before the
application container starts and configure the necessary environment for
the application containers to run. Init containers can have files that
application images do not have, for example setup scripts. Init
containers always run to completion and the application container does
not start if any of the init containers fails.

The Pod created by odo executes two Init Containers:

  - The `copy-supervisord` Init container.

  - The `copy-files-to-volume` Init container.

### `copy-supervisord`

The `copy-supervisord` Init container copies necessary files onto an
`emptyDir` Volume. The main application container utilizes these files
from the `emptyDir` Volume.

  - Binaries:
    
      - `go-init` is a minimal init system. It runs as the first process
        (PID 1) inside the application container. go-init starts the
        `SupervisorD` daemon which runs the developer code. go-init is
        required to handle orphaned processes.
    
      - `SupervisorD` is a process control system. It watches over
        configured processes and ensures that they are running. It also
        restarts services when necessary. For odo, `SupervisorD`
        executes and monitors the developer code.

  - Configuration files:
    
      - `supervisor.conf` is the configuration file necessary for the
        SupervisorD daemon to start.

  - Scripts:
    
      - `assemble-and-restart` is an OpenShift S2I concept to build and
        deploy user-source code. The assemble-and-restart script first
        assembles the user source code inside the application container
        and then restarts SupervisorD for user changes to take effect.
    
      - `Run` is an OpenShift S2I concept of executing the assembled
        source code. The `run` script executes the assembled code
        created by the `assemble-and-restart` script.
    
      - `s2i-setup` is a script that creates files and directories which
        are necessary for the `assemble-and-restart` and run scripts to
        execute successfully. The script is executed whenever the
        application container starts.

  - Directories:
    
      - `language-scripts`: Openshift S2I allows custom `assemble` and
        `run` scripts. A few language specific custom scripts are
        present in the `language-scripts` directory. The custom scripts
        provide additional configuration to make odo debug work.

The `emtpyDir Volume` is mounted at the `/opt/odo` mount point for both
the Init container and the application container.

### `copy-files-to-volume`

The `copy-files-to-volume` Init container copies files that are in
`/opt/app-root` in the S2I builder image onto the Persistent Volume. The
volume is then mounted at the same location (`/opt/app-root`) in an
application container.

Without the `PersistentVolume` on `/opt/app-root` the data in this
directory is lost when `PersistentVolumeClaim` is mounted at the same
location.

The `PVC` is mounted at the `/mnt` mount point inside the Init
container.

## Application container

Application container is the main container inside of which the
user-source code executes.

Application container is mounted with two Volumes:

  - `emptyDir` Volume mounted at `/opt/odo`

  - The `PersistentVolume` mounted at `/opt/app-root`

`go-init` is executed as the first process inside the application
container. The `go-init` process then starts the `SupervisorD` daemon.

`SupervisorD` executes and monitores the user assembled source code. If
the user process crashes, `SupervisorD` restarts it.

## `PersistentVolume` and `PersistentVolumeClaim`

`PersistentVolumeClaim` (`PVC`) is a volume type in Kubernetes which
provisions a `PersistentVolume`. The life of a `PersistentVolume` is
independent of a Pod lifecycle. The data on the `PersistentVolume`
persists across Pod restarts.

The `copy-files-to-volume` Init container copies necessary files onto
the `PersistentVolume`. The main application container utilizes these
files at runtime for execution.

The naming convention of the `PersistentVolume` is
\<component-name\>-s2idata.

| Container              | `PVC Mounted` at |
| ---------------------- | ---------------- |
| `copy-files-to-volume` | `/mnt`           |
| Application container  | `/opt/app-root`  |

## `emptyDir` Volume

An `emptyDir` Volume is created when a Pod is assigned to a node, and
exists as long as that Pod is running on the node. If the container is
restarted or moved, the content of the `emptyDir` is removed, Init
container restores the data back to the `emptyDir`. `emptyDir` is
initially empty.

The `copy-supervisord` Init container copies necessary files onto the
`emptyDir` volume. These files are then utilized by the main application
container at runtime for execution.

| Container             | `emptyDir Volume` Mounted at |
| --------------------- | ---------------------------- |
| `copy-supervisord`    | `/opt/odo`                   |
| Application container | `/opt/odo`                   |

## Service

Service is a Kubernetes concept of abstracting the way of communicating
with a set of Pods.

odo creates a Service for every application Pod to make it accessible
for communication.

# `odo push` workflow

This section describes `odo push` workflow. odo push deploys user code
on an OpenShift cluster with all the necessary OpenShift resources.

1.  Creating resources
    
    If not already created, `odo push` creates the following OpenShift
    resources:
    
      - Deployment config (DC):
        
          - Two init containers are executed: `copy-supervisord` and
            `copy-files-to-volume`. The init containers copy files onto
            the `emptyDir` and the `PersistentVolume` type of volumes
            respectively.
        
          - The application container starts. The first process in the
            application container is the `go-init` process with PID=1.
        
          - `go-init` process starts the SupervisorD daemon.
            
            <div class="note">
            
            The user application code has not been copied into the
            application container yet, so the `SupervisorD` daemon does
            not execute the `run` script.
            
            </div>
    
      - Service
    
      - Secrets
    
      - `PersistentVolumeClaim`

2.  Indexing files
    
      - A file indexer indexes the files in the source code directory.
        The indexer traverses through the source code directories
        recursively and finds files which have been created, deleted, or
        renamed.
    
      - A file indexer maintains the indexed information in an odo index
        file inside the `.odo` directory.
    
      - If the odo index file is not present, it means that the file
        indexer is being executed for the first time, and creates a new
        odo index JSON file. The odo index JSON file contains a file map
        - the relative file paths of the traversed files and the
        absolute paths of the changed and deleted files.

3.  Pushing code
    
    Local code is copied into the application container, usually under
    `/tmp/src`.

4.  Executing `assemble-and-restart`
    
    On a successful copy of the source code, the `assemble-and-restart`
    script is executed inside the running application container.
