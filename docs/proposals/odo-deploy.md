# Support build and deployment of application container image using odo

## Abstract
Add a new command (verb) to build a production-like/slim container image for the project and deploy the built image on the target platform.

There is an existing proposal for adding on outer-loop information (including support for multiple strategies for building and deploying the projects) to Devfile 2.1.0: 
https://github.com/devfile/kubernetes-api/issues/49 

It would be useful to start the design/development of a simpler version of `odo deploy` with devfile 2.0.0 that covers:
- single build strategy - Dockerfile built using `BuildConfig` or `Kaniko`.
- single deployment strategy - Kubernetes manifest deployed using `kubectl`.

## Motivation
`odo` is limited to development inner-loop for a project and there is no support for outer-loop - build a slim/production container image for your application and deploy the build container. It would be very useful for developers to be able to try inner-loop and then transition over to the outer-loop using odo. The outer-loop information could be provided by the application stack (devfile) to avoid developers having to worry about these aspects. 

`odo deploy` can be a good way to assure the developer that their application can be built and deployed successfully using the build/deploy guidance that comes from devfile.

It is not meant to replace GitOps/pipeline-based deployments for governed environments (test, stage, production). However, both `odo deploy` and `odo pipelines` must honour the build and deployment information provided by the application stack (devfile).

## User flow

This command will allow a user to test the transition from inner-loop t outer-loop, so the application is truly ready for pipelines to take over. Here's how a typical application development flow might look like: 

User flow: 
1. `odo create <component> <mycomponent>` - This initializes odo component in the current directory.
1. Edit the project source code to develop the application.
1. `odo url create` - This stores URL information (host, port etc.) for accessing the application on the cluster (if not done already).
1. `odo push` - This runs the application source code using inner-loop instructions from devfile.
1. Validate the running application is working as intended.
1. Iterate over steps 2 and beyond (as needed). 
1. `odo deploy` - This would build a new clean image and deploy it to the target cluster using outer-loop instructions from devfile and user-provided arguments.
1. Validate the deployed application is working as intended.
1. Iterate over steps 2 and beyond (as needed).
1. Optionally, run `odo deploy delete` to clean up resources created with `odo deploy`.
1. Push your code to Git - ready for sharing with your team and for CI/CD pipelines to take over.

## User stories

### Initial build and deploy support to odo - https://github.com/redhat-developer/odo/issues/3300

## Design overview
`odo deploy` could provide developers with a way to build a container image for their application and deploy it on a target Kubernetes deployment using the build/deploy guidance provided by the devfile.

This deployment is equivalent to a development version of your production and will be using the namespace and URL information from the inner-loop. This will ensure that it is not seen as a way to deploy real workloads in production.

`odo deploy delete` could provide developers with a way to clean up any existing deployment of the application. 

### High-level design:

#### Pre-requisites:
- The implementation would be under the experimental flag.
- Only supported for Devfile v2.0.0 components.
- Only supported for Kubernetes/OpenShift targets.

#### odo deploy 
This command will build a container image for the application and deploy it on the target Kubernetes environment. 

Flags:
 - `--tag`: The tag to be used for the built application container image - `<registry>/<org>/<name>:<tag>` (optional)
    - Openshift: the tag can be generated based on project details and internal image registry can be used
    - Kubernetes: the user must provide the fully qualified registry and tag details.
 - `--credentials`: The credentials needed to push the image to the container image registry (optional)
    - Openshift: default builder account for BuildConfig is used
    - Kubernetes: the user must provide the credentials
 - `--force`: Delete the existing deployment (if present) and re-create a new deployment.

#### odo deploy delete
This command will delete any resources created by odo deploy.

### Detailed design: 

### Devfile

For the initial implementation, we could use devfile v2.0.0 and capture basic outer-loop information as `metadata` on the devfile. `odo` could look for  specific keys, while other tools like Che could ignore them.

For example: 
```
metadata:
    alpha.build-dockerfile: <URI>
    alpha.deployment-manifest: <URI>
```

### Dockerfile
This could be any valid dockerfile.

### Deployment manifest
The deployment manifest could be templated to help with replacing key bits of information:
- {{.COMPONENT_NAME}}
- {{.CONTAINER_IMAGE}}
- {{.PORT}}

Examples: 
- Standard Kubernetes resources (deployment, service, route): https://github.com/groeges/devfile-registry/blob/master/devfiles/nodejs/deploy_deployment.yaml

- Operator based resources e.g. Runtime component operator (https://operatorhub.io/operator/): https://github.com/groeges/devfile-registry/blob/master/devfiles/nodejs/deploy_runtimecomponent.yaml

- Knative service: https://github.com/groeges/devfile-registry/blob/master/devfiles/nodejs/deploy_knative.yaml

### odo deploy
This command will perform the following actions:

#### Input validation
- Check if the devfile  is v2.0.0 and that it specifies the expected outer-loop metadata. 
    - If not provided, display a meaningful error message.
- Validate all arguments passed by the user. 
    - If argument values are invalid, display a meaningful error message.

#### Build
- If the cluster supports `BuildConfig`, use it:
    - Use BuildConfig to build and push a container image by using:
        - the source code in the user's workspace
        - dockerfile specified by the devfile
        - tag generated based on component and internal registry details
- If the cluster does not support `BuildConfig`, switch to Kaniko:
    - Use Kaniko to build and push a container image by using:
        - the source code in the user's workspace
        - dockerfile specified by the devfile
        - tag provided by the user
        - credentials provided by the user
- Cleanup all build resources created above. 

#### Deploy
- Delete existing deployment, (if invoked with --force flag)
- Fetch the deployment manifest using URI in the metadata of the devfile.
- Replace templated text in the deployment manifest with relevant values:
    - {{.COMPONENT_NAME}}: name of odo component/microservice
    - {{.CONTAINER_IMAGE}}: `tag` for the built image
    - {{.PORT}}: URL information in env.yaml
- Apply the new deployment manifest create/update the application deployment.
- Save the deployment manifest in `.odo/env/` folder.
- Provide the user with a URL for accessing the deployed application.

### odo deploy delete
This command will perform the following actions:
- Check if there is an existing deployment for the app (can use the saved deployment manifest in `.odo/env/` folder)
    - If found, delete the resources specified in the deployment manifest.
    - If not found, show a meaningful error message to the user. 

## Future evolution

- Devfile 2.1.0 should broaden the scope for the outer-loop support in devfiles. For example:
    - Support multiple build strategies - buildah, s2i, build v3 etc.
    - Support multiple deployment strategies - k8s manifest, native service, pod spec etc.
    - Any referenced assets should be immutable to ensure reproducible builds/deployments.

- If a devfile does not provide deployment manifest, odo can perhaps create a manifest in the way it does for inner-loop. This will mean devfile creators do not need to provide a deployment manifest if they do not care so much about the deployment aspect.

- Once `odo link` and service binding is supported by odo and devfiles v2, we could use the same service binding information for `odo deploy`.