---
title: Binding an external service with odo v3
author: Philippe Martin
author_url: https://github.com/feloy
author_image_url: https://github.com/feloy.png
tags: ["binding"]
slug: binding-external-service-with-odo-v3
---

How to bind an external service using odo v3

<!--truncate-->

When developers are working on a micro-service that needs to access a database or another service, 
they may want to provide to their application the address and the necessary credentials to access
this service as simply as possible.

In this article, we are saying that we are *binding* the service to the application.

Using the Service Binding Operator and creating some Kubernetes resources for
each service you want to be bindable, you can make the life easier for developers.

## Creating a Service resource to redirect to the external service

To expose an external service from inside a Kubernetes cluster, you can create a *Headless* Service,
and manually create the Enpoints to access this external service.

Here is an example, to connect to an external Redis service on IP 192.168.1.10 and port 6379:

```
kind: Service
apiVersion: v1
metadata:
  name: redis
  namespace: external-services
spec:
  type: ClusterIP
  ports:
  - port: 6379
    targetPort: 6379

---

kind: Endpoints
apiVersion: v1
metadata:
  name: redis
  namespace: external-services
subsets:
- addresses:
  - ip: 192.168.1.10
  ports:
  - port: 6379
```

Note that we have created these resources in a `external-services` namespace, which is a dedicated namespace
to store external services information, accessible by all developers.

