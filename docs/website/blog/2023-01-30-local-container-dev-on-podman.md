---
title: Local container development with Podman and odo
author: Armel Soro
author_url: https://github.com/rm3l
author_image_url: https://github.com/rm3l.png
tags: ["local", "container", "development", "podman", "container-dev"]
slug: local-container-development-with-podman-and-odo
---

<div>
<img
src={require('../static/img/odo_podman.png').default}
alt="odo and Podman"
style={{width: '50%', height: '50%', display: 'block', marginLeft: 'auto', marginRight: 'auto', marginBottom: '10px'}}
/>
</div>

So far, `odo` has been mainly focusing on container development on [Kubernetes](https://kubernetes.io/) and [OpenShift](https://www.redhat.com/en/technologies/cloud-computing/openshift) clusters.

In this post, we will showcase the experimental support we have recently added for [Podman](https://podman.io/).
We will see how `odo` can leverage Podman for local development in containers with no requirement whatsoever on any cluster — making it easier to iterate on the application locally and transition to Kubernetes or OpenShift later on.

<!--truncate-->

## Prerequisites

- [`odo`](https://odo.dev/docs/overview/installation) 3.3.0 or later. Support for Podman was added as an experimental feature in 3.3.0; 
so we recommend you [install the latest version](https://odo.dev/docs/overview/installation) of `odo`.
- [Podman](https://podman.io/getting-started/installation).
- [Podman Desktop](https://podman-desktop.io/), optional.

## Working locally with Podman

Let's revisit one of our quickstart guides, say the [Golang one](../../docs/user-guides/quickstart/go), to make it work with Podman.

### Step 0. Creating the initial source code (optional)

We will create the example source code by using some popular frameworks.

Before we begin, we will create a new directory and cd into it.
```shell
mkdir quickstart-demo && cd quickstart-demo
```

This is *optional* and you may use an existing project instead (make sure you `cd` into the project directory before running any odo commands) or a starter project from `odo init`.

For Go, we will create our own application using the standard library:

1. Create the following `main.go` file:

```go
package main

import (
  "fmt"
  "log"
  "net/http"
)

func main() {
  const addr = "0.0.0.0:8080"
  http.HandleFunc("/", HelloServer)
  log.Println("Up and running on", addr)
  http.ListenAndServe(addr, nil)
}

func HelloServer(w http.ResponseWriter, r *http.Request) {
  log.Println("New request:", *r)
  fmt.Fprintf(w, "Hello, %s!", r.URL.Path[1:])
}
```

2. Initialize a `go.mod` file:

```console
go mod init my.example.go.project
```
<details>
<summary>Example</summary>

```shell
$ go mod init my.example.go.project
go: creating new go.mod: module my.example.go.project
go: to add module requirements and sums:
	go mod tidy
```
</details>

Your source code has now been generated and created in the directory.

### Step 1. Initializing your application (`odo init`)

Now we'll initialize the application by creating a `devfile.yaml` to be deployed.

`odo` handles this automatically with the `odo init` command by auto-detecting the source code and downloading the appropriate Devfile.

**Note:** If you skipped *Step 0*, select a "starter project" when running `odo init`.

Let's run `odo init` and select `Go`:

```console
odo init
```

<details>
<summary>Sample Output</summary>

```console
$ odo init
  __
 /  \__     Initializing a new component
 \__/  \    Files: Source code detected, a Devfile will be determined based upon source code autodetection
 /  \__/    odo version: v3.6.0
 \__/

Interactive mode enabled, please answer the following questions:
Based on the files in the current directory odo detected
Language: Go
Project type: Go
The devfile "go:1.0.2" from the registry "Staging" will be downloaded.
? Is this correct? Yes
 ✓  Downloading devfile "go:1.0.2" from registry "Staging" [1s]

↪ Container Configuration "runtime":
  OPEN PORTS:
    - 8080
  ENVIRONMENT VARIABLES:

? Select container for which you want to change configuration? NONE - configuration is correct
? Enter component name: quickstart-demo

You can automate this command by executing:
   odo init --name quickstart-demo --devfile go --devfile-registry Staging --devfile-version 1.0.2

Your new component 'quickstart-demo' is ready in the current directory.
To start editing your component, use 'odo dev' and open this folder in your favorite IDE.
Changes will be directly reflected on the cluster.
```

:::note
If you skipped Step 0 and selected "starter project", your output will be slightly different.
:::

</details>

### Step 2. Enabling the experimental mode

Because the support for Podman is still experimental at the time of writing, we first need to explicitly opt-in.

Enabling the experimental mode can be done by setting the `ODO_EXPERIMENTAL_MODE` environment variable to `true` in the terminal session, like so:

```console
export ODO_EXPERIMENTAL_MODE=true
```

### Step 3. Iterating on your application locally on containers (`odo dev`)

Now that we've generated our code as well as our Devfile, let's start iterating on our application locally by starting a Development session with `odo dev`,
but targeting our local Podman.

`odo dev` on Podman will use the same [inner loop development](/docs/introduction#what-is-inner-loop-and-outer-loop) as for the cluster mode,
allowing you to code, build, run and test the application in a continuous workflow.

Once you run `odo dev --platform=podman`, you can freely edit the application code in your favorite IDE and watch as `odo` rebuilds and redeploys it.

Let's run `odo dev --platform=podman` to start development on your `Go` application:

```console
odo dev --platform=podman
```

<details>
<summary>Sample Output</summary>

```console
$ odo dev --platform=podman
============================================================================
⚠ Experimental mode enabled. Use at your own risk.
More details on https://odo.dev/docs/user-guides/advanced/experimental-mode
============================================================================

  __
 /  \__     Developing using the "quickstart-demo" Devfile
 \__/  \    Platform: podman
 /  \__/    odo version: v3.6.0
 \__/

↪ Running on podman in Dev mode
 ✓  Deploying pod [5s]
 ✓  Building your application in container (command: build) [693ms]
 •  Executing the application (command: run)  ...
 -  Forwarding from 127.0.0.1:20001 -> 8080

↪ Dev mode
 Status:
 Watching for changes in the current directory /tmp/test-go-podman/quickstart-demo

 Keyboard Commands:
[Ctrl+c] - Exit and delete resources from podman
     [p] - Manually apply local changes to the application on podman
```
</details>

You can now access the application at [127.0.0.1:20001](http://127.0.0.1:20001) in your local browser and start your development loop. `odo` will watch for changes and push the code for real-time updates.

<details>
<summary>Example</summary>

```console
$ curl http://127.0.0.1:20001/world

Hello, world!
```
</details>

We can optionally open the Podman Desktop application to take a look at the resources `odo` has created for our application on Podman:

<a href="/video/odo-dev-podman-demo.webm" target="_blank">
    <video style={{width:'100%', height:'100%'}} autoPlay loop muted><source src="/video/odo-dev-podman-demo.webm" type="video/webm"/></video>
</a>

## Wrapping Up

`odo` is now able to work with Podman to accelerate local development in containers, without requiring you to have access to any Kubernetes cluster.

Note that our support for Podman is still experimental, but we are working on improving the feature parity (as much as possible) with the cluster mode.

As such, [any feedback](https://github.com/redhat-developer/odo/wiki/Community:-Getting-involved) is highly appreciated.
