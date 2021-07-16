---
title: Basics
sidebar_position: 3
---
# odo concepts

odo abstracts Kubernetes concepts into a developer focused terminology. In this document, we will take a look at each of these concepts.

### Concepts

* **Application**
  
  An application in odo is a classic application that is used to perform a particular task, developed with a [cloud-native approach](https://www.redhat.com/en/topics/cloud-native-apps).
  Examples of applications: Online Video Streaming, Hotel Reservation System, Online Shopping.
  
  In the cloud-native architecture, an application is a collection of small, independent, and loosely coupled components.
  

* **Component** 

  In the cloud-native architecture, an application is a collection of small, independent, and loosely coupled components; and an odo component is one of these components.
  Examples of components: Warehouse API Backend, Inventory API, Web Frontend, Payment Backend.


* **Project**
  
  A project helps achieve multi-tenancy: several applications can be run in the same cluster by different teams.


* **Context**
  
  A context is the directory on the system that contains the source code, tests, libraries and odo specific config files for a single component.

* **URL**
  
  A URL exposes a component to be accessed outside the cluster.


* **Storage**
  
  A storage is persistent storage in the cluster: it persists the data across restarts and rebuilds of a component.


* **Service**
  
  A service is an external application that a component can connect to or depend on to gain an additional functionality.
  Example of services: MySQL, Redis.


* **Devfile**
  
  A Devfile is a portable YAML file containing the definition of a component and its related URLs, storages and services. See [Devfile](../architecture/devfile.md) to know more about devfile.
