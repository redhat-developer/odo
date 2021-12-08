---
title: Quickstart
sidebar_position: 3
---

In this guide, we will be using odo to set up a todo application based on Java Spring Boot framework for the backend/APIs, ReactJS for the frontend, and PostgreSQL database to store the todo items.

We will be performing following tasks using odo in this guide:
1. Create a project
2. Create an odo component for both the frontend and backend applications
3. Create an Operator backed service for PostgreSQL database
4. Link the backend component with the PostgreSQL service
5. Link the frontend component with the backend component

At the end of the guide, you will be able to list, add and delete todo items from the web browser.

## Prerequisites

* A [development Kubernetes](./cluster-setup/kubernetes.md) cluster with [Operator Lifecycle Manager](./cluster-setup/kubernetes#installing-the-operator-lifecycle-manager-olm) setup on it.
  * This guide is written for minikube users, hence you will notice the usage of `minikube ip` command to get the IP address of the Kubernetes cluster.
  * If you are using a Kubernetes cluster other than minikube, you will need to check with cluster administrator for the cluster IP to be used with `--host` flag.
  * If you are using [Code Ready Containers (CRC)](https://github.com/code-ready/crc) or another form of OpenShift cluster, you can skip the part of `odo url create` because odo automatically creates URL for the component using [OpenShift Routes](https://docs.openshift.com/container-platform/latest/networking/routes/route-configuration.html). 
* Install the [Crunchy Postgres Operator](https://operatorhub.io/operator/postgresql) on the cluster. Assuming you have admin privileges on the development Kubernetes cluster, you can install it using below command:
  ```shell
  kubectl create -f https://operatorhub.io/install/postgresql.yaml
  ```   
* Have the odo binary [installed](./installation.md) on your system.

## Create a project

We will create a project named `quickstart` on the cluster to keep quickstart related activities separate from rest of the cluster:
```shell
odo project create quickstart
```

## Clone the code

Clone [this git repository](https://github.com/dharmit/odo-quickstart/) and `cd` into it:
```shell
git clone https://github.com/dharmit/odo-quickstart
cd odo-quickstart
```

## Create the backend component

First we create a component for the backend application which is a Java Spring Boot based REST API. It will help us list, insert and delete todos from the database. Execute below steps:

```shell
cd backend
odo create java-springboot backend
odo url create --port 8080 --host `minikube ip`.nip.io
odo push
```

The `minikube ip` command helps get the IP address of the minikube instance. It is required to create a URL accesible from the web browser of the host system on which minikube is running.

## Create the Postgres database

In the [prerequisites](#prerequisites) section, we installed Postgres Operator. Before being able to create a service using it, first ensure that the Operator is installed correctly. You should see the Postgres Operator like in below output. Note that you might see more Operators in the output if there are other Operators installed on your cluster: 
```shell
odo catalog list services
```

```shell
$ odo catalog list services
Services available through Operators
NAME                        CRDs
postgresoperator.v5.0.3     PostgresCluster
```

If you don't see the Postgres Operator here, it might be still installing. Take a look at what you see in the `PHASE` column in below output:
```shell
kubectl get csv
```

```shell
$ kubectl get csv                         
NAME                      DISPLAY                           VERSION   REPLACES                  PHASE
postgresoperator.v5.0.3   Crunchy Postgres for Kubernetes   5.0.3     postgresoperator.v5.0.2   Succeeded
```

If the `PHASE` is something other than `Succeeded`, you won't see it in `odo catalog list services` output, and you won't be able to create a working Operator backed service out of it either.

Now create the service using:


```sh
odo service create --from-file ../postgrescluster.yaml
```

Example output:
```sh
$ odo service create --from-file ../postgrescluster.yaml
Successfully added service to the configuration; do 'odo push' to create service on the cluster
````

The `postgrescluster.yaml` file in the repository contains configuration that should help bring up a Postgres database. Do a push to create the database on the cluster:
```shell
odo push
```

## Link the backend component and the database

Next, we need to link the backend component with the database. Let's get the information about the database service first:

```shell
odo service list
```
Example output:
```shell
$ odo service list
NAME                      MANAGED BY ODO     STATE      AGE
PostgresCluster/hippo     Yes (backend)      Pushed     3m42s
```

Now, let's link the backend component with the above service using:
```shell
odo link PostgresCluster/hippo
odo push
```
Now, get the URL (`odo url list`) for the backend component, append `api/v1/todos` to it and open it on your browser:
```shell
odo url list
```

Example output:
```shell
$ odo url list
Found the following URLs for component backend
NAME         STATE      URL                                       PORT     SECURE     KIND
8080-tcp     Pushed     http://8080-tcp.192.168.39.117.nip.io     8080     false      ingress
```
In this case, the URL to load in browser would be `http://8080-tcp.192.168.39.117.nip.io/api/v1/todos`. Note that the URL would be different in your case depending on what the minikube VM's IP is. When you load the URL in the browser, you should see an empty list:
```shell
[]
```

## Create the frontend component

Our frontend component is a React application that communicates with the backend component. Create the frontend component:

```sh
cd ../frontend
odo create nodejs frontend
odo url create --port 3000 --host `minikube ip`.nip.io
odo push
```

Open the URL for the component in the browser, but note that you won't be able to add, remove or list the todos yet because we haven't linked the frontend and the backend components:
```shell
odo url list
```

## Link the frontend and backend components

To link the frontend component to backend:

```shell
odo link backend
odo push
```

Now reload the URL of frontend component and try adding and removing some todo items. The list of items appears by default on the same page just below the input box that reads `Add a new task`.