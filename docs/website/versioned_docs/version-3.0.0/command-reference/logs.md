---
title: odo logs
---

## Description

`odo logs` is used to display the logs for all the containers odo created for the component under current working 
directory.

## Running the command 

If you haven't already done so, you must [initialize](../command-reference/init) your source code with the `odo 
init` command. This document uses [this git project](https://github.com/odo-devfiles/nodejs-ex) and the following 
devfile to explain the flow of things in `odo logs` command:

```yaml
# This devfile has both inner and outer loop components. The outer loop components do start up on the cluster.
# It creates following resources on the cluster:
# - inner loop - Deployment for the component; a Pod for the k8s component named `innerloop-pod`
# - outer loop - Two Deployments named devfile-nodejs-deploy & devfile-nodejs-deploy-2; a Pod named for the k8s component named `outerloop-pod`
schemaVersion: 2.2.0
metadata:
   language: javascript
   name: devfile-nodejs-deploy
   projectType: nodejs
variables:
   CONTAINER_IMAGE: quay.io/tkral/devfile-nodejs-deploy:latest
commands:
   - id: install
     exec:
        commandLine: npm install
        component: runtime
        group:
           isDefault: true
           kind: build
        workingDir: $PROJECT_SOURCE
   - id: run
     exec:
        commandLine: npm start
        component: runtime
        group:
           isDefault: true
           kind: run
        workingDir: $PROJECT_SOURCE
   - id: build-image
     apply:
        component: prod-image
   - id: deploy-deployment
     apply:
        component: outerloop-deploy
   - id: deploy-another-deployment
     apply:
        component: another-deployment
   - id: outerloop-pod-command
     apply:
        component: outerloop-pod
   - id: deploy
     composite:
        commands:
           - build-image
           - deploy-deployment
           - deploy-another-deployment
           - outerloop-pod-command
        group:
           kind: deploy
           isDefault: true
