---
title: Using Helm with odo
author: Parthvi Vala
author_url: https://github.com/valaparthvi
author_image_url: https://github.com/valaparthvi.png
tags: []
slug: using-helm-with-odo
---

This blog will show how odo can now be used with tools such as Helm, Kustomize, etc. for the [outerloop](/docs/introduction/#what-is-inner-loop-and-outer-loop) development cycle.

:::note

This blog is an extension of [an earlier blog](./2022-06-30-binding-database-service-without-sbo.md) which focuses on the [innerloop](/docs/introduction/#what-is-inner-loop-and-outer-loop) development cycle.
:::

By the end of this blog, we will have deployed a CRUD REST mongodb application on a minikube cluster.

## Prerequisites:
1. [`odo` v3.8.0](https://github.com/redhat-developer/odo/releases/tag/v3.8.0)+
2. [Minikube cluster](https://minikube.sigs.k8s.io/docs/start/)
3. [Ingress enabled on the minikube cluster](https://kubernetes.io/docs/tasks/access-application-cluster/ingress-minikube/)


## 1. Fetch the project
```shell
git clone https://github.com/valaparthvi/restapi-mongodb-odo.git &&  cd restapi-mongodb-odo
```

## 2. Create namespace
Create a namespace called `restapi-mongodb`:

```shell
odo create namespace restapi-mongodb
```

<details>
<summary>Sample output:</summary>

```shell
$ odo create namespace
 ✓  Namespace "restapi-mongodb" is ready for use
 ✓  New namespace created and now using namespace: restapi-mongodb 
```
</details>

## 3. Initialize the component
Download the devfile to initialize an `odo` component with `odo init`.
```shell
odo init --devfile go --name places
```

<details>
<summary>Sample output:</summary>

```shell
$ odo init --devfile go --name places
  __
 /  \__     Initializing a new component
 \__/  \
 /  \__/    odo version: v3.9.0
 \__/

 ✓  Downloading devfile "go" [3s]

Your new component 'places' is ready in the current directory.
To start editing your component, use 'odo dev' and open this folder in your favorite IDE.
Changes will be directly reflected on the cluster.
```
</details>

## 4. Modify the Devfile

We will be using `odo deploy` for deploying our application, and for this we need to modify the Devfile by adding required commands and components.

### Add the `commands`
Let us begin by first adding a `deploy` command under the `commands` section.
```yaml
# This is the main "composite" command that will run all below commands
- id: deploy
  composite:
    commands:
      - k8s-serviceaccount-for-helm
      - k8s-role-for-helm
      - k8s-rolebinding-for-helm
      - deploy-db
      - build-image
      - k8s-deployment
      - k8s-service
      - k8s-url
    group:
      isDefault: true
      kind: deploy
```

`deploy` command is a composition of various other commands in the order in which we want them to be executed.
For e.g. before deploying the database with helm, we need to ensure a service account with the required permissions (made possible by role and rolebinding) has been created;
and so we run `k8s-serviceaccount-for-helm`, `k8s-role-for-helm` and `k8s-rolebinding-for-helm` before running `deploy-db` command.

Let us now add the individual commands.

We will first define the `deploy-db` command that is used to deploy the helm chart.

To use an external tool such as helm or kustomize, we need to ensure 2 things:
1. use an `exec` command; learn more [here](/docs/development/devfile#how-odo-runs-exec-commandsin-deploy-mode).
2. the container component referenced by this command uses an image that contains the required binary.

```yaml
- id: deploy-db
  exec:
    commandLine: helm repo add bitnami https://charts.bitnami.com/bitnami && helm repo update && helm install mongodb bitnami/mongodb
    component: deploy-db
```

We will now add the remaining commands.

```yaml
- id: k8s-serviceaccount-for-helm
  apply:
    component: outerloop-serviceaccount
- id: k8s-role-for-helm
  apply:
    component: outerloop-role
- id: k8s-rolebinding-for-helm
  apply:
    component: outerloop-rolebinding
- id: build-image
  apply:
    component: outerloop-build
- id: k8s-deployment
  apply:
    component: outerloop-deployment
- id: k8s-service
  apply:
    component: outerloop-service
- id: k8s-url
  apply:
      component: outerloop-url
```

### Add the `components`
Every command above references a `component`, and so we now add components under the `components` section.

We will first add the component referenced by `deploy-db` command.

```yaml
- name: deploy-db
  container:
    image: quay.io/tkral/devbox-demo-devbox
  attributes:
    pod-overrides:
      spec:
        serviceAccountName: my-go-app
```

The image used by this container component contains the Helm binary that we can use to deploy the helm chart.

The component is using a `pod-overrides` attribute that will override the service account used by the pod to deploy the helm chart to use the service account (`my-go-app`) we define in this Devfile.
If we do not do this, the pod will use the `default` service account that does not have the required permissions.

We will now add the remaining components.

```yaml
- name: outerloop-serviceaccount
  kubernetes:
    inlined: |
      apiVersion: v1
      kind: ServiceAccount
      metadata:
        name: {{RESOURCE_NAME}}
- name: outerloop-role
  kubernetes:
    inlined: |
      apiVersion: rbac.authorization.k8s.io/v1
      kind: Role
      metadata:
        name: {{RESOURCE_NAME}}
      rules:
      - apiGroups:
        - '*'
        resources:
        - '*'
        verbs:
        - '*'
- name: outerloop-rolebinding
  kubernetes:
    inlined: |
      apiVersion: rbac.authorization.k8s.io/v1
      kind: RoleBinding
      metadata:
        name: {{RESOURCE_NAME}}
      roleRef:
        apiGroup: rbac.authorization.k8s.io
        kind: Role
        name: {{RESOURCE_NAME}}
      subjects:
      - kind: ServiceAccount
        name: {{RESOURCE_NAME}}
# This will build the container image before deployment
- name: outerloop-build
  image:
    dockerfile:
      buildContext: ${PROJECT_SOURCE}
      rootRequired: false
      uri: ./Dockerfile
    imageName: "{{CONTAINER_IMAGE}}"
# This will create a Deployment in order to run your container image across the cluster.
# Note that we expose the env vars necessary to connect application with the mongodb service.
- name: outerloop-deployment
  kubernetes:
    inlined: |
      kind: Deployment
      apiVersion: apps/v1
      metadata:
        name: {{RESOURCE_NAME}}
      spec:
        replicas: 1
        selector:
          matchLabels:
            app: {{RESOURCE_NAME}}
        template:
          metadata:
            labels:
              app: {{RESOURCE_NAME}}
          spec:
            containers:
              - name: {{RESOURCE_NAME}}
 image: {{CONTAINER_IMAGE}}
 ports:
   - name: http
 containerPort: {{CONTAINER_PORT}}
 protocol: TCP
 env:
 - name: username
   value: {{USERNAME}}
 - name: host
   value: {{HOST}}
 - name: password
   valueFrom:
 secretKeyRef:
 name: mongodb
 key: mongodb-root-password
 resources:
   limits:
 memory: "1024Mi"
 cpu: "500m"
# This will create a Service so your Deployment is accessible.
# Depending on your cluster, you may modify this code so it's a
# NodePort, ClusterIP or a LoadBalancer service.
- name: outerloop-service
  kubernetes:
    inlined: |
      apiVersion: v1
      kind: Service
      metadata:
        name: {{RESOURCE_NAME}}
      spec:
        ports:
        - name: "{{CONTAINER_PORT}}"
          port: {{CONTAINER_PORT}}
          protocol: TCP
          targetPort: {{CONTAINER_PORT}}
        selector:
          app: {{RESOURCE_NAME}}
        type: NodePort
- name: outerloop-url
  kubernetes:
    inlined: |
      apiVersion: networking.k8s.io/v1
      kind: Ingress
      metadata:
        name: {{RESOURCE_NAME}}
      spec:
        rules:
          - host: "{{DOMAIN_NAME}}"
            http:
              paths:
 - path: "/"
   pathType: Prefix
   backend:
 service:
 name: {{RESOURCE_NAME}}
 port:
   number: {{CONTAINER_PORT}}
```

### Add the `variables`
Next, we add a `variables` section to the Devfile, so that we can make use of the same variables at multiple locations within the Devfile.

```yaml
# Add the following variables code anywhere in devfile.yaml
# This MUST be a container registry you are able to access
variables:
  CONTAINER_IMAGE: quay.io/MYUSERNAME/go-odo-example
  RESOURCE_NAME: my-go-app
  CONTAINER_PORT: "8080"
  DOMAIN_NAME: go.example.com
  USERNAME: root
  HOST: mongodb
```

:::note
Ensure that you replace `MYUSERNAME` in `CONTAINER_IMAGE` with your own username; or use a container registry that you have the write permissions to.

If you are using quay.io registry, you might have to change the repository permissions to Public to continue, otherwise you might see failures related to pulling the image.
:::

### Modify `schemaVersion`
One last thing is to change the `schemaVersion` of the Devfile since `deploy` commands are only supported in schema 2.2.0+.
```yaml
# Deploy "kind" ID's use schema 2.2.0+
schemaVersion: 2.2.0
```


<details>
<summary>Your final Devfile will look like the following:</summary>

```yaml showLineNumbers
commands:
- exec:
    commandLine: go build main.go
    component: runtime
    env:
      - name: GOPATH
        value: ${PROJECT_SOURCE}/.go
      - name: GOCACHE
        value: ${PROJECT_SOURCE}/.cache
    group:
      isDefault: true
      kind: build
    workingDir: ${PROJECT_SOURCE}
  id: build
- exec:
    commandLine: ./main
    component: runtime
    group:
      isDefault: true
      kind: run
    workingDir: ${PROJECT_SOURCE}
  id: run
  # highlight-start
- id: deploy-db
  exec:
    commandLine: helm repo add bitnami https://charts.bitnami.com/bitnami && helm repo update && helm install mongodb bitnami/mongodb
    component: deploy-db
- id: k8s-serviceaccount-for-helm
  apply:
    component: outerloop-serviceaccount
- id: k8s-role-for-helm
  apply:
    component: outerloop-role
- id: k8s-rolebinding-for-helm
  apply:
    component: outerloop-rolebinding

# This is the main "composite" command that will run all below commands
- id: deploy
  composite:
    commands:
      - k8s-sa
      - k8s-role
      - k8s-rolebinding
      - deploy-db
      - build-image
      - k8s-deployment
      - k8s-service
      - k8s-url
    group:
      isDefault: true
      kind: deploy
# Below are the commands and their respective components that they are "linked" to deploy
- id: build-image
  apply:
    component: outerloop-build
- id: k8s-deployment
  apply:
    component: outerloop-deployment
- id: k8s-service
  apply:
    component: outerloop-service
- id: k8s-url
  apply:
    component: outerloop-url
# highlight-end
components:
- container:
    args:
      - tail
      - -f
      - /dev/null
    endpoints:
      - name: http-go
        targetPort: 8080
    image: registry.access.redhat.com/ubi9/go-toolset:latest
    memoryLimit: 1024Mi
    mountSources: true
  name: runtime
# highlight-start
- name: deploy-db
  container:
    image: quay.io/tkral/devbox-demo-devbox
  attributes:
    pod-overrides:
      spec:
        serviceAccountName: my-go-app
- name: outerloop-serviceaccount
  kubernetes:
    inlined: |
      apiVersion: v1
      kind: ServiceAccount
      metadata:
        name: {{RESOURCE_NAME}}
- name: outerloop-role
  kubernetes:
    inlined: |
      apiVersion: rbac.authorization.k8s.io/v1
      kind: Role
      metadata:
        name: {{RESOURCE_NAME}}
      rules:
      - apiGroups:
        - '*'
        resources:
        - '*'
        verbs:
        - '*'
- name: outerloop-rolebinding
  kubernetes:
    inlined: |
      apiVersion: rbac.authorization.k8s.io/v1
      kind: RoleBinding
      metadata:
        name: {{RESOURCE_NAME}}
      roleRef:
        apiGroup: rbac.authorization.k8s.io
        kind: Role
        name: {{RESOURCE_NAME}}
      subjects:
      - kind: ServiceAccount
        name: {{RESOURCE_NAME}}
# This will build the container image before deployment
- name: outerloop-build
  image:
    dockerfile:
      buildContext: ${PROJECT_SOURCE}
      rootRequired: false
      uri: ./Dockerfile
    imageName: "{{CONTAINER_IMAGE}}"
# This will create a Deployment in order to run your container image across
# the cluster.
- name: outerloop-deployment
  kubernetes:
    inlined: |
      kind: Deployment
      apiVersion: apps/v1
      metadata:
        name: {{RESOURCE_NAME}}
      spec:
        replicas: 1
        selector:
          matchLabels:
            app: {{RESOURCE_NAME}}
        template:
          metadata:
            labels:
              app: {{RESOURCE_NAME}}
          spec:
            containers:
              - name: {{RESOURCE_NAME}}
 image: {{CONTAINER_IMAGE}}
 ports:
   - name: http
 containerPort: {{CONTAINER_PORT}}
 protocol: TCP
 env:
 - name: username
   value: {{USERNAME}}
 - name: host
   value: {{HOST}}
 - name: password
   valueFrom:
 secretKeyRef:
 name: mongodb
 key: mongodb-root-password
 resources:
   limits:
 memory: "1024Mi"
 cpu: "500m"
# This will create a Service so your Deployment is accessible.
# Depending on your cluster, you may modify this code so it's a
# NodePort, ClusterIP or a LoadBalancer service.
- name: outerloop-service
  kubernetes:
    inlined: |
      apiVersion: v1
      kind: Service
      metadata:
        name: {{RESOURCE_NAME}}
      spec:
        ports:
        - name: "{{CONTAINER_PORT}}"
          port: {{CONTAINER_PORT}}
          protocol: TCP
          targetPort: {{CONTAINER_PORT}}
        selector:
          app: {{RESOURCE_NAME}}
        type: NodePort
- name: outerloop-url
  kubernetes:
    inlined: |
      apiVersion: networking.k8s.io/v1
      kind: Ingress
      metadata:
        name: {{RESOURCE_NAME}}
      spec:
        rules:
          - host: "{{DOMAIN_NAME}}"
            http:
              paths:
 - path: "/"
   pathType: Prefix
   backend:
 service:
 name: {{RESOURCE_NAME}}
 port:
   number: {{CONTAINER_PORT}}
  # highlight-end
metadata:
  description:
    Go is an open source programming language that makes it easy to build
    simple, reliable, and efficient software.
  displayName: Go Runtime
  icon: https://raw.githubusercontent.com/devfile-samples/devfile-stack-icons/main/golang.svg
  language: Go
  name: places
  projectType: Go
  provider: Red Hat
  tags:
    - Go
  version: 1.0.2
# highlight-start
# Deploy "kind" ID's use schema 2.2.0+
schemaVersion: 2.2.0
# highlight-end
starterProjects:
  - description: A Go project with a simple HTTP server
    git:
      checkoutFrom:
        revision: main
      remotes:
        origin: https://github.com/devfile-samples/devfile-stack-go.git
    name: go-starter
# highlight-start
# Add the following variables code anywhere in devfile.yaml
# This MUST be a container registry you are able to access
variables:
  CONTAINER_IMAGE: quay.io/MYUSERNAME/go-odo-example
  RESOURCE_NAME: my-go-app
  CONTAINER_PORT: "8080"
  DOMAIN_NAME: go.example.com
  USERNAME: root
  HOST: mongodb
# highlight-end
```
</details>

## 5. Deploy
Now that the Devfile is ready, we can simply run `odo deploy`.
```shell
odo deploy
```

<details>
<summary>Sample output</summary>

```shell
$ odo deploy
 __
 /  \__     Running the application in Deploy mode using my-go-app Devfile
 \__/  \    Namespace: restapi-mongodb
 /  \__/    odo version: v3.9.0
 \__/

↪ Deploying Kubernetes Component: my-go-app
 ✓  Creating resource ServiceAccount/my-go-app 

↪ Deploying Kubernetes Component: my-go-app
 ✓  Creating resource Role/my-go-app 

↪ Deploying Kubernetes Component: my-go-app
 ✓  Creating resource RoleBinding/my-go-app 

↪ Executing command:
 ✓  Executing command in container (command: deploy-db) [18s]

↪ Building & Pushing Image: quay.io/pvala18/go-odo-example
 •  Building image locally
[1/2] STEP 1/5: FROM quay.io/redhat-developer/servicebinding-operator:builder-golang-1.16 AS builder  
[1/2] STEP 2/5: USER root   
--> Using cache 6caf9a75a7e8a27da9ebbddc7a4c7451033e53e588796c65e5a7683049927992
--> 6caf9a75a7e
[1/2] STEP 3/5: WORKDIR /workspace
--> Using cache 979b426e92aba26a7ec4c66b698516b35b164ea34c3c77f8b3ee52999009958a
--> 979b426e92a
[1/2] STEP 4/5: COPY / /workspace/
--> 28e87dd1e60
[1/2] STEP 5/5: RUN go build
go: downloading github.com/sirupsen/logrus v1.8.1
go: downloading github.com/spf13/viper v1.11.0   
go: downloading github.com/gorilla/mux v1.8.0    
go: downloading go.mongodb.org/mongo-driver v1.9.0       
go: downloading golang.org/x/sys v0.0.0-20220412211240-33da011f77ad             
go: downloading github.com/mitchellh/mapstructure v1.4.3 
go: downloading github.com/fsnotify/fsnotify v1.5.1      
go: downloading github.com/spf13/afero v1.8.2    
go: downloading github.com/spf13/cast v1.4.1     
go: downloading github.com/spf13/jwalterweatherman v1.1.0
go: downloading github.com/spf13/pflag v1.0.5 
go: downloading github.com/spf13/afero v1.8.2
go: downloading github.com/spf13/cast v1.4.1     
go: downloading github.com/spf13/jwalterweatherman v1.1.0
go: downloading github.com/spf13/pflag v1.0.5    
go: downloading golang.org/x/text v0.3.7         
go: downloading github.com/subosito/gotenv v1.2.0
go: downloading github.com/hashicorp/hcl v1.0.0
go: downloading gopkg.in/ini.v1 v1.66.4
go: downloading github.com/magiconair/properties v1.8.6
go: downloading github.com/pelletier/go-toml v1.9.4
go: downloading gopkg.in/yaml.v2 v2.4.0
go: downloading github.com/youmark/pkcs8 v0.0.0-20181117223130-1be2e3e5546d
go: downloading github.com/pkg/errors v0.9.1
go: downloading github.com/go-stack/stack v1.8.0
go: downloading github.com/golang/snappy v0.0.3
go: downloading github.com/klauspost/compress v1.13.6
go: downloading golang.org/x/crypto v0.0.0-20220411220226-7b82a4e95df4
go: downloading golang.org/x/sync v0.0.0-20210220032951-036812b2e83c 
go: downloading github.com/xdg-go/scram v1.0.2
go: downloading github.com/xdg-go/stringprep v1.0.2
go: downloading github.com/xdg-go/pbkdf2 v1.0.0
--> b596b2ffc4c
[2/2] STEP 1/6: FROM registry.access.redhat.com/ubi8-minimal
[2/2] STEP 2/6: WORKDIR /
--> Using cache 2f7a0096cef8c90c20c8091c6f1cea4660f100593c848ae2f2cd0b2283ca7f11
--> 2f7a0096cef
[2/2] STEP 3/6: COPY --from=builder /workspace/go-rest-mongodb .
--> Using cache 91eeb6e34d81e8ddbc5ef8f4b8ff3243400812e52a54a555a85343fad8a4caf9
--> 91eeb6e34d8
[2/2] STEP 4/6: COPY --from=builder /workspace/config.yml .
--> Using cache dab23c83ba939565242774abf7cac92849935a0d99c33609a0e7a16b32f34aeb
--> dab23c83ba9
[2/2] STEP 5/6: USER 65532:65532
--> Using cache 2faf2c79c0f92dfaac4cb88084beecbfa1df555512da42ec4e10db4208518cc6
--> 2faf2c79c0f
[2/2] STEP 6/6: ENTRYPOINT ["/go-rest-mongodb"]
--> Using cache 2271a27b9d4642a2af86ee5836797fc5161f49346b2251b7a6a0cc80c2d3089c
[2/2] COMMIT quay.io/pvala18/go-odo-example
--> 2271a27b9d4
Successfully tagged quay.io/pvala18/go-odo-example:latest
2271a27b9d4642a2af86ee5836797fc5161f49346b2251b7a6a0cc80c2d3089c
 ✓  Building image locally [23s]
 •  Pushing image to container registry  ...
Getting image source signatures
Copying blob 876fba3c71a7 skipped: already exists  
Copying blob 55ea6d5a354e skipped: already exists  
Copying blob a283f9ae821e skipped: already exists  
Copying config 2271a27b9d done  
Writing manifest to image destination
Storing signatures
 ✓  Pushing image to container registry [10s]

↪ Deploying Kubernetes Component: my-go-app
 ✓  Creating resource Deployment/my-go-app 

↪ Deploying Kubernetes Component: my-go-app
 ✓  Creating resource Service/my-go-app 

↪ Deploying Kubernetes Component: my-go-app
 ✓  Creating resource Route/my-go-app 

Your Devfile has been successfully deployed

```
</details>

## 6. Accessing the application

Run `odo describe component` to obtain access information.

```shell
odo describe component
```

<details>
<summary>Sample output</summary>

```shell
$ odo describe component
Name: places
Display Name: Go Runtime
Project Type: Go
Language: Go
Version: 1.0.2
Description: Go (version 1.18.x) is an open source programming language that makes it easy to build simple, reliable, and efficient software.
Tags: Go

Running in: Deploy

Running on:
 •  cluster: Deploy

Supported odo features:
 •  Dev: true
 •  Deploy: true
 •  Debug: false

Container components:
 •  runtime
    Source Mapping: /projects
 •  deploy-db
    Source Mapping: /projects

Kubernetes components:
 •  outerloop-serviceaccount
 •  outerloop-role
 •  outerloop-rolebinding
 •  outerloop-deployment
 •  outerloop-service
 •  outerloop-url

Kubernetes Ingresses:
 •  my-go-app: go.example.com/

```
</details>

Since we are using Ingress, we first need to check if an IP address has been set.

```sh
$ kubectl get ingress my-go-app
NAME        CLASS   HOSTS            ADDRESS          PORTS   AGE
my-go-app   nginx   go.example.com   192.168.59.124   80      7m4s
```

Once the IP address appears, you can now access the application at the following URL:

```
curl --resolve "go.example.com:80:192.168.59.124" -i http://go.example.com/api/places
```

<details>
<summary>Sample output</summary>

```shell
$ curl --resolve "go.example.com:80:192.168.59.124" -i http://go.example.com/api/places
HTTP/1.1 200 OK
Date: Thu, 27 Apr 2023 06:16:09 GMT
Content-Type: application/json
Content-Length: 4
Connection: keep-alive

null
```
</details>

This will return a _null_ response since the database is currently empty, but it also means that we have successfully connected to our database application.

:::note
You can add the following line to the `/etc/hosts` file of your computer to simply access the application at http://go.example.com.
Learn more about [using ingress to access an application](https://kubernetes.io/docs/tasks/access-application-cluster/ingress-minikube/#create-an-ingress).
```
192.168.59.124  go.example.com
```
:::

Add some data to the database:
```sh
curl --resolve "go.example.com:80:192.168.59.124" -i http://go.example.com/api/places -sSL -XPOST -d '{"title": "Agra", "description": "Land of Tajmahal"}'
```

<details>
<summary>Sample Output</summary>

```sh
$ curl --resolve "go.example.com:80:192.168.59.124" -i http://go.example.com/api/places -sSL -XPOST -d '{"title": "Agra", "description": "Land of Tajmahal"}'
HTTP/1.1 201 Created
Date: Thu, 27 Apr 2023 10:43:04 GMT
Content-Type: application/json
Content-Length: 86
Connection: keep-alive

{"id":"62c2a0659fa147e382a4db31","title":"Agra","description":"Land of Tajmahal"}
```
</details>

Fetch the list of places again:
```sh
$ curl --resolve "go.example.com:80:192.168.59.124" -i http://go.example.com/api/places
HTTP/1.1 201 Created
Date: Thu, 27 Apr 2023 10:41:09 GMT
Content-Type: application/json
Content-Length: 81
Connection: keep-alive

{"id":"62c2a0659fa147e382a4db31","title":"Agra","description":"Land of Tajmahal"}
```


### List of available API endpoints
- GET `/api/places` - List all places
- POST `/api/places` - Add a new place
- PUT `/api/places` - Update a place
- GET `/api/places/<id>` - Fetch place with id `<id>`
- DELETE `/api/places/<id>` - Delete place with id `<id>`

