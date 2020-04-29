# odo debug support proposal
odo debug command is meant for debugging code running as odo component.
First implementation in odo will just support Java and NodeJS components.


Table of Contents:
- [odo debug support proposal](#odo-debug-support-proposal)
  - [Design overview](#design-overview)
  - [new `DebugPort` component setting](#new-debugport-component-setting)
  - [New odo commands](#new-odo-commands)
    - [`odo debug port-forward`](#odo-debug-port-forward)
      - [`--local-port` flag](#local-port-flag)
      - [Examples](#examples)
        - [The `DebugPort` is set in the LocalConfig:](#the-debugport-is-set-in-the-localconfig)
        - [If the `DebugPort` is NOT set in the LocalConfig:](#if-the-debugport-is-not-set-in-the-localconfig)
    - [`odo debug info`](#odo-debug-info)
- [Languages support](#languages-support)
  - [NodeJS](#nodejs)
  - [Java](#java)

## Design overview

The debug mode inside the container is controlled by `DEV_MODE` and `DEBUG_PORT` environment variables.
- by default supported component (Java and NodeJS) are started with debug enabled. This can be optionaly disabled by setting `DEV_MODE=false`
- `DEBUG_PORT` is optional and controls the port of the remote debugger.

Odo sets up port-forwarding to allow a debugger to connect to the running process. 


## new `DebugPort` component setting
`DebugPort` is a new config option in `LocalConfig` in  `ComponentSettings`.

```yaml
kind: LocalConfig
apiversion: odo.openshift.io/v1
ComponentSettings:
  Type: nodejs
  SourceLocation: ./
  SourceType: local
  DebugPort: 9229
  Ports:
  - "8080"
  - "8088"
  Application: app
  Project: tkmyproject
  Name: nodejs-nodejs-rest-kagx
  Url:
  - Name: nodejs-nodejs-rest-kagx-8080
    Port: 8080
```

It can be set either by calling `odo config set DebugPort 9229` on a running component.
Or it can be set at component creation time `odo component create --debug-port 9229`.

`DebugPort` option directly maps to `DEBUG_PORT` env variable. That means that when `odo push` is creating or updating `DeploymentConfig` it puts the `DebugPort` value as a value for `DEBUG_PORT` env variable in `PodSpec` inside the DeploymentConfig.

If the `DebugPort` is not set the `5858` will be used as a default value. 
Even if `DebugPort` in local config is not set than the `DEBUG_PORT` env variable in `DeploymentConfig` should be populated with `5858` value (default value).


## New odo commands

### `odo debug port-forward`

If no other flag is provided the command sets the port forwarding from the remote port as specified by `DebugPort` config value to the local port with the same number.

#### `--local-port` flag
Optional flag, that controls the number of the local port. The value is not stored in any config.

#### Examples

##### The `DebugPort` is set in the LocalConfig:
```yaml
kind: LocalConfig
apiversion: odo.openshift.io/v1
ComponentSettings:
  Type: nodejs
  SourceLocation: ./
  SourceType: local
  DebugPort: 9229
  ...
  ...
```

- `odo push` will set `DEBUG_PORT` environment variable in component's DeploymentConfig to `9292`
- `odo debug port-forward` will setup port forwarding from local `9229` port to remote `9229` container port.
- `odo debug port-forward --local-port 9999` will setup port forwarding from local `9999` port to remote `9229` container port.

##### If the `DebugPort` is NOT set in the LocalConfig:
```yaml
kind: LocalConfig
apiversion: odo.openshift.io/v1
ComponentSettings:
  Type: nodejs
  SourceLocation: ./
  SourceType: local
  ...
  ...
```

- `odo push` will set `DEBUG_PORT` environment variable in component's DeploymentConfig to `5858`
- `odo debug port-forward` will setup port forwarding from local `5858` port to remote `5858` container port.
- `odo debug port-forward --local-port 9999` will setup port forwarding from local `9999` port to remote `5858` container port.



### `odo debug info`
// TODO

Informs the user if the debug mode is enabled for the component and if the port-forward process is running or not.

# Languages support

## NodeJS
Starting application using: `npx nodemon --inspect=$DEBUG_PORT`

NOTE:
`npx` takes some time to execute nodemon. The user needs to include nodemon in application's `package.json` (`npm install nodemon --save --dev`) to work around this issue.



## Java

Starting JVM with additional options: `-Xdebug -Xrunjdwp:server=y,transport=dt_socket,address=${DEBUG_PORT},suspend=n`

