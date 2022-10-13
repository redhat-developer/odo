---
title: Configuration
sidebar_position: 6
---
# Configuration

## Configuring odo global settings

The global settings for odo can be found in `preference.yaml` file; which is located by default in the `.odo` directory of the user's HOME directory.

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
Unsetting a preference key sets it to an empty value in the preference file. odo will use the [default value](./configure#preference-key-table) for such configuration.

### Preference Key Table

| Preference         | Description                                                                    | Default     |
| ------------------ | ------------------------------------------------------------------------------ | ----------- |
| UpdateNotification | Control whether a notification to update odo is shown                          | True        |
| Timeout            | Timeout for Kubernetes server connection check                                 | 1 second    |
| PushTimeout        | Timeout for waiting for a component to start                                   | 240 seconds |
| RegistryCacheTime  | For how long (in minutes) odo will cache information from the Devfile registry | 4 Minutes   |
| Ephemeral          | Control whether odo should create a emptyDir volume to store source code       | False       |
| ConsentTelemetry   | Control whether odo can collect telemetry for the user's odo usage             | False       |


## Managing Devfile registries

odo uses the portable *devfile* format to describe the components. odo can connect to various devfile registries to download devfiles for different languages and frameworks.

You can connect to publicly available devfile registries, or you can install your own [Devfile Registry](https://github.com/devfile/registry-support).

You can use the `odo preference <add/remove> registry` command to manage the registries used by odo to retrieve devfile information.

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

### Environment variables controlling odo behavior

| Variable                    | Usage                                                                                                               |
|-----------------------------|---------------------------------------------------------------------------------------------------------------------|
| `PODMAN_CMD`                | The command executed to run the local podman binary. `podman` by default                                            |
| `DOCKER_CMD`                | The command executed to run the local docker binary. `docker` by default                                            |
| `ODO_LOG_LEVEL`             | Useful for setting a log level to be used by odo commands.                                                          |
| `ODO_DISABLE_TELEMETRY`     | Useful for disabling telemetry collection.                                                                          |
| `GLOBALODOCONFIG`           | Useful for setting a different location of global preference file preference.yaml.                                  |
| `ODO_DEBUG_TELEMETRY_FILE`  | Useful for debugging telemetry. When set it will save telemetry data to a file instead of sending it to the server. |
| `DEVFILE_PROXY`             | Integration tests will use this address as Devfile registry instead of `https://registry.stage.devfile.io`          |
| `TELEMETRY_CALLER`          | Caller identifier passed to telemetry. Acceptable values: `vscode`, `intellij`, `jboss`.                            |
