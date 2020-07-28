# Connecting Applications with Operator-backed Services

<p align="center">
    <a alt="GoReport" href="https://goreportcard.com/report/github.com/redhat-developer/service-binding-operator">
        <img alt="GoReport" src="https://goreportcard.com/badge/github.com/redhat-developer/service-binding-operator">
    </a>
    <a href="https://godoc.org/github.com/redhat-developer/service-binding-operator">
        <img alt="GoDoc Reference" src="https://godoc.org/github.com/redhat-developer/service-binding-operator?status.svg">
    </a>
    <a href="https://codecov.io/gh/redhat-developer/service-binding-operator">
        <img alt="Codecov.io - Code Coverage" src="https://codecov.io/gh/redhat-developer/service-binding-operator/branch/master/graph/badge.svg">
    </a>
</p>

## Introduction

The goal of the Service Binding Operator is to enable application authors to
import an application and run it on OpenShift with operator-backed services
such as databases, without having to perform manual configuration of secrets,
configmaps, etc.

In order to make a service bindable, the operator provider needs to express
the information needed by applications to bind with the services provided by
the operator. In other words, the operator provider must express the
information that is “interesting” to applications.

There are multiple methods for making operator managed backing services
bindable, including the backing operator providing metadata in CRD
annotations. Details on the methods for making backing services bindable
are available in the [Operator Best Practices Guide](docs/OperatorBestPractices.md)

In order to make an imported application (for example, a NodeJS application)
connect to a backing service (for example, a database):

* The app author (developer) creates a `ServiceBindingRequest` and specifies:
  * The resource that needs the binding information. The resource can be
    specified by label selectors;
  * The backing service's resource reference that the imported application
    needs to be bound to;

* The Service Binding Controller then:
  * Reads backing service operator CRD annotations to discover the
    binding attributes
  * Creates a binding secret for the backing service, example, an operator-managed database;
  * Injects environment variables into the applications's `Deployment`, `DeploymentConfig`,
    `Replicaset`, `KnativeService` or anything that uses a standard PodSpec;

## Quick Start

Clone the repository and run `make local` in an existing `kube:admin` openshift
CLI session. Alternatively, install the operator using:

``` bash
cat <<EOS |kubectl apply -f -
---
apiVersion: operators.coreos.com/v1
kind: OperatorSource
metadata:
  name: redhat-developer-operators
  namespace: openshift-marketplace
spec:
  type: appregistry
  endpoint: https://quay.io/cnr
  registryNamespace: redhat-developer
EOS
```

## Getting Started

The best way to get started with the Service Binding Operator is to see it in action.

We've included a number of examples scenarios for using the operator in this
repo. The examples are found in the "/examples" directory. Each of these
examples illustrates a usage scenario for the operator. Each example also
includes a README file with step-by-step instructions for how to run the
example.

We'll add more examples in the future. The following section in this README
file includes links to the current set of examples.

## Example Scenarios

The following example scenarios are available:

[Binding an Imported app with an In-cluster Operator Managed PostgreSQL Database](examples/nodejs_postgresql/README.md)

[Binding an Imported app with an Off-cluster Operator Managed AWS RDS Database](examples/nodejs_awsrds_varprefix/README.md)

[Binding an Imported Java Spring Boot app with an In-cluster Operator Managed PostgreSQL Database](examples/java_postgresql_customvar/README.md)

[Binding an Imported Quarkus app deployed as Knative service with an In-cluster Operator Managed PostgreSQL Database](examples/knative_postgresql_customvar/README.md)

[Binding an Imported app with an In-cluster Operator Managed ETCD Database](examples/nodejs_etcd_operator/README.md)

[Binding an Imported app to an Off-cluster Operator Managed IBM Cloud Service](examples/nodejs_ibmcloud_operator/README.md)

[Binding an Imported app in one namespace with an In-cluster Managed PostgreSQL Database in another namespace](examples/nodejs_postgresql_namespaces/README.md)
