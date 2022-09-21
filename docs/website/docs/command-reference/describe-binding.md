---
title: odo describe binding
---

`odo describe binding` command is useful for getting information about service bindings.

This command supports the service bindings added with the command `odo add binding` and bindings added manually to the Devfile, using a `ServiceBinding` resource from one of these apiVersion:
- `binding.operators.coreos.com/v1alpha1`
- `servicebinding.io/v1alpha3`

## Running the Command

There are 2 ways to describe a service binding:
- [Describe with access to Devfile](#describe-with-access-to-devfile)
- [Describe without access to Devfile](#describe-without-access-to-devfile)

### Describe with access to Devfile

This command returns information extracted from the Devfile and, if possible, from the cluster.

The command lists the Kubernetes resources declared in the Devfile as a Kubernetes component,
with the kind `ServiceBinding` and one of these apiVersion:
- `binding.operators.coreos.com/v1alpha1`
- `servicebinding.io/v1alpha3`

For each of these resources, the following information is displayed:
- the resource name,
- the list of the services to which the component is bound using this service binding,
- for each service listed, the namespace containing the service, if any; otherwise, it means that the current namespace was used,
- if the variables are bound as files or as environment variables,
- the naming strategy used for binding names, if any,
- if the binding information is auto-detected.

```console
odo describe binding
```
When the service binding are not deployed yet to the cluster:

<details>
<summary>Example (not deployed)</summary>

```console
$ odo describe binding
ServiceBinding used by the current component:

Service Binding Name: my-nodejs-app-cluster-sample
Services:
 •  cluster-sample (Cluster.postgresql.k8s.enterprisedb.io) (namespace: shared-ns-1)
Bind as files: false
Detect binding resources: true
Naming strategy: uppercase
Available binding information: unknown

Service Binding Name: my-nodejs-app-redis-standalone
Services:
 •  redis-standalone (Redis.redis.redis.opstreelabs.in)
Bind as files: false
Detect binding resources: true
Available binding information: unknown

Binding information for one or more ServiceBinding is not available because they don't exist on the cluster yet.
Start "odo dev" first to see binding information.
```
</details>


When the resources have been deployed to the cluster, the command also extracts information from the status of the resources to display information about the variables that can be used from the component.


<details>
<summary>Example (after deploying on the cluster)</summary>

```console
$ odo describe binding 
ServiceBinding used by the current component:

Service Binding Name: my-nodejs-app-cluster-sample-2
Services:
 •  cluster-sample-2 (Cluster.postgresql.k8s.enterprisedb.io) (namespace: shared-ns-1)
Bind as files: false
Detect binding resources: true
Naming strategy: uppercase
Available binding information:
 •  CLUSTER_PASSWORD
 •  CLUSTER_PROVIDER
 •  CLUSTER_TLS.CRT
 •  CLUSTER_TLS.KEY
 •  CLUSTER_USERNAME
 •  CLUSTER_CA.KEY
 •  CLUSTER_CLUSTERIP
 •  CLUSTER_HOST
 •  CLUSTER_PGPASS
 •  CLUSTER_TYPE
 •  CLUSTER_CA.CRT
 •  CLUSTER_DATABASE

Service Binding Name: my-nodejs-app-redis-standalone
Services:
 •  redis-standalone (Redis.redis.redis.opstreelabs.in)
Bind as files: false
Detect binding resources: true
Available binding information:
 •  REDIS_CLUSTERIP
 •  REDIS_HOST
 •  REDIS_PASSWORD
 •  REDIS_TYPE
```
</details>


### Describe without access to Devfile

```console
odo describe binding --name <component_name>
```

<details>
<summary>Example</summary>

```shell
$ odo describe binding --name my-nodejs-app-redis-standalone
Service Binding Name: my-nodejs-app-redis-standalone
Services:
 •  redis-standalone (Redis.redis.redis.opstreelabs.in)
Bind as files: false
Detect binding resources: true
Available binding information:
 •  REDIS_CLUSTERIP
 •  REDIS_HOST
 •  REDIS_PASSWORD
 •  REDIS_TYPE
```
</details>

The command extracts information from the cluster.

The command searches for a resource in the current namespace with the given name, the kind `ServiceBinding` and one of these apiVersion:
- `binding.operators.coreos.com/v1alpha1`
- `servicebinding.io/v1alpha3`

If a resource is found, it displays information about the service binding and the variables that can be used from the component.
