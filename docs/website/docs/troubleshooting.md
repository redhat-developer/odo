---
sidebar_position: 4
title: Troubleshooting
toc_min_heading_level: 2
toc_max_heading_level: 3
---

This page documents possible solutions for the most common issues reported by users.

If your issue is not listed here, feel free to reach out to the team in the [#odo channel](https://kubernetes.slack.com/archives/C01D6L2NUAG) on the [Kubernetes Slack](http://slack.k8s.io/).
Or you can also [file an issue](https://github.com/redhat-developer/odo/issues/new/choose) or [start a new discussion](https://github.com/redhat-developer/odo/discussions).

[//]: # (## Generic issues)

## Authentication issues

### `odo deploy` is failing to push container image components due to 401 errors

#### Description

`odo` is failing to push the Image component from the Devfile. It looks like it is trying to push images to the `docker.io` registry.

<details>
    <summary>Example output</summary>

```shell
$ odo deploy
  __
 /  \__     Running the application in Deploy mode using go Devfile
 \__/  \    Namespace: my-test-project-1
 /  \__/    odo version: v3.9.0
 \__/

↪ Building & Pushing Image: go-image:latest
 •  Building image locally  ...
[...]
Successfully tagged localhost/go-image:latest
62cbfab7488bcb420404a7be564bb9a41dd2550c027e00fc9ca7037ae98cd193
 ✓  Building image locally [4s]
 •  Pushing image to container registry  ...
Getting image source signatures
Error: trying to reuse blob sha256:314640f419c581ddcac8f3618af39342a4571d5dc7a4e1f5b64d60f37e630b49 at destination: checking whether a blob sha256:314640f419c581ddcac8f3618af39342a4571d5dc7a4e1f5b64d60f37e630b49 exists in docker.io/library/go-image: errors:
denied: requested access to the resource is denied
error parsing HTTP 401 response body: unexpected end of JSON input: ""

 ✗  Pushing image to container registry [507ms]
 ✗  error running podman command: exit status 125
```

</details>

This most commonly happens with `odo deploy`, but might also happen with `odo dev` if an Image component is included in the Devfile as part of the inner-loop workflow. 

#### Possible Causes

The most common cause for this issue is that the Image component from the local Devfile is using a relative name in its `imageName` field.
For example:

```yaml
[...]
components:
- name: image-build
  image:
    # highlight-next-line
    imageName: go-image:latest
    dockerfile:
      uri: docker/Dockerfile
      buildContext: .
      rootRequired: false
```

Note that we recommend using relative image names to keep the Devfile portable.

#### Recommended Solution

As of `odo` [v3.11.0](/blog/odo-v3.11.0#handling-imagename-in-image-component-as-a-selector), we are handling relative image names as selectors, provided that you locally indicate to `odo` a registry where to push those relative images.
See [How `odo` handles image names](./development/devfile.md#how-odo-handles-image-names) for more details.

To fix this issue, use the command below to set your `ImageRegistry` preference before calling `odo deploy`.

```shell
odo preference set ImageRegistry $registry
```

<details>
    <summary>Example output</summary>

```shell
$ odo preference set ImageRegistry quay.io/$USER
 ✓  Value of 'imageregistry' preference was set to 'quay.io/user'
```

</details>

[//]: # (## Devfile issues)

## Dev Sessions issues

### The Image Pull Policy of the dev container is `Always` and I cannot change it

#### Description
When starting `odo dev`, the [image pull policy](https://kubernetes.io/docs/concepts/containers/images/#image-pull-policy) of the dev container is currently hardcoded to `Always`,
which may not be ideal for all platforms.

#### Recommended Solution
The image pull policy can be changed by declaring a [`container-overrides`](https://devfile.io/docs/2.2.0/overriding-pod-and-container-attributes#container-overrides) attribute in the `container` component in the Devfile, like so:

```yaml
components:
- name: runtime
# highlight-start
  attributes:
    container-overrides:
      imagePullPolicy: IfNotPresent
# highlight-end
  container:
    command: ['tail']
    args: ['-f', '/dev/null']
    endpoints:
    - name: http-node
      targetPort: 3000
    image: registry.access.redhat.com/ubi8/nodejs-16:latest
    mountSources: true
```

### I'm getting "Permission denied" errors when `odo dev` is syncing files

#### Description
When running `odo dev` against certain clusters, the `Syncing files into the container` stage fails due to "Permission denied" errors.

<details>
    <summary>Example output</summary>

```shell
$ odo dev
  __
 /  \__     Developing using the "places" Devfile
 \__/  \    Namespace: default
 /  \__/    odo version: v3.10.0
 \__/

 ⚠  You are using "default" namespace, odo may not work as expected in the default namespace.
 ⚠  You may set a new namespace by running `odo create namespace <name>`, or set an existing one by running `odo set namespace <name>`

↪ Running on the cluster in Dev mode
 •  Waiting for Kubernetes resources  ...
 ⚠  Pod is Pending
 ✓  Pod is Running
 ◐  Syncing files into the container ✗  Command 'tar xf - -C /projects --no-same-owner' in container failed.

 ✗  stdout:

 ✗  stderr: tar: main.go: Cannot open: Permission denied
tar: .gitignore: Cannot open: Permission denied
tar: README.md: Cannot open: Permission denied
tar: devfile.yaml: Cannot open: Permission denied
tar: go.mod: Cannot open: Permission denied
tar: Exiting with failure status due to previous errors


 ✗  err: error while streaming command: command terminated with exit code 2

 ✗  Syncing files into the container [4s]
Error occurred on Push - watch command was unable to push component: failed to sync to component with name places: failed to sync to component with name places: unable push files to pod: error while streaming command: command terminated with exit code 2


↪ Dev mode
 Status:
 Watching for changes in the current directory /tmp/go-app

 Keyboard Commands:
[Ctrl+c] - Exit and delete resources from the cluster
     [p] - Manually apply local changes to the application on the cluster
```

</details>

#### Possible Causes

Various factors are responsible for this:

- Storage Provisioner used for the cluster
- User set by the container image
- Location on the container where the files are to be synced
- Using Ephemeral vs Non-Ephemeral Volumes

#### Recommended Solution

Please refer to [Troubleshoot Storage Permission issues on managed cloud providers clusters](./user-guides/advanced/using-odo-with-other-clusters.md) for possible solutions to fix this.

### Orphan Devstate files

An `odo dev` session creates a `.odo/devstate.<PID>.json` file when the session starts, and deletes it at the end of the session.

If the session terminates abrupty, the state file won't be deleted, and will remain in the `.odo` directory.

You can delete such orphan devstate files using the command `odo delete component`.

<details>
    <summary>Example output</summary>

```shell
$ odo delete component
Searching resources to delete, please wait...
This will delete "go" from podman.
 •  The following pods and associated volumes will get deleted from podman:
 •  	- go-app

This will delete the following files and directories:
	- /home/user/projects/go/.odo/devstate.83932.json
	- /home/user/projects/go/.odo/devstate.json
```
</details>

## Podman Issues

### `odo` says it cannot access Podman

#### Description

`odo dev --platform podman` fails to start with the following error:

```
✗  unable to access podman.Do you have podman client installed and configured correctly?
cause: executable "podman" not recognized as podman client
```

This seems flaky however, because sometimes it works. 

#### Possible Causes

When initializing our connector to Podman, we have a default timeout of 1 second for the Podman executable to respond.
For some reason, `odo` sometimes might not be able to get a response from Podman during this short period of time.

#### Recommended Solution

First make sure the `podman version -f json` command returns in a timely manner on your system.
If the `podman` command cannot be found, please make sure Podman is installed correctly and available on your system path.
Or you can set the `PODMAN_CMD` environment variable to indicate where `odo` can find the `podman` executable, for example:

```shell
export PODMAN_CMD=/absolute/path/to/podman
```

If the `podman` executable is available, you can increase the timeout using the `PODMAN_CMD_INIT_TIMEOUT` environment variable.
You can increase this timeout before running `odo` and this should hopefully fix the issue, for example:

```shell
export PODMAN_CMD_INIT_TIMEOUT=10s
```

Please [file an issue](https://github.com/redhat-developer/odo/issues/new/choose) if the problem persists.

### I am using the remote Podman client and I am unable to reach the ports forwarded by `odo`

#### Description

Your local Podman is configured as a [remote client](https://github.com/containers/podman/blob/main/docs/tutorials/remote_client.md), interacting with another Podman Backend node via SSH.
For example, you are using the `podman-remote` binary, or are calling Podman with the `--remote` option.

The Podman remote client works fine, and you can start an `odo` Dev session leveraging it, e.g.:

```shell
$ odo dev --platform podman

[...]
 ✓  Waiting for the application to be ready [1s]
 -  Forwarding from 127.0.0.1:20001 -> 3000

↪ Dev mode
 Status:
 Watching for changes in the current directory /path/to/project

Web console accessible at http://localhost:20000/

Keyboard Commands:
[Ctrl+c] - Exit and delete resources from podman
     [p] - Manually apply local changes to the application on podman
```

However, the port forwarded by `odo dev` seems unreachable, e.g.:
```shell
$ curl http://127.0.0.1:20001

curl: (7) Failed to connect to localhost port 20001 after 0 ms: Connection refused
```

#### Possible causes

Podman provides the ability to use a local client interacting with a Podman backend node through an SSH connection.
However, as explained in [this discussion](https://github.com/containers/podman/discussions/20027#discussioncomment-7046636), all resources are only created on the backend node.
This includes any ports that might need to be forwarded.

#### Recommended Solution

Since you have already configured the SSH connection between the Podman remote client and the backend node, one possible workaround
could be to manually open up an SSH tunnel (using the same credentials) right after starting the `odo dev` session.
This way, SSH will do the work of forwarding ports between the machine running `odo dev` (along with the Podman remote client) and the Podman backend node.

Example:
```shell
$ ssh -v -i /path/to/ssh/key -NL 20001:127.0.0.1:20001 -l $user $podman_backend_host
```

Right after creating the SSH Tunnel in a separate terminal , you will be able to reach the port displayed by `odo` for port forwarding.

More details on [SSH Tunneling](https://www.ssh.com/academy/ssh/tunneling).
