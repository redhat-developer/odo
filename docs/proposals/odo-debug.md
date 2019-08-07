# odo debug support proposal

Table of Contents:
- [odo debug support proposal](#odo-debug-support-proposal)
  - [Design overview](#design-overview)
    - [New odo commands](#new-odo-commands)
      - [`odo debug start`](#odo-debug-start)
      - [`odo debug stop`](#odo-debug-stop)
      - [`odo debug status`](#odo-debug-status)
    - [`odo debug port-forward LOCAL_PORT:REMOTE_PORT`](#odo-debug-port-forward-localportremoteport)
- [Languages support](#languages-support)
  - [NodeJS](#nodejs)
    - [Problem 1](#problem-1)
      - [Solution 1](#solution-1)
      - [Solution 2](#solution-2)
  - [Java](#java)

## Design overview

The debug mode inside the container is controlled by `DEBUG_MODE` and `DEBUG_PORT` env variable.
- if `DEBUG_MODE == true` than the code should be executed with remote debugging capabilities enabled.
- `DEBUG_PORT` is optional and controls the port of the remote debugger.

Odo sets up port-forwarding to allow a debugger to connect to the running process. 
Debugger client will always connect to `localhost:$DEBUG_PORT`
### New odo commands
#### `odo debug start`
Enables debug mode for the component. It does this by adding `DEBUG_MODE = true` env variable.

After adding env variables the component (Deployment Config) is restarted.
Once the pod is in running state `odo debug start` executes `odo debug port-forward $DEBUG_PORT:$DEBUG_PORT`.
This process executed on the background and PID of the process is recorded in `~/.odo/port-forward.pid` file.


Flags:
- `--port`: Optional. Controls value of `DEBUG_PORT` env variable
- `--context`: Optional. The same as in other commands (see `odo push --context` for example)
- `--no-port-forward`: Optional. Just enable debugging but don't start forwarding port. User can use `odo debug port-forward` manually or use other mechanisms.

#### `odo debug stop`
Stop port-forward process if running and remove `DEBUG_PORT` and `DEBUG_MODE` env variables from Deployment Config.

First, terminate and cleanup port-forwarding.
If `~/.odo/port-forward.pid` exists and process with PID is running terminate the process and remove the file.
If the process is not running just remove the file (warn the user).
If the file doesn't exist it is ok, as `odo debug start --no-port-forward` might have been used.

Second, remove `DEBUG_MODE` and `DEBUG_PORT` env variables from Deployment Config and wait for the pod to be the ready state after a restart.


#### `odo debug status`
Informs the user if the debug mode is enabled for the component and if the port-forward process is running or not.


### `odo debug port-forward LOCAL_PORT:REMOTE_PORT`
Forwards `LOCAL_PORT` to container port `REMOTE_PORT`.
This blocks the terminal. It will be used for `odo debug start`
As long as this command is running the port forward is enable (same as `oc port-forward`).



# Languages support

## NodeJS

To run the NodeJS application in the debug mode it has to start with `--inspect` (`node --inspect [host:port] script.js`) flag.

### Problem 1 
The command requires the entry point file. So odo has to know what is the `.js` file.
#### Solution 1
We can assume some default like `server.js`, and allow users to overwrite this using env variable.
#### Solution 2
User will provide their own debug command in `package.json`, the same way s2i builder image already assumes that there is a `start` script in `package.json`
```
 "scripts": {
 "start": "node server.js",
 "debug": "node --inspect server.js"
 }
```

## Java

