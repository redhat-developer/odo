# Support execution of the commands based on lifecycle events

## Abstract
Devfile 2.0.0 allows binding of commands to specific lifecycle events: https://github.com/devfile/kubernetes-api/issues/32. These commands are meant to be executed when those specific lifecycle events are triggered.

The lifecycle events that have been implemented in Devfile 2.0.0 are generic. Their effect/meaning might be slightly different in the context a workspace (Che) or a single project (odo). However, odo should aim to provide a similar user experience within the scope of a single project.

## Motivation
Devfiles support commands that can be triggered based on development lifecycle events. With the implementation of lifecycle events, stack creators will be able to leverage all the capabilities of devfiles when building the correct experience/features for their stacks. `odo` should support/execute these commands at appropriate times within `odo` development flow.

## User Stories
Each lifecycle event would be suited to individual user stories. We propose that at least one issue is made for each of them for tracking and implementation purposes.

- Add support for `preStart` commands: [preStart lifecycle event support](https://github.com/redhat-developer/odo/issues/3565)
- Add support for `postStart` commands: [container initialization support](https://github.com/redhat-developer/odo/issues/2936)
- Add support for `preStop` commands:  [preStop lifecycle event support](https://github.com/redhat-developer/odo/issues/3566)
- Add support for `postStop` commands: [postStop lifecycle event support](https://github.com/redhat-developer/odo/issues/3577)

## Design overview
Some of the events might not be useful for `odo` to adopt. We should only support the ones that have a clear use-case, and then add more as other use-cases emerge.

- `preStart` - support executing these commands via init containers
- `postStart` - support executing these commands as part of component initialization
- `preStop` - support executing these commands as part of the component delete
- `postStop` - support executing these commands via standalone pod or job and delete it once complete

As per the Devfile 2.0.0 design specification: 
- Commands associated with a lifecycle binding should provide a mechanism to figure out when they have completed. That means that the command should be terminating or having a readiness probe.

- Commands associated with the same lifecycle binding do not guarantee to be executed sequentially or in any order. To run commands in a given order a user should use a composite.

The flow for initializing and deleting components should be updated to add execution of commands bound to relevant lifecycle event while initializing and destroying any containers.

### preStart
 - `preStart` command(s) that are specified in the devfile will be translated into entry points for their specified containers, and added to the pod spec as init containers. 
 - These are commands that the stack creator intends to run before the main containers are created/initialized. 

### Initialize containers
- Identify which containers have been created as part of creating/updating the component deployment
  - Initial implementation can assume containers are only initialised during the component creation (first push). We can improve this in future to allow execution of postStart commands at a more granular level.
- Sync the source code to the containers
- Iterate over the list of `postStart` commands and identify the ones that are associated with newly started containers
- Execute the identified commands one by one
  - execute the command in the container
    - on failure, display a meaningful error message and stop
    - on success, execute next command in the list (if any)
- Consider the component has been successfully initialised and continue as usual (e.g. execute build command)

  Note: This functionality would act as a replacement for `devInit` in devfile v1.0.

### Destroy containers
- Identify which containers will be deleted as of a result of deleting the component deployment
- Prepare a list of `preStop` commands that are associated with the containers being destroyed
- Execute the preStop commands one by one
  - execute the command in the container
    - on failure, display a meaningful error message and stop
    - on success, execute next command in the list (if any)
- Destroy/delete the component and continue as usual

### postStop
- Execute the `postStop` commands one by one
  - via standalone pod or job and delete it once complete

## Example devfile with lifecycle binding:
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

 The example flow for **odo delete**, in this case, would be:
 - execute disconnectDatabase
 - clean up the pod and deployment etc. (as done today)
 
## Future Evolution

The initial implementation of postStart events assumes all containers are initilised during component startup (first push). In future we could make this granular and allow executing these postStart commands if containers are restarted by other commands/events too. 

Example: A potential flow could be to check for postStart commands associated with a specific container when initializing it, storing those commands in an Object(?) and then executing only those postStart commands when the containers have all finished initializing (in the order that they have been defined within the devfile). This would benefit cases where an individual container has been re-initialized, but not the whole component. For example, if a build container has been restarted as a part of **odo push --force** and there is a postStart command associated with the build container in the devfile that is required to re-run, that single postStart command would be executed before continuing to the build/run phase of the push command.