components:
   - container:
        endpoints:
           - name: http-3000
             targetPort: 3000
        image: registry.access.redhat.com/ubi8/nodejs-14:latest
        memoryLimit: 1024Mi
        mountSources: true
     name: runtime
   - name: prod-image
     image:
        imageName: "{{CONTAINER_IMAGE}}"
        dockerfile:
           uri: ./Dockerfile
           buildContext: ${PROJECT_SOURCE}
   - name: outerloop-deploy
     kubernetes:
        inlined: |
           kind: Deployment
           apiVersion: apps/v1
           metadata:
             name: devfile-nodejs-deploy
           spec:
             replicas: 1
             selector:
               matchLabels:
                 app: devfile-nodejs-deploy
             template:
               metadata:
                 labels:
                   app: devfile-nodejs-deploy
               spec:
                 containers:
                   - name: main
                     image: "{{CONTAINER_IMAGE}}"
   - name: another-deployment
     kubernetes:
        inlined: |
           kind: Deployment
           apiVersion: apps/v1
           metadata:
             name: devfile-nodejs-deploy-2
           spec:
             replicas: 1
             selector:
               matchLabels:
                 app: devfile-nodejs-deploy-2
             template:
               metadata:
                 labels:
                   app: devfile-nodejs-deploy-2
               spec:
                 containers:
                   - name: main
                     image: "{{CONTAINER_IMAGE}}"
   - name: innerloop-pod
     kubernetes:
        inlined: |
           apiVersion: v1
           kind: Pod
           metadata:
             name: myapp
           spec:
             containers:
             - name: main
               image: "{{CONTAINER_IMAGE}}"
   - name: outerloop-pod
     kubernetes:
        inlined: |
           apiVersion: v1
           kind: Pod
           metadata:
             name: myapp
           spec:
             containers:
             - name: main
               image: "{{CONTAINER_IMAGE}}"
 ```

Notice that multiple containers have been named as `main` to show how `odo logs` would display logs when more than one
container have the same name.

### Dev mode (Inner loop)

Run the `odo dev` command so that odo can create the resources on the Kubernetes cluster. When you execute `odo dev`,
odo creates following resources on the Kubernetes cluster:
1. Deployment for the component `devfile-nodejs-deploy` itself. Containers for this are created using `.components.
   container`.
2. Pod named `innerloop-pod`.

When you run `odo logs --dev`, you should see logs from all the containers started by `odo dev` command. Each line 
is prefixed with `<container-name>:` to easily distinguish which the container the logs belong to:
```shell
$ odo logs --dev
runtime: time="2022-06-14T09:03:32Z" level=info msg="create process:devrun" 
runtime: time="2022-06-14T09:03:32Z" level=info msg="create process:debugrun" 
runtime: time="2022-06-14T09:03:32Z" level=info msg="try to start program" program=devrun 
runtime: time="2022-06-14T09:03:32Z" level=info msg="success to start program" program=devrun 
runtime: time="2022-06-14T09:03:33Z" level=debug msg="no auth required" 
runtime: time="2022-06-14T09:03:33Z" level=debug msg="wait program exit" program=devrun 
runtime: time="2022-06-14T09:03:33Z" level=info msg="program stopped with status:exit status 0" program=devrun 
runtime: time="2022-06-14T09:03:33Z" level=info msg="Don't start the stopped program because its autorestart flag is false" program=devrun 
runtime: time="2022-06-14T09:03:35Z" level=debug msg="no auth required" 
runtime: time="2022-06-14T09:03:35Z" level=debug msg="succeed to find process:devrun" 
runtime: time="2022-06-14T09:03:35Z" level=info msg="try to start program" program=devrun 
runtime: time="2022-06-14T09:03:35Z" level=info msg="success to start program" program=devrun 
runtime: ODO_COMMAND_RUN is npm start
runtime: Changing directory to $PROJECT_SOURCE
runtime: Executing command cd $PROJECT_SOURCE && npm start
runtime: 
runtime: > nodejs-starter@1.0.0 start /projects
runtime: > node server.js
runtime: 
runtime: App started on PORT 3000
runtime: time="2022-06-14T09:03:36Z" level=debug msg="wait program exit" program=devrun 
runtime: time="2022-06-14T09:03:37Z" level=debug msg="no auth required" 
main: 
main: > nodejs-starter@1.0.0 start /opt/app-root/src
main: > node server.js
main: 
main: App started on PORT 3000
```

### Deploy mode

If you are using the git repo mentioned at the beginning of this document, note that it doesn't have a Dockerfile to 
build the image which odo does when you do `odo deploy`. To skip image creation and still be able to use Deploy mode 
for the component, run `odo deploy` using below:

```shell
PODMAN_CMD=echo DOCKER_CMD=echo odo deploy
```

odo creates following resources on the Kubernetes cluster when you run `odo deploy`:
1. Deployment named `devfile-nodejs-deploy`.
2. Another Deployment named `devfile-nodejs-deploy-2`.
3. Pod named `myapp`.

When you run `odo logs --deploy`, you should see logs from all the containers started by `odo deploy` command. Each 
line is prefixed with `<container-name>:` to easily distinguish which the container the logs belong to:
```shell
$ odo logs --deploy
main: 
main: > nodejs-starter@1.0.0 start /opt/app-root/src
main: > node server.js
main: 
main: App started on PORT 3000
main[1]: 
main[1]: > nodejs-starter@1.0.0 start /opt/app-root/src
main[1]: > node server.js
main[1]: 
main[1]: App started on PORT 3000
main[2]: 
main[2]: > nodejs-starter@1.0.0 start /opt/app-root/src
main[2]: > node server.js
main[2]: 
main[2]: App started on PORT 3000
```
odo helps distinguish between multiple containers named `main` by appending a numeric value to it.

### `odo logs` without specifying the mode

When you run `odo logs` without specifying the Dev mode (using `--dev` flag) or Deploy mode (using `--deploy` flag), 
it shows logs for all the modes that a component is running in. Since we have started both Dev mode as well as Deploy 
mode here, it will show logs for all containers running in both modes:

```shell
$ odo logs
runtime: time="2022-06-14T09:30:35Z" level=info msg="create process:debugrun" 
runtime: time="2022-06-14T09:30:35Z" level=info msg="create process:devrun" 
runtime: time="2022-06-14T09:30:35Z" level=info msg="try to start program" program=devrun 
runtime: time="2022-06-14T09:30:35Z" level=info msg="success to start program" program=devrun 
runtime: time="2022-06-14T09:30:36Z" level=debug msg="no auth required" 
runtime: time="2022-06-14T09:30:36Z" level=debug msg="wait program exit" program=devrun 
runtime: time="2022-06-14T09:30:36Z" level=info msg="program stopped with status:exit status 0" program=devrun 
runtime: time="2022-06-14T09:30:36Z" level=info msg="Don't start the stopped program because its autorestart flag is false" program=devrun 
runtime: time="2022-06-14T09:30:39Z" level=debug msg="no auth required" 
runtime: time="2022-06-14T09:30:39Z" level=debug msg="succeed to find process:devrun" 
runtime: time="2022-06-14T09:30:39Z" level=info msg="try to start program" program=devrun 
runtime: time="2022-06-14T09:30:39Z" level=info msg="success to start program" program=devrun 
runtime: ODO_COMMAND_RUN is npm start
runtime: Changing directory to $PROJECT_SOURCE
runtime: Executing command cd $PROJECT_SOURCE && npm start
runtime: 
runtime: > nodejs-starter@1.0.0 start /projects
runtime: > node server.js
runtime: 
runtime: App started on PORT 3000
runtime: time="2022-06-14T09:30:40Z" level=debug msg="wait program exit" program=devrun 
runtime: time="2022-06-14T09:30:41Z" level=debug msg="no auth required" 
main: 
main: > nodejs-starter@1.0.0 start /opt/app-root/src
main: > node server.js
main: 
main: App started on PORT 3000
main[1]: 
main[1]: > nodejs-starter@1.0.0 start /opt/app-root/src
main[1]: > node server.js
main[1]: 
main[1]: App started on PORT 3000
main[2]: 
main[2]: > nodejs-starter@1.0.0 start /opt/app-root/src
main[2]: > node server.js
main[2]: 
main[2]: App started on PORT 3000
main[3]: 
main[3]: > nodejs-starter@1.0.0 start /opt/app-root/src
main[3]: > node server.js
main[3]: 
main[3]: App started on PORT 3000

```