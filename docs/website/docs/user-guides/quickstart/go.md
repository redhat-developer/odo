---
title: Developing with Go
sidebar_position: 4
---

## Step 0. Creating the initial source code (optional)

import InitialSourceCodeInfo from './_initial_source_code.mdx';

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

Your source code has now been generated and created in the directory.

## Step 1. Connect to your cluster and create a new namespace or project

import ConnectingToCluster from './_connecting_to_cluster.mdx';

<ConnectingToCluster/>

## Step 2. Creating your application (`odo init`)

import CreatingApp from './_creating_app.mdx';

<CreatingApp name="go" port="8080" language="go" framework="Go"/>

## Step 3. Developing your application continuously (`odo dev`)

import RunningCommand from './_running_command.mdx';

<RunningCommand name="go" port="8080" language="go" framework="Go"/>