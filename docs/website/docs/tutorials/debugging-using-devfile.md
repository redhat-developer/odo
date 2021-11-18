---
title: Debugging using devfile
sidebar_position: 6
---
### Debugging a component

Debugging your component involves port forwarding with the Kubernetes pod. Before you start, it is required that you have a `kind: debug` step located within your `devfile.yaml`.

The following `devfile.yaml` contains a `debug` step under the `commands` key:

```yaml
commands:
  - id: install
    exec:
      component: runtime
      commandLine: npm install
      workingDir: /project
      group:
        kind: build
        isDefault: true
  - id: run
    exec:
      component: runtime
      commandLine: npm start
      workingDir: /project
      group:
        kind: run
        isDefault: true
  - id: debug
    exec:
      component: runtime
      commandLine: npm run debug
      workingDir: /project
      group:
        kind: debug
        isDefault: true
```

### Debugging your devfile component via CLI

We will use the official [nodejs](https://github.com/odo-devfiles/registry/tree/master/devfiles/nodejs) example in our debugging session which includes the necessary `debug` step within `devfile.yaml`.

1. Download the example application:
  ```shell
  odo create nodejs --starter nodejs-starter
  ```
  Example:
  ```shell
  $ odo create nodejs --starter nodejs-starter
  Validation
   ✓  Checking devfile existence [11498ns]
   ✓  Checking devfile compatibility [15714ns]
   ✓  Creating a devfile component from registry: DefaultDevfileRegistry [17565ns]
   ✓  Validating devfile component [113876ns]
  
  Starter Project
   ✓  Downloading starter project nodejs-starter from https://github.com/odo-devfiles/nodejs-ex.git [428ms]
  
  Please use `odo push` command to create the component with source deployed
  ```

2. Push with the `--debug` flag which is required for all debugging deployments:
  ```shell
  odo push --debug
  ```
  Example:
  ```shell
  $ odo push --debug
  
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
  NOTE: A custom debug command may be chosen via the `--debug-command="custom-step"` flag.

3. Port forward to the local port in order to access the debugging interface:
  ```shell
  odo debug port-forward
  ```
  Example:
  ```shell
  $ odo debug port-forward
  Started port forwarding at ports - 5858:5858
  ```

  NOTE: A specific port may be specified using the `--local-port` flag

4. Open a separate terminal window and check if the debug session is running.
  ```shell
  odo debug info
  ```
  
  Example:
  ```shell
  $ odo debug info
  Debug is running for the component on the local port : 5858
  ```

5. Accessing the debugger:
   The debugger is accessible through an assortment of tools. An example of setting up a debug interface would be through [VSCode's debugging interface](https://code.visualstudio.com/docs/nodejs/nodejs-debugging#_remote-debugging).

  ```json
  {
    "type": "node",
    "request": "attach",
    "name": "Attach to remote",
    "address": "TCP/IP address of process to be debugged",
    "port": 5858
  }
  ```