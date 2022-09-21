---
title: Connecting to a Service
sidebar_position: 1
---

Building on top of the [Go quickstart guide](../quickstart/go.md), this guide will connect a service to the Go application.

## Prerequisites
1. [Install the Service Binding Operator via Operator Hub](https://operatorhub.io/operator/service-binding-operator). [Read why this is required.](../../command-reference/add-binding.md#description)

2. [Install Percona Server Mongodb Operator via Operator Hub](https://operatorhub.io/operator/percona-server-mongodb-operator).

    :::note
    This operator is namespace scoped, so it will only be accessible in the namespace where it has been installed.

    When you install the operator by following Operator Hub instructions, the operator will be installed in a new namespace called "my-percona-server-mongodb-operator".
    :::

3. Create a MongoDB service.
:::note
The service should be created in the namespace where Percona Server Mongodb Operator is installed, in this case, it will be "my-percona-server-mongodb-operator".
:::

import CreateMongodbService from './_create-mongodb-service.mdx';

<CreateMongodbService/>

## Step 1. Implement the code logic
:::note
If you are already running `odo dev` in the terminal, press "Ctrl+c" and exit.
:::

We will extend the current code logic to obtain the connection information (username, password, and host) from the environment and then use it to connect to the MongoDB service and ping it.

The environment variables will be exposed when we link our application with the MongoDB service.

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
We will be using the MongoDB client library. Update the go.mod dependency by running the following command: 
```shell
go get go.mongodb.org/mongo-driver/mongo
```

## Step 2. Run the application
Run the application on the cluster with `odo dev`.
```shell
odo dev
```
<details>
<summary>Sample output:</summary>

```shell
$ odo dev
  __
 /  \__     Developing using the my-go-app Devfile
 \__/  \    Namespace: odo-dev
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
</details>


### Check the connection

Once the application is running, query the URL and check the response.
```shell
curl 127.0.0.1:40001
```
<details>
<summary>Sample output:</summary>

```shell
$ curl 127.0.0.1:40001
failed to connect: error validating uri: username required if URI contains user info
```
</details>


You should have received an error in response which is expected as we have not yet exposed the connection information in our application.

## Step 3. Connect the application to the MongoDB service
Connect the application to the MongoDB service with `odo add binding`.

From a new terminal, run the following command that will add necessary data to `devfile.yaml`:
```shell
odo add binding \
  --service mongodb-instance/PerconaServerMongoDB \
  --service-namespace=my-percona-server-mongodb-operator \
  --name my-go-app-mongodb-instance \
  --bind-as-files=false
```

<details>
<summary>Sample output:</summary>

```shell
$ odo add binding --service mongodb-instance/PerconaServerMongoDB --service-namespace=my-percona-server-mongodb-operator --name my-go-app-mongodb-instance --bind-as-files=false
 ✓  Successfully added the binding to the devfile.
Run `odo dev` to create it on the cluster.
```
</details>


:::note
Our code logic relies on obtaining connection information from the environment variables in the system and `--binding-as-files=false` ensures that. Read [here](../../command-reference/add-binding.md#understanding-bind-as-files) to know more about this flag.
:::

Wait for `odo dev` to detect the new changes to `devfile.yaml`. 

<details>
<summary>Sample output:</summary>

```shell
$ odo dev
  __
 /  \__     Developing using the my-go-app Devfile
 \__/  \    Namespace: odo-dev
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
</details>

Once `odo dev` has completed its task, you can run `odo describe binding` from a new terminal to obtain all the MongoDB service related information that has been exposed to your application:
```shell
odo describe binding --name my-go-app-mongodb-instance
```
<details>
<summary>Sample output:</summary>

```shell
$ odo describe binding --name my-go-app-mongodb-instance
Service Binding Name: my-go-app-mongodb-instance
Services:
 •  mongodb-instance (PerconaServerMongoDB.psmdb.percona.com) (namespace: my-percona-server-mongodb-operator)
Bind as files: false
Detect binding resources: true
Available binding information:
 •  PERCONASERVERMONGODB_MONGODB-KEY
 •  PERCONASERVERMONGODB_MONGODB_CLUSTER_MONITOR_USER
 •  PERCONASERVERMONGODB_USERNAME
 •  PERCONASERVERMONGODB_CLUSTERIP
 •  PERCONASERVERMONGODB_MONGODB_CLUSTER_MONITOR_PASSWORD
 •  PERCONASERVERMONGODB_MONGODB_USER_ADMIN_USER
 •  PERCONASERVERMONGODB_PASSWORD
 •  PERCONASERVERMONGODB_HOST
 •  PERCONASERVERMONGODB_MONGODB_USER_ADMIN_PASSWORD
 •  PERCONASERVERMONGODB_PROVIDER
 •  PERCONASERVERMONGODB_TYPE
 •  PERCONASERVERMONGODB_MONGODB_BACKUP_PASSWORD
 •  PERCONASERVERMONGODB_MONGODB_BACKUP_USER
 •  PERCONASERVERMONGODB_MONGODB_CLUSTER_ADMIN_PASSWORD
 •  PERCONASERVERMONGODB_MONGODB_CLUSTER_ADMIN_USER

```
</details>

## Step 4. Check the connection again
Query the URL again for a successful connection: 
```shell
curl 127.0.0.1:40001
```

<details>
<summary>Sample output:</summary>

```shell
$ curl 127.0.0.1:40001
Successfully connected and pinged.
```
</details>


## Step 5. Exit and cleanup
* Press `Ctrl+c` to exit `odo dev`.

* Remove the binding from `devfile.yaml` with `odo remove binding`.
    ```shell
    odo remove binding --name my-go-app-mongodb-instance
    ```

* Delete the MongoDB instance that we created at the beginning.

import DeleteMongodbService from './_delete-mongodb-service.mdx';

<DeleteMongodbService/>
