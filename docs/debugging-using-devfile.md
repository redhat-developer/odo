---
# Page settings
layout: default
keywords:
comments: false

# Hero section
title: Debugging applications in odo
description: Learn how to debug an application in odo CLI and IDE

# Micro navigation
micro_nav: true

---
`odo` uses devfiles to build and deploy components. More information on
devifles : [Introduction to
devfile](https://redhat-developer.github.io/devfile/)

In order to use `odo debug` your devfile is required to have a
`debugrun` step. Example of a nodejs devfile with a debugrun step:

``` yaml
schemaVersion: 2.0.0
metadata:
  name: nodejs
  version: 1.0.0
starterProjects:
  - name: nodejs-starter
    git:
      remotes:
        origin: "https://github.com/odo-devfiles/nodejs-ex.git"
components:
  - name: runtime
    container:
      image: registry.access.redhat.com/ubi8/nodejs-12:1-45
      memoryLimit: 1024Mi
      mountSources: true
      sourceMapping: /project
      endpoints:
        - name: http-3000
          targetPort: 3000
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

  - Now we need to create the component using `odo create nodejs`

  - Next we enable remote debugging for the component using `odo push
    --debug`. We can also use a custom step as the debugrun step using
    `odo push --debug --debug-command="custom-step"`

  - Next we port forward a local port for debugging using `odo debug
    port-forward`. The default local port used for debugging is 5858. If
    5858 is occupied, odo will automatically pick up a local port. We
    can also specify the local port using, `odo debug port-forward
    --local-port 5858`

  - Next we need to attach the debugger to the local port. Here’s a
    guide to do it for VS Code : [Remote
    Debugging](https://code.visualstudio.com/docs/nodejs/nodejs-debugging#_remote-debugging)

# Check if a debugging session is running

We can check if a debugging session is running for a component by using
`odo debug info`

    odo debug info
    Debug is running for the component on the local port : 5858
