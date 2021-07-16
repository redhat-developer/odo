---
title: Basics
sidebar_position: 3
---
# odo concepts

odo abstracts Kubernetes concepts into a developer focused terminology. In this document, we will take a look at each of these concepts.

### Concepts

* **Component** 
  
  A component is like a microservice. Each component can be run and deployed separately. Examples of components: Warehouse API Backend, Inventory API, Web Frontend, Payment Backend

* **Application**
  
  An application is a group of one or more components that work individually to build the entire application. Examples of applications: Online Video Streaming, Hotel Reservation System, Online Booking


* **Project**
  
  A project is a separate single unit that provides a scope for names and helps in dividing cluster resources between users. A resource name must be unique within the project but not across multiple projects.


* **Context**
  
  A context is the directory where the source code, tests, libraries and odo specific config files for your component resides. A single context can only contain one component.

* **URL**
  
  A URL exposes your component to be accessed outside the cluster.


* **Storage**
  
  A storage is a way to claim persistent storage in the cluster. A storage can persist data across restarts and rebuilds of a component.


* **Service**
  
  A service is an external application that a component can connect to or depend on to gain some functionality. Example of services: MySQL, Redis


* **Devfile**
  
  A portable YAML file responsible for your entire reproducible development environment. See [Devfile](../architecture/devfile.md) to know more about devfile.
  

### odo services vs. Kubernetes services
Service in terms of odo is an external application that provides an additional functionality, odo service can only be created with [Operators](https://operatorframework.io/what/).

Service in terms of Kubernetes is an abstract way of exposing a microservice as a set of pods. Learn more about [Kubernetes Service](https://kubernetes.io/docs/concepts/services-networking/service/).

Deploying a service created with odo will create a set of resources in the Kubernetes cluster to establish a successful connection between the odo component and the external application.