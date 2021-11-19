---
title: Usage Data Collection
sidebar_position: 11
---
With user approval, we collect [pseudonymized](https://en.wikipedia.org/wiki/Pseudonymization) user data that will help us improve odo.

To learn more about what data is being collected and how to configure this collection, see [here](https://github.com/openshift/odo/blob/main/USAGE_DATA.md).


Considering the user has approved to data collection, everytime a command is executed, it triggers a background process that sends the data to our servers.

This background process is an odo command called `telemetry`.

Before the user-run command exits, data about the command execution is collected and `odo telemetry <jsonPayload>` is triggered as a background process.

If the command, for example `odo create nodejs` is successful, the following data will be sent -
```
{
    "event": "odo create",
    "properties": {
        "duration": "100(ms)",
        "success": true,
        "tty": true,
        "version": "odo v2.2.1 (e85f5e83c)"
        "cmdProperties": {
            "componentType": "nodejs"
        }
    }
}
```

If the command, for example `odo create xyz` is unsuccessful, the following data will be sent -
```
{
    "event": "odo create",
    "properties": {
        "duration": "100(ms)",
        "success": false,
        "tty": true,
        "version": "odo v2.2.1 (e85f5e83c)"
        "error": "component type \"xyz\" not found"
        "errortype": "*errors.fundamental"
        "cmdProperties": {
            "componentType": "nodejs"
        }
    }
}
```

Before sending the data to the server, error message is sanitized of any PII(Personally Identifiable Information) data which includes username, file paths, URLs, and pod exec commands executed via odo commands.

Note that these commands do not include `--help` commands. We do not collect data about help commands.

We use [Segment](https://segment.io) as our data platform and [Woopra](https://www.woopra.com) as our analytics tool.
