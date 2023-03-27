---
title: Using Devfile Registries declared in the cluster
sidebar_position: 6
tags: ["devfile-registry", "registry", "in-cluster"]
slug: using-in-cluster-devfile-registry
---

Besides getting the list of Devfile Registries to work with from the [local configuration file](../../overview/configure.md#managing-devfile-registries),
`odo` can automatically detect Devfile Registries declared in the current cluster, and use them.

It does so by detecting the presence of the following [Custom Resources](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/) in the current cluster:
- `ClusterDevfileRegistriesList`: installable at the cluster level
- `DevfileRegistriesList`: installable at the namespace level

More details on the [Devfile Registry Operator documentation](https://github.com/devfile/registry-operator/blob/main/REGISTRIES_LISTS.md).

Registries detected from the cluster are added automatically to the top of the list of registries usable by `odo`, and `odo` will use them in the following priority order:
1. registries from the current namespace (declared in the `DevfileRegistriesList` resource)
2. cluster-wide registries (declared in the `ClusterDevfileRegistriesList` resource) 
3. all other registries configured in the local configuration file

This behavior applies to all `odo` commands interacting with Devfile registries, such as:
- [`odo preference view`](../../command-reference/preference.md)
- [`odo registry`](../../command-reference/registry.md)
- [`odo analyze`](../../command-reference/json-output.md#odo-analyze--o-json)
- [`odo init`](../../command-reference/init.md)
- [`odo dev`](../../command-reference/dev.md) and [`odo deploy`](../../command-reference/deploy.md) when there is no Devfile in the current directory

You can use the `odo preference view` command at any time to see the registries sorted by priority.

<details>
<summary>Example output:</summary>

```shell
$ odo preference view
[...]                

Devfile registries:
 NAME                      URL                                                   SECURE 
 ns-devfile-registry       http://my-devfile-registry.my-ns.172.17.0.1.nip.io    No     
 ns-devfile-staging        https://registry.stage.devfile.io                     Yes    
 cluster-devfile-registry  http://my-devfile-registry.cluster.172.17.0.1.nip.io  No     
 cluster-devfile-staging   https://registry.stage.devfile.io                     Yes    
 cluster-devfile-prod      https://registry.devfile.io                           Yes    
 Staging                   https://registry.stage.devfile.io                     Yes     
 DefaultDevfileRegistry    https://registry.devfile.io                           Yes     

```
</details>
