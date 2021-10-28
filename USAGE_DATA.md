Usage Data
---

If the user has consented to `odo` collecting usage data, the following data will be collected when a command is executed -

* command's ID
* command's duration time
* command's pseudonymized error message and error type (in case of failure)
* whether the command was run from a terminal
* OS type
* `odo` version in use

Note that these commands do not include `--help` commands. We do not collect data about help commands.

###  Enable/Disable preference

#### Enable
`odo preference set ConsentTelemetry true`

#### Disable
`odo preference set ConsentTelemetry false`

Alternatively you can _disable_ telemetry by setting `ODO_DISABLE_TELEMETRY` environment variable to `true`.
This environment variable will override the `ConsentTelemetry` value set by `odo preference`.

The following table describes the additional information collected by odo commands.

|Event                  | Data                         | Type
| :-: | :-: | :-- |
|**Component Create** | Component Type | Devfile component type |
|**Component Push**| Component Type| Devfile component type|
| | **Cluster Type** | Openshift 4 / Kubernetes |
|**Project Create**| Cluster Type |Openshift 4 / Kubernetes |
|**Project Set**| Cluster Type |Openshift 4 / Kubernetes |
|**Preference Change** | Preference Key| UpdateNotification/NamePrefix/Timeout/BuildTimeout/PushTimeout/Ephemeral/ConsentTelemetry |


