# Multiple Services Binding Example

This document describes step-by-step the actions to create the required
infrastructure to demonstrate multiple services binding built on existing
Postgres and ETCD examples.

As *Cluster Administrator*, the reader will install both the "PostgreSQL
Database" and the "ETCD" operators, as described below.

Once the cluster setup is finished, the reader will create a Postgres
database and a ETCD cluster, and bind services to a Node application as
a *Developer*.

## Cluster Configuration

### Create a New Project

Create a new project, in this example it is called `multiple-services-demo`.

### Install the Postgres Operator

Switch to the *Administrator* perspective.

Add an extra OperatorSource by pushing the "+" button on the top right corner
and pasting the following:

```yaml
---
apiVersion: operators.coreos.com/v1
kind: OperatorSource
metadata:
  name: db-operators
  namespace: openshift-marketplace
spec:
  type: appregistry
  endpoint: https://quay.io/cnr
  registryNamespace: pmacik
```

Go to "Operators > OperatorHub", search for "Postgres" and install "PostgreSQL
Database" provided by Red Hat.

Select "A specific namespace on the cluster" in "Installation Mode", select the
"multiple-services-demo" namespace in "Installed Namespace" and push "Subscribe".

### Install the ETCD Operator

Go to "Operators > OperatorHub", search for "etcd" and install "etcd" provided by
CNCF.

Select "A specific namespace on the cluster" in "Installation Mode", select the
"multiple-services-demo" namespace in "Installed Namespace" and push "Subscribe".

## Application Configuration

Switch to the *Developer* perspective.

Create the Postgres database `db-demo` by pushing the "+" button on the top right
corner and pasting the following:

```yaml
---
apiVersion: postgresql.baiju.dev/v1alpha1
kind: Database
metadata:
  name: db-demo
spec:
  image: docker.io/postgres
  imageName: postgres
  dbName: db-demo
```

Create the ETCD cluster `etcd-demo` by pushing the "+" button on the top right
corner and pasting the following:

```yaml
---
apiVersion: etcd.database.coreos.com/v1beta2
kind: EtcdCluster
metadata:
 name: etcd-demo
spec:
 size: 3
 version: "3.2.13"
```

Import the application by pushing the "+Add" button on the left side of the
screen, and then the "From Git" button. Fill the "Git Repo URL" with
`https://github.com/akashshinde/node-todo.git`; the repository will be
validated and the appropriate "Builder Image" and "Builder Image Version"
will be selected. Push the "Create" button to create the application.

Create the ServiceBindingRequest `node-todo-git` by pushing the "+" button
on the top right corner and pasting the following:

```yaml
---
apiVersion: apps.openshift.io/v1alpha1
kind: ServiceBindingRequest
metadata:
  name: node-todo-git
spec:
  applicationSelector:
    resourceRef: node-todo-git
    group: apps
    version: v1
    resource: deployments
  backingServiceSelectors:
  - group: postgresql.baiju.dev
    version: v1alpha1
    kind: Database
    resourceRef: db-demo
  - group: etcd.database.coreos.com
    version: v1beta2
    kind: EtcdCluster
    resourceRef: etcd-demo
  detectBindingResources: true
```

Once the binding is processed, the secret can be verified by executing
`kubectl get secrets node-todo-git -o yaml`:

```yaml
apiVersion: v1
data:
  DATABASE_CLUSTERIP: MTcyLjMwLjcyLjg5
  DATABASE_CONFIGMAP_DB_HOST: MTcyLjMwLjcyLjg5
  DATABASE_CONFIGMAP_DB_NAME: ZGItZGVtbw==
  DATABASE_CONFIGMAP_DB_PASSWORD: cGFzc3dvcmQ=
  DATABASE_CONFIGMAP_DB_PORT: NTQzMg==
  DATABASE_CONFIGMAP_DB_USERNAME: cG9zdGdyZXM=
  DATABASE_CONFIGMAP_PASSWORD: cGFzc3dvcmQ=
  DATABASE_CONFIGMAP_USER: cG9zdGdyZXM=
  DATABASE_DB_HOST: MTcyLjMwLjcyLjg5
  DATABASE_DB_NAME: ZGItZGVtbw==
  DATABASE_DB_PASSWORD: cGFzc3dvcmQ=
  DATABASE_DB_PORT: NTQzMg==
  DATABASE_DB_USERNAME: cG9zdGdyZXM=
  DATABASE_DBCONNECTIONIP: MTcyLjMwLjcyLjg5
  DATABASE_DBCONNECTIONPORT: NTQzMg==
  DATABASE_DBNAME: ZGItZGVtbw==
  DATABASE_SECRET_PASSWORD: cGFzc3dvcmQ=
  DATABASE_SECRET_USER: cG9zdGdyZXM=
  ETCDCLUSTER_CLUSTERIP: MTcyLjMwLjYyLjUy
kind: Secret
metadata:
  annotations:
    service-binding-operator.apps.openshift.io/binding-name: node-todo-git
    service-binding-operator.apps.openshift.io/binding-namespace: multiple-services-demo
  creationTimestamp: "2020-02-14T11:58:29Z"
  name: node-todo-git
  namespace: multiple-services-demo
  resourceVersion: "257567"
  selfLink: /api/v1/namespaces/multiple-services-demo/secrets/node-todo-git
  uid: 15aafcae-d334-49d8-be4c-2331f9c7cffe
type: Opaque
```
