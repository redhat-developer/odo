# Support execution of the commands based on lifecycle events.

## Abstract
Add support for the lifecycle events which can be defined in 2.0 devfiles.

Our proposed solution is to perform the postStart event as a part of the **odo push** command, and the preStop event as a part of the **odo delete** command.

Lifecycle bindings for specific events - https://github.com/devfile/kubernetes-api/issues/32


## Motivation
Devfiles support commands that can be triggered based on dev lifecycle events. Odo will need to support/execute these commands at appropriate times within the flow.

With the implementation of lifecycle events, stack creators will be able to leverage all the capabilities of devfiles when building the correct experience/features for their stacks.

## User Stories
Each lifecycle event would be suited to individual user stories. We propose that at least one issue is made for each of them for tracking and implementation purposes.

An issue has already been created for postStart: [container initiliazation support](https://github.com/openshift/odo/issues/2936)

## Design overview

The lifecycle events that have been implemented in Devfile 2.0 are generic. So their effect/meaning might be slightly different when applying them to a workspace (Che) or a single project (odo). However, this proposal will aim to provide a similar user experience within the scope of a single project.

Our proposal is that the postStart event will run as a part of **odo push**, and that the preStop event will run as a part of **odo delete**.


### The flow for **odo push** including the *postStart* lifecycle event will be as follows:
 - Start the container(s) (as done today)
 - postStart - Exec the specified command(s) in the container - one time only
 - Rest of the command execution (as we do today - exec commands for build/run/test/debug)


**Starting the container(s)**
 - The containers specified in the Devfile are initialised and created if the component doesn’t already exist.

**postStart**
 - If the component is newly created, the postStart events are executed sequentially within the containers given in the Devfile. 
 - If any of the *postStart* commands do fail, then we would exit the **odo push** and report the error that had occurred. If it had failed to execute one of the postStart commands, the **odo push** might behave differently to what the stack creator had intended, so we would bail out early.
 - This functionality would also act as a replacement for `devInit` in devfile v1.0.

**Command Execution**
 - After postStart, we’d run the usual build/run/test/debug commands (depending on what sort of odo push parameters the user has provided) as usual.


### The flow for **odo delete** including the *preStop* lifecycle event will be as follows:
 - preStop - Exec the specified command(s) in the container
 - Clean up the pod and deployment etc. (as done today)

**preStop**
 - Exec the specified command(s) in their respective containers before deleting the deployment and any clean up begins.
  - If any of the *preStop* commands do fail, then we would exit the **odo delete** and report the error that had occurred. If it had failed to execute one of the preStop commands, the **odo delete** might behave differently to what the stack creator had intended, so we would bail out early.


### Flow with example devfile:
```
schemaVersion: "2.0.0"
metadata:
  name: test-devfile
projects:
  - name: nodejs-web-app
    git: 
      location: "https://github.com/che-samples/web-nodejs-sample.git"
components:
  - container:
      id: tools
      image: quay.io/eclipse/che-nodejs10-ubi:nightly
      name: "tools"
  - container:
      id: runtime
      image: quay.io/eclipse/che-nodejs10-ubi:nightly
      name: "runtime"
commands:
  - exec:
      id: download dependencies
      commandLine: "npm install"
      component: tools
      workingDir: ${CHE_PROJECTS_ROOT}/nodejs-web-app/app
      group:
        kind: build
  - exec:
      id: run the app
      commandLine: "nodemon app.js"
      component: runtime
      workingDir: ${CHE_PROJECTS_ROOT}/nodejs-web-app/app 
      group:
        kind: run  
  - exec:
      id: firstPostStartCmd
      commandLine: echo I am the first PostStart
      component: tools
      workingDir: ${CHE_PROJECTS_ROOT}/
  - exec:
      id: secondPostStartCmd
      commandLine: echo I am the second PostStart
      component: tools
      workingDir: ${CHE_PROJECTS_ROOT}/
  - exec:
      id: disconnectDatabase
      commandLine: echo disconnecting from the database
      component: tools
      workingDir: ${CHE_PROJECTS_ROOT}/
events:
  postStart:
   - "firstPostStartCmd"
   - "secondPostStartCmd"
  preStop:
   - "disconnectDatabase"
    
```

The example flow for **odo push** in this case would be:
 - create the containers
 - execute firstPostStartCmd within the tools container
 - execute secondPostStartCmd within the tools container
 - execute build command
 - execute run command

 The example flow for **odo delete** in this case would be:
 - execute disconnectDatabase
 - clean up the pod and deployment etc. (as done today)


### Conclusions:
 - We think that the most important events are the postStart, and the preStop because they have clear, reasonable use case within odo’s flow. 
 - We aren’t fully conclusive on the necessity of preStart and postStop. They do not currently have a useful case within odo.
 
 ## Future Evolution

 - *preStop* and *postStop* events are supported in Devfile 2.0, but aren't currently applicable to odo. In the future, a reason could arise where would benefit from these events.
 

