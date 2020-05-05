---
# Page settings
layout: default
keywords:
comments: false

# Hero section
title: Creating a single-component application with odo
description: Get started deploying your first component

# Micro navigation
micro_nav: true

# Page navigation
page_nav:
    prev:
        content: Understanding odo
        url: '/docs/understanding-odo'
    next:
        content: Creating a multicomponent application with odo
        url: '/docs/creating-a-multicomponent-application-with-odo'
---
With `odo`, you can create and deploy applications on OpenShift
clusters.

  - `odo` is installed.

  - You have a running OpenShift cluster. You can use [CodeReady
    Containers
    (CRC)](https://cloud.redhat.com/openshift/install/crc/installer-provisioned?intcmp=7013a000002CtetAAC)
    to deploy a local OpenShift cluster quickly.

# Creating a project

Create a project to keep your source code, tests, and libraries
organized in a separate single unit.

1.  Log in to an OpenShift cluster:
    
        $ odo login -u developer -p developer

2.  Create a project:
    
        $ odo project create myproject
         ✓  Project 'myproject' is ready for use
         ✓  New project created and now using project : myproject

# Creating a Node.js application with odo

To create a Node.js component, download the Node.js application and push
the source code to your cluster with `odo`.

1.  Create a directory for your components:
    
        $ mkdir my_components $$ cd my_components

2.  Download the example Node.js application:
    
        $ git clone https://github.com/openshift/nodejs-ex

3.  Change the current directory to the directory with your application:
    
        $ cd <directory name>

4.  Add a component of the type Node.js to your application:
    
        $ odo create nodejs
    
    > **Note**
    > 
    > By default, the latest image is used. You can also explicitly
    > specify an image version by using `odo create openshift/nodejs:8`.

5.  Push the initial source code to the component:
    
        $ odo push
    
    Your component is now deployed to OpenShift.

6.  Create a URL and add an entry in the local configuration file as
    follows:
    
        $ odo url create --port 8080

7.  Push the changes. This creates a URL on the cluster.
    
        $ odo push

8.  List the URLs to check the desired URL for the component.
    
        $ odo url list

9.  View your deployed application using the generated URL.
    
        $ curl <URL>

# Modifying your application code

You can modify your application code and have the changes applied to
your application on OpenShift.

1.  Edit one of the layout files within the Node.js directory with your
    preferred text editor.

2.  Update your component:
    
        $ odo push

3.  Refresh your application in the browser to see the changes.

# Adding storage to the application components

Persistent storage keeps data available between restarts of odo. You can
add storage to your components with the `odo storage` command.

  - Add storage to your
        components:
    
        $ odo storage create nodestorage --path=/opt/app-root/src/storage/ --size=1Gi

Your component now has 1 GB storage.

# Adding a custom builder to specify a build image

With OpenShift, you can add a custom image to bridge the gap between the
creation of custom images.

The following example demonstrates the successful import and use of the
`redhat-openjdk-18` image:

  - The OpenShift CLI (oc) is installed.

<!-- end list -->

1.  Import the image into OpenShift:
    
        $ oc import-image openjdk18 \
        --from=registry.access.redhat.com/redhat-openjdk-18/openjdk18-openshift \
        --confirm

2.  Tag the image to make it accessible to odo:
    
        $ oc annotate istag/openjdk18:latest tags=builder

3.  Deploy the image with odo:
    
        $ odo create openjdk18 --git \
        https://github.com/openshift-evangelists/Wild-West-Backend

# Connecting your application to multiple services using OpenShift Service Catalog

The OpenShift service catalog is an implementation of the Open Service
Broker API (OSB API) for Kubernetes. You can use it to connect
applications deployed in OpenShift to a variety of services.

  - You have a running OpenShift cluster.

  - The service catalog is installed and enabled on your cluster.

<!-- end list -->

  - To list the services:
    
        $ odo catalog list services

  - To use service catalog-related operations:
    
        $ odo service <verb> <servicename>

# Deleting an application

> **Important**
> 
> Deleting an application will delete all components associated with the
> application.

1.  List the applications in the current project:
    
        $ odo app list
            The project '<project_name>' has the following applications:
            NAME
            app

2.  List the components associated with the applications. These
    components will be deleted with the application:
    
        $ odo component list
            APP     NAME                      TYPE       SOURCE        STATE
            app     nodejs-nodejs-ex-elyf     nodejs     file://./     Pushed

3.  Delete the application:
    
        $ odo app delete <application_name>
            ? Are you sure you want to delete the application: <application_name> from project: <project_name>

4.  Confirm the deletion with `Y`. You can suppress the confirmation
    prompt using the `-f` flag.
