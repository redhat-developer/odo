---
title: odo deploy
sidebar_position: 2
---

odo can be used to deploy components in a similar manner they would be deployed by a CI/CD system, 
by first building the images of the containers to deploy, then by deploying the Kubernetes resources
necessary to deploy the components.

When running the command `odo deploy`, odo searches for the default command of kind `deploy` in the devfile, and executes this command.
The kind `deploy` is supported by the devfile format starting from version 2.2.0.

The `deploy` command is typically a *composite* command, composed of several *apply* commands:
- a command referencing an `image` component that, when applied, will build the image of the container to deploy, and push it to its registry,
- a command referencing a [`kubernetes` component](https://devfile.io/docs/devfile/2.2.0/user-guide/adding-a-kubernetes-or-openshift-component-to-a-devfile) that, when applied, will create a Kubernetes resource in the cluster.

With the following example `devfile.yaml` file, a container image will be built by using the `Dockerfile` present in the directory,
the image will be pushed to its registry and a Kubernetes Deployment will be created in the cluster, using this freshly built image.

```
schemaVersion: 2.2.0
[...]
variables:
  CONTAINER_IMAGE: quay.io/phmartin/myimage
commands:
  - id: build-image
    apply:
      component: outerloop-build
  - id: deployk8s
    apply:
      component: outerloop-deploy
  - id: deploy
    composite:
      commands:
        - build-image
        - deployk8s
      group:
        kind: deploy
        isDefault: true
components:
  - name: outerloop-build
    image:
      imageName: "{{CONTAINER_IMAGE}}"
      dockerfile:
        uri: ./Dockerfile
        buildContext: ${PROJECTS_ROOT}
  - name: outerloop-deploy
    kubernetes:
      inlined: |
        kind: Deployment
        apiVersion: apps/v1
        metadata:
          name: my-component
        spec:
          replicas: 1
          selector:
            matchLabels:
              app: node-app
          template:
            metadata:
              labels:
                app: node-app
            spec:
              containers:
                - name: main
                  image: {{CONTAINER_IMAGE}}
```
