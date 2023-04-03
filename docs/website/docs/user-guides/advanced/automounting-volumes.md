---
title: Automounting Volumes
sidebar_position: 8
---

Existing configmaps, secrets, and persistent volume claims on the cluster can be mounted automatically to all containers created by `odo`. These resources can be configured by applying the appropriate labels.

To mark a resource for mounting to containers created by `odo`, apply the following label to the resource:

```yaml
metadata:
  labels:
    controller.devfile.io/mount-to-containers: "true"
```

By default, resources will be mounted based on the resource name:

- Secrets will be mounted to `/etc/secret/<secret-name>`

- Configmaps will be mounted to `/etc/config/<configmap-name>`

- Persistent volume claims will be mounted to `/tmp/<pvc-name>`

Mounting resources can be additionally configured via annotations:

- `controller.devfile.io/mount-path`: configure where the resource should be mounted

- `controller.devfile.io/mount-as`: for secrets and configmaps only, configure how the resource should be mounted to the container

    - If `controller.devfile.io/mount-as: file`, the configmap/secret will be mounted as files within the mount path. This is the default behavior.

    - If `controller.devfile.io/mount-as: subpath`, the keys and values in the configmap/secret will be mounted as files within the mount path using subpath volume mounts.

    - If `controller.devfile.io/mount-as: env`, the keys and values in the configmap/secret will be mounted as environment variables in all containers.

    When `file` is used, the configmap is mounted as a directory within the containers, erasing any files/directories already present. When `subpath` is used, each key in the configmap/secret is mounted as a subpath volume mount in the mount path, leaving existing files intact but preventing changes to the secret/configmap from propagating into the containers without a restart.

- `controller.devfile.io/read-only`: for persistent volume claims, mount the resource as read-only

