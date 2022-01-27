
# `odo preference`

Configures odo behavior like timeouts or registries that odo uses

The behavior is mostly the same as in v2. We can just remove some fields from the preference.
And the `odo registry` command should be integrated into `odo preference` as a `odo preference registry` subcommand.

- NamePrefix          - remove
- BuildTimeout        - remove
- Timeout             - add explanations
- PushTimeout         - add explanations
- Ephemeral           - keep
- ConsentTelemetry    - keep
- UpdateNotification  - keep


`odo preference view` should include list of the configured registries

## examples

```
odo preference registry add CheRegistry https://che-devfile-registry.openshift.io

odo preference registry list

odo preference view

odo preference set UpdateNotification false
```
