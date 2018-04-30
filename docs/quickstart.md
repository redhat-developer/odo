---
layout: default
permalink: /quickstart/
redirect_from: 
  - /docs/quickstart.md/
---

# Kedge - Concise Application Definitions for Kubernetes

Simplify your Kubernetes deployment by using Kedge. Reduce your technical debt by investing in a simple and concise definition!

Kedge is a deployment tool for Kubernetes artifacts by using a simplified version of the Kubernetes spec (a Kedge formatted YAML file).

In two steps, we will go from a super-simple YAML file to a full-fledged Kubernetes deployment:

__1. Using an example [httpd.yaml](https://raw.githubusercontent.com/kedgeproject/kedge/master/examples/httpd/httpd.yaml) file__

```yaml
name: httpd

containers:
- image: centos/httpd

services:
- name: httpd
  type: LoadBalancer
  portMappings: 
    - 8080:80
```

__2. Now run the apply command to deploy to Kubernetes!__

```sh
$ kedge apply -f httpd.yaml
deployment "httpd" created
service "httpd" created
```

__2.1. Alternatively, you can generate the raw Kubernetes artifact files and deploy it with `kubectl`.__

```sh
$ kedge generate -f httpd.yaml > output.yaml
$ kubectl apply -f output.yaml
```

__View the deployed service__

Now that your service has been deployed, let's access it.

If you're already using `minikube` for your development process:

```sh
$ minikube service httpd
Opening kubernetes service default/httpd in default browser...
```

Otherwise, let's look up what IP your service is using!

```sh
$ kubectl describe svc httpd
Name:                   httpd
Namespace:              default
Labels:                 app=httpd
Selector:               app=httpd
Type:                   NodePort
IP:                     10.0.0.34
Port:                   <unset> 8080/TCP
NodePort:               <unset> 31511/TCP
Endpoints:              172.17.0.4:80
Session Affinity:       None
No events.
```

__Next steps__

That's it! There's more examples in our [repository](https://github.com/kedgeproject/kedge/tree/master/examples). Check out the further documentation such as the [user guide](/docs/user_guide.md) or our [file reference](/docs/file-reference.md).
