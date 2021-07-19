---
title: Configuration
sidebar_position: 6
---
# Configuring odo global settings

The global settings for odo can be found in `preference.yaml` file; which is located by default in the `.odo` directory of the user's HOME directory.

A  different location can be set for the `preference.yaml` by exporting `GLOBALODOCONFIG` in the user environment.

### View the configuration
To view the current configuration, run `odo preference view`.

```shell
$ odo preference view
PARAMETER             CURRENT_VALUE
UpdateNotification
NamePrefix
Timeout
BuildTimeout
PushTimeout
Experimental
Ephemeral
ConsentTelemetry
```

### Set a configuration
To set a value for a preference key, run `odo preference set <key> <value>`.
```shell
$ odo preference set updatenotification false
Global preference was successfully updated
```

### Unset a configuration
To unset a value of a preference key, run `odo preference unset <key>`.
```shell
$ odo preference unset updatednotification
? Do you want to unset updatenotification in the preference (y/N) y
Global preference was successfully updated
```

Unsetting a preference key sets it back to its default value.

### Preference Key Table

| Preference | Description | Default |
| ----------- | ----------- | ----------- |
| UpdateNotification | Control if an update notification is shown or not | True |
| NamePrefix | Set a default name prefix | Current directory name |
| Timeout | Timeout for server connection check | 1 second |
| BuildTimeout | Timeout for waiting for a build of the git component to complete | 300 seconds |
| PushTimeout | Timeout for waiting for a Pod to come up |  240 seconds |
| Experimental | Expose features in development/experimental mode | False |
| Ephemeral | Control whether odo should create a emptyDir volume to store source code | False |
| ConsentTelemetry | Control whether odo can collect telemetry for the user's odo usage | False |