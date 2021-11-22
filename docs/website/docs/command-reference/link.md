---
title: odo link
sidebar_position: 4
---

`odo link` command helps link an odo component to an Operator backed service or another odo component. It does this by using [Service Binding Operator](https://github.com/redhat-developer/service-binding-operator). At the time of writing this, odo makes use of the Service Binding library and not the Operator itself to achieve the desired functionality.

In this document we will cover various options to create link between a component & a service, and a component & another component. The steps in this document are going to be based on the [odo quickstart project](https://github.com/dharmit/odo-quickstart/) that we covered in [Quickstart guide](/docs/getting-started/quickstart). The outputs mentioned in this document are based on commands executed on [minikube cluster](/docs/getting-started/cluster-setup/kubernetes).

This document assumes that you know how to [create components](/docs/command-reference/create) and [services](/docs/command-reference/service). It also assumes that you have cloned the [odo quickstart project](https://github.com/dharmit/odo-quickstart/). Terminology used in this document:

- *quickstart project*: git clone of the odo quickstart project having below directory structure:
    ```shell
    $ tree -L 1
    .
    ├── backend
    ├── frontend
    ├── postgrescluster.yaml
    ├── quickstart.code-workspace
    └── README.md
    
    2 directories, 3 files
    ```
- *backend component*: `backend` directory in above tree structure
- *frontend component*: `frontend` directory in above tree structure
- *Postgres service*: Operator backed service created from *backend component* using the `odo service create --from-file ../postgrescluster.yaml` command.

## Various linking options

odo provides various options to link a component with an Operator backed service or another odo component. All these options (or flags) can be used irrespective of whether you are linking a component to a service or another component.

### Default behaviour

By default, `odo link` creates a directory named `kubernetes/` in your component directory and stores the information (YAML manifests) about services and links in it. When you do `odo push`, odo compares these manifests with the state of the things on the Kubernetes cluster and decides whether it needs to create, modify or destroy resources to match what is specified by the user.

### The `--inlined` flag

If you specified `--inlined` flag to the `odo link` command, odo will store the link information inline in the `devfile.yaml` in the component directory instead of creating a file under `kubernetes/` directory. The behaviour of `--inlined` flag is similar in both the `odo link` and `odo service create` commands. This flag is helpful if you would like everything to be stored in a single `devfile.yaml`. You will have to remember to use `--inlined` flag with each `odo link` and `odo service create` commands that you execute for the component.

### The `--map` flag

At times, you might want to add more binding information to the component than what is available by default. For example, if you are linking the component with a service and would like to bind some information from the service's spec (short for specification), you could use the `--map` flag. Note that odo doesn't do any validation against the spec of the service/component being linked. Using this flag is recommended only if you are comfortable with reading the Kubernetes YAML manifests.

## Examples

### Default `odo link`

We will link the backend component with the Postgres service using default `odo link` command. For the backend component, make sure that your component and service are pushed to the cluster:

```shell
$ odo list
APP     NAME        PROJECT       TYPE       STATE      MANAGED BY ODO
app     backend     myproject     spring     Pushed     Yes


$ odo service list
NAME                      MANAGED BY ODO     STATE      AGE
PostgresCluster/hippo     Yes (backend)      Pushed     59m41s

```

Now, run `odo link` to link the backend component with the Postgres service:
```shell
odo link PostgresCluster/hippo
```
Example output:
```shell
$ odo link PostgresCluster/hippo
 ✓  Successfully created link between component "backend" and service "PostgresCluster/hippo"

To apply the link, please use `odo push`

$ odo push
```
And then run `odo push` for the link to actually get created on the Kubernetes cluster.

Upon successful `odo push`, you can notice a few things:
1. When you open the URL for the application deployed by backend component, it shows you a list of todo items in the database. For example, for below `odo url list` output, we will append the path where todos are listed:
  ```shell
  $ odo url list
  Found the following URLs for component backend
  NAME         STATE      URL                                       PORT     SECURE     KIND
  8080-tcp     Pushed     http://8080-tcp.192.168.39.112.nip.io     8080     false      ingress
  
  ```
  The correct path for such URL would be - http://8080-tcp.192.168.39.112.nip.io/api/v1/todos. Note that exact URL would be different for your setup. Also note that there are no todos in the database unless you add some, so the URL might just show an empty JSON object.
2. You can see binding information related to Postgres service injected into the backend component. This binding information is injected, by default, as environment variables, so you can check it out using:
  ```shell
  # below command filters environment variables with either HIPPO or POSTGRES in their name
  $ odo exec -- env | grep -e "HIPPO\|POSTGRES" 
  ```
  Few of these variables are used in the backend component's `src/main/resources/application.properties` file so that the Java Springboot application can connect to the Postgres database service.
3. odo has created a directory called `kubernetes/` in your backend component's directory which contains below files. 
  ```shell
  $ ls kubernetes 
  odo-service-backend-postgrescluster-hippo.yaml  odo-service-hippo.yaml
  ```
  This files contains the information (YAML manifests) about two things:
    1. `odo-service-hippo.yaml` - the Postgres service we created using `odo service create --from-file ../postgrescluster.yaml` command.
    2. `odo-service-backend-postgrescluster-hippo.yaml` - the link we created using `odo link` command.
  
### `odo link` with `--inlined`

Using `--inlined` flag with `odo link` command does the exact same thing to our application (that is, injects binding information) as an `odo link` command without the flag does. However, the subtle difference is that in above case we saw two manifest files under `kubernetes/` directory — one for the Postgres service and other for the link between the backend component and this service — but when we pass `--inlined` flag, odo does not create a file under `kubernetes/` directory to store the YAML manifest, but stores it inline in the `devfile.yaml` file.

To see this, let's unlink our component from the Postgres service first:

```shell
odo unlink PostgresCluster/hippo
```
Example output:
```shell
$ odo unlink PostgresCluster/hippo
 ✓  Successfully unlinked component "backend" from service "PostgresCluster/hippo"

To apply the changes, please use `odo push`
```
To unlink them on the cluster, run `odo push`. Now if you take a look at the `kubernetes/` directory, you'll see only one file in it:
```shell
$ ls kubernetes 
odo-service-hippo.yaml
```
Next, let's use the `--inlined` flag to create a link:
```shell
odo link PostgresCluster/hippo --inlined
```
Example output:
```shell
$ odo link PostgresCluster/hippo --inlined
 ✓  Successfully created link between component "backend" and service "PostgresCluster/hippo"

To apply the link, please use `odo push`
```
Just like the time without `--inlined` flag, you need to do `odo push` for the link to get created on the cluster. But where did odo store the configuration/manifest required to create this link? odo stores this in `devfile.yaml`. You can see an entry like below in this file:
```yaml
 kubernetes:
    inlined: |
      apiVersion: binding.operators.coreos.com/v1alpha1
      kind: ServiceBinding
      metadata:
        creationTimestamp: null
        name: backend-postgrescluster-hippo
      spec:
        application:
          group: apps
          name: backend-app
          resource: deployments
          version: v1
        bindAsFiles: false
        detectBindingResources: true
        services:
        - group: postgres-operator.crunchydata.com
          id: hippo
          kind: PostgresCluster
          name: hippo
          version: v1beta1
      status:
        secret: ""
  name: backend-postgrescluster-hippo
```
Now if you were to do `odo unlink PostgresCluster/hippo`, odo would first remove the link information from the `devfile.yaml` and then a subsequent `odo push` would delete the link from the cluster.

## Custom bindings

`odo link` accepts the flag `--map` which can inject custom binding information into the component. Such binding information will be fetched from the manifest of the resource we are linking to our component. For example, speaking in context of the backend component and Postgres service, we can inject information from the Postgres service's manifest ([`postgrescluster.yaml` file](https://github.com/dharmit/odo-quickstart/blob/main/postgrescluster.yaml)) into the backend component.

Considering the name of your `PostgresCluster` service is `hippo` (check the output of `odo service list` if your PostgresCluster service is named differently), if we wanted to inject the value of `postgresVersion` from that YAML definition into our backend component:
```shell
odo link PostgresCluster/hippo --map pgVersion='{{ .hippo.spec.postgresVersion }}'
```
Note that, if the name of your Postgres service is different from `hippo`, you will have to specify that in the above command in place `.hippo`. For example, if your `PostgresCluster` service is named as `database`, you would change the link command to as shown below:

```shell
$ odo service list
NAME                      MANAGED BY ODO     STATE      AGE
PostgresCluster/database     Yes (backend)      Pushed     2h5m43s

$ odo link PostgresCluster/hippo --map pgVersion='{{ .database.spec.postgresVersion }}'
```

After a link operation, do `odo push` as usual. Upon successful completion of push operation, you can run below command from your backend component directory to validate if custom mapping got injected properly:

```shell
odo exec -- env | grep pgVersion
```
Example output:
```shell
$ odo exec -- env | grep pgVersion
pgVersion=13
```

### To inline or not?

You can stick to the default behaviour wherein `odo link` will generate a manifest file for the link under `kubernetes/` directory, or you could use `--inlined` flag if you prefer to store everything in a single `devfile.yaml` file. It doesn't matter what you use for this functionality of adding custom mappings.