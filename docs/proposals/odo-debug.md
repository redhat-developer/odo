# odo debug support proposal
odo debug command is meant for debugging code running as odo component.
First implementation in odo will just support Java and NodeJS components.


Table of Contents:
- [odo debug support proposal](#odo-debug-support-proposal)
  - [Design overview](#design-overview)
    - [New odo commands](#new-odo-commands)
      - [`odo debug info`](#odo-debug-info)
    - [`odo debug port-forward LOCAL_PORT:REMOTE_PORT`](#odo-debug-port-forward-localportremoteport)
      - [Open Question](#open-question)
- [Languages support](#languages-support)
  - [NodeJS](#nodejs)
  - [Java](#java)

## Design overview

The debug mode inside the container is controlled by `DEBUG_MODE` and `DEBUG_PORT` environment variables.
- by default supported component (Java and NodeJS) are started with debug enabled -  `DEBUG_MODE == true`
- `DEBUG_PORT` is optional and controls the port of the remote debugger.

Odo sets up port-forwarding to allow a debugger to connect to the running process. 
Debugger client will always connect to `localhost:$DEBUG_PORT`

### New odo commands

#### `odo debug info`
Informs the user if the debug mode is enabled for the component and if the port-forward process is running or not.


### `odo debug port-forward LOCAL_PORT:REMOTE_PORT`
Forwards `LOCAL_PORT` to container port `REMOTE_PORT`.
This blocks the terminal. It will be used in `odo debug start`
As long as this command is running the port forward is enable (same as `oc port-forward`).

#### Open Question
To provide a good user experience user shouldn't be forced to think about ports.
`LOCAL_PORT` and `REMOTE_PORT` shouldn't be required. 
`odo debug port-forward` command should work without any additional arguments. 


# Languages support

## NodeJS

Starting application using: `npx nodemon --inspect=$DEBUG_PORT`


## Java

Starting JVM with additional options: `-Xdebug -Xrunjdwp:server=y,transport=dt_socket,address=${DEBUG_PORT},suspend=n`

