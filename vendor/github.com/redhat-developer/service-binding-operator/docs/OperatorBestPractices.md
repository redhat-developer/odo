# Operator Best Practices

## Introduction

The goals of the Service Binding Operator is to make it easier for
applications developers to bind applications with needed backing
services, without having to perform manual configuration of secrets,
configmaps, etc. and to assist operator providers in promoting and
expanding the adoption of their operators.

When a ServiceBindingRequest is created the Service Binding Operator
collects binding information and shares it with application. The
Binding Service Operator's controller injects the binding information
into the application's "DeploymentConfig", "Deployment" or “Replicaset"
as environment variables via an intermediate Secret called "binding-request".
The binding also works with Knative services as it works with any "deployment"
controller which has the podspec defined in the its jsonpath as
"spec.template.spec.containers".

This document provides "best practices" guidelines for the development of
Operators that manage backing services to be bound together with applications
by the Service Binding Operator.

## Making an Operator Managed Backing Service Bindable

In order to make a service bindable, the operator provider needs to express
the information needed by applications to bind with the services provided by
the operator. In other words, the operator provider must express the
information that is “interesting” to applications.

There are three methods for making Operator Managed Backing Service Bindable:

* [Operator Providing Metadata in CRD Annotations](#operator-providing-metadata-in-crd-annotations)
* [Operator Providing Metadata in OLM](#operator-providing-metadata-in-olm)
* [Operator Not Providing Metadata](#operator-not-providing-metadata)

### Operator Providing Metadata in CRD Annotations

This feature enables operator providers who do not use OLM (Operator Lifecycle
Manager) to provide metadata outside of an OLM descriptor. In this method,
the binding information is provided as annotations in the CRD of the operator
that manages the backing service. The Service Binding Operator extracts the
annotations to bind the application together with the backing service.

For example, this is a *bind-able* operator's annotations in its CRD for a
PostgreSQL database backing operator.
``` yaml
---
[...]
kind: CustomResourceDefinition
apiVersion: apiextensions.k8s.io/v1beta1
metadata:
  name: databases.postgresql.baiju.dev
  annotations:
    servicebindingoperator.redhat.io/status.dbConfigMap.password: 'binding:env:object:secret'
    servicebindingoperator.redhat.io/status.dbConfigMap.username: 'binding:env:object:configmap'
    servicebindingoperator.redhat.io/status.dbName: 'binding:env:attribute'
    servicebindingoperator.redhat.io/spec.Token.private: 'binding:volumemount:secret'
spec:
  group: postgresql.baiju.dev
  version: v1alpha1
```

### Operator Providing Metadata in OLM

This feature enables operator providers to specify binding information an
operator's OLM (Operator Lifecycle Manager) descriptor. The Service Binding
Operator extracts to bind the application together with the backing service.
The information may be specified in the "status" and/or "spec" section of the
OLM in plaintext or as a reference to a secret.

For example, this is a *bind-able* operator OLM Descriptor for a
PostgreSQL database backing operator.
``` yaml
---
[...]
statusDescriptors:
  description: Name of the Secret to hold the DB user and password
    displayName: DB Password Credentials
    path: dbCredentials
    x-descriptors:
      - urn:alm:descriptor:io.kubernetes:Secret
      - binding:env:object:secret:user
      - binding:env:object:secret:password
  description: Database connection IP address
    displayName: DB IP address
    path: dbConnectionIP
    x-descriptors:
      - binding:env:attribute
```

### Operator Not Providing Metadata

This feature enables operators that manage backing services but which don't
have any metadata in their CSV to use the Service Binding Operator to bind
together the service and applications. The Service Binding Operator binds all
sub-resources defined in the backing service CR by populating the binding
secret with information from Routes, Services, ConfigMaps, and Secrets owned
by the backing service CR.

[This is how resource and sub-resource relationships are set in
Kubernetes.](https://kubernetes.io/docs/concepts/workloads/controllers/garbage-collection/#owners-and-dependents)

The binding is initiated by the introduction of this API option in the backing service CR:
``` yaml
detectBindingResources : true
```
When this API option is set to true, the Service Binding Operator
automatically detects Routes, Services, ConfigMaps, and Secrets owned by
the backing service CR.

## Reference Operators

Reference backing service operators are available [here.](https://github.com/operator-backing-service-samples)

A set of examples, each of which illustrates a usage scenario for the
Service Binding Operator, is being developed in parallel with the Operator.
Each example makes use of one of the reference operators and includes
instructions for deploying the reference operators to a cluster, either
through the command line or client web console UI. The examples are
available [here.](https://github.com/redhat-developer/service-binding-operator/blob/master/README.md#example-scenarios)
