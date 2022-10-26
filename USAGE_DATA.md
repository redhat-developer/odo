Usage Data
---

You can help improve `odo` by allowing it to collect usage data.
Read more about our privacy statement in this article on [developers.redhat.com](https://developers.redhat.com/article/tool-data-collection).

If the user has consented to `odo` collecting usage data, the following data will be collected when a command is executed -

* Command Name
* Command Duration
* Command Success
* Pseudonymized error message and error type (in case of failure)
* Whether the command was run from a terminal
* `odo` version in use

In addition to this, the following data about user's identity is also noted - 
* OS type
* Timezone
* Locale

The following tables describe the additional information collected by `odo` commands.

**odo v3**

| Command                           | Data                                                                          |
|-----------------------------------|-------------------------------------------------------------------------------|
| odo init                          | Component Type, Devfile Name, Language, Project Type, Interactive Mode (bool) |
| odo dev                           | Component Type, Devfile Name, Language, Project Type                          |
| odo deploy                        | Component Type, Devfile Name, Language, Project Type                          |
| odo <create/set/delete> namespace | Cluster Type (Possible values: OpenShift 3, OpenShift 4, Kubernetes)          |

**odo v2**

| Command                  | Data                                                                 |
|--------------------------|----------------------------------------------------------------------|
| odo create               | Component Type, Devfile name                                         |
| odo push                 | Component Type, Cluster Type, Language, Project Type                 |
| odo project <create/set> | Cluster Type (Possible values: OpenShift 3, OpenShift 4, Kubernetes) |


All the data collected above is pseudonymized to keep the user information anonymous.

Note: Telemetry data is not collected when you run `--help` for commands.

###  Enable/Disable preference

#### Enable
`odo preference set ConsentTelemetry true`

#### Disable
`odo preference set ConsentTelemetry false`

Alternatively you can _disable_ telemetry by setting the `ODO_TRACKING_CONSENT` environment variable to `no`.
This environment variable will override the `ConsentTelemetry` value set by `odo preference`.
