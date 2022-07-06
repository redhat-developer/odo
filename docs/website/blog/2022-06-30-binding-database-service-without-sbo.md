---
title: Binding to a database service without the Service Binding Operator
author: Parthvi Vala
author_url: https://github.com/valaparthvi
author_image_url: https://github.com/valaparthvi.png
tags: ["binding"]
slug: binding-database-service-without-sbo
---

How to bind your application to a database service without the Service Binding Operator.

<!--truncate-->

There are a few ways of binding your application to a database service with the help of odo. The recommended way is with the help of Service Binding Operator(SBO), but it is also possible to bind without it, and this blog will show you how.


## Architecture
We have a simple CRUD application built in Go that can create/list/update/delete a place. This application requires connecting to a MongoDB database in order to function correctly, which will be deployed as a microservice on the cluster.

## Prerequisites:
This blog assumes:
- [odo v3.0.0-beta1](https://github.com/redhat-developer/odo/releases/tag/v3.0.0-beta1)
- you have access to a Kubernetes or OpenShift cluster to be able to install the MongoDB operator.
- you have _Helm_ installed on your system. See https://helm.sh/docs/intro/install/ for installation instructions.

## (Optional) Setting up the namespace
0. We will create a new namespace to deploy our application in, with the help of odo.
```sh
odo create namespace restapi-mongodb
```

## Setting up the MongoDB microservice
We are going to use the Bitnami's helm charts for creating our MongoDB database.

1. Add the Bitnami's Helm charts repository and make your Helm client up to date with it:
```sh
helm repo add bitnami https://charts.bitnami.com/bitnami && helm repo update
```

2. Export the necessary environment variables:
```sh
export MY_MONGODB_ROOT_USERNAME=root
export MY_MONGODB_ROOT_PASSWORD=my-super-secret-root-password
export MY_MONGODB_USERNAME=my-app-username
export MY_MONGODB_PASSWORD=my-app-super-secret-password
export MY_MONGODB_DATABASE=my-app
```

3. Create the MongoDB service.
```sh
helm install mongodb bitnami/mongodb \
  --set auth.rootPassword=$MY_MONGODB_ROOT_PASSWORD \
  --set auth.username=$MY_MONGODB_USERNAME \
  --set auth.password=$MY_MONGODB_PASSWORD \
  --set auth.database=$MY_MONGODB_DATABASE
```

<details>
<summary>Expected output:</summary>

```sh
$ helm install mongodb bitnami/mongodb \
  --set auth.rootPassword=$MY_MONGODB_ROOT_PASSWORD \
  --set auth.username=$MY_MONGODB_USERNAME \
  --set auth.password=$MY_MONGODB_PASSWORD \
  --set auth.database=$MY_MONGODB_DATABASE
NAME: mongodb
LAST DEPLOYED: Tue Jul  5 15:53:40 2022
NAMESPACE: restapi-mongodb
STATUS: deployed
REVISION: 1
TEST SUITE: None
NOTES:
CHART NAME: mongodb
CHART VERSION: 12.1.24
APP VERSION: 5.0.9

** Please be patient while the chart is being deployed **

MongoDB&reg; can be accessed on the following DNS name(s) and ports from within your cluster:

    mongodb.restapi-mongodb.svc.cluster.local

To get the root password run:

    export MONGODB_ROOT_PASSWORD=$(kubectl get secret --namespace restapi-mongodb mongodb -o jsonpath="{.data.mongodb-root-password}" | base64 -d)

To get the password for "my-app-username" run:

    export MONGODB_PASSWORD=$(kubectl get secret --namespace restapi-mongodb mongodb -o jsonpath="{.data.mongodb-passwords}" | base64 -d | awk -F'
,' '{print $1}')

To connect to your database, create a MongoDB&reg; client container:

    kubectl run --namespace restapi-mongodb mongodb-client --rm --tty -i --restart='Never' --env="MONGODB_ROOT_PASSWORD=$MONGODB_ROOT_PASSWORD" --
image docker.io/bitnami/mongodb:5.0.9-debian-11-r1 --command -- bash

Then, run the following command:
    mongosh admin --host "mongodb" --authenticationDatabase admin -u root -p $MONGODB_ROOT_PASSWORD

To connect to your database from outside the cluster execute the following commands:

    kubectl port-forward --namespace restapi-mongodb svc/mongodb 27017:27017 &
    mongosh --host 127.0.0.1 --authenticationDatabase admin -p $MONGODB_ROOT_PASSWORD

```
</details>

Notice the resources(sevice, deployment, and secrets) that are deployed.

Wait for the pods to come up, this might take a few minutes:
```sh
$ kubectl get pods
NAME                       READY   STATUS    RESTARTS   AGE
mongodb-85fff797f6-fnwvl   1/1     Running   0          63s
```

## Setting up the application
4. Clone the repository, and cd into it.
```sh
git clone https://github.com/valaparthvi/restapi-mongodb-odo.git && cd restapi-mongodb-odo
```


## Download the devfile.yaml
5. Run `odo init` to fetch the necessary devfile.
```sh
odo init --devfile go --name places
```


## Adding the connection information to devfile.yaml
There are 3 changes that we will need to make to our devfile:

6.1 Change the `schemaVersion` of devfile to 2.2.0.
```yaml
schemaVersion: 2.2.0
```
Please note that this change is only necessary because we are using [devfile variable substitution](../../command-reference/dev/#substituting-variables).

6.2 Add a `variables` field in the devfile.
```yaml
variables:
  PASSWORD: password
  USERNAME: user
  HOST: host
```
6.3 Edit the 'runtime' container component in devfile to add information such as username, password, and host required to connect to the MongoDB service.
```yaml
components:
- container:
    ...
    ...
    env:
    - name: username
      value: "{{USERNAME}}"
    - name: password
      value: "{{PASSWORD}}"
    - name: host
      value: "{{HOST}}"
  name: runtime
```

The values for _username_, _password_, and _host_ will be passed to devfile.yaml when we run the `odo dev` command.


<details>
<summary>Your final devfile.yaml should look something like this:</summary>

```yaml
commands:
- exec:
    commandLine: GOCACHE=${PROJECT_SOURCE}/.cache go build main.go
    component: runtime
    group:
      isDefault: true
      kind: build
    hotReloadCapable: false
    workingDir: ${PROJECT_SOURCE}
  id: build
- exec:
    commandLine: ./main
    component: runtime
    group:
      isDefault: true
      kind: run
    hotReloadCapable: false
    workingDir: ${PROJECT_SOURCE}
  id: run
components:
- container:
    dedicatedPod: false
    endpoints:
    - name: port-3000-tcp
      protocol: tcp
      secure: false
      targetPort: 3000
    image: golang:latest
    memoryLimit: 1024Mi
    mountSources: true
    env:
    - name: username
      value: "{{USERNAME}}"
    - name: password
      value: "{{PASSWORD}}"
    - name: host
      value: "{{HOST}}"
  name: runtime
variables:
  PASSWORD: password
  USERNAME: user
  HOST: host
metadata:
  description: Stack with the latest Go version
  displayName: Go Runtime
  icon: https://raw.githubusercontent.com/devfile-samples/devfile-stack-icons/main/golang.svg
  language: go
  name: restapi
  projectType: go
  tags:
  - Go
  version: 1.0.0
schemaVersion: 2.2.0
starterProjects:
- git:
    checkoutFrom:
      revision: main
    remotes:
      origin: https://github.com/devfile-samples/devfile-stack-go.git
  name: go-starter
```
</details>


## Deploy the application
7. Run `odo dev` to deploy the application on the cluster.
```sh
odo dev --var PASSWORD=$MY_MONGODB_ROOT_PASSWORD --var USERNAME=$MY_MONGODB_ROOT_USERNAME  --var HOST="mongodb"
```

The value for _host_ is name of the service that belongs to our database application, in this case it is a service resource called "mongodb", you might have noticed it when we deployed the helm chart.

<details>
<summary>Expected output:</summary>

```sh
$ odo dev --var PASSWORD=$MY_MONGODB_ROOT_PASSWORD --var USERNAME=$MY_MONGODB_ROOT_USERNAME --var HOST="mongodb"
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
</details>


## Accessing the application
8. Run the following curl command to test the application:
```sh
curl 127.0.0.1:40001/api/places
```
This will return a _null_ response since the database is currently empty, but it also means that we have successfully connected to our database application.

9. Add some data to the database:
```sh
curl -sSL -XPOST -d '{"title": "Agra", "description": "Land of Tajmahal"}' 127.0.0.1:40001/api/places
```

10. Fetch the list of places again:
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
To conclude this blog, it is possible to connect your application with another microservice without the Service Binding Operator if you have the correct connection information. Using the Service Binding Operator with a [Bindable Operator](https://github.com/redhat-developer/service-binding-operator#known-bindable-operators) makes it easy for you to not care about finding the connection information and ease the binding.

### Related articles on binding:
* [Binding an external service with odo v3](./2022-06-14-binding-external-service.md)
* [odo add binding](../docs/command-reference/add-binding.md)