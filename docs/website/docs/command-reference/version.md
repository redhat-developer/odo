---
title: odo version
---

## Description
The `odo version` command returns the version information about `odo`, cluster and podman client.

## Running the Command
The command takes an optional `--client` flag that only returns version information about `odo`. It will only print Openshift version if it is available.
```shell
odo version [--client] [-o json]
```

<details>
<summary>Example</summary>

```shell
$ odo version
odo v3.11.0 (a9e6cdc34)

Server: https://ab0bc42973f0043e7a2b9c24f5acddd6-9c1554c20c1ec323.elb.us-east-1.amazonaws.com:6443
OpenShift: 4.13.0
Kubernetes: v1.27.2+b451817
Podman Client: 4.5.1
```
</details>
