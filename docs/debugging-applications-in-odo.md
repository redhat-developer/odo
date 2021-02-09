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
With `odo`, you can attach a debugger to remotely debug your application. This feature is only supported for NodeJS and Java components.

Components created with `odo` run in the debug mode by default. A debugger agent runs on the component, on a specific port. To start debugging your application, you must start port forwarding and attach the local debugger bundled in your Integrated development environment (IDE).

# Debugging an application

You can debug your application in `odo` with the `odo debug` command.

1.  Download the sample application that contains the necessary `debugrun` step within its devfile:
    
    ``` terminal
    $ odo create nodejs --starter
    ```
    
    **Example output.**
    
    ``` terminal
    Validation
     ✓  Checking devfile existence [11498ns]
     ✓  Checking devfile compatibility [15714ns]
     ✓  Creating a devfile component from registry: DefaultDevfileRegistry [17565ns]
     ✓  Validating devfile component [113876ns]
    
    Starter Project
     ✓  Downloading starter project nodejs-starter from https://github.com/odo-devfiles/nodejs-ex.git [428ms]
    
    Please use `odo push` command to create the component with source deployed
    ```

2.  Push the application with the `--debug` flag, which is required for all debugging deployments:
    
    ``` terminal
    $ odo push --debug
    ```
    
    **Example output.**
    
    ``` terminal
    Validation
     ✓  Validating the devfile [29916ns]
    
    Creating Kubernetes resources for component nodejs
     ✓  Waiting for component to start [38ms]
    
    Applying URL changes
     ✓  URLs are synced with the cluster, no changes are required.
    
    Syncing to component nodejs
     ✓  Checking file changes for pushing [1ms]
     ✓  Syncing files to the component [778ms]
    
    Executing devfile commands for component nodejs
     ✓  Executing install command "npm install" [2s]
     ✓  Executing debug command "npm run debug" [1s]
    
    Pushing devfile component nodejs
     ✓  Changes successfully pushed to component
    ```
    
    > **Note**
    > 
    > You can specify a custom debug command by using the `--debug-command="custom-step"` flag.

3.  Port forward to the local port to access the debugging interface:
    
    ``` terminal
    $ odo debug port-forward
    ```
    
    **Example output.**
    
    ``` terminal
    Started port forwarding at ports - 5858:5858
    ```
    
    > **Note**
    > 
    > You can specify a port by using the `--local-port` flag.

4.  Check that the debug session is running in a separate terminal window:
    
    ``` terminal
    $ odo debug info
    ```
    
    **Example output.**
    
    ``` terminal
    Debug is running for the component on the local port : 5858
    ```

5.  Attach the debugger that is bundled in your IDE of choice. Instructions vary depending on your IDE, for example: [VSCode debugging interface](https://code.visualstudio.com/docs/nodejs/nodejs-debugging#_remote-debugging).

# Configuring debugging parameters

You can specify a remote port with `odo config` command and a local port with the `odo debug` command.

  - To set a remote port on which the debugging agent should run, run:
    
    ``` terminal
    $ odo config set DebugPort 9292
    ```
    
    > **Note**
    > 
    > You must redeploy your component for this value to be reflected on the component.

  - To set a local port to port forward, run:
    
    ``` terminal
    $ odo debug port-forward --local-port 9292
    ```
    
    > **Note**
    > 
    > The local port value does not persist. You must provide it every time you need to change the port.
