---
title: odo add binding
---

## Description
The `odo add binding` command can add a link between an operator-backed service and a component. odo uses the Service Binding Operator to create this link. Running this command will make the necessary changes to the Devfile, and once pushed to the cluster, it creates an instance of the `ServiceBinding` resource.

Currently, it only allows connecting to the Operator-backed services which support binding via the Service Binding Operator.
To know about the Operators supported by the Service Binding Operator, read its [README](https://github.com/redhat-developer/service-binding-operator#known-bindable-operators).

## Running the Command

### Pre-requisites
* A directory containing a Devfile; if you don't have one, see [odo init](init.md) on obtaining a devfile.
* A cluster with the Service Binding Operator installed, along with the Operator whose service you want to bind to

### Interactive Mode
In the interactive mode, you will be guided to choose:
* a service from the list of bindable service instances as supported by the Service Binding Operator,
* option to bind the service as a file (see [Understanding Bind as Files](#understanding-bind-as-files) for more information on this),
* a name for the binding.

```shell
# Add binding between a service, and the component present in the working directory in the interactive mode
odo add binding
```

### Non-interactive mode
In the non-interactive mode, you will have to specify the following required information through the command-line:
* `--service` flag to specify the service you want to bind to,
* `--name` flag to specify a name for the binding (see [Understanding Bind as Files](#understanding-bind-as-files) for more information on this)
* `--bind-as-files` flag to specify if the service should be bound as a file; this flag is set to true by default.


```shell
# Add binding between a service named 'cluster-sample',
# and the component present in the working directory in the non-interactive mode
odo add binding --name mybinding --service myRedisService.Redis
```

#### Understanding Bind as Files
To connect your component with a service, you need to store some data (e.g. username, password, host address) on your component's container.
If the service is bound as files, this data will be written to a file and stored on the container, else it will be injected as Environment Variables inside the container.

Note that every piece of data is stored in its own individual file or environment variable.
For example, if your data includes a username and password, then 2 separate files, or 2 environment variables will be created to store them both.

#### Formats supported by the `--service` flag
The `--service` flag supports the following formats to specify the service name:
* `<name>`
* `<name>.<kind>`
* `<name>.<kind>.<apigroup>`
* `<name>/<kind>`
* `<name>/<kind>.<apigroup>`

The above formats are helpful when multiple services with the same name exist on the cluster.

### Using different formats
```shell
# Add binding between a service named 'cluster-sample', and the component present in the working directory
odo add binding --service cluster-sample --name restapi-cluster-sample

# Add binding between service named 'cluster-sample' of kind 'Cluster', and APIGroup 'postgresql.k8s.enterprisedb.io',
# and the component present in the working directory 
odo add binding --service cluster-sample/Cluster.postgresql.k8s.enterprisedb.io --name restapi-cluster-sample
odo add binding --service cluster-sample.Cluster.postgresql.k8s.enterprisedb.io --name restapi-cluster-sample

# Add binding between service named 'cluster-sample' of kind 'Cluster',
# and the component present in the working directory
odo add binding --service cluster-sample/Cluster --name restapi-cluster-sample
odo add binding --service cluster-sample.Cluster --name restapi-cluster-sample
```