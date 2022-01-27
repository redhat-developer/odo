---
title: Hands-on odo guide
sidebar_position: 5
---

# Hands-on Guide

In this guide, we will be using odo to create a to-do list application, with the following:
* ReactJS for the frontend
* Java Spring Boot for the backend
* PostgreSQL to store all persistent data

At the end of the guide, you will be able to list, add and delete to-do items from the web browser.

## Prerequisites

* Have the odo binary [installed](./installation.md).
* A [Kubernetes cluster](/docs/getting-started/cluster-setup/kubernetes) set up with a [ingress controller](/docs/getting-started/cluster-setup/kubernetes#installing-an-ingress-controller), [operator lifecycle manager](/docs/getting-started/cluster-setup/kubernetes#installing-the-operator-lifecycle-manager-olm) and (optional) [service binding operator](/docs/getting-started/cluster-setup/kubernetes#installing-the-service-binding-operator).
* Or a [OpenShift cluster](/docs/getting-started/cluster-setup/openshift) set up with the (optional) [service binding operator](/docs/getting-started/cluster-setup/openshift#installing-the-service-binding-operator)

## Clone the quickstart guide

Clone the [quickstart](https://operatorhub.io/) repo from GitHub:
```shell
git clone https://github.com/odo-devfiles/odo-quickstart
cd odo-quickstart
```

## Create a project

We will create a project named `quickstart` on the cluster to keep all quickstart-related activities separate from rest of the cluster:
```shell
odo project create quickstart
```

## Create the frontend Node.JS component

Our frontend component is a React application that communicates with the backend component. 

We will use the catalog command to list all available components and find `nodejs`:
```shell
odo catalog list components
```

Example output of `odo catalog list components`:
```shell
Odo Devfile Components:
NAME                             DESCRIPTION                                                         REGISTRY
nodejs                           Stack with Node.js 14                                               DefaultDevfileRegistry
nodejs-angular                   Stack with Angular 12                                               DefaultDevfileRegistry
nodejs-nextjs                    Stack with Next.js 11                                               DefaultDevfileRegistry
nodejs-nuxtjs                    Stack with Nuxt.js 2                                                DefaultDevfileRegistry
...
```

Pick `nodejs` to create the frontend component:
```shell
cd frontend
odo create nodejs frontend
```

Create a URL in order to access the component in the browser:
```shell
odo url create --port 3000 --host <CLUSTER-HOSTNAME>
```

**Minikube users:** Use `minikube ip` to find out the hostname and then use `<MINIKUBE-HOSTNAME>.nip.io`  for `--host`.


Push the component to the cluster:
```shell
odo push
```

The URL will be listed in the `odo push` output, or can be found in `odo url list`.

Browse the site and try it out! Note that you will not be able to add, remove or list the to-dos yet, as we have not linked the frontend and the backend components yet.

## Create the backend Java component

The backend application is a Java Spring Boot based REST API which will list, insert and delete to-dos from the database.

Find `java-springboot` in the catalog:
```shell
odo catalog list components
```

Example output of `odo catalog list components`:
```shell
Odo Devfile Components:
NAME                             DESCRIPTION                                                         REGISTRY
java-quarkus                     Quarkus with Java                                                   DefaultDevfileRegistry
java-springboot                  Spring Boot® using Java                                             DefaultDevfileRegistry
java-vertx                       Upstream Vert.x using Java                                          DefaultDevfileRegistry
...
```

Let's create the component below:
```shell
cd ../backend
odo create java-springboot backend
odo url create --port 8080 --host <CLUSTER-HOSTNAME>.nip.io
odo push
```

Note, you will not be able to access `http://<YOUR-URL>/api/v1/todos` yet until we link the backend component to the database service.

## Create the Postgres service

Use `odo catalog list services` to list all available operators.

By default, [Operator Lifecycle Manager (OLM)](/docs/getting-started/cluster-setup/kubernetes#installing-the-operator-lifecycle-manager-olm) includes no Operators and they must be installed via [Operator Hub](https://operatorhub.io/)

Install the [Postgres Operator](https://operatorhub.io/operator/postgresql) on the cluster:
```shell
kubectl create -f https://operatorhub.io/install/postgresql.yaml
```

Find `postgresql` in the catalog:
```shell
odo catalog list services
```

Example output of `odo catalog list services`:
```shell
Services available through Operators
NAME                        CRDs
postgresoperator.v5.0.3     PostgresCluster
```

If you don't see the PostgreSQL Operator listed yet, it may still be installing. Check out our [Operator troubleshooting guide](/docs/getting-started/cluster-setup/kubernetes#checking-to-see-if-an-operator-has-been-installed) for more information.

[//]: # (This needs to fixed in the future and a parameter-based command added rather than a .yaml file)
[//]: # (Right now this is blocked on: https://github.com/redhat-developer/odo/issues/5215)
Create the service usng the provided `postgrescluster.yaml` file from [CrunchyData's Postgres guide](https://access.crunchydata.com/documentation/postgres-operator/5.0.0/tutorial/create-cluster/):
```sh
odo service create --from-file ../postgrescluster.yaml
````

The service from `postgrescluster.yaml` should now be added to your `devfile.yaml`, do a push to create the database on the cluster:
```shell
odo push
```

## Link the backend component and the service

Now we will link the the backend component (Java API) to the service (Postgres).

First, see if the service has been deployed:

```shell
odo service list
NAME                      MANAGED BY ODO     STATE      AGE
PostgresCluster/hippo     Yes (backend)      Pushed     3m42s
```

Link the backend component with the above service:
```shell
odo link PostgresCluster/hippo
```

Push the changes and `odo` will link the service to the component:
```shell
odo push
```

Now your service is linked to the backend component!

## Link the frontend and backend components


For our last step, we will now link the backend Java component (which also uses the Postgres service) and the frontend Node.JS component.

This will allow both to communicate with each other in order to store persistent data.


Change to the `frontend` component directory and link it to the backend:
```shell
cd ../frontend
odo link backend
```

Push the changes:
```shell
odo push
```

We're done! Now it's time to test your new multi-component and service application.

## Testing your application

### Frontend Node.JS component

Find out what URL is being used by the frontend:
```shell
odo url list
Found the following URLs for component frontend
NAME          STATE      URL                           PORT     SECURE     KIND
http-3000     Pushed     http://<URL-OUTPUT>           3000     false      ingress
```

Visit the link and type in some to-dos!


### Backend Java component

Let's see if each to-do is being stored in the backend api and database.

Find out what URL is being used by the backend:
```shell
odo url list
Found the following URLs for component backend
NAME         STATE      URL                                       PORT     SECURE     KIND
8080-tcp     Pushed     http://<URL-OUTPUT>                       8080     false      ingress
```

When you `curl` or view the URL on your browser, you'll now see the list of your to-dos:

```yaml
curl http://<URL-OUTPUT>/api/v1/todos
[{"id":1,"description":"hello"},{"id":2,"description":"world"}]
```

## Further reading

Want to learn what else `odo` can do? Check out the [Tutorials](/docs/intro) on the sidebar.
