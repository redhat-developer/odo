---
# Page settings
layout: default
keywords:
comments: false

# Hero section
title: Migrating S2I (Source-to-Image) components to Devfile components
description: Use odo's built-in tool to convert your S2I deployment to devfile

# Micro navigation
micro_nav: true
---
# Migrating from S2I (Source-to-Image) components to Devfile components

[Devfiles](https://devfile.github.io/) are new way of deploying
component with odo.

Existing users using S2I based components can migrate to a devfile
component by following the below steps:

1.  Change to the component directory you want to migrate:
    
    ``` sh
     $ cd <directory-name>
    ```

2.  List the component status
    
    ``` sh
     $ odo list
    
    Openshift Components:
    APP     NAME      PROJECT        TYPE       SOURCETYPE     STATE
    app     myapp     myproject      nodejs     local          Pushed
    ```

3.  Convert the S2I component to a devfile component by running the
    following command:
    
    ``` sh
     $ odo utils convert-to-devfile
    
     devfile.yaml is available in the current directory.
    ```

4.  Push the newly generated devfile component:
    
    ``` sh
     $ odo push
    
     Validation
      ✓  Validating the devfile [40136ns]
    
     Creating Kubernetes resources for component myapp
      ✓  Waiting for component to start [12s]
    
     Applying URL changes
      ✓  URL myapp-8080: http://myapp-8080-example.com/ created
    
     Syncing to component myapp
      ✓  Checking files for pushing [499419ns]
      ✓  Syncing files to the component [155ms]
    
     Executing devfile commands for component myapp
      ✓  Executing s2i-assemble command "/opt/odo/bin/s2i-setup && /opt/odo/bin/assemble-and-restart" [47s]
      ✓  Executing s2i-run command "/opt/odo/bin/run" [1s]
    
     Pushing devfile component myapp
      ✓  Changes successfully pushed to component
    ```

5.  Check the component status, there should be two components running
    both S2I and devfile:
    
    ``` sh
     $ odo list
    
     Devfile Components:
     APP     NAME      PROJECT        TYPE      STATE
     app     myapp     myproject      myapp     Pushed
    
     Openshift Components:
     APP     NAME      PROJECT        TYPE       SOURCETYPE     STATE
     app     myapp     myproject      nodejs     local          Pushed
    ```

6.  Delete the unused S2I component:
    
    ``` sh
     $ odo delete --s2i --all --force
     ✓  Deleting component myapp [277ms]
     ✓  Component myapp from application app has been deleted
     ✓  Config for the Component myapp has been deleted
    ```
