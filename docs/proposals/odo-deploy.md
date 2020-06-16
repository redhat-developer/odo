# Support build and deployment of application container image using odo

## Abstract
Add a new command (verb) to build a production-like/slim container image for the project and deploy the built image on the target platform.

There is a proposal for adding on outer-loop information (including support for multiple strategies for building and deploying the projects) to Devfile 2.1.0: 
https://github.com/devfile/kubernetes-api/issues/49 

It would be useful to start the design/development of a simpler version of `odo deploy` with devfile 2.0.0 that covers:
- single build strategy - Dockerfile built using `buildah`
- single deployment strategy - Kubernetes manifest deployed using `kubectl`.

## Motivation
`odo` provides a great way to develop containerized applications on Docker and Kubernetes platforms using applications stacks based on devfiles. However, this is limited to development inner-loop for a project and there is no support for outer-loop - build a slim/production container image for your application and deploy the build container.

It would be very useful for application developers to be able to build a production-like container image and perform a Kubernetes deployment for their project. With devfile v2.0.0 this information could be provided by the application stack (devfile) to avoid developers having to worry about these aspects. This command can be a good way to assure the developer that their application can be built and deployed successfully using the build/deploy guidance that comes from devfile.

It is not meant to replace GitOps/pipeline-based deployments for governed environments (test, stage, production). However, both `odo deploy` and `odo pipelines` must honour the build and deployment information provided by the application stack (devfile).

## User Stories

This command will allow a user to perform inner-loop and then test the outer-loop, so the application is truly ready for checking into git and for pipelines to take over. Here's how a typical application development flow might look like: 

User flow: 
1. `odo create <component> <mycomponent>` - This initializes odo component in the current directory.
1. Edit the project source code to develop the application.
1. `odo URL create` - This stores URL information (host, port etc.) for accessing the application on the cluster.
1. `odo push` - This runs the application source code using inner-loop instructions from devfile.
1. Validate the running application is working as intended.
1. Iterate over steps 2 and beyond (as needed). 
1. `odo deploy` - This would build a new clean image and deploy it to the target cluster using outer-loop instructions from devfile.
1. Validate the deployed application is working as intended.
1. Iterate over steps 2 and beyond (as needed).
1. `odo pipelines bootstrap` - This would bootstrap the GitOps repository using outer-loop instructions from the devfile. 

Additional Notes: 
- `odo deploy` will not make any changes to the user's project.
- 
### Initial build and deploy support to odo - https://github.com/openshift/odo/issues/3300
### Delete deployment support in odo - <Need an issue on this>

## Design overview
`odo deploy` could provide developers with a way to build a container image for their application and deploy it on a target Kubernetes deployment using the build/deploy guidance provided by the stack (devfile 2.0.0).

This deployment will be using the namespace, service binding, URL information from the inner-loop artefacts. This will ensure that `odo deploy` is not seen as a way to deploy workloads in production and leave more complex production gitops options to `odo pipelines`.

`odo deploy delete` could provide developers with a way to clean up any existing deployment of the application. This should also be done as part of `odo delete` to make sure that we do not leave behind deployed apps.

Command and flags:
- `odo deploy` - build a container image for their application and deploy it on a target Kubernetes deployment.
 - `--tag`: The tag to be used for the built application container image - `<registry>/<repo>/<name>/<tag>` (mandatory).
 - `--credentials`: The credentials for the container image registry where the application image will be pushed/pulled from (mandatory).

- `odo deploy delete` and `odo delete` - delete any existing deployment of the application.
 
### Specify outer-loop information in devfile v2.0.0

There is a proposal for adding outer-loop information (including support for multiple strategies for building and deploying the projects) to Devfile 2.1.0: 
https://github.com/devfile/kubernetes-api/issues/49 

For the initial implementation, we could use devfile v2.0.0 and capture basic outer-loop information as `metadata` on the devfile. 

For example: 
```
metadata:
 dockerfile: <URI>
 deployment-manifest: <URI>
```

This should not need any change to the devfile v2 schema or the parser.

The deployment manifest could be templated to help with replacing key bits of information:
- PROJECT_NAME
- CONTAINER_IMAGE
- PORT

For example: 
```
apiVersion: app.stacks/v1beta1
kind: RuntimeComponent
metadata:
 name: PROJECT_NAME
spec:
 applicationImage: CONTAINER_IMAGE 
 service:
 type: ClusterIP
 port: PORT
 expose: true
 storage:
 size: 2Gi
 mountPath: "/logs"
```

### Utilize outer-loop information from devfiles

`odo deploy` will perform the following actions:

#### Startup
- Check if the devfile specifies the expected outer-loop metadata, If not provided, display a meaningful error message to the user.

- Validate all arguments passed by the user, and display a meaningful error message if argument values are invalid.

#### Build
- Start a new pod with a build container (using stable buildah image).
- Mount the image registry credentials passed by the user.
- Copy the application source code into the build container.
- Fetch the dockerfile using URI in the metadata of the devfile and make it available in the container.
- Execute `buildah bud` command with right src code, dockerfile and tag information.
- Push the built image using the tag and credentials.
- Delete the build container and the pod.

#### Deploy
- Fetch the deployment manifest using URI in the metadata of the devfile and make it available in the container.
- Replace templated text in the deployment manifest with relevant values:
 - PROJECT_NAME: name of odo project
 - CONTAINER_IMAGE: `tag` for the built image
 - PORT: URL information in env.yaml
- Inject any service binding environment variables from the user's project into the deployment manifest.
- Check if there is an existing deployment for the app. If that is the case, then delete the existing deployment (will only be true when running `odo deploy` multiple times).
- Apply the deployment manifest.
- Wait for the application container to start.
- Provide the user with a URL they can use to access the deployed application. 

`odo deploy delete` and will perform the following actions:
- Check if there is an existing deployment for the app, then delete it, else show a meaningful error message to the user. 

## Future evolution

- Devfile 2.1.0 should broaden the scope for the outer-loop support in devfiles. For example:
    - Support multiple build strategies - buildah, s2i, build v3 etc.
    - Support multiple deployment strategies - k8s manifest, native service, pod spec etc.
    - Any referenced assets should be immutable to ensure reproducible builds/deployments.

- If a devfile does not provide deployment manifest, odo can perhaps create a manifest in the way it does for inner-loop. This will mean devfile creators do not need to provide a deployment manifest if they do not care so much about the deployment aspect.