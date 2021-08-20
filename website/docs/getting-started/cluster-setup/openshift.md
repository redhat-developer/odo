---
title: OpenShift
sidebar_position: 2
---

# Setup an OpenShift cluster
*Note that this guide is only helpful in setting up a development environment; this setup is not recommended for a production environment.*

## Prerequisites
* You have an OpenShift cluster setup, this could for example be a [crc](https://crc.dev/crc/#installing-codeready-containers_gsg) cluster.
* You have admin privileges to the cluster, since Operator installation is only possible with an admin user.

[//]: # (Move this section to Architecture > Service Binding or create a new Operators doc)
**What are Operators?**
>The Operator pattern aims to capture the key aim of a human operator who is managing a service or set of services. Human operators who look after specific applications and services have deep knowledge of how the system ought to behave, how to deploy it, and how to react if there are problems.
>
>People who run workloads on Kubernetes often like to use automation to take care of repeatable tasks. The Operator pattern captures how you can write code to automate a task beyond what Kubernetes itself provides.
> [(Source)](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/#motivation)
[//]: # (Move until here)

## Installing the Service Binding Operator
odo uses [Service Binding Operator](https://operatorhub.io/operator/service-binding-operator) to provide the `odo link` feature which helps connect an odo component to a service or another component.

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

## Installing an Operator
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

## Verifying the Operator installation
Once the Operator is successfully installed on the cluster, you can also use `odo` to verify the Operator installation and see the CRDs associated with it; run the following command:
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
