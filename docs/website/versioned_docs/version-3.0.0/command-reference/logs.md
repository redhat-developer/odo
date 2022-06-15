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

Running the `odo dev` command for aforementioned git repo and devfile creates following resources on the 
Kubernetes cluster:
1. Deployment for the component `devfile-nodejs-deploy` itself. Containers for this are created using `.components.
   container`.
2. Pod named `innerloop-pod`.


Running `odo deploy` (use `PODMAN_CMD=echo DOCKER_CMD=echo odo deploy` to skip building the image) command for 
aforementioned git repo and devfile creates following resources on the Kubernetes cluster:
1. Deployment named `devfile-nodejs-deploy`.
2. Another Deployment named `devfile-nodejs-deploy-2`.
3. Pod named `myapp`.

`odo logs` command can be used with below flags:
* Use `odo logs --dev` to see the logs for the containers created by `odo dev` command.
* Use `odo logs --deploy` to see the logs for the containers created by `odo deploy` command.
* Use `odo logs` (without any flag) to see the logs of all the containers created by both `odo dev` and `odo deploy`.

Note that since multiple containers are named the same (`main`), the `odo logs` output appends a number to container 
name to help differentiate between the containers.