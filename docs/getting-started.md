---
layout: default
permalink: /getting-started/
redirect_from: 
  - /docs/getting-started.md/
---

# Getting Started

* TOC
{:toc}

These guide(s) will show you how to get started with your favourite container orchestrator as well as Kedge!

There are three different guides depending on your container orchestrator as well as operating system.

For beginners and the most compatibility, follow the __Minikube and Kedge__ guide.

## Minikube and Kedge

In this guide, we'll deploy a sample kedge file `httpd.yaml` to a Kubernetes cluster.

Requirements:
  - [minikube](https://github.com/kubernetes/minikube)
  - [kedge](https://github.com/kedgeproject/kedge)

__Start `minikube`:__

If you don't already have a Kubernetes cluster running, [minikube](https://github.com/kubernetes/minikube) is the best way to get started.

```sh
$ minikube start
Starting local Kubernetes v1.7.5 cluster...
Starting VM...
Getting VM IP address...
Moving files into cluster...
Setting up certs...
Connecting to cluster...
Setting up kubeconfig...
Starting cluster components...
Kubectl is now configured to use the cluster
```

__Download the [httpd.yaml](https://raw.githubusercontent.com/kedgeproject/kedge/master/examples/httpd/httpd.yaml) example file, or another example from the [examples](https://github.com/kedgeproject/kedge/tree/master/examples) GitHub directory:__

```sh
curl -LO https://raw.githubusercontent.com/kedgeproject/kedge/master/examples/httpd/httpd.yaml
```

__Deploy directly to your Kubernetes cluster with `kedge apply`:__

Run `kedge apply -f httpd.yaml` in the same directory as your example file.

```sh
$ kedge apply -f httpd.yaml 
service "httpd" created
deployment "httpd" created
```

__Access the newly deployed service:__

Now that your service has been deployed, let's access it.

If you're using `minikube` you may access it via the `minikube service` command.

```sh
$ minikube service httpd
Opening kubernetes service default/httpd in default browser...
Created new window in existing browser session.
```

Otherwise, use `kubectl` to see what IP address the service is using:

```sh
$ kubectl describe svc httpd
Name:                   httpd
Namespace:              default
Labels:                 app=httpd
Annotations:            kubectl.kubernetes.io/last-applied-configuration={"apiVersion":"v1","kind":"Service","metadata":{"annotations":{},"creationTimestamp":null,"labels":{"app":"httpd"},"name":"httpd","namespace":"default"...
Selector:               app=httpd
Type:                   LoadBalancer
IP:                     10.0.0.249
Port:                   httpd-8080      8080/TCP
NodePort:               httpd-8080      30692/TCP
Endpoints:              172.17.0.4:80
Session Affinity:       None
Events:                 <none>
```

## Minishift and Kedge

In this guide, we'll deploy an OpenShift-compatible Kedge file to an OpenShift cluster.

Requirements:
  - [minishift](https://github.com/minishift/minishift)
  - [kedge](https://github.com/kedgeproject/kedge)
  - An OpenShift route created

__Note:__ The service will NOT be accessible until you create an OpenShift route with `oc expose`. You must also have a virtualization environment setup. By default, `minishift` uses KVM.

__Start `minishift`:__

[Minishift](https://github.com/minishift/minishift) is a tool that helps run OpenShift locally using a single-node cluster inside of a VM. Similar to [minikube](https://github.com/kubernetes/minikube).

```sh
$ minishift start
Starting local OpenShift cluster using 'kvm' hypervisor...
-- Checking OpenShift client ... OK
-- Checking Docker client ... OK
-- Checking Docker version ... OK
-- Checking for existing OpenShift container ... OK
...
```

__Login as developer within `minishift`:__

Login as developer if you haven't done so already within `minishift`.


```sh
▶ oc login -u developer          
Server [https://localhost:8443]: https://192.168.42.113:8443
The server uses a certificate signed by an unknown authority.
You can bypass the certificate check, but any data you send to the server could be intercepted by others.
Use insecure connections? (y/n): y 

Authentication required for https://192.168.42.113:8443 (openshift)
Username: developer
Password: 
Login successful.

You have one project on this server: "myproject"

Using project "myproject".
Welcome! See 'oc help' to get started.
```

__Download the [Guestbook Demo files](https://github.com/kedgeproject/kedge/tree/master/examples/guestbook-demo) from GitHub:__

Due to OpenShift using non-root containers, we will be using an OpenShift-compatible demo. In particular, the highly-used "Guestbook" Kubernetes demo.

```sh
curl -LO https://raw.githubusercontent.com/kedgeproject/kedge/master/examples/guestbook-demo/backend.yaml
curl -LO https://raw.githubusercontent.com/kedgeproject/kedge/master/examples/guestbook-demo/frontend.yaml
curl -LO https://raw.githubusercontent.com/kedgeproject/kedge/master/examples/guestbook-demo/db.yaml
```

__Deploy directly to your OpenShift cluster with `kedge apply`:__

Run `kedge apply -f backend.yaml -f frontend.yaml -f db.yaml` in the same directory as your example files.

```sh
$ kedge apply -f backend.yaml -f frontend.yaml -f db.yaml
service "guestbook" created
deployment "guestbook" created
service "backend" created
deployment "backend" created
persistentvolumeclaim "mongodb-data" created
service "database" created
secret "mongodb-admin" created
secret "mongodb-user" created
configmap "mongodb-user" created
deployment "database" created
```

__Access the newly deployed service:__

After deployment, you must create an OpenShift route in order to access the service.

If you're using `minishift`, you'll use a combination of `oc` and `minishift` commands to access the service.

Create a route for the `frontend` service using `oc`:

```sh
$ oc expose service frontend
route "frontend" exposed
```

Access the `frontend` service with `minishift`:

```sh
$ minishift openshift service frontend --namespace=myproject
Opening the service myproject/frontend in the default browser...
```

You can also access the GUI interface of OpenShift for an overview of the deployed containers:

```sh
$ minishift console
Opening the OpenShift Web console in the default browser...
```
