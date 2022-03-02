---
title: OpenShift
sidebar_position: 2
---

# Setting up a OpenShift cluster

## Introduction
This guide is helpful in setting up a development environment intended to be used with `odo`; this setup is not recommended for a production environment.

## Prerequisites
* You have a OpenShift cluster set up (such as [crc](https://crc.dev/crc/#installing-codeready-containers_gsg))
* You have admin privileges to the cluster

## Summary
* An Operator in order to use `odo service`
* (Optional) Service Binding Operator in order to use `odo link`

## Installing an Operator

Installing an Operator allows you to install a service such as PostgreSQL, Redis or DataDog.

To install an Operator from the OpenShift web console:
1. Login to the OpenShift web console with admin, and navigate to Operators > OperatorHub.
2. Make sure that the Project is set to All Projects.
3. Search for an Operator of your choice in the search box under **All Items**.
4. Click on the Operator; this should open a side pane.
5. Click on the **Install** button on the side pane; this should open an **Install Operator** page.
6. Set the **Installation mode**, **Installed Namespace** and **Approval Strategy** as per your requirement.
7. Click on the **Install** button.
8. Wait until the Operator is installed.
9. Once the Operator is installed, you should see _**Installed operator - ready for use**_, and a **View Operator** button appears on the page.
10. Click on the **View Operator** button; this should take you to Operators > Installed Operators > Operator details page, and you should be able to see details of your Operator.

### Verifying the Operator installation

Once the Operator is successfully installed on the cluster, you can use `odo` to verify the Operator installation and see the CRDs associated with it; run the following command:

```shell
odo catalog list services
```

The output may look similar to:

```shell
odo catalog list services
Services available through Operators
NAME                                CRDs
datadog-operator.v0.6.0             DatadogAgent, DatadogMetric, DatadogMonitor
service-binding-operator.v0.9.1     ServiceBinding, ServiceBinding
```

## (Optional) Installing the Service Binding Operator

`odo` uses [Service Binding Operator](https://operatorhub.io/operator/service-binding-operator) to provide the `odo link` feature which helps to connect an odo component to a service or another component.

The Service Binding Operator is _optional_ and is used to provide extra metadata support for `odo` deployments.

To install the Service Binding Operator from the OpenShift web console:
1. Login to the OpenShift web console with admin, and navigate to Operators > OperatorHub.
2. Make sure that the Project is set to All Projects.
3. Search for _**Service Binding Operator**_ in the search box under **All Items**.
4. Click on the **Service Binding Operator**; this should open a side pane.
5. Click on the **Install** button on the side pane; this should open an **Install Operator** page.
6. Make sure the **Installation mode** is set to "_All namespaces on the cluster(default)_"; **Installed Namespace** is set to "_openshift-operators_"; and **Approval Strategy** is "_Automatic_".
7. Click on the **Install** button.
8. Wait until the Operator is installed.
9. Once the Operator is installed, you should see **_Installed operator - ready for use_**, and a **View Operator** button appears on the page.
10. Click on the **View Operator** button; this should take you to Operators > Installed Operators > Operator details page, and you should be able to see details of your Operator.

