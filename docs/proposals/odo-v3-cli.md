# odo v3 CLI

## TODO:

- [ ] define JSON outputs for each command


# Table of Contents
- [odo v3 CLI](#odo-v3-cli)
  - [TODO:](#todo)
- [Table of Contents](#table-of-contents)
  - [odo v2 commands that should be removed](#odo-v2-commands-that-should-be-removed)
  - [Command that won't change much in v3](#command-that-wont-change-much-in-v3)
  - [Commands that should be present in v3.0.0-alpha1](#commands-that-should-be-present-in-v300-alpha1)
    - [general rules for odo cli](#general-rules-for-odo-cli)
    - [`odo login`](#odo-login)
    - [`odo logout`](#odo-logout)
    - [`odo init`](#odo-init)
      - [Interactive mode](#interactive-mode)
        - [Example](#example)
      - [Flags](#flags)
      - [Command behavior and error states](#command-behavior-and-error-states)
    - [`odo deploy`](#odo-deploy)
      - [Interactive mode](#interactive-mode-1)
      - [Flags](#flags-1)
      - [Command behavior and error states](#command-behavior-and-error-states-1)
        - [When there is no devfile in the current directory yet](#when-there-is-no-devfile-in-the-current-directory-yet)
        - [When devfile exists in the current directory](#when-devfile-exists-in-the-current-directory)
    - [`odo dev`](#odo-dev)
      - [Interactive mode](#interactive-mode-2)
      - [Flags](#flags-2)
      - [Command behavior and error states](#command-behavior-and-error-states-2)
        - [When there is no devfile in the current directory yet](#when-there-is-no-devfile-in-the-current-directory-yet-1)
        - [When devfile exists in the current directory](#when-devfile-exists-in-the-current-directory-1)
    - [`odo list`](#odo-list)
      - [`odo list components`](#odo-list-components)
      - [Flags](#flags-3)
        - [exmaple](#exmaple)
      - [`odo list namespaces`](#odo-list-namespaces)
        - [flags](#flags-4)
      - [`odo list endpoints`](#odo-list-endpoints)
      - [`odo list bindings`](#odo-list-bindings)
      - [`odo list services`](#odo-list-services)
    - [`odo preference`](#odo-preference)
    - [`odo project`](#odo-project)
    - [`odo registry`](#odo-registry)
    - [`odo describe`](#odo-describe)
    - [`odo exec`](#odo-exec)
    - [`odo status`](#odo-status)
    - [`odo build-images`](#odo-build-images)
  - [Commands that will be added in v3.0.0-alpha2](#commands-that-will-be-added-in-v300-alpha2)
    - [`odo create`](#odo-create)
      - [`odo create binding`](#odo-create-binding)
      - [`odo create service`](#odo-create-service)
      - [`odo create endpoint`](#odo-create-endpoint)
    - [`odo create component`](#odo-create-component)
      - [Interactive mode](#interactive-mode-3)
      - [Flags](#flags-5)
      - [Command behavior and error states](#command-behavior-and-error-states-3)
        - [When devfile exists in the current directory](#when-devfile-exists-in-the-current-directory-2)
        - [When there is no devfile in the current directory yet](#when-there-is-no-devfile-in-the-current-directory-yet-2)
    - [`odo delete`](#odo-delete)
      - [`odo delete binding`](#odo-delete-binding)
      - [`odo delete service`](#odo-delete-service)
      - [`odo delete endpoint`](#odo-delete-endpoint)
      - [`odo delete component`](#odo-delete-component)
    - [`odo app`](#odo-app)
    - [`odo catalog`](#odo-catalog)
    - [`odo config`](#odo-config)
    - [`odo build-images`](#odo-build-images-1)
    - [`odo utils`](#odo-utils)
    - [`odo version`](#odo-version)

Over the years of odo development we picked up a lot of commands.
Odo v3 will introduce new commands (`odo dev` #5299, `odo init` #5297).
This will change the command flow from what we currently have in v2. We need to make sure that the whole odo CLI is consistent, and all commands follow the same pattern.

There are also some commands that are there since the original odo v1 and were originally designed for s2i approach only, those commands or flags should be removed, or reworked to better fit Devfile

## odo v2 commands that should be removed

- `odo component` every command from “odo component *” already exists as a root command.
- `odo watch` will be replaced with `odo dev`
- `odo push` will be replaced with `odo dev`
- `odo unlink` will be replaced with `odo delete binding`
- `odo link` will be replaced with `odo create binding`
- `odo url` will be replaced with `odo create endpoint`
- `odo test` will be replaced `odo run test` ?
- `odo service` will be replaced with `odo create service`
- `odo storage` if needed it will be replaced with `odo create storage`
- `odo env` should be integrated with `odo config`
- `odo debug` will be replaced with `odo run debug`
- `odo registry` will be replaced with `odo preference ....`?

## Command that won't change much in v3

- `odo login`
- `odo logout`
- `odo preference` (maybe renamed to `odo config`? TODO: need to figure out right naming for configuring devfile component and configuring odo)
- `odo build-images`
- `odo deploy`
- `odo exec`


## Commands that should be present in v3.0.0-alpha1

### general rules for odo cli

- Once even a single flag is provided it ruins in non-interactive mode. All required information needs to be provided via flags
- If command is executed without flags it can enter interactive mode


### `odo login`

`odo login` should work exactly the same way as `oc login` command. It should have the same arguments and flags.

### `odo logout`

`odo login` should work exactly the same way as `oc login` command. It should have the same arguments and flags.

### `odo init`

[#5297](https://github.com/redhat-developer/odo/issues/5297)

#### Interactive mode

1. **"Select language:"**

   Shows list of all values of `metadata.language` fields from all devfiles in the current active Devfile registry. (every language only once)

2. **"Select project type:"**

   Select all possible values of `metadata.projectType` fields from all Devfiles that have selected language.
    If there is a Devfile that doesn't have `metadata.projectType` it should display its `metadata.name`.

    If there there is more than one devfile with the same projectType the list item should include the `metadata.name` and registry name. For example  if there are the same devfiles in multiple registries

    ```
    SpringBoot (java-springboot, registry: DefaultRegistry)
    SpringBoot (java-springboot, registry: MyRegistry)
    ```

    or if there is the same projectType in mulitple Devfiles

    ```
    SpringBoot (java-maven-springboot, registry: MyRegistry)
    SpringBoot (java-gradle-springboot, registry: MyRegistry)
    ```

3. **"Which starter project do you want to use:"**

    At this point, the previous answers should be enough to uniquely select one Devfile from registry.
    List of all starter projects defined in selected devfile.


4. **"Enter component name:"**
    Name of the component. This should be saved in the local `devfile.yaml` as a value for `metadata.name` field.

##### Example
```
$ odo init
TODO: Intro text (Include  goal as well as the steps that they are going to take ( including terminology ))

? Select language:  [Use arrows to move, type to filter]
> dotnet
  go
  java
  javascript
  typescript
  php
  python

? Select project type:  [Use arrows to move, type to filter]
  .NET 5.0
> .NET 6.0
  .NET Core 3.1
  ** GO BACK ** (not implemented)

? Which starter project do you want to use?  [Use arrows to move, type to filter]
> starter1
  starter2

? Enter component name: mydotnetapp

⠏ Downloading "dotnet60". DONE
⠏ Downloading starter project "starter1" ... DONE
Your new component "mydotnetapp" is ready in the current directory.
To start editing your component, use “odo dev” and open this folder in your favorite IDE.
Changes will be directly reflected on the cluster.
To deploy your component to a cluster use “odo deploy”.
```

Proof of concept was implemented in https://github.com/kadel/odo-v3-prototype/blob/main/cmd/init.go



#### Flags
- `--name` (string) - name of the component (required)
- `--devfile` (string) - name of the devfile in devfile registry (required if `--devfile-path` is not defined)
- `--devfile-registry`  (string)- name of the devfile registry (as configured in `odo registry`). It can be used in combination with `--devfile`, but not with `--devfile-path` (optional)
- `--starter`  (string) - name of the  starter project (optional)
- `--devfile-path` (string) - path to a devfile. This is alternative to using devfile from Devfile registry. It can be local filesystem path or http(s) URL (required if `--devfile` is not defined)
- `-o` (string) output information in a specified format (json).



If no flag is specified it should enter interactive mode.
If even a single optional flag is specified then run in non-interactive mode and requires all required flags.


#### Command behavior and error states

- the result of running `odo init` command should be local devfile.yaml saved in the current directory, and starter project extracted in the current directory (if user picked one)
- running `odo init` in non-empty directory exits with error

  ```
  The current directory is not empty. You can bootstrap new component only in empty directory.
  If you have existing code that you want to deploy use `odo deploy` or use `odo dev` command to quickly iterate on your component.
  ```

- command should use registry as configured in `odo registry` command. If there is multiple registries configured it should use all of them.




### `odo deploy`

Deploying application (outer loop)
[#5298](https://github.com/redhat-developer/odo/issues/5298)

#### Interactive mode
```
$ odo deploy
There is no devfile.yaml in the current directory.

Based on the files in the current directory odo detected
Language: Java
Project type: SpringBoot

? Is this correct? Yes

Current component configuration:
Opened ports:
- 8080
- 8084
Environment variables:
- FOO = bar
- FOO1 = bar1

? What configuration do you want change?  [Use arrows to move, type to filter]
> NOTHING - configuration is correct
  Delete port "8080"
  Delete port "8084"
  Add new port
  Delete environment variable "FOO"
  Delete environment variable "FOO1"
  Add new environment variable

⠏ Downloading "java-quarkus". DONE
Deploying your component to cluster ...
⠏ Building container image locally ... DONE
⠏ Pushing image to container registry ... DONE
⠏ Waiting for Kubernetes resources ... DONE
Your component is running on cluster.
You can access it at https://example.com
```

If there is no devfile in the current directory `odo` should use ([Alizer](https://github.com/redhat-developer/alizer/pull/55/)) to get corresponding Devfile for code base in the current directory.
After the successful detection it will show  Language and Project Type information to users and ask for confirmation.
If user answers that the information is not correct, odo should ask "Select language" and "Select project type" questions (see: `odo init` interactive mode).

The configuration part helps users to modify most common configurations done on Devfiles.
"? What configuration do you want change? " question is repeated over and over again until user confirms that the configuration is done and there is nothing else to change.

#### Flags

- `-o` (string) output information in a specified format (json).



#### Command behavior and error states

The result of successful execution of `odo deploy` is Devfile component deployed in outer loop mode.


##### When there is no devfile in the current directory yet

When no flags provided, starts in interactive mode to guide user through the devfile selection.
When some flags provided it errors out: "No devfile.yaml in the current directory. Use `odo create component` to get devfile.yaml for your application first."

##### When devfile exists in the current directory

Deploy application in outer loop mode using the information form `devfile.yaml` in the current directory. 


### `odo dev`

Running the application on the cluster for **development** (inner loop)

[#5299](https://github.com/redhat-developer/odo/issues/5298)

#### Interactive mode
```
$ odo dev
There is no devfile.yaml in the current directory.

Based on the files in the current directory odo detected
Language: Java
Project type: SpringBoot

? Is this correct? Yes

Current component configuration:
Opened ports:
- 8080
- 8084
Environment variables:
- FOO = bar
- FOO1 = bar1

? What configuration do you want change?  [Use arrows to move, type to filter]
> NOTHING - configuration is correct
  Delete port "8080"
  Delete port "8084"
  Add new port
  Delete environment variable "FOO"
  Delete environment variable "FOO1"
  Add new environment variable


⠏ Downloading "java-quarkus". DONE
Starting your application on cluster in developer mode ...
⠏ Waiting for Kubernetes resources ... DONE
⠏ Syncing files into the container ... DONE
⠏ Building your application in container on cluster ... DONE
⠏ Execting the application ... DONE
Your application is running on cluster.
You can access it at https://example.com

⠏ Watching for changes in the current directory ... DONE
Change in main.java detected.
⠏ Syncing files into the container ... DONE
⠏ Reloading application ... DONE
⠏ Watching for changes in the current directory ...

Press Ctrl+c to exit and clean up resources from cluster.

<ctrl+c>

⠏ Cleaning up ... DONE
```

Questions and their behavior is the same as for [`odo deploy`](#odo-deploy) command

#### Flags

- `-o` (string) output information in a specified format (json).
- `--watch` (boolean) Run command in watch mode. In this mode command is watching the current directory for file changes and automatically sync change to the container where it rebuilds and reload the application.
  By default, this is `true` (`--watch=true`). You can disable watch using `--watch=false`
- `--cleanup` (boolean). default is `true` when user presses `ctrl+c` it deletes all resource that it created on the cluster.

#### Command behavior and error states

The result of successful execution of `odo dev` is Devfile component deployed in inner loop mode.
By default command deletes all resources that it created before it exits. 

When users runs command with `--cleanup=false` there will be no cleanup before existing. The message "Press Ctrl+c to exit and clean up resources from cluster." will be only "Press Ctrl+c to exit."

If user executed `odo dev --cleanup=false` and then run this command again. The first line of the output should display warning: "Reusing already existing resources".


##### When there is no devfile in the current directory yet

When no flags provided, starts in interactive mode to guide user through the devfile selection.
When some flags provided it errors out: "No devfile.yaml in the current directory. Use `odo create component` to get devfile.yaml for your application first."

##### When devfile exists in the current directory

Deploy application in inner loop mode using the information form `devfile.yaml` in the current directory. 


### `odo list`

Listing "stuff" created by odo.

#### `odo list components`
list devfile components deployed to the cluster in the current namespace.


#### Flags
- `--namespace` - list components from the given namespace instead of current namespace.
- `--path` - find and list all components that are in a given path or in its subdirectories.

##### exmaple
```
$ odo list
Components in the "mynamspace" namespace:

  NAME            APPLICATION    TYPE         MANAGED BY ODO   RUNNING IN
* frontend        myapp          nodejs       Yes              Dev,Deploy
  backend         myapp          springboot   Yes              Deploy
  created-by-odc  asdf           python       No               Unknown
```

```
$ odo list --path /home/user/my-components/
Components present in the /home/user/my-components/ path

  NAME            APPLICATION    TYPE         MANAGED BY ODO   RUNNING IN  PATH
  frontend        myapp          nodejs       Yes              Dev         frontend
  backend         myapp          springboot   Yes              Deploy      backend
  backend         myapp          springboot   Yes              None        asdf

```

- row/component marked with `*` at the begging of the line is the one that is also in the current directory.
- `TYPE` corresponds to the `langauge` field in `devfile.yaml` tools, this should also correspond to `odo.dev/project-type` label.
- `RUNNING IN` indicates in what modes the component is running. `Dev` means the component is running in development mode (`odo dev`). `Deploy` indicates that the component is running in deploy mode (`odo deploy`), `None` means that the component is currently not running on cluster. `Unknown` indicates that odo can't detect in what mode is component running. `Unknown` will also be used for compoennt that are running on the cluster but are not managed by odo.
- `PATH` column is displayed only if the command was executed with `--path` flag. It shows the path in which the component "lives". This is relative path to a given `--path`.



#### `odo list namespaces`
Lists all namespaces that user has access to.
This is the same as `kubectl get namespace` or `oc projects`
The output shoudl indicate the active namespace.

##### flags
- `-o json` Output information in json format


#### `odo list endpoints`
list endpoints defined in local devfile and in cluster

#### `odo list bindings`
Will be implemented in v3.0.0-alpha2

list bindings defined in local devfile and in cluster

#### `odo list services`
Will be implemented in v3.0.0-alpha2

list services defined in local devfile and in cluster


### `odo preference`

Configures odo behavior like timeouts.
- UpdateNotification  - keep
- NamePrefix          - remove
- Timeout             - add explanations
- BuildTimeout        - remove
- PushTimeout         - add explanations
- Ephemeral           - keep
- ConsentTelemetry    - keep



### `odo project`
mostly as it is

### `odo registry`
mostly as it is just drop support "github" registry

### `odo describe`
TODO

### `odo exec`
mostly as it is

### `odo status`
mostly as it is

### `odo build-images`
mostly as it is in v2

## Commands that will be added in v3.0.0-alpha2

### `odo create`

#### `odo create binding`
Will be implemented in v3.0.0-alpha2

Generate new ServiceBinding. Based on the devfile existence save it to the devfile or just save it into the yaml file.

#### `odo create service`
Will be implemented in v3.0.0-alpha2

Generate custom resource for operator backed service. Based on the devfile existence save it to the devfile or just save it into the yaml file.

#### `odo create endpoint`

Create new endpoint (url) for. Saves the endpoint information to the `devfile.yaml`


### `odo create component`

Create new Devfile component in the current directory.
Some users might prefer creating components separately before running `odo dev` or `odo deploy`.
But this command is mainly intended to be used by IDE plugins or scripts.


#### Interactive mode
```
$ odo create

Based on the files in the current directory odo detected
Language: Java
Project type: SpringBoot

? Is this correct? Yes

Current component configuration:
Opened ports:
- 8080
- 8084
Environment variables:
- FOO = bar
- FOO1 = bar1

? What configuration do you want change?  [Use arrows to move, type to filter]
> NOTHING - configuration is correct
  Delete port "8080"
  Delete port "8084"
  Add new port
  Delete environment variable "FOO"
  Delete environment variable "FOO1"
  Add new environment variable

⠏ Downloading "java-quarkus". DONE

Your component is ready in the current directory.
To deploy it to the cluster you can run `odo deploy`.
```

The questions for interactive mode are identical to questions in `odo dev` and `odo deploy` command.

#### Flags

- `--name` - name of the component (required)
- `-o` output information in a specified format (json).
- `--devfile` - name of the devfile in devfile registry (required if `--devfile-path` is not defined)
- `--devfile-registry` - name of the devfile registry (as configured in `odo registry`). It can be used in combination with `--devfile`, but not with `--devfile-path` (optional)
- `--devfile-path` - path to a devfile. This is alternative to using devfile from Devfile registry. It can be local file system path or http(s) URL (required if `--devfile` is not defined)



#### Command behavior and error states

##### When devfile exists in the current directory

Error out with the message: "Unable to create new component in the current directory. The current directory already contains `devfile.yaml` file."

##### When there is no devfile in the current directory yet

Based on the information provided by user download devfile.yaml and place it into the current directory.


### `odo delete`

#### `odo delete binding`
Will be implemented in v3.0.0-alpha2

#### `odo delete service`
Will be implemented in v3.0.0-alpha2

#### `odo delete endpoint`

#### `odo delete component`


### `odo app`
TODO

### `odo catalog`
TODO

### `odo config`
TODO

### `odo build-images`
mostly as it is



### `odo utils`
mostly as it is

### `odo version`
mostly as it is

