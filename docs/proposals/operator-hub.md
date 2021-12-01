# Operator Hub integration
odo currently supports creation, listing and deletion of services created from
the Service Catalog. With newer versions of OCP, Service Catalog is being
deprecated in favour of Operator Hub.

This proposal discusses adding similar support for Operator Hub as is currently
available for Service Catalog. Idea is to be able to create, list and delete
the services (basically Custom Resources in Operator framework terminology) on
an OpenShift or Kubernetes cluster. Considering the ongoing `kclient` and DevFile
efforts, we intend to add support for Kubernetes from day 1. However, there
will be certain prerequisites that need to be satisfied in case of Kubernetes.

Table of Contents
- [Which Operators do we support](#which-operators-do-we-support)
- [Installing Operators](#installing-operators)
- [List the services/operators](#list-the-servicesoperators)
  - [Both Service Catalog and Operator Hub are enabled](#both-service-catalog-and-operator-hub-are-enabled)
  - [Only Service Catalog is enabled](#only-service-catalog-is-enabled)
  - [Only Operator Hub is enabled](#only-operator-hub-is-enabled)
  - [Kubernetes](#kubernetes)
- [Link components with operands](#link-components-with-operands)
- [Listing the services](#listing-the-services)
- [Describe deployed services](#describe-deployed-services)
- [Delete the service](#delete-the-service)
- [Testing on Kubernetes](#testing-on-Kubernetes)
- [Questions](#questions)
- [GitHub issues related to this task](#github-issues-related-to-this-task)

## Which Operators do we support
Operator Hub is not the only place to find/list Operators that could be
installed on a OpenShift/k8s cluster. There’s also KUDO. However, to start
things off, we must focus on a limited number of Operators and iterate from
there. Hence, we will focus only on OperatorHub (kind of obvious 😉).

There are three main categories of Operators as described in the OpenShift
docs. We could focus on one of those categories or select specific Operators
from each of the categories. OpenShift comes pre-installed with [Operator
Lifecycle
Manager](https://github.com/operator-framework/operator-lifecycle-manager/)
which helps in creating Operator backed services. We will levarage this to spin
up services using odo.

From a development perspective, we might just take a CRD definition and throw
it to the Kubernetes API. The controller running in background would take care
of spinning up the service. So, it wouldn’t matter a lot to differentiate
between what we support and what we don’t. However, limiting our scope to a few
Operators will keep us from being overwhelmed.

## Installing Operators
There’s an open issue about adding capability in odo to install Operators and
installing “Service Binding Operator” (which helps bind applications with
Operator-backed services) as a one-off command.

As far as creating services from Service Catalog is concerned, we expect the
users to enable Service Catalog in their Minishift or OpenShift cluster before
being able to see any services in `odo catalog list services` output. Why should
this be different for Operators? We should mention in the documentation about
using `kube:admin` user for CRC and/or contacting administrator to install
required Operators in OpenShift cluster.

As for using Service Binding Operator, we also should explore the possibility
of linking services to components without using the Operator.

## List the services/operators
Taking this from [Tomas’s
comment](https://github.com/redhat-developer/odo/issues/2461#issuecomment-566577064)
on the issue. Added column for available plans. But do Operators have “PLANS”?
We should prune the PLANS column and let that piece be handled by `odo catalog
describe service <service-name>`.

### Both Service Catalog and Operator Hub are enabled
```
$ odo catalog list services
NAME                          PLANS                    PROVISIONER
mariadb-ephemeral             default                  ServiceCatalog
mongodb-ephemeral             default                  ServiceCatalog
mysql-ephemeral               default                  ServiceCatalog
mongodb-enterprise.v1.2.4     prod                     OperatorHub
mariadb-operator-0.1.3-6      ephemeral,persistent     OperatorHub
```
### Only Service Catalog is enabled
```
$ odo catalog list services
NAME                          PLANS                    PROVISIONER
mariadb-ephemeral             default                  ServiceCatalog
mongodb-ephemeral             default                  ServiceCatalog
mysql-ephemeral               default                  ServiceCatalog
```

### Only Operator Hub is enabled
```
$ odo catalog list services
NAME                          PLANS                    PROVISIONER
mongodb-enterprise.v1.2.4     prod                     OperatorHub
mariadb-operator-0.1.3-6      ephemeral,persistent     OperatorHub
```

When working on OpenShift, we can get a list of enabled Operators in a
namespace (project) by doing `oc get csvs`. 

### Kubernetes
Listing Operators in Kubernetes might involve some extra tinkering. Operator
Hub is not pre-installed on a minikube cluster. In absence of any Operators,
`kubectl get crds` still works and shows No resource found  but `kubectl get
csvs` doesn’t work since there’s no API resources with that name in Kubernetes.

One needs to manually install OLM from its
[repo](https://github.com/operator-framework/operator-lifecycle-manager/tree/00eab85df0fd570891754cc743bc3d5831e9dd62/deploy/upstream/quickstart)
to make ClusterServiceVersion objects available in Kubernetes.

## Create a service
At the moment, we don't have a lot of clarity around how the exact `odo service
create` command will look for Operator Hub backed stuff. But it could be
something like:

```sh
$ odo service create <operator-name> <service-name> --crd <crd-name> -p parameter1=value1 -p parameter2=value2
```

## Link components with operands
[GitHub issue](https://github.com/redhat-developer/odo/issues/2463)

In its current form, linking a component with another component/service
involves making the Secrets (environment variables) of the component/service
being linked available to the component it is linked with. For example, if
`component A` needs to be connected to `service A`  then the environment variables
of `service A` will be made available in `component A`.

Although Service Binding Operator could provide an easy mechanism to link the
components with operands created from an operator, we intend to explore linking
the two without adding the dependency on this operator. This should help us be
agnostic of backend cluster being OpenShift or Kubernetes.

## Listing the services
[GitHub issue](https://github.com/redhat-developer/odo/issues/2479)

At the moment, we use `odo service list --app <app-name> --project
<project-name>` to list the services in a given application and project. We
could augment this to have the information about the services created from the
Operators. To make a distinction between the source of the services in the
cluster, we could add a column to the output just like we plan to do for
listing available/installed operators via `odo catalog list services`.

```sh
$ odo services list
NAME                          TYPE                  PLANS                    PROVISIONER
mongodb-enterprise.v1.2.4     dh-postgresql-apb     prod                     OperatorHub
mariadb-operator-0.1.3-6      mariadb               persistent               ServiceCatalog
```

For Operator Hub instances, `TYPE` column could indicate the CRD used to spin
up the CR/service.

## Describe deployed services
[GitHub issue](https://github.com/redhat-developer/odo/issues/2480)

At the moment there’s no command in odo that describes a deployed service. We
have  a command to describe a service in the Service Catalog: `odo catalog
describe service <service-name>`. But there’s no way to describe a deployed
service. We can only list the deployed services with `odo service list`.

If we intend to add a command to describe the deployed service(s), we need to:
1. Check with current users (CLI users, IDE plugin teams, GSS) if this is
something that would help them.
2. Define what “describing a deployed service” exactly means
3. Decide what should be shown to the user as the output on CLI
4. Implement the command for both ServiceCatalog and OperatorHub

## Delete the service
[GitHub issue](https://github.com/redhat-developer/odo/issues/2481)

At the moment, we use `odo service delete <service-name>` to delete the service
deployed in OpenShift cluster. We should be able to delete the service
deployed using Operator Hub with the same command.

## Testing on Kubernetes
Our current CI tests odo against different OCP versions. How do we make sure
that our code/tooling is working on Kubernetes as well? How do we add this to
our existing CI?

## Questions
1. How do the CustomResourceDefinitions oc get crds get populated in the
   cluster?  It looks like this is because Operator Hub is installed by default
   on OCP.
2. Kubernetes doesn’t have Operator Hub installed by default like OCP has. How
   do we list Operators on it? It doesn’t have ClusterServiceVersions (CSVs)
   either.  And there are other Operator Registries for it like KUDO operators.
3. In future, we might want to add support for services to DevFile so as to
   spin up components linked with services

## TODO
1. Need more clarity around how to create the service/operand using the CRDs in
   interactive and non-interactive mode.

## GitHub issues related to this task

1. [Validating the cluster for operator and service catalog capability and odo
   service list](https://github.com/redhat-developer/odo/issues/2461)
2. [odo service describe for operators +
   services](https://github.com/redhat-developer/odo/issues/2480)
3. [Instantiating operators using
   CRDs](https://github.com/redhat-developer/odo/issues/2462)
4. [Status command for operators](https://github.com/redhat-developer/odo/issues/2479)
5. [Use the service binding operator to link operator and a
   component](https://github.com/redhat-developer/odo/issues/2463)
6. [Provide admin commands to install operators and one-off
   setup](https://github.com/redhat-developer/odo/issues/2464)
7. [Enhance the odo service delete command to allow deletion of
   operators](https://github.com/redhat-developer/odo/issues/2481)
