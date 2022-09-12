---
title: odo build-images
---

odo can build container images based on Dockerfiles, and push these images to their registries.

When running the command `odo build-images`, odo searches for all components in the `devfile.yaml` with the `image` type, for example:

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

For each image component, odo executes either `podman` or `docker` (the first one found, in this order), to build the image with the specified Dockerfile, build context and arguments.

If the `--push` flag is passed to the command, the images will be pushed to their registries after they are built.

## Running the command
### Pre-requisites
* Login to an image registry(quay.io, hub.docker.com, etc)
* Dockerfile

```shell
odo build-images
```
```shell
$ odo build-images

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
--> 84f475ad011
STEP 6/7: EXPOSE 8080
--> 12af8468cd0
STEP 7/7: CMD [ "node", "server.js" ]
COMMIT quay.io/pvala18/myimage
--> 58c0731e9a1
Successfully tagged quay.io/pvala18/myimage:latest
58c0731e9a110e8dbb2dbe4bdb55a15bdbbce1b78e121d350e23de79f33c3dde
 ✓  Building image locally [2s]
```

### Faking the image build
You can also fake the image build by exporting `PODMAN_CMD=echo` to your environment.