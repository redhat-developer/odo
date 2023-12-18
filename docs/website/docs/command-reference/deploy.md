---
title: odo deploy
---

`odo` can be used to deploy components in a similar manner they would be deployed by a CI/CD system, 
by first building the images of the containers to deploy, then by deploying the Kubernetes resources
necessary to deploy the components.

When running the command `odo deploy`, `odo` searches for the default command of kind `deploy` in the devfile, and executes this command.
The kind `deploy` is supported by the devfile format starting from version 2.2.0.

The `deploy` command is typically a *composite* command, composed of several *apply* and *exec* commands:
- an `apply` command referencing an `image` component that, when applied, will build the image of the container to deploy, and push it to its registry,
- an `apply` command referencing a [`kubernetes` component](https://devfile.io/docs/2.2.0/defining-kubernetes-resources) that, when applied, will create a Kubernetes resource in the cluster.
- an `exec` command referencing a container component that, when applied, will run the command defined by `commandLine` inside a container started by a Kubernetes Job; read more about it [here](../development/devfile.md#how-odo-runs-exec-commands-in-deploy-mode).

- With the following example `devfile.yaml` file, a container image will be built by using the `Dockerfile` present in the directory,
the image will be pushed to its registry and a Kubernetes Deployment will be created in the cluster, using this freshly built image.

```yaml
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
  - id: deploy-db
    exec:
      commandLine: helm repo add bitnami https://charts.bitnami.com/bitnami && helm install my-db bitnami/postgresql
      component: tools
  - id: deploy
    composite:
      commands:
        - build-image
        - deployk8s
        - deploy-db
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
  - name: tools
    container:
      image: quay.io/tkral/devbox-demo-devbox
```

:::note
The `uri` for the Dockerfile could also be an HTTP or HTTPS URL.
It may also point to a [`Containerfile`](https://www.mankier.com/5/Containerfile).
:::

import Note from '../_imageregistrynote.mdx';

<Note />

## Running the command
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

↪ Building & Pushing Container: quay.io/phmartin/myimage
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
COMMIT quay.io/phmartin/myimage
--> 7578e3e3667
Successfully tagged quay.io/phmartin/myimage:latest
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

↪ Executing command:
 ✓  Executing command in container (command: deploy-db) [12s]

Your Devfile has been successfully deployed

```
</details>

### Passing extra args to Podman or Docker when building images

Similarly to how [`odo build-images`](build-images.md#passing-extra-args-to-podman-or-docker) works, you can set the [`ODO_IMAGE_BUILD_ARGS` environment variable](../overview/configure.md#environment-variables-controlling-odo-behavior),
which is a semicolon-separated list of extra arguments to pass to Podman or Docker when building images.
See [this section](build-images.md#passing-extra-args-to-podman-or-docker) for further details.

```shell
ODO_IMAGE_BUILD_ARGS='arg1=value1;arg2=value2;...;argN=valueN' odo deploy
```

<details>
<summary>Example</summary>

```shell
$ ODO_IMAGE_BUILD_ARGS='--platform=linux/amd64;--build-arg=MY_ARG=my_value' odo deploy

  __                                                                                                                                                                                           
 /  \__     Running the application in Deploy mode using my-nodejs-app Devfile                                                                                                          
 \__/  \    Namespace: default                                                                                                                                                                 
 /  \__/    odo version: v3.10.0                                                                                                                                                               
 \__/
                                                                                                  
 ⚠  You are using "default" namespace, odo may not work as expected in the default namespace.                                                         
 ⚠  You may set a new namespace by running `odo create namespace <name>`, or set an existing one by running `odo set namespace <name>`

↪ Building Image: localhost:5000/nodejs-odo-example
 •  Building image locally  ...
[1/2] STEP 1/4: FROM registry.access.redhat.com/ubi8/nodejs-14:latest
[1/2] STEP 2/4: RUN echo XXX $MY_ARG
--> Using cache cbd3ef1317b96dbef4c9ab3646df49d3770831516c3b5c9f1e15687d67bc8803
--> cbd3ef1317b9
[1/2] STEP 3/4: COPY package*.json ./
--> Using cache de4a08bf2632ef49339beeda4ba50eb6e8a9b7524ffd5717fdcc372c15003b61
--> de4a08bf2632
[1/2] STEP 4/4: RUN npm install --production
--> Using cache 5a37e2783e140582da7ac4e241790e6e2052826c07f46cc0053801f4580e728c
--> 5a37e2783e14
[2/2] STEP 1/6: FROM registry.access.redhat.com/ubi8/nodejs-14-minimal:latest
[2/2] STEP 2/6: COPY --from=0 /opt/app-root/src/node_modules /opt/app-root/src/node_modules
--> Using cache 8779f5d3753baec5961b5ae017d8246b2674eb70f3c5607e4060f6b38e07c182
--> 8779f5d3753b
[2/2] STEP 3/6: COPY . /opt/app-root/src
--> 6ea250968b12
[2/2] STEP 4/6: ENV NODE_ENV production
--> 0bf4dd6605e9
[2/2] STEP 5/6: ENV PORT 3000
--> deea4247dd08
[2/2] STEP 6/6: CMD ["npm", "start"]
[2/2] COMMIT localhost:5000/nodejs-odo-example
--> eebc7c012506
Successfully tagged localhost:5000/nodejs-odo-example:latest
eebc7c01250682bf4e1e9544de1434d5edb90a51cf2d3e96f0faab354918bedb
 ✓  Building image locally [3s]
 •  Pushing image to container registry  ...
Getting image source signatures
Copying blob 4577dc5a9258 skipped: already exists  
Copying blob 6c30129af541 skipped: already exists  
Copying blob e48a40635da9 skipped: already exists  
Copying blob 1a5c88cd67e6 skipped: already exists  
Copying config a4bda6ab2b done  
Writing manifest to image destination
Storing signatures
 ✓  Pushing image to container registry [72ms]

↪ Deploying Kubernetes Component: my-component
 ✓  Creating kind Deployment 

↪ Executing command:
 ✓  Executing command in container (command: deploy-db) [12s]

Your Devfile has been successfully deployed
```
</details>


## Substituting variables

The Devfile can define variables to make the Devfile parameterizable. The Devfile can define values for these variables, and you 
can override the values for variables from the command line when running `odo deploy`, using the `--var` and `--var-file` options.

See [Substituting variables in `odo` dev](dev.md#substituting-variables) for more information.
