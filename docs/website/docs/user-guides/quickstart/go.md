---
title: Developing with Go
sidebar_position: 4
---

## Step 0. Creating the initial source code (optional)

import InitialSourceCodeInfo from './docs-mdx/initial_source_code_description.mdx';

<InitialSourceCodeInfo/>

For Go, we will create our own application using the standard library:

1. Create the following `main.go` file:

```go
package main

import (
  "fmt"
  "net/http"
)

func main() {
  http.HandleFunc("/", HelloServer)
  http.ListenAndServe("0.0.0.0:8080", nil)
}

func HelloServer(w http.ResponseWriter, r *http.Request) {
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

## Step 1. Connect to your cluster and create a new namespace or project

import ConnectingToCluster from './docs-mdx/connecting_to_the_cluster_description.mdx';

<ConnectingToCluster/>

## Step 2. Initializing your application (`odo init`)

import InitSampleOutput from './docs-mdx/go/go_odo_init_output.mdx';
import InitDescription from './docs-mdx/odo_init_description.mdx';

<InitDescription framework="Go" initout=<InitSampleOutput/> />

## Step 3. Developing your application continuously (`odo dev`)

import DevSampleOutput from './docs-mdx/go/go_odo_dev_output.mdx';

import DevDescription from './docs-mdx/odo_dev_description.mdx';

<DevDescription framework="Go" devout=<DevSampleOutput/> />


_You can now follow the [advanced guide](../advanced/deploy/go.md) to deploy the application to production._
