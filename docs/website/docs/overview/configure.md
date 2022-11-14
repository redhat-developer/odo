---
title: Configuration
sidebar_position: 6
---
# Configuration

## Configuring global settings

The global settings for `odo` can be found in `preference.yaml` file; which is located by default in the `.odo` directory of the user's HOME directory.

Example:

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

<Tabs
defaultValue="linux"
values={[
{label: 'Linux', value: 'linux'},
{label: 'Windows', value: 'windows'},
{label: 'Mac', value: 'mac'},
]}>

<TabItem value="linux">

```sh
/home/userName/.odo/preference.yaml
```

</TabItem>

<TabItem value="windows">

```sh
C:\\Users\userName\.odo\preference.yaml
```

</TabItem>

<TabItem value="mac">

```sh
/Users/userName/.odo/preference.yaml
```

</TabItem>
</Tabs>

---
A  different location can be set for the `preference.yaml` by exporting `GLOBALODOCONFIG` in the user environment.

### View the configuration
To view the current configuration, run the following command:

```shell
odo preference view
```
<details>
<summary>Example</summary>

```shell
$ odo preference view
Preference parameters:
 PARAMETER           VALUE
 ConsentTelemetry    true
 Ephemeral           true
 PushTimeout
 RegistryCacheTime
 Timeout
 UpdateNotification

Devfile registries:
 NAME             URL                                SECURE
 StagingRegistry  https://registry.stage.devfile.io  No

```
</details>

### Set a configuration
To set a value for a preference key, run the following command:
```shell
odo preference set <key> <value>
```
<details>
<summary>Example</summary>

```shell
$ odo preference set updatenotification false
Global preference was successfully updated
```
</details>

Note that the preference key is case-insensitive.

### Unset a configuration
To unset a value of a preference key, run the following command:
```shell
odo preference unset <key> [--force]
```

<details>
<summary>Example</summary>

```shell
$ odo preference unset updatednotification
? Do you want to unset updatenotification in the preference (y/N) y
Global preference was successfully updated
```
</details>

You can use the `--force` (or `-f`) flag to force the unset.
Unsetting a preference key sets it to an empty value in the preference file. `odo` will use the [default value](./configure#preference-key-table) for such configuration.

### Preference Key Table

| Preference         | Description                                                              | Default     |
| ------------------ |--------------------------------------------------------------------------| ----------- |
| UpdateNotification | Control whether a notification to update `odo` is shown                    | True        |
| Timeout            | Timeout for Kubernetes server connection check                           | 1 second    |
| PushTimeout        | Timeout for waiting for a component to start                             | 240 seconds |
| RegistryCacheTime  | Duration for which `odo` will cache information from the Devfile registry  | 4 Minutes   |
| Ephemeral          | Control whether `odo` should create a emptyDir volume to store source code | False       |
| ConsentTelemetry   | Control whether `odo` can collect telemetry for the user's `odo` usage       | False       |


## Managing Devfile registries

`odo` uses the portable *devfile* format to describe the components. `odo` can connect to various devfile registries to download devfiles for different languages and frameworks.

You can connect to publicly available devfile registries, or you can install your own [Devfile Registry](https://devfile.io/docs/2.1.0/building-a-custom-devfile-registry).

You can use the `odo preference <add/remove> registry` command to manage the registries used by `odo` to retrieve devfile information.

### Adding a registry

To add a registry, run the following command:

```
odo preference add registry <name> <url>
```

<details>
<summary>Example</summary>

```
$ odo preference add registry StageRegistry https://registry.stage.devfile.io
New registry successfully added
```
</details>

### Deleting a registry

To delete a registry, run the following command:

```
odo preference remove registry <name> [--force]
```
<details>
<summary>Example</summary>

```
$ odo preference remove registry StageRegistry
? Are you sure you want to delete registry "StageRegistry" Yes
Successfully deleted registry
```
</details>

You can use the `--force` (or `-f`) flag to force the deletion of the registry without confirmation.


:::tip **Updating a registry**
To update a registry, you can delete it and add it again with the updated value.
:::

## Advanced configuration

This is a configuration that normal `odo` users don't need to touch.
Options here are mostly used for debugging and testing `odo` behavior.

### Environment variables controlling `odo` behavior

| Variable                   | Usage                                                                                                                                                                                                                                                                                                                                                                          | Since         | Example                         |
|----------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|---------------|---------------------------------|
| `PODMAN_CMD`               | The command executed to run the local podman binary. `podman` by default                                                                                                                                                                                                                                                                                                       | v2.4.2        | `podman`                        |
| `DOCKER_CMD`               | The command executed to run the local docker binary. `docker` by default                                                                                                                                                                                                                                                                                                       | v2.4.2        | `docker`                        |
| `ODO_LOG_LEVEL`            | Useful for setting a log level to be used by `odo` commands. Takes precedence over the `-v` flag.                                                                                                                                                                                                                                                                              | v1.0.2        | 3                               |
| `ODO_DISABLE_TELEMETRY`    | Useful for disabling [telemetry collection](https://github.com/redhat-developer/odo/blob/main/USAGE_DATA.md). **Deprecated in v3.2.0**. Use `ODO_TRACKING_CONSENT` instead.                                                                                                                                                                                                    | v2.1.0        | `true`                          |
| `GLOBALODOCONFIG`          | Useful for setting a different location of global preference file `preference.yaml`.                                                                                                                                                                                                                                                                                           | v0.0.19       | `~/.config/odo/preference.yaml` |
| `ODO_DEBUG_TELEMETRY_FILE` | Useful for debugging [telemetry](https://github.com/redhat-developer/odo/blob/main/USAGE_DATA.md). When set it will save telemetry data to a file instead of sending it to the server.                                                                                                                                                                                         | v3.0.0-alpha1 | `/tmp/telemetry_data.json`      |
| `DEVFILE_PROXY`            | Integration tests will use this address as Devfile registry instead of `registry.stage.devfile.io`                                                                                                                                                                                                                                                                             | v3.0.0-beta3  | `my-registry.example.com`       |
| `TELEMETRY_CALLER`         | Caller identifier passed to [telemetry](https://github.com/redhat-developer/odo/blob/main/USAGE_DATA.md). Case-insensitive. Acceptable values: `vscode`, `intellij`, `jboss`.                                                                                                                                                                                                  | v3.1.0        | `intellij`                      |
| `ODO_TRACKING_CONSENT`     | Useful for controlling [telemetry](https://github.com/redhat-developer/odo/blob/main/USAGE_DATA.md). Acceptable values: `yes` ([enables telemetry](https://github.com/redhat-developer/odo/blob/main/USAGE_DATA.md) and skips consent prompt), `no` (disables telemetry and consent prompt). Takes precedence over the [`ConsentTelemetry`](#preference-key-table) preference. | v3.2.0        | `yes`                           |
| `ODO_EXPERIMENTAL_MODE`    | Whether to enable experimental features. See [Experimental Mode](../user-guides/advanced/experimental-mode) for more details. Acceptable values: `true`                                                                                                                                                                                                                        | v3.3.0        | `true`                          |
