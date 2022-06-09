---
title: odo logs
---

## Description

`odo logs` is used to display the logs for all the containers odo created for the component under current working 
directory.

## Running the command 

If you haven't already done so, you must [initialize](../command-reference/init) your source code with the `odo 
init` command. Next, run the `odo dev` command so that odo can create the resources on the Kubernetes cluster.

Consider a devfile.yaml like below which was used to create inner loop resources using `odo dev`. Notice that 
multiple containers have been named as `main` to show how `odo logs` would display logs when more than one 
containers have the same name:
```yaml
metadata:
  description: Stack with Node.js 14
  displayName: Node.js Runtime
  icon: https://nodejs.org/static/images/logos/nodejs-new-pantone-black.svg
  language: javascript
  name: node
  projectType: nodejs
  tags:
    - NodeJS
    - Express
    - ubi8
  version: 1.0.1
schemaVersion: 2.0.0
starterProjects:
  - git:
      remotes:
        origin: https://github.com/odo-devfiles/nodejs-ex.git
    name: nodejs-starter
commands:
- exec:
    commandLine: npm install
    component: runtime
    group:
      isDefault: true
      kind: build
    workingDir: ${PROJECT_SOURCE}
  id: install
- exec:
    commandLine: npm start
    component: runtime
    group:
      isDefault: true
      kind: run
    workingDir: ${PROJECT_SOURCE}
  id: run
- exec:
    commandLine: npm run debug
    component: runtime
    group:
      isDefault: true
      kind: debug
    workingDir: ${PROJECT_SOURCE}
  id: debug
- exec:
    commandLine: npm test
    component: runtime
    group:
      isDefault: true
      kind: test
    workingDir: ${PROJECT_SOURCE}
  id: test
components:
- container:
    endpoints:
    - name: http-3000
      targetPort: 3000
    image: registry.access.redhat.com/ubi8/nodejs-14:latest
    memoryLimit: 1024Mi
    mountSources: true
  name: runtime
- name: infinitepodone
  kubernetes:  
    inlined: |
      apiVersion: v1
      kind: Pod
      metadata:
        name: infinitepodone
      spec:
        containers:
          - name: main
            image: docker.io/dharmit/infiniteloop
- name: infinitedeployment
  kubernetes:
    inlined: |
      apiVersion: apps/v1
      kind: Deployment
      metadata:
        name: infinitedeployment
      spec:
        replicas: 1
        selector:
          matchLabels:
            app: infinite
        template:
          metadata:
            labels:
              app: infinite
          spec:
            containers:
            - name: main
              image: docker.io/dharmit/infiniteloop
 ```
When you do `odo dev`, odo creates pods for:
1. The component named `node` itself. Containers for this are created using `.components.container`.
2. Kubernetes component named `infinitepodone`
3. Kubernetes component named `infinitedeployment`. As can be seen under `.spec.template.spec.containers` for this 
   particular component, it creates two containers for it.

When you run `odo logs`, you should see logs from all these containers. Each line is prefixed with 
`<container-name>:` to easily distinguish which the container the logs belong to. Since we named multiple 
containers in the `devfile.yaml` as `container`, `odo logs` has distinguished these containers as `container`, 
`container1` and `container2`:

```shell
$ odo logs
main: Fri May 27 06:17:30 UTC 2022 - this is infinite for loop
main: Fri May 27 06:17:31 UTC 2022 - this is infinite for loop
main: Fri May 27 06:17:32 UTC 2022 - this is infinite for loop
main: Fri May 27 06:17:33 UTC 2022 - this is infinite for loop
main: Fri May 27 06:17:34 UTC 2022 - this is infinite for loop
main: Fri May 27 06:17:35 UTC 2022 - this is infinite for loop
main: Fri May 27 06:17:36 UTC 2022 - this is infinite for loop
main: Fri May 27 06:17:37 UTC 2022 - this is infinite for loop
main: Fri May 27 06:17:38 UTC 2022 - this is infinite for loop
main: Fri May 27 06:17:39 UTC 2022 - this is infinite for loop
main: Fri May 27 06:17:40 UTC 2022 - this is infinite for loop
main: Fri May 27 06:17:41 UTC 2022 - this is infinite for loop
main: Fri May 27 06:17:42 UTC 2022 - this is infinite for loop
main: Fri May 27 06:17:44 UTC 2022 - this is infinite for loop
main: Fri May 27 06:17:45 UTC 2022 - this is infinite for loop
main: Fri May 27 06:17:46 UTC 2022 - this is infinite for loop
main: Fri May 27 06:17:47 UTC 2022 - this is infinite for loop
runtime: time="2022-05-27T06:17:36Z" level=info msg="create process:devrun" 
runtime: time="2022-05-27T06:17:36Z" level=info msg="create process:debugrun" 
runtime: time="2022-05-27T06:17:36Z" level=info msg="try to start program" program=devrun 
runtime: time="2022-05-27T06:17:36Z" level=info msg="success to start program" program=devrun 
runtime: time="2022-05-27T06:17:37Z" level=debug msg="no auth required" 
runtime: time="2022-05-27T06:17:37Z" level=debug msg="wait program exit" program=devrun 
runtime: time="2022-05-27T06:17:37Z" level=info msg="program stopped with status:exit status 0" program=devrun 
runtime: time="2022-05-27T06:17:37Z" level=info msg="Don't start the stopped program because its autorestart flag is false" program=devrun 
runtime: time="2022-05-27T06:17:41Z" level=debug msg="no auth required" 
runtime: time="2022-05-27T06:17:41Z" level=debug msg="succeed to find process:devrun" 
runtime: time="2022-05-27T06:17:41Z" level=info msg="try to start program" program=devrun 
runtime: time="2022-05-27T06:17:41Z" level=info msg="success to start program" program=devrun 
runtime: ODO_COMMAND_RUN is npm start
runtime: Changing directory to ${PROJECT_SOURCE}
runtime: Executing command cd ${PROJECT_SOURCE} && npm start
runtime: 
runtime: > nodejs-starter@1.0.0 start /projects
runtime: > node server.js
runtime: 
runtime: App started on PORT 3000
runtime: time="2022-05-27T06:17:42Z" level=debug msg="wait program exit" program=devrun 
runtime: time="2022-05-27T06:17:43Z" level=debug msg="no auth required" 
main1: Fri May 27 06:17:34 UTC 2022 - this is infinite for loop
main1: Fri May 27 06:17:35 UTC 2022 - this is infinite for loop
main1: Fri May 27 06:17:36 UTC 2022 - this is infinite for loop
main1: Fri May 27 06:17:37 UTC 2022 - this is infinite for loop
main1: Fri May 27 06:17:38 UTC 2022 - this is infinite for loop
main1: Fri May 27 06:17:39 UTC 2022 - this is infinite for loop
main1: Fri May 27 06:17:40 UTC 2022 - this is infinite for loop
```