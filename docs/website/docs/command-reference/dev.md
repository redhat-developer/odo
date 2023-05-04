---
title: odo dev
---

`odo dev` is used in order to quickly and effectively iterate through your development process for building an application.

This is [inner loop](../introduction#what-is-inner-loop-and-outer-loop) development and allows you to code, build, run and test the application in a continuous workflow.

## Running the Command

If you haven't already done so, you must [initialize](../command-reference/init) your source code with the `odo init` command.

Afterwards, run `odo dev`:

```console
odo dev
```
<details>
<summary>Example</summary>

```console
$ odo dev
  __
 /  \__     Developing using the my-nodejs-app Devfile
 \__/  \    Namespace: default
 /  \__/    odo version: v3.0.0-alpha1
 \__/

↪ Running on the cluster in Dev mode
 ✓  Waiting for Kubernetes resources [3s]
 ✓  Syncing files into the container [335ms]
 ✓  Building your application in container on cluster [2s]
 ✓  Executing the application [1s]

Your application is now running on the cluster
 - Forwarding from 127.0.0.1:40001 -> 3000

Watching for changes in the current directory /Users/user/nodejs

[Ctrl+c] - Exit and delete resources from the cluster
     [p] - Manually apply local changes to the application on the cluster
```
</details>


In the above example, three things have happened:
  * Your application has been built and deployed to the cluster
  * `odo` has port-forwarded your application for local accessibility
  * `odo` will watch for changes in the current directory and rebuild the application when changes are detected

You can press Ctrl-c at any time to terminate the development session. The command can take a few moment to terminate, as it
will first delete all resources deployed into the cluster for this session before terminating.

### Applying local changes to the application on the cluster

By default, the changes made by the user to the Devfile and source files are applied directly.

The flag `--no-watch` can be used to change this behaviour: when the user changes the devfile or any source file, the changes
won't be applied immediately, but the next time the user presses the `p` key.

Depending on the local changes, different events can occur on the cluster:

- if source files are modified, they are pushed to the container running the application, and:
  - if the `build` command is marked as `HotReloadCapable`, the application is responsible for building the application with the new changes
  - if the `build` command is not marked as `HotReloadCapable`, the `build` command is executed again
  - if the `run` command is marked as `HotReloadCapable`, the application is responsible for applying the new changes
  - if the `run` command is not marked as `HotReloadCapable`, the application is stopped, then restarted by odo using the `run` command again.
- if the Devfile is modified, the deployment of the application is modified with the new changes. In some circumstances, this may
  cause the restart of the container running the application and therefore the application itself.


### Running an alternative command

#### Running an alternative build command
By default, `odo dev` builds the application using the default Build command defined in the Devfile,
i.e, the command with a group `kind` set to `build` and with `isDefault` set to `true`, if any.

Passing the `build-command` flag allows to override this behavior by running any other command, provided it is in the `build` group in the Devfile.

For example, given the following excerpt from a Devfile:
```yaml
- id: my-build
  exec:
    commandLine: go build main.go
    component: tools
    workingDir: ${PROJECT_SOURCE}
    group:
      isDefault: true
      kind: build

- id: my-build-with-version
  exec:
    commandLine: go build -ldflags="-X main.version=v1.0.0" main.go
    component: tools
    workingDir: ${PROJECT_SOURCE}
    group:
      kind: build
```

- running `odo dev` will build the application using the default `my-build` command.
- running `odo dev --build-command my-build-with-version` will build the application using the `my-build-with-version` command:
```console
odo dev --build-command my-build-with-version
```

<details>
<summary>Example</summary>

```console
$ odo dev --build-command my-build-with-version

  __
 /  \__     Developing using the my-sample-go Devfile
 \__/  \    Namespace: default
 /  \__/    odo version: v3.0.0-alpha3
 \__/

↪ Running on the cluster in Dev mode
 ✓  Waiting for Kubernetes resources [39s]
 ✓  Syncing files into the container [84ms]
 ✓  Building your application in container on cluster (command: my-build-with-version) [456ms]
 •  Executing the application (command: run)  ...

Your application is now running on the cluster
 - Forwarding from 127.0.0.1:40001 -> 8080

Watching for changes in the current directory /path/to/my/sources/go-app

[Ctrl+c] - Exit and delete resources from the cluster
     [p] - Manually apply local changes to the application on the cluster
```
</details>


#### Running an alternative run command

By default, `odo dev` executes the default Run command defined in the Devfile, 
i.e, the command with a group `kind` set to `run` and its `isDefault` field set to `true`.

Passing the `run-command` flag allows to override this behavior by running any other non-default command, provided it is in the `run` group in the Devfile.

For example, given the following excerpt from a Devfile:
```yaml
- id: my-run
  exec:
    commandLine: mvn spring-boot:run
    component: tools
    workingDir: ${PROJECT_SOURCE}
    group:
      isDefault: true
      kind: run

- id: my-run-with-postgres
  exec:
    commandLine: mvn spring-boot:run -Dspring-boot.run.profiles=postgres
    component: tools
    workingDir: ${PROJECT_SOURCE}
    group:
      isDefault: false
      kind: run
```

- running `odo dev` will run the default `my-run` command
- running `odo dev --run-command my-run-with-postgres` will run the `my-run-with-postgres` command:
```console
odo dev --run-command my-run-with-postgres
```
<details>
<summary>Example</summary>

```console
$ odo dev --run-command my-run-with-postgres

  __
 /  \__     Developing using the my-java-springboot-app Devfile
 \__/  \    Namespace: default
 /  \__/    odo version: v3.0.0-alpha3
 \__/

↪ Running on the cluster in Dev mode
 ✓  Added storage m2 to my-java-springboot-app
 ✓  Creating kind ServiceBinding [8ms]
 ✓  Waiting for Kubernetes resources [39s]
 ✓  Syncing files into the container [84ms]
 ✓  Building your application in container on cluster (command: build) [51s]
 •  Executing the application (command: my-run-with-postgres)  ...

Your application is now running on the cluster
 - Forwarding from 127.0.0.1:40002 -> 8080

Watching for changes in the current directory /path/to/my/sources/java-springboot-app

[Ctrl+c] - Exit and delete resources from the cluster
     [p] - Manually apply local changes to the application on the cluster

```
</details>


### Substituting variables

The Devfile can define variables to make the Devfile parameterizable. The Devfile can define values for these variables, and you 
can override the values for variables from the command line when running `odo dev`, using the `--var` and `--var-file` options.

The `--var` option is a repeatable option that takes a `variable=value` pair, where `=value` is optional. If the `=value` is omitted, the value is extracted from the environment variable named `variable`. In this case, if the environment variable with this name is not defined, the value defined into the Devfile will be used.

The `--var-file` option takes a filename as argument. The file contains a line-separated list of `variable=value` pairs, with the same behaviour as before.  

Note that the values passed with the `--var` option overrides the values obtained with the `--var-file` option.

#### Examples

Considering the Devfile contains this `variables` field:

```
variables:
  USER: anonymous
  DEBUG: false
```


This command will override the `USER` Devfile variable with the value of the `USER` environment variable, if it is defined.
It will also override the value of the `DEBUG` Devfile variable with the `true` value.

```shell
odo dev --var USER --var DEBUG=true
```

If you create a file `config.vars` containing:

```
USER
DEBUG=true
```

The following command will have the same behaviour as the previous one:

```shell
odo dev --var-file config.vars
```

The following command will override the `USER` Devfile variable with the `john` value:


```shell
odo dev --var USER=john --var-file config.vars
```


### Using custom port mapping for port forwarding
Custom local ports can be passed for port forwarding with the help of the `--port-forward` flag. This feature is supported on both podman and cluster.

This feature can be helpful when you want to provide consistent and predictable port numbers and avoid being assigned a potentially different port number every time `odo dev` is run.

Supported formats for this flag include:
1. `<LOCAL_PORT>:<CONTAINER_PORT>`
2. `<LOCAL_PORT>:<CONTAINER_NAME>:<CONTAINER_PORT>` - This format is necessary when multiple container components of a Devfile have the same port number.

The flag accepts a stringArray, so `--port-forward` flag can be defined multiple times.

If a custom port mapping is not defined for a port, `odo` will assign a free port in the range of 20001-30001.

```shell
odo dev --port-forward <LOCAL_PORT_1>:<CONTAINER_PORT_1> --port-forward <LOCAL_PORT_2>:<CONTAINER_NAME>:<CONTAINER_PORT_2>
```

<details>
<summary>Example</summary>

```shell
$ odo dev --port-forward 3000:runtime:3000 --port-forward 5000:5858 --debug
  __
 /  \__     Developing using the "my-nodejs-app" Devfile
 \__/  \    Namespace: default
 /  \__/    odo version: v3.9.0
 \__/

 ⚠  You are using "default" namespace, odo may not work as expected in the default namespace.
 ⚠  You may set a new namespace by running `odo create namespace <name>`, or set an existing one by running `odo set namespace <name>`

↪ Running on the cluster in Dev mode
 •  Waiting for Kubernetes resources  ...
 ⚠  Pod is Pending
 ✓  Pod is Running
 ✓  Syncing files into the container [152ms]
 ✓  Building your application in container (command: install) [27s]
 •  Executing the application (command: debug)  ...
 ✓  Waiting for the application to be ready [1s]
 -  Forwarding from 127.0.0.1:8000 -> 3000

 -  Forwarding from 127.0.0.1:5000 -> 5858


↪ Dev mode
 Status:
 Watching for changes in the current directory /tmp/nodejs-debug-2

 Keyboard Commands:
[Ctrl+c] - Exit and delete resources from the cluster
     [p] - Manually apply local changes to the application on the cluster
```
</details>

Note that `--random-ports` flag cannot be used with `--port-forward` flag.

### Using custom address for port forwarding
A custom address can be passed for port forwarding with the help of `--address` flag. This feature is supported on both podman and cluster.
The default value is 127.0.0.1.

```shell
odo dev --address <IP_ADDRESS>
```

<details>
<summary>Example</summary>

```shell
$ odo dev --address 127.0.10.3
  __
 /  \__     Developing using the "my-nodejs-app" Devfile
 \__/  \    Namespace: default
 /  \__/    odo version: v3.9.0
 \__/

 ⚠  You are using "default" namespace, odo may not work as expected in the default namespace.
 ⚠  You may set a new namespace by running `odo create namespace <name>`, or set an existing one by running `odo set namespace <name>`

↪ Running on the cluster in Dev mode
 •  Waiting for Kubernetes resources  ...
 ⚠  Pod is Pending
 ✓  Pod is Running
 ✓  Syncing files into the container [123ms]
 ✓  Building your application in container (command: install) [15s]
 •  Executing the application (command: run)  ...
 ✓  Waiting for the application to be ready [1s]
 -  Forwarding from 127.0.10.3:20001 -> 3000


↪ Dev mode
 Status:
 Watching for changes in the current directory /tmp/nodejs-debug-2

 Keyboard Commands:
[Ctrl+c] - Exit and delete resources from the cluster
     [p] - Manually apply local changes to the application on the cluster
```
</details>

:::note
If you are on macOS and using a Cluster platform, you may not be able to run multiple Dev sessions in parallel on address 0.0.0.0 without defining a custom port mapping, or using a different or default address.

For more information, see the following issues:
1. [Cannot start 2 different Dev sessions on Podman due to conflicting host ports](https://github.com/redhat-developer/odo/issues/6612)
2. [[MacOS] Cannot run 2 dev sessions simultaneously on cluster](https://github.com/redhat-developer/odo/issues/6744)
:::

### Running on Podman

Instead of deploying the container into a Kubernetes cluster, `odo dev` can leverage the podman installation on your system to deploy the container.

You need to use the `--platform podman` flags to run the component using podman instead of a Kubernetes cluster.

```console
odo dev --platform podman
```
<details>
<summary>Example</summary>

```console

$ odo dev --platform podman
  __
 /  \__     Developing using the "my-nodejs-app" Devfile
 \__/  \    Platform: podman
 /  \__/    odo version: v3.7.0
 \__/

↪ Running on podman in Dev mode
 ✓  Deploying pod [4s]
 ✓  Building your application in container (command: install) [3s]
 •  Executing the application (command: run)  ...
 -  Forwarding from 127.0.0.1:20001 -> 3000

↪ Dev mode
 Status:
 Watching for changes in the current directory /path/to/project/nodejs

 Keyboard Commands:
[Ctrl+c] - Exit and delete resources from podman
     [p] - Manually apply local changes to the application on podman
```
</details>

## Devfile (Advanced Usage)

### Devfile Overview

When `odo dev` is ran, it first looks in the `devfile.yaml` for the instructions to be executed, specifically: `components` and `commands`.

Each command has a group `kind` key which correspond to either: `build`, `run`, `test`, or `debug`.

These instructions make up the development cycle of `odo dev`.

With the following example `devfile.yaml` file generated by `odo init` and selecting `nodejs`, a container image will be pushed to the cluster, as well as your source code in order to start the development inner loop cycle.

A much more descriptive explanation on each part of a Devfile can be found on the [Devfile API reference](https://devfile.io/docs/2.2.0/devfile-schema) site.

Below, we'll explain each section of the corresponding Devfile:

### `metadata`

Descriptive registry information with regards to the devfile being used.

```yaml
metadata:
  description: Stack with Node.js 14
  displayName: Node.js Runtime
  icon: https://nodejs.org/static/images/logos/nodejs-new-pantone-black.svg
  language: javascript
  name: my-nodejs-app
  projectType: nodejs
  tags:
  - NodeJS
  - Express
  - ubi8
  version: 1.0.1
schemaVersion: 2.0.0
```

### `starterProjects`
The starter projects available for this devfile, shown when running `odo init`.
```yaml
starterProjects:
- git:
    remotes:
      origin: https://github.com/odo-devfiles/nodejs-ex.git
  name: nodejs-starter
```

### `commands`

The list of commands to be used when running `odo dev`.

Each command contains what will be ran within the container.

The four groups are: build, run, test and debug.

Build is what is initially ran when deploying to the cluster. It is what builds the command from the sources synced to the container.

Run executes the command after it has been built from sources. 

Debug is used instead of Run when running `odo dev --debug`.

Test executes any tests which are available and part of your application. (This is NOT-YET-IMPLEMENTED in odo)

All of these commands are needed to: build the program, execute the program, debug the application and run any applicable tests.
```yaml
commands:
- exec:
    commandLine: npm install
    component: runtime
    group:
      isDefault: true
      kind: build
    workingDir: ${PROJECT_SOURCE}
  id: install
- exec:
    commandLine: npm start
    component: runtime
    group:
      isDefault: true
      kind: run
    workingDir: ${PROJECT_SOURCE}
  id: run
- exec:
    commandLine: npm run debug
    component: runtime
    group:
      isDefault: true
      kind: debug
    workingDir: ${PROJECT_SOURCE}
  id: debug
- exec:
    commandLine: npm test
    component: runtime
    group:
      isDefault: true
      kind: test
    workingDir: ${PROJECT_SOURCE}
  id: test
```

### `components`

Components can include containers as well as Kubernetes yaml.

In our example, the nodejs container is being used on port 3000.
```yaml
components:
- name: runtime
  container:
    endpoints:
    - name: http-3000
      targetPort: 3000
    image: registry.access.redhat.com/ubi8/nodejs-14:latest
    memoryLimit: 1024Mi
    mountSources: true
```

Note that `odo` will set the container entrypoint to `tail -f /dev/null` if no `command` or `args` fields are explicitly defined for this component in the Devfile.
This is a temporary workaround that allows `odo` to start non-terminating containers in which the Devfile commands will get executed.

### Full Example

```yaml
metadata:
  description: Stack with Node.js 14
  displayName: Node.js Runtime
  icon: https://nodejs.org/static/images/logos/nodejs-new-pantone-black.svg
  language: javascript
  name: my-nodejs-app
  projectType: nodejs
  tags:
  - NodeJS
  - Express
  - ubi8
  version: 1.0.1
schemaVersion: 2.0.0

starterProjects:
- git:
    remotes:
      origin: https://github.com/odo-devfiles/nodejs-ex.git
  name: nodejs-starter

commands:
- exec:
    commandLine: npm install
    component: runtime
    group:
      isDefault: true
      kind: build
    workingDir: ${PROJECT_SOURCE}
  id: install
- exec:
    commandLine: npm start
    component: runtime
    group:
      isDefault: true
      kind: run
    workingDir: ${PROJECT_SOURCE}
  id: run
- exec:
    commandLine: npm run debug
    component: runtime
    group:
      isDefault: true
      kind: debug
    workingDir: ${PROJECT_SOURCE}
  id: debug
- exec:
    commandLine: npm test
    component: runtime
    group:
      isDefault: true
      kind: test
    workingDir: ${PROJECT_SOURCE}
  id: test

components:
- name: runtime
  container:
    endpoints:
    - name: http-3000
      targetPort: 3000
    image: registry.access.redhat.com/ubi8/nodejs-14:latest
    memoryLimit: 1024Mi
    mountSources: true
```

## State File

When the command `odo dev` is executed, the state of the command is saved to the file `.odo/devstate.json`. 

This state file contains the forwarded ports:

```json
{
 "forwardedPorts": [
  {
   "containerName": "runtime",
   "localAddress": "127.0.0.1",
   "localPort": 40001,
   "containerPort": 3000
  }
 ]
}
```
