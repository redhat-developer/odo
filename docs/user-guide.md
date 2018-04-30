---
layout: default
permalink: /user-guide/
redirect_from: 
  - /docs/user-guide.md/
---

# User Guide

* TOC
{:toc}

## Kedge Apply

Deploy directly to Kubernetes, either updating or creating the artifacts if they don't exist.

```sh
$ kedge apply -f httpd.yaml
deployment "httpd" created
service "httpd" created
```

## Kedge Create

Deploy directly to Kubernetes without creating the artifacts. Internally, Kedge will generate the artifacts and then create it using the `kubectl` command.

__Deploy to Kubernetes:__

```sh
$ kedge create -f httpd.yaml
deployment "httpd" created
service "httpd" created
```

## Kedge Generate

Generate Kubernetes artifacts based upon your Kedge YAML file, see our [examples](/examples) or the [file reference](/docs/file-reference.md) on how to create said file.

In these examples, we use the [simplest of examples](/examples/simplest/httpd.yaml)

__Convert to Kubernetes artifacts:__

```sh
$ kedge generate -f httpd.yaml
---
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  creationTimestamp: null
  labels:
    app: httpd
  name: httpd
spec:
  strategy: {}
  template:
    metadata:
      creationTimestamp: null
      labels:
        app: httpd
      name: httpd
    spec:
      containers:
      - image: centos/httpd
        name: httpd
        resources: {}
status: {}
---
apiVersion: v1
kind: Service
metadata:
  creationTimestamp: null
  labels:
    app: httpd
  name: httpd
spec:
  ports:
  - port: 8080
    targetPort: 80
  selector:
    app: httpd
  type: NodePort
status:
  loadBalancer: {}
```

__Generate and deploy directly to Kubernetes:__

Generation commands can also be "piped" to Kubernetes

```sh
$ kedge generate -f httpd.yaml | kubectl create -f -
deployment "httpd" created
service "httpd" created
```

## Kedge Delete

Deletes Kubernetes artifacts

```sh
$ kedge delete -f httpd.yaml
deployment "httpd" deleted
service "httpd" deleted
```
## Kedge Version

Outputs the current Kedge version

__Version:__

```sh
$ kedge version
```

## Kedge Init

__Getting started:__

```bash
$ kedge init --name web --image centos/httpd --ports 80
```
This will create a `kedge.yml` file for an `httpd` web server with container
image `centos/httpd` exposed on port 80.

__Create a different file:__

```bash
$ kedge init --out myapp.yml --name web --image centos/httpd --ports 80
```

This will create a different file named `myapp.yml` as opposed to the default
`kedge.yml`.

__Create a different controller type:__

By default the controller type that is created is [Kubernetes Deployments](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/),
which is meant for long running processes. But if you want to create a controller
type which is [Kubernetes Job](https://kubernetes.io/docs/concepts/workloads/controllers/jobs-run-to-completion/),
you can do that using the flag `--controller`.

```bash
$ kedge init --name myjob --image jobimage --controller Job
```

## Kedge Build

__Build:__

Build container image with image name `username/myapp:version`

```console
$ kedge build -i username/myapp:version
``` 

Here you might wanna replace the `username` with container image registry URL and then
`username`.


__Build & Push:__

```console
$ kedge build -i username/myapp:version -p
``` 

If you want to build image and also push it to the registry then use the flag `-p`.

**Note**: You should have access to push image to that container registry. Read more about
`docker login` on official [docs](https://docs.docker.com/engine/reference/commandline/login/).

__Dockerfile and context are different:__

If you have a file structure like this, where your `Dockerfile` resides in a directory and
your code context is different.

```console
$ tree
.
├── main.go
├── parsego.go
├── scripts
│   ├── Dockerfile
│   ├── entrypoint.sh
│   └── k8s-release
...
```

To do builds in environment like above, run following command:

```console
$ kedge build -i surajd/json-schema -c . -f scripts/Dockerfile 
INFO[0000] Building image 'surajd/json-schema' from directory 'json-schema-generator' 
INFO[0001] Image 'surajd/json-schema' from directory 'json-schema-generator' built successfully
```

__Build in minikube/minishift:__

If you are running Kubernetes in a environment like minikube or minishift run following
command before running this build command:

```console
$ eval $(minikube docker-env)
```

### Build using source to image

You can hand off the image building work to OpenShift's source to image utility, without having
to write any Dockerfile. 

To do build using OpensShift's s2i run following command:

```console
$ kedge build --s2i --image pyappth -b centos/python-35-centos7:3.5
```

Deconstructing above command, we can see that to enable using OpenShift's s2i use the boolean
flag `--s2i` then give the output name of the ImageStream that will be created using flag
`--image` and finally also provide what builder image to use to do builds of the code using
flag `-b`.

**Note**: Flag `-b` is valid only when using s2i mode. And build flag `-p` is not valid when
in s2i mode.
