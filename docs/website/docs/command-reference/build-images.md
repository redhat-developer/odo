---
title: odo build-images
---

`odo` can build container images based on Dockerfiles, and push these images to their registries.

When running the command `odo build-images`, `odo` searches for all components in the `devfile.yaml` with the `image` type, for example:

```
components:
- image:
    imageName: quay.io/myusername/myimage
    dockerfile:
      uri: ./Dockerfile
      buildContext: ${PROJECTS_ROOT}
  name: component-built-from-dockerfile
```

The `uri` field indicates the relative path of the Dockerfile to use, relative to the directory containing the `devfile.yaml`. 
As indicated in the Devfile specification, `uri` could also be an HTTP or HTTPS URL.

The `buildContext` indicates the directory used as build context. The default value is `${PROJECT_SOURCE}`.

For each image component, `odo` executes either `podman` or `docker` (the first one found, in this order), to build the image with the specified Dockerfile, build context and arguments.

If the `--push` flag is passed to the command, the images will be pushed to their registries after they are built.

## Running the command
### Pre-requisites
* Login to an image registry ([quay.io](https://docs.quay.io/guides/login.html), [hub.docker.com](https://hub.docker.com/), etc)
* Dockerfile

```shell
odo build-images
```
<details>
<summary>Example</summary>

```shell
$ odo build-images

↪ Building & Pushing Container: quay.io/user/myimage
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
--> 84f475ad011
STEP 6/7: EXPOSE 8080
--> 12af8468cd0
STEP 7/7: CMD [ "node", "server.js" ]
COMMIT quay.io/user/myimage
--> 58c0731e9a1
Successfully tagged quay.io/user/myimage:latest
58c0731e9a110e8dbb2dbe4bdb55a15bdbbce1b78e121d350e23de79f33c3dde
 ✓  Building image locally [2s]
```
</details>

### Passing extra args to Podman or Docker

You can set the [`ODO_IMAGE_BUILD_ARGS` environment variable](../overview/configure.md#environment-variables-controlling-odo-behavior),
which is a semicolon-separated list of extra arguments to pass to Podman or Docker when building images.

A typical use case for this is to build images for a platform different from the one `odo` is running on.
For example, building images on Mac with Apple Silicon, with the intent to use them on a cluster supporting a different architecture or operating system.

```shell
ODO_IMAGE_BUILD_ARGS='arg1=value1;arg2=value2;...;argN=valueN' odo build-images
```

<details>
<summary>Example</summary>

```shell
$ ODO_IMAGE_BUILD_ARGS='--platform=linux/amd64;--build-arg=MY_ARG=my_value' odo build-images

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
 ✓  Building image locally [4s]

```
</details>

### Faking the image build
You can also fake the image build by exporting `PODMAN_CMD=echo` or `DOCKER_CMD=echo` to your environment. Read [environment variables controlling `odo` behaviour](../overview/configure.md#environment-variables-controlling-odo-behavior) for more information.

## Substituting variables

The Devfile can define variables to make the Devfile parameterizable. The Devfile can define values for these variables, and you
can override the values for variables from the command line when running `odo build-images`, using the `--var` and `--var-file` options.

See [Substituting variables in `odo` dev](dev.md#substituting-variables) for more information.

