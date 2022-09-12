---
title: Connecting to a service
sidebar_position: 1
---

This tutorial will show you how you can connect your Go application to a mongodb service.

Building on top of the [Go quickstart guide](../quickstart/go.md), we will extend the application to check if it is connected to a mongodb service.

### Pre-requisites
1. Install Service Binding Operator on your cluster.  See [installation instruction](https://operatorhub.io/operator/service-binding-operator).

2. Install Percona Server Mongodb operator on your cluster. See [installation instruction](https://operatorhub.io/operator/percona-server-mongodb-operator).
:::note
The operator will be installed in a new namespace called "my-percona-server-mongodb-operator" and for the sake of simplicity, we will use this namespace for this guide. 
:::
3. Create a mongodb service.

import CreateMongodbService from './_create-mongodb-service.mdx';

<CreateMongodbService/>

### Implement the code logic
:::note
If you're already running `odo dev` in a terminal, exit it and start afresh.
:::

The new code is simple. We obtain the connection information (username, password, and host) from the environment, and use it to connect to the mongodb service and ping it. 

Replace the content of your `main.go` with the following content:
```go
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Connection URI
var (
	USERNAME = os.Getenv("PERCONASERVERMONGODB_MONGODB_USER_ADMIN_USER")
	PASSWORD = os.Getenv("PERCONASERVERMONGODB_MONGODB_USER_ADMIN_PASSWORD")
	HOST     = os.Getenv("PERCONASERVERMONGODB_HOST")
	uri      = fmt.Sprintf("mongodb://%s:%s@%s:27017/?maxPoolSize=20&w=majority", USERNAME, PASSWORD, HOST)
)

func main() {
	http.HandleFunc("/", HelloServer)
	http.ListenAndServe("0.0.0.0:8080", nil)
}

func HelloServer(w http.ResponseWriter, r *http.Request) {
	// Create a new client and connect to the server
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))

	if err != nil {
		fmt.Fprintf(w, "failed to connect: %s", err.Error())
		return
	}
	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			fmt.Fprintf(w, "failed to disconnect: %s", err.Error())
			return
		}
	}()

	// Ping the primary
	if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
		fmt.Fprintf(w, "unable to connect to the server: %s", err.Error())
		return
	}

	fmt.Fprintf(w, "Successfully connected and pinged.")
}
```
We will be using the mongodb client library, so let's update the go.mod with this dependency by running the following command: 
```shell
go get go.mongodb.org/mongo-driver/mongo
```

### Run the application on the cluster
Run this application on the cluster with `odo dev`.
```shell
odo dev
```
```shell
$ odo dev
  __
 /  \__     Developing using the my-go-app Devfile
 \__/  \    Namespace: my-percona-server-mongodb-operator
 /  \__/    odo version: v3.0.0-rc1
 \__/

↪ Deploying to the cluster in developer mode
 •  Waiting for Kubernetes resources  ...
 ⚠  Pod is Pending
 ✓  Pod is Running
 ✓  Syncing files into the container [152ms]
 ✓  Building your application in container on cluster (command: build) [15s]
 •  Executing the application (command: run)  ...
 -  Forwarding from 127.0.0.1:40001 -> 8080


Watching for changes in the current directory /tmp/go
Press Ctrl+c to exit `odo dev` and delete resources from the cluster

Pushing files...


File /tmp/go/.odo changed
 •  Waiting for Kubernetes resources  ...
 ✓  Syncing files into the container [1ms]

Watching for changes in the current directory /tmp/go
Press Ctrl+c to exit `odo dev` and delete resources from the cluster
```

### Check the connection

Once the application is running, try calling the URL and check the response.
```shell
curl 127.0.0.1:40001
```
```shell
$ curl 127.0.0.1:40001
failed to connect: error validating uri: username required if URI contains user info
```

This response is expected because we have not yet exposed the connection information to our cluster environment and hence have failed to connect to the mongodb service.

### Connect the application to the mongodb service
Let's now connect our application to the mongodb service with `odo add binding`.

From a new terminal, run the following command that will add necessary data to devfile.yaml:
```shell
odo add binding \
  --service mongodb-instance/PerconaServerMongoDB \
  --name my-go-app-mongodb-instance \
  --bind-as-files=false
```
```shell
$ odo add binding --service mongodb-instance/PerconaServerMongoDB --name my-go-app-mongodb-instance --bind-as-files=false
 ✓  Successfully added the binding to the devfile.
Run `odo dev` to create it on the cluster.
```

:::note
`--binding-as-files=false` because our code logic relies on obtaining environment variables from the system instead of reading data from files.
:::

Wait for `odo dev` to detect the new changes to the devfile.yaml. 
```shell
$ odo dev
  __                                                                                                                                              
 /  \__     Developing using the my-go-app Devfile                                                                                                
 \__/  \    Namespace: my-percona-server-mongodb-operator                                                                                         
 /  \__/    odo version: v3.0.0-rc1
 \__/

...
...
...
Pushing files...


File /tmp/go/devfile.yaml changed
 •  Waiting for Kubernetes resources  ...
 ✓  Creating kind ServiceBinding 
Error occurred on Push - watch command was unable to push component: some servicebindings are not injected

Updating Component...

 •  Waiting for Kubernetes resources  ...
Error occurred on Push - watch command was unable to push component: some servicebindings are not injected

 ⚠  Pod is Terminating
 •  Waiting for Kubernetes resources  ...
 ✗  Finished executing the application (command: run) [1m]
 ⚠  No pod exists
 ⚠  Pod is Pending
 ✓  Pod is Running
 ✓  Syncing files into the container [170ms]
 ✓  Building your application in container on cluster (command: build) [192ms]
 •  Executing the application (command: run)  ...
 -  Forwarding from 127.0.0.1:40001 -> 8080
                                                                                                                                                  

Watching for changes in the current directory /tmp/go
Press Ctrl+c to exit `odo dev` and delete resources from the cluster

Pushing files...


File /tmp/go/devfile.yaml changed

File /tmp/go/.odo/devstate.json changed
 •  Waiting for Kubernetes resources  ...
 ✓  Syncing files into the container [1ms]

Watching for changes in the current directory /tmp/go
Press Ctrl+c to exit `odo dev` and delete resources from the cluster


```

### Check the connection again
Once it is done, call the URL again. We should now be connected.
```shell
curl 127.0.0.1:40001
```
```shell
$ curl 127.0.0.1:40001
Successfully connected and pinged.
```

### Exit and cleanup
Press `Ctrl+c` to exit `odo dev`.

Delete the mongodb instance that we had created.

import DeleteMongodbService from './_delete-mongodb-service.mdx';

<DeleteMongodbService/>
