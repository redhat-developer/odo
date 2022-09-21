---
title: odo deploy
---

odo can be used to deploy components in a similar manner they would be deployed by a CI/CD system, 
by first building the images of the containers to deploy, then by deploying the Kubernetes resources
necessary to deploy the components.

When running the command `odo deploy`, odo searches for the default command of kind `deploy` in the devfile, and executes this command.
The kind `deploy` is supported by the devfile format starting from version 2.2.0.

The `deploy` command is typically a *composite* command, composed of several *apply* commands:
- a command referencing an `image` component that, when applied, will build the image of the container to deploy, and push it to its registry,
- a command referencing a [`kubernetes` component](https://devfile.io/docs/2.2.0-alpha/defining-kubernetes-resources) that, when applied, will create a Kubernetes resource in the cluster.

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

Note that the `uri` for the Dockerfile could also be an HTTP or HTTPS URL.

### Running the command
```shell
odo deploy
```
<details>
<summary>Example</summary>

```shell
$ odo deploy
  __
 /  \__     Deploying the application using my-nodejs Devfile
 \__/  \    Namespace: my-percona-server-mongodb-operator
 /  \__/    odo version: v3.0.0-rc1
 \__/

↪ Building & Pushing Container: quay.io/pvala18/myimage
 •  Building image locally  ...
STEP 1/7: FROM quay.io/phmartin/node:17
STEP 2/7: WORKDIR /usr/src/app
--> Using cache b18c8d9f4c739a91e5430f235b7beaac913250bec8bfcae531a8e93c750cea87
--> b18c8d9f4c7
STEP 3/7: COPY package*.json ./
--> Using cache cd151181cd9b2c69fc938eb89f3f71d0327d27ffba53c54247a105733cb36217
--> cd151181cd9
STEP 4/7: RUN npm install
--> Using cache 72b79a4f76ab0f9665653a974f5c667b1cb964c89c58e71aa4817b1055b1c473
--> 72b79a4f76a
STEP 5/7: COPY . .
--> 5c81f92690e
STEP 6/7: EXPOSE 8080
--> 9892b562a8a
STEP 7/7: CMD [ "node", "server.js" ]
COMMIT quay.io/pvala18/myimage
--> 7578e3e3667
Successfully tagged quay.io/pvala18/myimage:latest
7578e3e36676418853c579063dd190c9d736114ca414e28c8646880b446a1618
 ✓  Building image locally [2s]
 •  Pushing image to container registry  ...
Getting image source signatures
Copying blob 0b3c02b5d746 skipped: already exists
Copying blob 62a747bf1719 skipped: already exists
Copying blob 650b52851ab5 done
Copying blob 013fc0144002 skipped: already exists
Copying blob aef6a4d33347 skipped: already exists
Copying config 7578e3e366 done
Writing manifest to image destination
Storing signatures
 ✓  Pushing image to container registry [22s]

↪ Deploying Kubernetes Component: my-component
 ✓  Creating kind Deployment 

Your Devfile has been successfully deployed

```
</details>

## Substituting variables

The Devfile can define variables to make the Devfile parameterizable. The Devfile can define values for these variables, and you 
can override the values for variables from the command line when running `odo deploy`, using the `--var` and `--var-file` options.

See [Substituting variables in odo dev](dev.md#substituting-variables) for more information.
