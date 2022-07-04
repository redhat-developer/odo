---
title: Binding to a database service without SBO
author: Parthvi Vala
author_url: https://github.com/valaparthvi
author_image_url: https://github.com/valaparthvi.png
tags: ["binding"]
slug: binding-database-service-without-sbo
---

How to bind to your application to a database service without SBO.

<!--truncate-->

There are a few ways of binding your application to a database service with the help of odo. The recommended way is with the help of Service Binding Operator, but it is also possible to bind without it, and this blog will show you how.


## Architecture
We have a simple CRUD application built in Go that can create/list/update/delete a place. This application requires connecting to a MongoDB database in order to function correctly, which will be deployed as a microservice on the cluster.

## Prerequisites:
This blog assumes:
- you have admin access to a Kubernetes or OpenShift cluster to be able to install the MongoDB operator.
- you have _Helm_ installed on your system. See https://helm.sh/docs/intro/install/ for installation instructions.

## Setting up the application
1. Clone the repository, and cd into it.
```sh
git clone https://github.com/valaparthvi/restapi-mongodb-odo.git && cd restapi-mongodb-odo
```

## Setting up the namespace
2. We will create a new namespace to deploy our application in, with the help of odo.
```sh
odo create namespace restapi-mongodb
```


## Setting up the MongoDB microservice
We are going to use the Percona's operator for creating our MongoDB database. For the sake of simplicity, we will use Helm to deploy our MongoDB operator and the service.

3. Add the Percona’s Helm charts repository and make your Helm client up to date with it:
```sh
helm repo add percona https://percona.github.io/percona-helm-charts/ && helm repo update
```

4. Install Percona Operator for MongoDB:
```sh
helm install my-op percona/psmdb-operator
```
Wait for the pod to come up:
```sh
$ kubectl get pods
NAME                                    READY   STATUS    RESTARTS   AGE
my-op-psmdb-operator-69d88f479c-7hj5m   1/1     Running   0          3m34s
```

5. Install Percona server for MongoDB from our local `psmdb-db` Helm chart.
```sh
helm install my-db ./psmdb-db
```

```sh
$ helm install my-db ./psmdb-db

NAME: my-db
LAST DEPLOYED: Mon Jul  4 13:49:26 2022
NAMESPACE: restapi-mongodb
STATUS: deployed
REVISION: 1
TEST SUITE: None
NOTES:
To get a MongoDB prompt inside your new cluster you can run:

  ADMIN_USER=$(kubectl -n restapi-mongodb get secrets minimal-cluster -o jsonpath="{.data.MONGODB_USER_ADMIN_USER}" | base64 --decode)
  ADMIN_PASSWORD=$(kubectl -n restapi-mongodb get secrets minimal-cluster -o jsonpath="{.data.MONGODB_USER_ADMIN_PASSWORD}" | base64 --decode)

And then for replica set:
  $ kubectl run -i --rm --tty percona-client --image=percona/percona-server-mongodb:4.4 --restart=Never -- mongo "mongodb+srv://${ADMIN_USER}:${ADMIN_PASSWORD}@minimal-cluster-rs0.restapi-mongodb.svc.cluster.local/admin?replicaSet=rs0&ssl=false"

Or for sharding setup:
  $ kubectl run -i --rm --tty percona-client --image=percona/percona-server-mongodb:4.4 --restart=Never -- mongo "mongodb://${ADMIN_USER}:${ADMIN_PASSWORD}@minimal-cluster-mongos.restapi-mongodb.svc.cluster.local/admin?ssl=false"
```

Wait for the pods to come up, this might take a few minutes:
```sh
$ kubectl get pods
NAME                                    READY   STATUS    RESTARTS   AGE
minimal-cluster-cfg-0                   1/1     Running   0          3m6s
minimal-cluster-mongos-0                1/1     Running   0          3m5s
minimal-cluster-rs0-0                   1/1     Running   0          3m5s
my-op-psmdb-operator-69d88f479c-7hj5m   1/1     Running   0          3m34s
```

## Download the devfile.yaml
6. Run `odo init` to fetch the necessary devfile.
```sh
odo init --devfile go --name places
```

7. Edit the 'runtime' container component in devfile to add information such as username, password, and host required to connect to the MongoDB service.
```yaml
components:
- container:
    ...
    ...
    env:
    - name: username
      value: userAdmin
    - name: password
      value: userAdmin123456
    - name: host
      value: minimal-cluster-mongos
  name: runtime
```

The _username_, and _password_ values are hard-coded here as a part of the helm deployment, but their value can be obtained from the secret resource called "minimal-cluster" that was deployed via our local helm chart "psmdb-db".

The value for _host_ is name of the service that belongs to our database application, in this case it is a service resource called "minimal-cluster-mongos".
Optionally, you can run `kubectl get psmdb/minimal-cluster -ojsonpath='{.status.host}'` to obtain the host's value.

## Deploy the application
8. Run `odo dev` to deploy the application on the cluster.
```sh
$ odo dev
  __
 /  \__     Developing using the restapi Devfile
 \__/  \    Namespace: restapi-mongodb
 /  \__/    odo version: v3.0.0-alpha3
 \__/

↪ Deploying to the cluster in developer mode
 ✓  Waiting for Kubernetes resources [52s]
 ✓  Syncing files into the container [844ms]
 ✓  Building your application in container on cluster (command: build) [5s]
 •  Executing the application (command: run)  ...

Your application is now running on the cluster

 - Forwarding from 127.0.0.1:40001 -> 3000

Watching for changes in the current directory /home/pvala/restapi-mongodb-odo
Press Ctrl+c to exit `odo dev` and delete resources from the cluster
```

## Accessing the application
9. Run the following curl command to test the application:
```sh
curl 127.0.0.1:40001/api/places
```
This will return a _null_ response since the database is currently empty, but it also means that we have successfully connected to our database application.

10. Add some data to the database:
```sh
curl -sSL -XPOST -d '{"title": "Agra", "description": "Land of Tajmahal"}' 127.0.0.1:40001/api/places
```

11. Fetch the list of places again:
```sh
$ curl 127.0.0.1:40001/api/places

{"id":"62c2a0659fa147e382a4db31","title":"Agra","description":"Land of Tajmahal"}
```

### List of available API endpoints
- GET `/api/places` - List all places
- POST `/api/places` - Add a new place
- PUT `/api/places` - Update a place
- GET `/api/places/<id>` - Fetch place with id `<id>`
- DELETE `/api/places/<id>` - Delete place with id `<id>`

## Conclusion
To conclude this blog, it is possible to connect your application with another microservice without the Service Binding Operator if you have the correct connection information. Using SBO with a bindable operator makes it easy for you to not care about finding the connection information and ease the binding.