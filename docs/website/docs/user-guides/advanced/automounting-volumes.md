---
title: Automounting Volumes
sidebar_position: 8
---

Existing [ConfigMaps](https://kubernetes.io/docs/concepts/configuration/configmap/), [Secrets](https://kubernetes.io/docs/concepts/configuration/secret/), and [Persistent Volume Claims](https://kubernetes.io/docs/concepts/storage/persistent-volumes/) on the cluster can be mounted automatically to all containers created by `odo`. These resources can be configured by applying the appropriate labels.

To mark a resource for mounting to containers created by `odo`, apply the following label to the resource:

```yaml
metadata:
  labels:
    devfile.io/auto-mount: "true"
```

By default, resources will be mounted based on the resource name:

- Secrets will be mounted to `/etc/secret/<secret-name>`

- Configmaps will be mounted to `/etc/config/<configmap-name>`

- Persistent volume claims will be mounted to `/tmp/<pvc-name>`

Mounting resources can be additionally configured via annotations:

- `devfile.io/mount-path`: configure where the resource should be mounted

- `devfile.io/mount-as`: for secrets and configmaps only, configure how the resource should be mounted to the container

    - If `devfile.io/mount-as: file`, the configmap/secret will be mounted as files within the mount path. This is the default behavior.

    - If `devfile.io/mount-as: subpath`, the keys and values in the configmap/secret will be mounted as files within the mount path using subpath volume mounts.

    - If `devfile.io/mount-as: env`, the keys and values in the configmap/secret will be mounted as environment variables in all containers.

  When `file` is used, the configmap is mounted as a directory within the containers, erasing any files/directories already present. When `subpath` is used, each key in the configmap/secret is mounted as a subpath volume mount in the mount path, leaving existing files intact but preventing changes to the secret/configmap from propagating into the containers without a restart.

- `devfile.io/read-only`: for persistent volume claims, mount the resource as read-only

- `devfile.io/mount-access-mode`: for  secret/configmap, can be used to configure file permissions on mounted files
