---
title: Development in an air-gapped environment
sidebar_position: 6
tags: ["air-gapped", "disconnected", "offline", "environment", "container", "kubernetes", "openshift", "development"]
slug: container-based-application-development-air-gapped-environment
---

# Container-based application development in an air-gapped environment

An air-gapped environment (otherwise known as offline or disconnected environment) will prevent both `odo` and the application
developed with `odo` to access the internet directly.

In this case, you will need to apply some configuration and/or deploy some elements in the local network to be able to work with `odo`.

The development with `odo` is done in two phases. The first phase is to obtain a Devfile from a Devfile Registry, and the second phase
is to run the application into a cluster, based on the configuration defined in the Devfile.

## Accessing a Devfile Registry

If your environment provides an HTTPS proxy giving access to the upstream Devfile Registry (https://registry.devfile.io/),
you can define the `HTTPS_PROXY` environment variable when executing `odo init`. For example:

```
$ HTTPS_PROXY=https://your_proxy odo init
```

If you cannot rely on the upstream Devfile Registry, you can install a registry in your local network.
The following procedures provide the instructions to build a container image for an offline registry
(where all Devfiles reference container images and starter projects in your local environment),
and how to deploy the image you have built in a local Kubernetes cluster:
- [Installation of in-cluster offline devfile registry](https://devfile.io/docs/2.2.0/installation-of-in-cluster-offline-devfile-registry)
- [Deploying a devfile registry](https://devfile.io/docs/2.2.0/deploying-a-devfile-registry)

## Running the application

To execute the application in the cluster, `odo` first deploys a Pod into this cluster, based on a container image defined in the Devfile. This container image will be pulled by the cluster to instantiate the pod.

Then, the source files will be synchronized into the container and the application will be built from inside the container. Depending on the language and/or framework used for your application, the build may need to access a dependency registry (NPM registry, Maven repository, etc).

### Accessing the cluster's control-plane

To create resources into the cluster, `odo` needs to communicate with the cluster's control-plane through its API.

If the cluster is not accessible directly from the air-gapped environment but accessible through an HTTPS proxy, you can define the `HTTPS_PROXY` environment variable when executing `odo dev` and `odo deploy`. For example:

```
$ HTTPS_PROXY=https://your_proxy odo dev
```

### Pulling the container image

If this image is accessible from a local container registry without any authentication,
you don't need to add any configuration.

If the container registry requires some authentication, you will need to pass the credentials
to the cluster, using an `ImagePullSecret`. Instructions to work with `ImagePullSecret` resources
are provided here: [Pull an Image from a Private Registry](https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/).

As described in the previous instructions, you will need to define an `ImagePullSecret` for the Pod
that will be deployed by `odo`. For this, you can use the [`pod-overrides`](https://devfile.io/docs/2.2.0/overriding-pod-and-container-attributes#pod-overrides) feature provided by the Devfile, for example:

```yaml
[...]
components:
- container:
    args:
    - tail
    - -f
    - /dev/null
    endpoints:
    - name: http-node
      targetPort: 3000
    - exposure: none
      name: debug
      targetPort: 5858
    env:
    - name: DEBUG_PORT
      value: "5858"
    image: your_secure_registry/nodejs-16:latest
    memoryLimit: 1024Mi
    mountSources: true
  name: runtime
# highlight-start
attributes:
  pod-overrides:
    spec:
      imagePullSecrets:
      - name: regcred
# highlight-end
```

### Accessing a dependency registry

If you are using a local registry, you will need to configure the build to pass it the address of the local registry. You can either modify the command-line for the build, or define additional environment variables, depending on your needs.

To modify the build command-line, you can edit the `install` and/or `build` command in the Devfile. For example:

```yaml
[...]
commands:
- exec:
# highlight-start
    commandLine: npm install --registry https://your_local_registry
# highlight-end
    component: runtime
    group:
      isDefault: true
      kind: build
    workingDir: ${PROJECT_SOURCE}
  id: install
```

Or:

```yaml
[...]
commands:
- exec:
# highlight-start
    commandLine: npm config set registry https://your_local_registry && npm install
# highlight-end
    component: runtime
    group:
      isDefault: true
      kind: build
    workingDir: ${PROJECT_SOURCE}
  id: install
```

You can also provide additional environment variables to the container:

```yaml
[...]
components:
- container:
    args:
    - tail
    - -f
    - /dev/null
    endpoints:
    - name: http-node
      targetPort: 3000
    - exposure: none
      name: debug
      targetPort: 5858
    env:
    - name: DEBUG_PORT
      value: "5858"
# highlight-start
    - name: npm_config_registry
      value: https://your_local_registry
# highlight-end
    image: your_secure_registry/nodejs-16:latest
    memoryLimit: 1024Mi
    mountSources: true
  name: runtime
```
