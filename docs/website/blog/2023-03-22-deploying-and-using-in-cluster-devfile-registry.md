---
title: Deploying and using an in-cluster Devfile registry
author: Armel Soro
author_url: https://github.com/rm3l
author_image_url: https://github.com/rm3l.png
tags: ["devfile-registry", "registry", "registry-operator"]
slug: deploying-and-using-in-cluster-devfile-registry
---

<div>
<img
src={require('../static/img/devfile.png').default}
alt="in-cluster Devfile registries with odo"
style={{display: 'block', marginLeft: 'auto', marginRight: 'auto', marginBottom: '10px'}}
/>
</div>

Starting with [v3.8.0](2023-03-08-odo-v3.8.0.md#detecting-in-cluster-devfile-registries),
`odo` can detect Devfile registries declared into the current cluster and use them preferably.

In this article, we'll explore how we can deploy our own Devfile Registry into a Kubernetes or OpenShift cluster and how we can use it automatically with `odo`.

<!--truncate-->

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

## Prerequisites

- [odo](/docs/overview/installation) [3.8.0](https://github.com/redhat-developer/odo/releases/tag/v3.8.0) or later
- A Kubernetes cluster with an Ingress Controller (like [ingress-nginx](https://github.com/kubernetes/ingress-nginx) or [Traefik](https://doc.traefik.io/traefik/providers/kubernetes-ingress/)) or an OpenShift cluster
- A user in the cluster with permission to install Custom Resource Definitions (CRDs). Or ask a cluster administrator to install [those resources](#installing-the-devfile-registry-operator-custom-resource-definitions)
- [`kubectl`](https://kubernetes.io/docs/tasks/tools/#kubectl) or [`oc`](https://docs.openshift.com/container-platform/4.12/cli_reference/openshift_cli/getting-started-cli.html) CLIs
- [`Helm`](https://helm.sh/) CLI, version 3 or higher

## Deploying a Devfile Registry in the cluster

In this section, we'll leverage Helm to install a Devfile Registry.

The [Devfile Registry Chart](https://github.com/devfile/registry-support/tree/main/deploy/chart/devfile-registry) currently deploys an [Ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/) resource on Kubernetes,
hence the requirement of having an Ingress Controller and a reachable Ingress domain.
On OpenShift, it will deploy a [Route](https://docs.openshift.com/container-platform/4.12/networking/routes/route-configuration.html), which will provide an automatic public HTTP URL for accessing the registry.

1. Clone the `registry-support` repository containing the Helm Chart we will deploy:

```shell
git clone --depth=1 https://github.com/devfile/registry-support
```

2. Install the [Helm Chart](https://github.com/devfile/registry-support/tree/main/deploy/chart/devfile-registry) into the current cluster.

<Tabs groupId="devfile-registry-helm">
  <TabItem value="kubernetes" label="Kubernetes">

An Ingress Controller should have been installed with a domain for Ingress resources.

```console
helm install my-devfile-registry \
    ./registry-support/deploy/chart/devfile-registry \
    --set global.ingress.domain=<domain> \
    --set global.ingress.class=<ingress-class>
```

<details>
<summary>Example output:</summary>

```shell
$ helm install my-devfile-registry \
    ./registry-support/deploy/chart/devfile-registry \
    --set global.ingress.domain=$(minikube ip).nip.io \
    --set global.ingress.class=nginx

NAME: my-devfile-registry
LAST DEPLOYED: Fri Mar 24 15:50:18 2023
NAMESPACE: default
STATUS: deployed
REVISION: 1
TEST SUITE: None

```
</details>

  </TabItem>
  <TabItem value="openshift" label="OpenShift">

```console
helm install my-devfile-registry \
    ./registry-support/deploy/chart/devfile-registry \
    --set global.isOpenShift=true
```

<details>
<summary>Example output:</summary>

```shell
$ helm install my-devfile-registry \
    ./registry-support/deploy/chart/devfile-registry \
    --set global.isOpenShift=true    
         
NAME: my-devfile-registry
LAST DEPLOYED: Fri Mar 24 15:54:42 2023
NAMESPACE: default
STATUS: deployed
REVISION: 1
TEST SUITE: None

```
</details>
</TabItem>
</Tabs>


## Determining the Devfile Registry URL

<Tabs groupId="devfile-registry-helm">
  <TabItem value="kubernetes" label="Kubernetes">

On Kubernetes, the Helm Chart installed above will create an Ingress reachable via a DNS domain named as follows: `devfile-registry-<namespace>.<ingressDomain>`.

You can get the actual Host by getting the Ingress Resource, via the following command: `kubectl get ingress <registryName>`.

<details>
<summary>Example output:</summary>

```shell
$ kubectl get ingress my-devfile-registry \
    -o custom-columns='URL:.spec.rules[*].host' \
    --no-headers

devfile-registry-default.172.20.0.2.nip.io
```

In this example, the Devfile Registry is now reachable at http://devfile-registry-default.172.20.0.2.nip.io

</details>
  </TabItem>
  <TabItem value="openshift" label="OpenShift">

On OpenShift, the Helm Chart installed above will create a Route with a URL generated automatically.

You can get the actual URL by getting the Route Resource, via the following command: `oc get route <registryName>`.

<details>
<summary>Example output:</summary>

```shell
$ oc get route my-devfile-registry \
    -o custom-columns='URL:.spec.host' \
    --no-headers

my-devfile-registry-default.apps.4fa297b23808ddc3612a.hypershift.aws-2.ci.openshift.org
```

In this example, the Devfile Registry is reachable at the following URLs:
- https://my-devfile-registry-default.apps.4fa297b23808ddc3612a.hypershift.aws-2.ci.openshift.org
- http://my-devfile-registry-default.apps.4fa297b23808ddc3612a.hypershift.aws-2.ci.openshift.org

</details>

  </TabItem>
</Tabs>

## Installing the Devfile Registry Operator Custom Resource Definitions

Devfile Registries declared in a `DevfileRegistriesList` or `ClusterDevfileRegistriesList` custom resource are automatically included in the list of registries
that `odo` can use.
To be able to create such resources, we need to install their definitions as Custom Resource Definitions (CRDs) in the cluster.
You can do so by applying the [Kustomize](https://kustomize.io/) project available [here](https://github.com/devfile/registry-operator/config/crd), using the command below:

```
kubectl apply -k https://github.com/devfile/registry-operator/config/crd
```

<details>
<summary>Example output:</summary>

```shell
$ kubectl apply -k https://github.com/devfile/registry-operator/config/crd

customresourcedefinition.apiextensions.k8s.io/clusterdevfileregistrieslists.registry.devfile.io created
customresourcedefinition.apiextensions.k8s.io/devfileregistries.registry.devfile.io created
customresourcedefinition.apiextensions.k8s.io/devfileregistrieslists.registry.devfile.io created
```

</details>

## Declaring the Devfile Registry

Now that the Custom Resource Definitions are installed, we are ready to declare the Devfile Registry we deployed by listing it as part of a
`DevfileRegistriesList` or `ClusterDevfileRegistriesList` Custom Resource.

A `DevfileRegistriesList` resource is scoped at the namespace level, while `ClusterDevfileRegistriesList` is a cluster-wide resource.

We will go on with creating a `DevfileRegistriesList` resource in the current namespace, but it is also possible to create a `ClusterDevfileRegistriesList`
instead if we have the appropriate permissions in the cluster.

Make sure you replace `<devfileRegistryUrl>` in the `url` field with the Devfile Registry Host (and protocol) we got from the previous sections.

:::caution
Due to [#6635](https://github.com/redhat-developer/odo/issues/6635), `odo` cannot be forced to work with HTTPS registries exposed over insecure or self-signed certificates.
So we will need to communicate with the Registry over HTTP for now.
:::

```shell
cat <<EOF | kubectl apply -f -               
apiVersion: registry.devfile.io/v1alpha1
kind: DevfileRegistriesList
metadata:
  name: ns-devfile-registries
spec:
  devfileRegistries:
    - name: my-devfile-registry
      url: <devfileRegistryUrl>
EOF
```

<details>
<summary>Example output:</summary>

```shell
$ cat <<EOF | kubectl apply -f -               
apiVersion: registry.devfile.io/v1alpha1
kind: DevfileRegistriesList
metadata:
  name: ns-devfile-registries
spec:
  devfileRegistries:
    - name: my-devfile-registry
      url: 'http://devfile-registry-default.172.20.0.2.nip.io'
EOF

devfileregistrieslist.registry.devfile.io/ns-devfile-registries created
```

</details>

:::note
There can be only one `ClusterDevfileRegistriesList` resource per cluster and only one `DevfileRegistriesList` resource per namespace.

Also, the registry URLs listed in those resources need to be valid and reachable URLs at the time they are created into the cluster.

These rules will be enforced if you have the [Devfile Registry Operator](https://github.com/devfile/registry-operator) installed in the cluster.
:::

## Using the in-cluster registry with odo

With the `DevfileRegistriesList` resource installed, `odo` will start using the registries listed there first.
You can check this by running `odo preference view`.

<details>
<summary>Example output:</summary>

```shell
$ odo preference view

Preference parameters:
[...]

Devfile registries:
 NAME                    URL                                                SECURE 
 my-devfile-registry     http://devfile-registry-default.172.20.0.2.nip.io  Yes
 DefaultDevfileRegistry  https://registry.devfile.io                        No
```

</details>

As a rule of thumb, `odo` will combine Devfile registries from the cluster and those listed in the local settings, and use them in the following priority order:
1. registries from the current namespace (declared in the `DevfileRegistriesList` resource)
2. cluster-wide registries (declared in the `ClusterDevfileRegistriesList` resource)
3. all other registries configured in the local configuration file

More details in this [guide](/docs/user-guides/advanced/using-in-cluster-devfile-registry).

This behavior applies to all `odo` commands interacting with registries, such as:
- [`odo preference view`](/docs/command-reference/preference)
- [`odo registry`](/docs/command-reference/registry)
- [`odo analyze`](/docs/command-reference/json-output#odo-analyze--o-json)
- [`odo init`](/docs/command-reference/init)
- [`odo dev`](/docs/command-reference/dev) and [`odo deploy`](/docs/command-reference/deploy) when there is no Devfile in the current directory

For example, we can list Stacks coming from our in-cluster Devfile Registry, with the `odo registry` command.

<details>
<summary>Example output:</summary>

```shell
$ odo registry --devfile-registry my-devfile-registry

 NAME                          REGISTRY             DESCRIPTION                                  VERSIONS     
 dotnet50                      my-devfile-registry  Stack with .NET 5.0                          1.0.3        
 dotnet60                      my-devfile-registry  Stack with .NET 6.0                          1.0.2        
 dotnetcore31                  my-devfile-registry  Stack with .NET Core 3.1                     1.0.3        
 go                            my-devfile-registry  Go is an open source programming languag...  1.0.2, 2.0.0 
 java-maven                    my-devfile-registry  Upstream Maven and OpenJDK 11                1.2.0        
 java-openliberty              my-devfile-registry  Java application Maven-built stack using...  0.9.0        
 java-openliberty-gradle       my-devfile-registry  Java application Gradle-built stack usin...  0.4.0        
 java-quarkus                  my-devfile-registry  Quarkus with Java                            1.3.0        
 java-springboot               my-devfile-registry  Spring Boot using Java                       1.2.0, 2.0.0 
 java-vertx                    my-devfile-registry  Upstream Vert.x using Java                   1.2.0        
 java-websphereliberty         my-devfile-registry  Java application Maven-built stack using...  0.9.0        
 java-websphereliberty-gradle  my-devfile-registry  Java application Gradle-built stack usin...  0.4.0        
 java-wildfly                  my-devfile-registry  Upstream WildFly                             1.1.0        
 java-wildfly-bootable-jar     my-devfile-registry  Java stack with WildFly in bootable Jar ...  1.1.0        
 nodejs                        my-devfile-registry  Stack with Node.js 16                        2.1.1        
 nodejs-angular                my-devfile-registry  Angular is a development platform, built...  2.0.2        
 nodejs-nextjs                 my-devfile-registry  Next.js gives you the best developer exp...  1.0.3        
 nodejs-nuxtjs                 my-devfile-registry  Nuxt is the backbone of your Vue.js proj...  1.0.3        
 nodejs-react                  my-devfile-registry  React is a free and open-source front-en...  2.0.2        
 nodejs-svelte                 my-devfile-registry  Svelte is a radical new approach to buil...  1.0.3        
 nodejs-vue                    my-devfile-registry  Vue is a JavaScript framework for buildi...  1.0.2        
 php-laravel                   my-devfile-registry  Laravel is an open-source PHP framework,...  1.0.1        
 python                        my-devfile-registry  Python is an interpreted, object-oriente...  2.1.0, 3.0.0 
 python-django                 my-devfile-registry  Django is a high-level Python web framew...  2.1.0        

```

</details>

## Wrapping Up

In this article, we have walked through deploying a Devfile Registry into our cluster, and have seen how `odo` can automatically
use registries that are declared in the cluster.

With a Devfile Registry deployed in the cluster, declaring it in a namespace-scoped `DevfileRegistriesList` or
cluster-wide `ClusterDevfileRegistriesList` resource will make `odo` automatically discover it and try to use it preferably.
This can be useful for example in an [air-gapped environment](/docs/user-guides/advanced/container-based-application-development-air-gapped-environment).

This shows how `odo` can automatically adapt to the environment of the cluster it is used against.

As usual, any feedback on this feature is appreciated.
