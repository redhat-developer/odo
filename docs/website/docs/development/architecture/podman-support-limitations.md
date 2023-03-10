---
title: Podman Support limitations
sidebar_position: 40
---

The `odo dev` command is able to work either on Podman or on a Kubernetes cluster. 

The motivation behind the support for the Podman platform is to lower the learning curve
for developers working on containerized applications, and to limit the physical resources 
necessary for development.

As a matter of fact, Podman is simpler to apprehend, install and maintain than a Kubernetes cluster, and can run with a minimal overhead on the developer machine.

Thanks to the support for the **Kubernetes Pod** abstraction by Podman, `odo`, and 
the user, can work on both Podman and Kubernetes on top of this abstraction.

Here are a list of limitations when `odo` is working on Podman:

## Commands working on Podman

- `odo dev --platform podman`

  This command will run the component in development mode on Podman. If you omit to use the `--platform` flag, `odo dev` works on cluster.

- `odo logs --platform podman`

  This command will display the component's logs from Podman. If you omit to use the `--platform` flag, `odo logs` get the logs from cluster.

- `odo list component [--platform podman]`

  This command without the `--platform` flag will list components from both the cluster and Podman. You can use the `--platform` flag to limit the search from a specific platform, either `cluster` or `podman`.

- `odo describe component [--platform podman]`

  This command without the `--platform` flag will describe a component from both the cluster and Podman. You can use the `--platform` flag to limit the search from a specific platform, either `cluster` or `podman`.

- `odo delete component [--platform podman]`

  This command without  the `--platform` flag will delete components from both the cluster and Podman. You can use the `--platform` flag to limit the deletion from a specific platform, either `cluster` or `podman`.


## Apply command are not supported

A Devfile `Apply` command gives the possibility to "apply" any Kubernetes resource to the cluster. As Podman only supports a limited number of Kubernetes resources, `Apply` commands are not executed by `odo` when running on Podman.

## Component listening on localhost not forwarded

When working on a cluster, `odo dev` forwards the ports opened by the application to the developer's machine. This port forwarding works when the application is listening either on localhost or on `0.0.0.0` address.

Podman is natively not able to forward ports bound to localhost. In this situation, you may have two solutions:
- you can change your application to listen on `0.0.0.0`. This will be necessary for the ports giving access to the application or, in Production, this port would not be available (this port will most probably be exposed through an Ingress or a Route in Production, and these methods need the port to be bound to `0.0.0.0`),
- you can keep the port bound to `localhost`. This is the best choice for the Debug port, to restrict access to this Debug port. In this case, you can use the flag `--forward-localhost` when running `odo dev` on Podman. This way, you keep the Debug port secure on cluster.

## Pod not updated when Devfile changes

When running `odo dev` on cluster, if you make changes to the Devfile affecting the definition of the deployed Pod (for example the memory or CPU requests or limits), the Pod will be recreated with its new definition. This behaviour is not supported yet for Podman.

## Pre-Stop events not supported

Pre-Stop events defined in the Devfile are not triggered when running `odo dev` on Podman.