You can find more information about creating Service resources to access external services [here](https://docs.openshift.com/dedicated/3/dev_guide/integrating_external_services.html) or [here](https://www.youtube.com/watch?v=fvpq4jqtuZ8).

## Storing the credentials into a Secret resource

The Redis instance is protected by a password, and you may want to store this password into a Secret resource,
so it can be used by applications.

The developers may want to *mount* this Secret into their application's Pod, but Secrets are mountable only
from Pods in the same namespace, and you would like to share these credentials with all the developers 
of the team, without creating several instances of this Secret (one in each developer's namespace), but only one
in the `external-services` namespace.

Here is, as an example, the secret to store the Redis password.

```
kind: Secret
apiVersion: v1
metadata:
  name: redis-credentials
  namespace: external-services
stringData:
  password: MyEasyPassword
```

## Adding SBO Annotations to the Service resource

To be able to *mount* the values of this secret from any namespace, you can use the *Service Binding Operator* (SBO for short), so each developer can define a ServiceBinding resource
between the service and its application, and get the values of the secret (and other values) mounted into its application's Pod.

You can find information about the Service Binding Operator [here](/docs/getting-started/cluster-setup/kubernetes#optional-installing-the-service-binding-operator).

A ServiceBinding defines a binding between an *Application* and a *Service*. The credentials injected into the application
can be defined in diffent ways:
- if the service is an Operator-backed service running on the cluster, the details of the injected credentials can be set
as annotations of the CRD associated with the Operator-backed service,
- in any case, the details of the injected credentials can be set in the resource itself (not the CRD). 
- in any case, the details of the injected credentials can be set in the ServiceBinding resource itself.

In this article, we are not using an Operator-backed service, but an external service referenced by a Service resource.
As the Service resource is a native Kubernetes resource, we cannot add annotations to its CRD, so we will add annotations to
the Service resource itself.

You can modifiy the definition of the Service, by adding the following annotations:

```
kind: Service
apiVersion: v1
metadata:
  name: redis
  annotations:
    service.binding/host: path={.metadata.name}.{.metadata.namespace}.svc.cluster.local
    service.binding: path={.metadata.name}-credentials,objectType=Secret
spec:
  type: ClusterIP
  ports:
  - port: 6379
    targetPort: 6379
```

In this snippet, the first annotation

```
service.binding/host: path={.metadata.name}.{.metadata.namespace}.svc.cluster.local
```

indicates to the SBO to inject a `host` variable into the application, with a value computed based on the
`metadata.name` and `metadata.namespace` of the Service resource. In this example, the value `redis.external-services.svc.cluster.local`
will be given to the `host` variable.


The second annotation

```
service.binding: path={.metadata.name}-credentials,objectType=Secret
```

indicates to the SBO to inject the values defined in the Secret, whose name is the name of the Service resource
followed by `-credentials` (`redis-credentials` in our example), into the application. In this example, the variable `password` 
with a value `MyEasyPassword` will be injected into the applications's Pod.

## Adding a ServiceBinding to the Devfile

To define a ServiceBinding, we need information (group, version, kind, name and namespace) about the Application and the Service.

In our example, the service is a Kubernetes Service (group "", version "v1" and kind "Service") named `redis`
in the `external-services` namespace, 

The application will be the Deployment resource (group "apps", version "v1", kind "Deployment") created by odo when you run `odo dev`.
By convention, the Deployment name will be the name of the Devfile (in the `.metadata.name` field) followed by `-app` (`my-nodejs-app-app` in our example).
You don't have to specify the namespace, as the Deployment will be in the same namespace as the ServiceBinding.

The option `bindAsFiles` indicates to the SBO to create files into the Pod's container, each file having the name 
of a credential, and containing the value of the credential.

```
apiVersion: binding.operators.coreos.com/v1alpha1
kind: ServiceBinding
metadata:
  name: binding-to-redis
spec:
  application:
    group: apps
    version: v1
    kind: Deployment
    name: my-nodejs-app-app
  services:
  - group: ""
    version: v1
    kind: Service
    name: redis
    namespace: external-services
  bindAsFiles: true
```

You can create a file `kubernetes/redis.yaml` in your directory containing this snippet,
and add a Kubernetes component into your Devfile referring to this YAML file:

```
metadata:
  name: my-nodejs-app
[...]
components:
[...]
- name: binding-to-redis
  kubernetes:
    uri: kubernetes/redis.yaml
```

By adding this Kubernetes component to your Devfile, when you run `odo dev`, the ServiceBinding resource defined
in the `kubernetes/redis.yaml` file will be created into the cluster, and the Service Binding Operator will inject
into the application's Pod the `host` and `password` necessary to connect to the Redis external service.

## Using the variables into the application's code

The Devfile is now ready, and the developer can start accessing the external service from the code. 

The first step to know how the credentials are exposed into the application's container is to start the `odo dev` 
command and to execute the `odo describe binding` command.

Running `odo dev`, you can see that the ServiceBinding resource is deployed to the cluster.

```
$ odo dev
[...]
↪ Deploying to the cluster in developer mode
 ✓  Creating kind ServiceBinding [60ms]
 ✓  Waiting for Kubernetes resources [10s]
 ✓  Syncing files into the container [740ms]
 ✓  Building your application in container on cluster [4s]
 ✓  Executing the application [1s]
[...]
```

From another terminal, running `odo describe binding` shows you the status of the ServiceBinding:

```
$ odo describe binding
ServiceBinding used by the current component:

Service Binding Name: binding-to-redis
Services:
 •  redis (Service.)
Bind as files: true
Detect binding resources: false
Available binding information:
 •  ${SERVICE_BINDING_ROOT}/binding-to-redis/host
 •  ${SERVICE_BINDING_ROOT}/binding-to-redis/password
```

This output shows that two files `host` and `password` are present in the application's container, at the mentioned paths.

You can leverage a library to help you access
these files, for example the [Python pyservicebinding library](https://github.com/baijum/pyservicebinding) or the [Go servicebinding library](https://github.com/baijum/servicebinding).


## Troubleshooting

If the output of `odo describe binding` shows an unknown status:

```
Available binding information: unknown
```

- first check if `odo dev` is still running. `odo` is not able to know
the bound credentials if the ServiceBinding resource is not deployed by `odo dev`.
- if `odo dev` is running, you can check that the ServiceBinding resource is deployed to the cluster, and if its status is `ApplicationsBound`, with the command:
  ```
  kubectl get servicebindings.binding.operators.coreos.com
  ```
- if the status of the ServiceBinding resource displayed in the list is not `ApplicationsBound`, you can get an error message with the command:
  ```
  kubectl describe servicebindings.binding.operators.coreos.com <service-binding-name>
  ```

