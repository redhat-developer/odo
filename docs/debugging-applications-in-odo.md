---
# Page settings
layout: default
keywords:
comments: false

# Hero section
title: Debugging applications in odo
description: Learn how to debug an application in odo

# Micro navigation
micro_nav: true
---
> **Important**
> 
> Interactive debugging in odo is a Technology Preview feature only.
> Technology Preview features are not supported with Red Hat production
> service level agreements (SLAs) and might not be functionally
> complete. Red Hat does not recommend using them in production. These
> features provide early access to upcoming product features, enabling
> customers to test functionality and provide feedback during the
> development process.
> 
> For more information about the support scope of Red Hat Technology
> Preview features, see
> <https://access.redhat.com/support/offerings/techpreview/>.

With `odo`, you can attach a debugger to remotely debug your
application. This feature is only supported for NodeJS and Java
components.

Components created with `odo` run in the debug mode by default. A
debugger agent runs on the component, on a specific port. To start
debugging your application, you must start port forwarding and attach
the local debugger bundled in your Integrated development environment
(IDE).

# Debugging an application

You can debug your application on in `odo` with the `odo debug` command.

1.  After an application is deployed, start the port forwarding for your
    component to debug the application:
    
        $ odo debug port-forward

2.  Attach the debugger bundled in your IDE to the component.
    Instructions vary depending on your IDE.

# Configuring debugging parameters

You can specify a remote port with `odo config` command and a local port
with the `odo debug` command.

  - To set a remote port on which the debugging agent should run, run:
    
        $ odo config set DebugPort 9292
    
    > **Note**
    > 
    > You must redeploy your component for this value to be reflected on
    > the component.

  - To set a local port to port forward, run:
    
        $ odo debug port-forward --local-port 9292
    
    > **Note**
    > 
    > The local port value does not persist. You must provide it every
    > time you need to change the port.
