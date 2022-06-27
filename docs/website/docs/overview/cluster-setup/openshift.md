---
title: OpenShift
sidebar_position: 2
---

# Setting up a OpenShift cluster

## Introduction
This guide is helpful in setting up a development environment intended to be used with `odo`; this setup is not recommended for a production environment.

## Requirements
* You have a OpenShift cluster set up (such as [crc](https://crc.dev/crc/#installing-codeready-containers_gsg))
* You have admin privileges to the cluster

## (OPTIONAL) Installing the Service Binding Operator

Service Binding Operator is required to bind an application with microservices.

Visit the [official documentation](https://redhat-developer.github.io/service-binding-operator/userguide/getting-started/installing-service-binding.html#installing-the-service-binding-operator-from-the-openshift-container-platform-web-ui) of Service Binding Operator to see how you can install it on your OpenShift cluster.
