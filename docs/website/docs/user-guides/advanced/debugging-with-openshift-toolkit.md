---
title: Debugging an Application with OpenShift Toolkit
sidebar_position: 7
---

Debugging is an unavoidable part of development, and it can prove even more difficult when developing an application that runs remotely.

However, this task is made absurdly simple with the help of OpenShift Toolkit.

## OpenShift Toolkit
[OpenShift Toolkit](https://github.com/redhat-developer/intellij-openshift-connector) is an IDE plugin that allows you to do all things that `odo` does, i.e. create, test, debug and deploy cloud-native applications on a cloud-native environment in simple steps.
`odo` enables this plugin to do what it does.

## Prerequisites
1. [You have logged in to your cluster](../quickstart/nodejs.md#step-1-connect-to-your-cluster-and-create-a-new-namespace-or-project).
2. You have [initialized an application with `odo`](/docs/command-reference/init), for example [the Node.JS quickstart application](../quickstart/nodejs.md#step-2-initializing-your-application-odo-init).
:::note
 This tutorial uses a Node.js application, but you can use any application that has Devfile with debug command defined in it. If your Devfile does not contain a debug command, refer to [Configure Devfile to support debugging](#configure-devfile-to-support-debugging).
:::
3. You have [installed](/docs/overview/installation#ide-installation) the OpenShift Toolkit Plugin in your preferred VS Code or a JetBrains IDE.
4. You have opened the application in the IDE.

In the plugin window, you should be able to see the cluster you are logged into in "APPLICATION EXPLORER" section, and your component "my-nodejs-app" in "COMPONENTS" section.

![Pre-requisite Setup](../../assets/user-guides/advanced/Prerequisite%20Setup.png)

## Step 1. Start the Dev session to run the application on cluster

1. Right click on "my-nodejs-app" and select "Start on Dev".

![Starting Dev session](../../assets/user-guides/advanced/Start%20Dev%20Session.png)

2. Wait until the application is running on the cluster, i.e. until you see "Keyboard Commands" appear in your "TERMINAL" window.

![Wait until Dev session finishes](../../assets/user-guides/advanced/Wait%20until%20Dev%20Session%20finishes.png)

Our application is now available at 127.0.0.1:20001. The debug server is available at 127.0.0.1:20002.

## Step 2. Start the Debugging session

1. Right click on "my-nodejs-app" and select "Debug".

![Select Debug](../../assets/user-guides/advanced/Select%20Debug%20Session.png)

2. Debug session should have started successfully in the container at the debug port, in this case, 5858. And you must be looking at the "DEBUG CONSOLE".

![Debug session starts](../../assets/user-guides/advanced/Debug%20Session%20Starts.png)

## Step 3. Set Breakpoints in the application

Now that the debug session is running, we can set breakpoints in the code.

1. Open 'server.js' file if you haven't opened it already. We will set a breakpoint on Line 55 by clicking the red dot that appears right next to line numbers.

![Add breakpoint](../../assets/user-guides/advanced/Add%20Breakpoint.png)

2. From a new terminal, or a browser window, ping the URL at which the application is available, in this case, it is 127.0.0.1:20001.

![Ping Application](../../assets/user-guides/advanced/Ping%20Application.png)

3. The debug session should halt execution at the breakpoint, at which point you can start debugging the application.

![Application Debugged](../../assets/user-guides/advanced/Application%20Debugged.png)


## Configure Devfile to support debugging
Here, we are taking example of a Go devfile that currently does not have a debug command out-of-the-box.
<details>
<summary>Sample Go Devfile</summary>

```yaml
schemaVersion: 2.1.0
metadata:
  description: "Go is an open source programming language that makes it easy to build simple, reliable, and efficient software."
  displayName: Go Runtime
  icon: https://raw.githubusercontent.com/devfile-samples/devfile-stack-icons/main/golang.svg
  name: go
  projectType: Go
  provider: Red Hat
  language: Go
  tags:
    - Go
  version: 1.0.2
starterProjects:
  - name: go-starter
    description: A Go project with a simple HTTP server
    git:
      checkoutFrom:
        revision: main
      remotes:
        origin: https://github.com/devfile-samples/devfile-stack-go.git
components:
  - container:
      endpoints:
        - name: http-go
          targetPort: 8080
      image: registry.access.redhat.com/ubi9/go-toolset:latest
      args: ["tail", "-f", "/dev/null"]
      memoryLimit: 1024Mi
      mountSources: true
    name: runtime
commands:
  - exec:
      env:
        - name: GOPATH
          value: ${PROJECT_SOURCE}/.go
        - name: GOCACHE
          value: ${PROJECT_SOURCE}/.cache
      commandLine: go build main.go
      component: runtime
      group:
        isDefault: true
        kind: build
      workingDir: ${PROJECT_SOURCE}
    id: build
  - exec:
      commandLine: ./main
      component: runtime
      group:
        isDefault: true
        kind: run
      workingDir: ${PROJECT_SOURCE}
    id: run
```
</details>

1. Add an exec command with `group`:`kind` set to `debug`. The debugger tool you use must be able to start a debug server that we can later on connect to. The binary for your debugger tool should be made available by the container component image.
```yaml
commands:
- exec:
    env:
    - name: GOPATH
      value: ${PROJECT_SOURCE}/.go
    - name: GOCACHE
      value: ${PROJECT_SOURCE}/.cache
    commandLine: |
        dlv \
          --listen=127.0.0.1:${DEBUG_PORT} \
          --only-same-user=false \
          --headless=true \
          --api-version=2 \
          --accept-multiclient \
          debug --continue main.go
    component: runtime
    group:
      isDefault: true
      kind: debug
    workingDir: ${PROJECT_SOURCE}
  id: debug
```
For the example above, we use [`dlv`](https://github.com/go-delve/delve) debugger for debugging a Go application and it listens to the port exposed by the environment variable *DEBUG_PORT* inside the container. The debug command references a container component called "runtime".

2. Add Debug endpoint to the container component's [`endpoints`](https://devfile.io/docs/2.2.0/defining-endpoints) with `exposure` set to `none` so that it cannot be accessed from outside, and export the debug port number via `DEBUG_PORT` `env` variable.

The debug endpoint name must be named **debug** or be prefixed by **debug-** so that `odo` can recognize it as a debug port.

```yaml
components:
  - container:
      endpoints:
      - name: http-go
        targetPort: 8080
        # highlight-start
      - exposure: none
        name: debug
        targetPort: 5858
        # highlight-end
      image: registry.access.redhat.com/ubi9/go-toolset:latest
      args: ["tail", "-f", "/dev/null"]
    # highlight-start
      env:
      - name: DEBUG_PORT
        value: '5858'
    # highlight-end
      memoryLimit: 1024Mi
      mountSources: true
    name: runtime
```

For the example above, we assume that the "runtime" container's `image` provides the binary for delve debugger. We also add an endpoint called "debug" with `targetPort` set to *5858* and `exposure` set to `none`. We also export debug port number via `env` variable called `DEBUG_PORT`.

The final Devfile should look like the following:
<details>
<summary>Go Devfile configured for debugging</summary>

```yaml showLineNumbers
commands:
- exec:
    commandLine: go build main.go
    component: runtime
    env:
    - name: GOPATH
      value: ${PROJECT_SOURCE}/.go
    - name: GOCACHE
      value: ${PROJECT_SOURCE}/.cache
    group:
      isDefault: true
      kind: build
    workingDir: ${PROJECT_SOURCE}
  id: build
- exec:
    commandLine: ./main
    component: runtime
    group:
      isDefault: true
      kind: run
    workingDir: ${PROJECT_SOURCE}
  id: run
# highlight-start
- exec:
    env:
    - name: GOPATH
      value: ${PROJECT_SOURCE}/.go
    - name: GOCACHE
      value: ${PROJECT_SOURCE}/.cache
    commandLine: |
        dlv \
          --listen=127.0.0.1:${DEBUG_PORT} \
          --only-same-user=false \
          --headless=true \
          --api-version=2 \
          --accept-multiclient \
          debug --continue main.go
    component: runtime
    group:
      isDefault: true
      kind: debug
    workingDir: ${PROJECT_SOURCE}
  id: debug
# highlight-end
components:
- container:
    args:
    - tail
    - -f
    - /dev/null
    endpoints:
    - name: http-go
      targetPort: 8080
# highlight-start
    - name: debug
      exposure: none
      targetPort: 5858
    env:
    - name: DEBUG_PORT
      value: '5858'
# highlight-end
    image: registry.access.redhat.com/ubi9/go-toolset:latest
    memoryLimit: 1024Mi
    mountSources: true
  name: runtime
metadata:
  description: Go is an open source programming language that makes it easy to build
    simple, reliable, and efficient software.
  displayName: Go Runtime
  icon: https://raw.githubusercontent.com/devfile-samples/devfile-stack-icons/main/golang.svg
  language: Go
  name: my-go-app
  projectType: Go
  provider: Red Hat
  tags:
  - Go
  version: 1.0.2
schemaVersion: 2.1.0
starterProjects:
- description: A Go project with a simple HTTP server
  git:
    checkoutFrom:
      revision: main
    remotes:
      origin: https://github.com/devfile-samples/devfile-stack-go.git
  name: go-starter
```
</details>

## Extra Resources
To learn more about running and debugging an application on cluster with OpenShift Toolkit, see the links below.
1. [Using OpenShift Toolkit - project with existing devfile](https://www.youtube.com/watch?v=2jfV0QqG8Sg)
2. [Using OpenShift Toolkit with two microservices](https://www.youtube.com/watch?v=8SpV6UZ23_c)
3. [Using OpenShift Toolkit - project without devfile](https://www.youtube.com/watch?v=sqqznqoWNSg)
