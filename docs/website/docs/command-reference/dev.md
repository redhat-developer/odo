---
title: odo dev
---

## Description

`odo dev` is used in order to quickly and effectively iterate through your development process for building an application.

This is [inner loop](../introduction#what-is-inner-loop-and-outer-loop) development and allows you to code, build, run and test the application in a continuous workflow.

## Running the Command

If you haven't already done so, you must [initialize](../command-reference/init) your source code with the `odo init` command.

Afterwards, run `odo dev`:

```sh
$ odo dev
  __
 /  \__     Developing using the my-nodejs-app Devfile
 \__/  \    Namespace: default
 /  \__/    odo version: v3.0.0-alpha1
 \__/

↪ Deploying to the cluster in developer mode
 ✓  Waiting for Kubernetes resources [3s]
 ✓  Syncing files into the container [335ms]
 ✓  Building your application in container on cluster [2s]
 ✓  Executing the application [1s]

Your application is now running on the cluster
 - Forwarding from 127.0.0.1:40001 -> 3000

Watching for changes in the current directory /Users/user/nodejs
Press Ctrl+c to exit `odo dev` and delete resources from the cluster
```

In the above example, three things have happened:
  * Your application has been built and deployed to the cluster
  * `odo` has port-forwarded your application for local accessability
  * `odo` will watch for changes in the current directory and rebuild the application when changes are detected

You can press Ctrl-c at any time to terminate the development session. The command can take a few moment to terminate, as it
will first delete all resources deployed into the cluster for this session before terminating.

### Running an alternative command

#### Running an alternative build command
By default, `odo dev` builds the application using the default Build command defined in the Devfile,
i.e, the command with a group `kind` set to `build` and with `isDefault` set to `true`., if any.

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
```shell
$ odo dev --build-command my-build-with-version

  __
 /  \__     Developing using the my-sample-go Devfile
 \__/  \    Namespace: default
 /  \__/    odo version: v3.0.0-alpha3
 \__/

↪ Deploying to the cluster in developer mode
 ✓  Waiting for Kubernetes resources [39s]
 ✓  Syncing files into the container [84ms]
 ✓  Building your application in container on cluster (command: my-build-with-version) [456ms]
 •  Executing the application (command: run)  ...

Your application is now running on the cluster
 - Forwarding from 127.0.0.1:40001 -> 8080

Watching for changes in the current directory /path/to/my/sources/go-app
Press Ctrl+c to exit `odo dev` and delete resources from the cluster

```

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
```shell
$ odo dev --run-command my-run-with-postgres

  __
 /  \__     Developing using the my-java-springboot-app Devfile
 \__/  \    Namespace: default
 /  \__/    odo version: v3.0.0-alpha3
 \__/

↪ Deploying to the cluster in developer mode
 ✓  Added storage m2 to my-java-springboot-app
 ✓  Creating kind ServiceBinding [8ms]
 ✓  Waiting for Kubernetes resources [39s]
 ✓  Syncing files into the container [84ms]
 ✓  Building your application in container on cluster (command: build) [51s]
 •  Executing the application (command: my-run-with-postgres)  ...

Your application is now running on the cluster
 - Forwarding from 127.0.0.1:40002 -> 8080

Watching for changes in the current directory /path/to/my/sources/java-springboot-app
Press Ctrl+c to exit `odo dev` and delete resources from the cluster

```

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
$ odo dev --var USER --var DEBUG=true
```

If you create a file `config.vars` containing:

```
USER
DEBUG=true
```

The following command will have the same behaviour as the previous one:

```shell
$ odo dev --var-file config.vars
```

The following command will override the `USER` Devfile variable with the `john` value:


```shell
$ odo dev --var USER=john --var-file config.vars
```

## Devfile (Advanced Usage)

### Devfile Overview

When `odo dev` is ran, it first looks in the `devfile.yaml` for the instructions to be executed, specifically: `components` and `commands`.

Each command has a group `kind` key which correspond to either: `build`, `run`, `test`, or `debug`.

These instructions make up the development cycle of `odo dev`.

With the following example `devfile.yaml` file generated by `odo init` and selecting `nodejs`, a container image will be pushed to the cluster, as well as your source code in order to start the development inner loop cycle.

A much more descriptive explanation on each part of a Devfile can be found on the [Devfile API reference](https://devfile.io/docs/devfile/2.0.0/user-guide/api-reference/) site.

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
