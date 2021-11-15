---
title: odo registry
sidebar_position: 5
---
# odo registry

odo uses the portable *devfile* format to describe the components. odo can connect to various devfile registries to download devfiles for different languages and frameworks.

You can connect to publicly available devfile registries, or you can install your own [Secure Registry](/docs/architecture/secure-registry).

You can use the `odo registry` command to manage the registries used by odo to retrieve devfile information.

## Listing the registries

You can use the following command to list the registries currently contacted by odo:

```
odo registry list
```

For example:

```
$ odo registry list
NAME                       URL                             SECURE
DefaultDevfileRegistry     https://registry.devfile.io     No
```

`DefaultDevfileRegistry` is the default registry used by odo; it is provided by the [devfile.io](https://devfile.io) project.

## Adding a registry

You can use the following command to add a registry:

```
odo registry add
```

For example:

```
$ odo registry add StageRegistry https://registry.stage.devfile.io
New registry successfully added
```

If you are deploying your own Secure Registry, you can specify the personal access token to authenticate to the secure registry with the `--token` flag:

```
$ odo registry add MyRegistry https://myregistry.example.com --token <access_token>
New registry successfully added
```

## Deleting a registry

You can delete a registry with the command:

```
odo registry delete
```

For example:

```
$ odo registry delete StageRegistry
? Are you sure you want to delete registry "StageRegistry" Yes
Successfully deleted registry
```

You can use the `--force` (or `-f`) flag to force the deletion of the registry without confirmation.

## Updating a registry

You can update the URL and/or the personal access token of a registry already registered with the command:

```
odo registry update
```

For example:

```
$ odo registry update MyRegistry https://otherregistry.example.com --token <other_access_token>
? Are you sure you want to update registry "MyRegistry" Yes
Successfully updated registry
```

You can use the `--force` (or `-f`) flag to force the update of the registry without confirmation.

