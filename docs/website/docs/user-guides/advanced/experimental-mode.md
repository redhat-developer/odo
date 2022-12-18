---
title: Experimental Mode
sidebar_position: 5
---

Before making certain features generally available, `odo` might expose them as experimental first.

This gives you an opportunity to try out our newest features sooner.
In return, [your feedback](https://github.com/redhat-developer/odo/wiki/Community:-Getting-involved) helps us make sure that our new features
are reliable and useful, and we appreciate any and all feedback you want to provide.

If a feature is labeled as Experimental, this specifically means:
- The feature is in early or intermediate stages of development and [your feedback](https://github.com/redhat-developer/odo/wiki/Community:-Getting-involved) is highly appreciated.
- The feature may have known and undiscovered bugs.
- The feature may be changed, deprecated or removed altogether.
- The documentation for the feature can be incomplete, missing or may contain errors.
- The feature will have to be deliberately enabled.

## Enabling the experimental mode

Experimental mode is currently opt-in. You can enable it by setting the `ODO_EXPERIMENTAL_MODE` environment variable to `true` prior to running `odo`.
Doing so unlocks all experimental commands and flags.

Example:
```shell
$ ODO_EXPERIMENTAL_MODE=true odo dev --run-on=some-platform

============================================================================
⚠ Experimental mode enabled. Use at your own risk.
More details on https://odo.dev/docs/user-guides/advanced/experimental-mode
============================================================================

...
-  Forwarding from 127.0.0.1:40001 -> 8080
```

:::info NOTE
Running `odo` with an experimental command or flag without enabling the experimental mode returns an error.

```shell
$ odo dev --run-on=some-platform
...
...
 ✗  unknown flag: --run-on
```
:::

## List of experimental features

### Generic `--run-on` flag

This is a generic flag that allows running `odo` on any supported platform (other than the default Kubernetes or OpenShift cluster mode).

The supported platforms are `cluster` and `podman`.

By default, if you do not use the `--run-on` flag, or if you do not activate the experimental mode, the `cluster` platform is used.

The `cluster` platform uses the current Kubernetes or OpenShift cluster.

The `podman` platform uses the local installation of `podman`. It relies on the `podman` binary to be installed on your system.

These commands support the `--run-on`  flag:

- `odo dev`
