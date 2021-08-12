---
title: OpenShift
sidebar_position: 2
---

# Setup an OpenShift cluster
*Note that this guide is only helpful in setting up a development environment; this setup is not recommended for a production environment.*

## Prerequisites
* This guide assumes that you have an OpenShift cluster setup, this could also be a [crc](https://crc.dev/crc/#installing-codeready-containers_gsg) cluster.
* This guide also assumes that you have admin privileges to the cluster, since Operator installation is only possible with an admin user.

## Install the Service Binding Operator
odo uses [Service Binding Operator](https://operatorhub.io/operator/service-binding-operator) to provide the `odo link` feature which helps connect an odo component to a service or another component.

1. Login to the OpenShift web console with admin, and navigate to Operators > OperatorHub.
2. Make sure that the Project is set to All Projects.
3. Search for `Service Binding Operator` in the search box under `All Items`.
4. Click on the `Service Binding Operator`; this should open a side pane.
5. Click on the `Install` button on the side pane; this should open an `Install Operator` page.
6. Make sure the `Installation mode` is set to `All namespaces on the cluster(default)`; `Installed Namespace` is set to `openshift-operators`; and `Approval Strategy` is `Automatic`.
7. Click on the `Install` button.
8. Wait until the operator is installed.
9. Once the operator is installed, you should see **Installed operator - ready for use**, and a **View Operator** button appears on the page.
10. Click on the **View Operator** button; this should take you to Operators > Installed Operators > Operator details page, and you should be able to see details of your operator.

## Install an Operator
1. Login to the OpenShift web console with admin, and navigate to Operators > OperatorHub.
2. Make sure that the Project is set to All Projects.
3. Search for an operator of your choice in the search box under `All Items`.
4. Click on the operator; this should open a side pane.
5. Click on the `Install` button on the side pane; this should open an `Install Operator` page.
6. Set the `Installation mode`, `Installed Namespace` and `Approval Strategy` as per your requirement.
7. Click on the `Install` button.
8. Wait until the Operator is installed.
9. Once the Operator is installed, you should see `Installed operator - ready for use`, and a `View Operator` button appears on the page.
10. Click on the `View Operator` button; this should take you to Operators > Installed Operators > Operator details page, and you should be able to see details of your operator.

## Verify the operator installation
Once the operator is successfully installed on the cluster, you can also use `odo` to verify the operator installation and see the CRDs associated with it; run the following command:
```shell
odo catalog list services
```
The output can look similar to:
```shell
$ odo catalog list services
Services available through Operators
NAME                                CRDs
datadog-operator.v0.6.0             DatadogAgent, DatadogMetric, DatadogMonitor
service-binding-operator.v0.9.1     ServiceBinding, ServiceBinding
```
