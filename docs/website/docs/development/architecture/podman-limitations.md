---
title: Podman limitations
sidebar_position: 40
---

The `odo dev` command is able to work either on Podman or on a Kubernetes cluster. 

The motivation behind the support for the Podman platform is to lower the learning curve
for developers working on containerized applications. As a matter of fact, Podman is simpler
to apprehend, install and maintain than a Kubernetes cluster.

Thanks to the support for the **Kubernetes Pod** abstraction by Podman, `odo`, and 
the user, can work on both Podman and Kubernetes on top of this abstraction.

Here are a list of limitations when `odo` is working on Podman:

## Apply command are not supported

A Devfile `Apply` command gives the possibility to "apply" any Kubernetes resource to the cluster. As Podman only supports a very limited number of Kubernetes resources, `Apply` commands are not executed by `odo` when running on Podman.

## Component listening on localhost

When working on a cluster, `odo dev` forwards the ports opened by the application to the developer's machine. This port forwarding works when the application is listening either on localhost or on `0.0.0.0` address.

Podman is natively not able to forward ports bound to localhost. In this situation, you may have two solutions:
- you can change your application to listen on `0.0.0.0`. This will be necessary for the ports giving access to the application or, in Production, this port would not be available (this port will most probably be exposed through an Ingress or a Route in Production, and these methods need the port to be bound to `0.0.0.0`),
- you can keep the port bound to `localhost`. This is the best choice for the Debug port, to restrict access to this Debug port. In this case, you can use the flag `--forward-localhost` when running `odo dev` on Podman. This way, you keep the Debug port secure on cluster.
