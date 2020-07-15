# Switching from inner-loop to outer-loop

## Definitions



### Inner loop
The Inner loop stage is the development stage where the developer creates an application and it goes through the interation of coding, debugging, testing, locally on the workstation.

### Outer loop validation
The Outer loop validation stage is the development stage where the developer creates a platform-native build artifact ready to be deployed and distributed on Kubernetes. In case of the Kubernetes platform, a developer would build one or more images from source and deploy them as a Kubernetes Deployment, OpenShift DeploymentConfig or a Knative Service.

### Outer loop Gitops
The outer loop Gitops stage is the development stage where the developer has validated outerloop by having a built image deployed, and now wishes to deploy the app across multiple environments/stages using CI/CD


## End-to-end user experience


> As a user of a tool which understands devfiles, 
I wish to build one or more images from my source code and deploy them on a Kubernetes cluster using the outer-loop guidance from a relevant devfile.


The user validates her code in the hot-reload mode in the cluster using: 

```
odo push
```

The user validates the Kubernetes-style image-based deployment 

```
odo deploy
```

The user chooses to generate deployment manifests, which she would be using to manage Git-driven deployments across multiple stages using CI/CD pipelines


```
odo pipelines bootstrap
```

## Outer-loop guidance in the devfile

### Build guidance

In the context of outer-loop development, unless otherwise specified, "Build guidance" refers to the guidance a developer or a runtime stack author would provide to `odo` for building an image using commonly-known image build strategies like 
* Dockerfile, 
* Buildpacks-v3 and 
* source-to-image.

### Deploy guidance

In the context of outer-loop development, unless otherwise specified, "Deploy guidance" refers to the guidance a developer or runtime stack author would provide to `odo` for deploying the built image using

* Helm charts
* Kubernetes manifests, etc.



## Usage of Build and Deploy guidance in CI/CD and Gitops


### Summary 


`odo push` syncs user's source code to a workspace container  run using the guidance from a runtime-specific `devfile`.

Example, a Sprintboot project, when associated with a Springboot specific devfile, provides `odo`  the know-how to run the inner-loop container with workspace dependencies
like debug tools and SDKs.

`odo deploy` would read guidance from the devfile on how to build or more images and deploy the same. The devfile would either be a registry-provided devfile based on the user's choice of runtime or a project-specific devfile that was present at the root of the project.

`odo pipelines` would generate configuration and manifests locally on the filesystem to eventually manage the application component from Git using CI/CD pipelines. The manifests are to be generated based on the build and deploy guidance from the devfile.

### The outerloop and Gitops story


In the initial iterations of supporting outerloop, the `odo pipelines` command would be used to generate Tekton and Argo manifests for CI and CD, respectively, for enabling users to drive Gitops-based workflows, organized in an opinionated directory structure in Git.

To enable generation of runtime-specific near-accurate pipelines, `odo` would read the image build and deploy guidance from the devfile, the same way `odo deploy` would 
do an outerloop build and image deployment on the cluster using guidance from the devfile.

#### Build
As an example, for a devfile associated with an application component, a dockerfile-based image build guidance in the `devfile` would be consumed by the `odo pipelines` command to
* generate [Tekton Task yaml manifests](https://github.com/tektoncd/catalog/blob/master/task/buildah/0.1/buildah.yaml) which does a Dockerfile build, or 
* generate an equivalent [k8s-build manifest](https://github.com/redhat-developer/build/blob/master/samples/build/build_buildah_cr.yaml), orchestrated by Tekton Pipelines.
* directly use a Tekton Task referenced in the `devfile`.

#### Deploy
Similarly, the `odo pipelines` command would generate 
* The necessary 'yaml' manifests to deploy the built image using the guidance from the devfile.
* The Argo CRs that represent apps that would be reconciled by Argo on a given cluster.

The generated manifests may optionally be consumed outside the context of `odo` using plain old `kubectl` and `kustomize`. Overtime, the user may choose to modify them and commit them
to Git.

## Enhancements to the devfile

The build and deploy guidances would primarily feature as new devfile components in `devfile` v2.1.0.

Example, the current proposal to add a devfile component called `dockerfile` looks like: 


```
schemaVersion: 2.1.0
metadata:
  name: nodejs
  version: 1.0.0
projects:
  - name: nodejs-starter
    git:
      location: "https://github.com/odo-devfiles/nodejs-ex.git"
components:
  - container:
      name: runtime
      image: registry.access.redhat.com/ubi8/nodejs-12:1-36
      memoryLimit: 1024Mi
      mountSources: true
      endpoints:
        - name: http-3000
          targetPort: 3000
          configuration:
            protocol: tcp
            scheme: http
            type: terminal

  - dockerfile:
      name: dockerfile-build
      source: 
         sourceDir: "src"
         location: "https://github.com/ranakan19/golang-ex.git"
      dockerfilePath: "https://raw.githubusercontent.com/wtam2018/test/master/nodejs-dockerfiile"
      destination: 
```

The full list of proposed enhancements could be found [here](https://github.com/devfile/kubernetes-api/issues/49).
