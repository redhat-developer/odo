---
title: Configuration
sidebar_position: 6
---
# Configuring odo global settings

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
To view the current configuration, run `odo preference view`.

```shell
odo preference view
```
Example:
```shell
$ odo preference view
PARAMETER             CURRENT_VALUE
UpdateNotification
Timeout
PushTimeout
RegistryCacheTime
Ephemeral
ConsentTelemetry
```
### Set a configuration
To set a value for a preference key, run `odo preference set <key> <value>`.
```shell
odo preference set updatenotification false
```
Example:
```shell
$ odo preference set updatenotification false
Global preference was successfully updated
```
Note that the preference key is case-insensitive.

### Unset a configuration
To unset a value of a preference key, run `odo preference unset <key>`; use `-f` flag to skip the confirmation.
```shell
odo preference unset updatednotification
```
Example:
```shell
$ odo preference unset updatednotification
? Do you want to unset updatenotification in the preference (y/N) y
Global preference was successfully updated
```

Unsetting a preference key sets it to an empty value in the preference file. odo will use the [default value](./configure#preference-key-table) for such configuration.

### Preference Key Table

| Preference         | Description                                                                    | Default                |
|--------------------|--------------------------------------------------------------------------------|------------------------|
| UpdateNotification | Control whether a notification to update odo is shown                          | True                   |
| Timeout            | Timeout for Kubernetes server connection check                                 | 1 second               |
| PushTimeout        | Timeout for waiting for a component to start                                   | 240 seconds            |
| RegistryCacheTime  | For how long (in minutes) odo will cache information from the Devfile registry | 4 Minutes              |
| Ephemeral          | Control whether odo should create a emptyDir volume to store source code       | True                   |
| ConsentTelemetry   | Control whether odo can collect telemetry for the user's odo usage             | False                  |


## Managing Devfile registries

odo uses the portable *devfile* format to describe the components. odo can connect to various devfile registries to download devfiles for different languages and frameworks.

You can connect to publicly available devfile registries, or you can install your own [Devfile Registry](https://github.com/devfile/registry-support).

You can use the `odo preference registry` command to manage the registries used by odo to retrieve devfile information.

### Adding a registry

You can use the following command to add a registry:

```
odo preference registry add
```

For example:

```
$ odo preference registry add StageRegistry https://registry.stage.devfile.io
New registry successfully added
```

### Deleting a registry

You can delete a registry with the command:

```
odo preference registry delete
```

For example:

```
$ odo preference registry delete StageRegistry
? Are you sure you want to delete registry "StageRegistry" Yes
Successfully deleted registry
```

You can use the `--force` (or `-f`) flag to force the deletion of the registry without confirmation.

### Updating a registry

You can update the URL and/or the personal access token of a registry already registered with the command:

```
odo preference registry update
```

For example:

```
$ odo preference registry update MyRegistry https://otherregistry.example.com
? Are you sure you want to update registry "MyRegistry" Yes
Successfully updated registry
```

You can use the `--force` (or `-f`) flag to force the update of the registry without confirmation.

