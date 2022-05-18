---
title: Using Devfile Lifecycle Events
sidebar_position: 5
---

odo uses devfile to build and deploy components. You can also use devfile events with a component during its lifecycle. The four different types of devfile events are `preStart`, `postStart`, `preStop` and `postStop`

Each event is an array of devfile commands to be executed. The devfile command to be executed should be of type `exec` or `composite`:

```yaml
components:
  - name: runtime
    container:
      image: quay.io/eclipse/che-nodejs10-ubi:nightly
      memoryLimit: 1024Mi
      endpoints:
        - name: "3000/tcp"
          targetPort: 3000 
      mountSources: true
      command: ['tail']
      args: [ '-f', '/dev/null']
  - name: "tools"
    container:
      image: quay.io/eclipse/che-nodejs10-ubi:nightly
      mountSources: true
      memoryLimit: 1024Mi
commands:
  - id: copy
    exec:
      commandLine: "cp /tools/myfile.txt tools.txt"
      component: tools
      workingDir: /
  - id: initCache
    exec:
      commandLine: "./init_cache.sh"
      component: tools
      workingDir: /
  - id: connectDB
    exec:
      commandLine: "./connect_db.sh"
      component: runtime
      workingDir: /
  - id: disconnectDB
    exec:
      commandLine: "./disconnect_db.sh"
      component: runtime
      workingDir: /
  - id: cleanup
    exec:
      commandLine: "./cleanup.sh"
      component: tools
      workingDir: /
  - id: postStartCompositeCmd
    composite:
      label: Copy and Init Cache
      commands:
        - copy
        - initCache
      parallel: true
events:
  preStart:
    - "connectDB"
  postStart:
    - "postStartCompositeCmd" 
  preStop:
    - "disconnectDB"
  postStop:
    - "cleanup"
```

### preStart

PreStart events are executed as init containers for the project pod in the order they are specified.

The devfile command's `commandLine` and `workingDir` become the init container's command and as a result the devfile component container's `command` and `args` or the container image's `Command` and `Args` are overwritten. If a composite command with `parallel: true` is used, it will be executed sequentially as Kubernetes init containers only execute in sequence.

In the above example, PreStart is going to execute the devfile command `connectDB` as an init container for the odo component's main pod.

Caution should be exercised when using preStart with devfile container component that mount sources. File operations with preStart on the project sync directory may result in inconsistent behaviour.

Note that odo currently does not support preStart events.

### postStart

PostStart events are executed when the Kubernetes deployment for the odo component is created. 

In the above example, PostStart is going to execute the composite command `postStartCompositeCmd` once the odo component's deployment is created and the pod is up and running. The composite command `postStartCompositeCmd` has sub-commands `copy` and `initCache` which will be executed in parallel.

### preStop

PreStop events are executed before the Kubernetes deployment for the odo component is deleted. 

In the above example, PreStop is going to execute the devfile command `disconnectDB` before the odo component deployment is deleted.

### postStop

PostStop events are executed after the Kubernetes deployment for the odo component is deleted.

In the above example, PostStop will execute the devfile command `cleanup` after the component has been deleted.
Note that odo currently does not support postStop events.
