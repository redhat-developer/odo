---
title: Basics
sidebar_position: 3
---
# odo concepts

odo abstracts Kubernetes concepts into a developer focused terminology. In this document, we will take a look at each of these concepts and also their Kubernetes equivalent.
> The quoted text will talk about the Kubernetes equivalent of these concepts.

### Concepts

* **Component** 
  
  A component is like a microservice. Each component can be run and deployed separately. Examples of components: Warehouse API Backend, Inventory API, Web Frontend, Payment Backend

  > Deploying a component created with odo will create resources such as deployments or deploymentconfigs (which in turn creates replicasets and pods), services, and storage which are necessary to run the microservice in the Kubernetes cluster.

* **Application**
  
  An application is a group of one or more components that work individually to build an entire application. Examples of applications: e-Shop, Hotel Reservation System, Online Booking
  
  > An application can be considered as an equivalent of labels in Kubernetes that helps in grouping a set of resource.


* **Project**
  
  A project is a separate single unit that provides a scope for names and helps in dividing cluster resources between users. A resource name must be unique within the project but not across multiple projects.
  > A project is an equivalent of a namespace in Kubernetes. Creating a project with odo will create a namespace in the Kubernetes cluster with the same name.


* **Context**
  
  A context is the directory where the source code, tests, libraries and odo specific config files for your component resides. A single context can only contain one component.
  > There is no Kubernetes equivalent of a context because it is merely a directory.

* **URL**
  
  A URL exposes your component to be accessed outside the cluster.
  > A URL is an equivalent of Ingress in Kubernetes. Deploying a URL with odo will create an Ingress resource in the Kubernetes cluster. 


* **Storage**
  
  A storage is a way to "claim" persistent storage in the cluster environment. A storage can persist data across restarts and rebuilds of a component.
  > A storage is an equivalent of PVC in Kubernetes. Deploying a storage with odo will create a PVC resource in the Kubernetes cluster.


* **Service**
  
  A service is an external application that a component can connect to or depend on to gain some functionality. Example of services: MySQL, Redis
  > Deploying a service created with odo, creates necessary resources in the Kubernetes cluster to establish a proper connection between the component and the external application. odo service is not the same as a Kubernetes service, see [odo services vs. Kubernetes services](/basics#odo-services-vs-kubernetes-services).


* **Devfile**
  
  A portable YAML file responsible for your entire reproducible development environment. See [Devfile](../architecture/devfile.md) to know more about devfile.
  > A devfile can be considered similar to a manifest file in Kubernetes, but there is no equivalent of it in Kubernetes.
  

### Component vs. Application
A component is akin to a microservice, while an application is a group of components.


### odo services vs. Kubernetes services
Service in terms of odo is an external application that provides an additional functionality, odo service can only be created with [Operators](https://operatorframework.io/what/).
Service in terms of Kubernetes is a way of exposing a microservice accessible from the cluster to the outside world. Learn more about [Kubernetes Service](https://kubernetes.io/docs/concepts/services-networking/service/).


Deploying a service created with odo will create a set of resources in the Kubernetes cluster to establish a successful connection between the odo component and the external application.