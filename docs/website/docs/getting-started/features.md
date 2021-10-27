---
title: Features
sidebar_position: 2
---

# Features provided by odo

By using odo, application developers can develop, test, debug, and deploy microservices based applications on Kubernetes without having a deep understanding of the platform.

odo follows "create and push" workflow for almost everything. It means, as a user, when you "create", something the information (or manifest) is stored in a configuration file, and then upon doing a "push" it gets created  on the Kubernetes cluster. You can take an existing git repository and create an odo component from it, which can be pushed to a Kubernetes cluster.

odo helps "deploy and link" multiple components and services with each other. Using odo, developers can create and deploy services based on [Kubernetes Operators](https://github.com/operator-framework/) in their development cluster. These services can be created using any of the Operators available on [OperatorHub.io](https://operatorhub.io). Next, upon linking this service, odo injects the service configuration into the microservice created using odo. Your application can use this configuration to communicate with the Operator backed service.


### What can odo do?

odo uses container images to run the microservices in the cluster.

Full details of what each odo command is capable of doing can be found in the "Command Reference" sections.
Below is a summary of odo's most important capabilities:
* Create a manifest to deploy applications on Kubernetes cluster; odo creates the manifest for existing projects as well as new ones.
* No need to interact with YAML configurations; odo provides commands to create and update the manifest.
* Securely expose the application running on Kubernetes cluster to access it from developer's machine.
* Add and remove additional storage to the application on Kubernetes cluster.
* Create [Operator](https://github.com/operator-framework/) backed services and link with them.
* Create a link between multiple microservices deployed as odo components.
* Debug remote applications deployed using odo from the IDE.
* Run tests on the applications deployed on Kubernetes.

Take a look at "Using odo" section for guides on doing various things using odo.

### What features to expect in odo?

We are working on some exciting features like:
* Linking to services created using Helm package manager.
* Create `odo deploy` command to transition from inner loop to outer loop.
* Support for Knative eventing.

For a quick high level summary of the features we are planning to add, take a look at odo's [milestones on GitHub](https://github.com/openshift/odo/milestones).