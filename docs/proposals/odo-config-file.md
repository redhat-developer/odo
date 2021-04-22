# `odo` config file

# !! OUTDATED !!
see https://docs.google.com/document/d/1RKva9lvIR6dLmDm2tlqmp8V4nFjlrXdulOQ4FSmLkr8/


## Abstract

When developers work on creating an application, there's 2 phases they go through. First phase is creating the application initial implementation from scratch, and then iterate development on an existing application. While the developer works on the application, he might take some decisions on how the application should behave that should be preserved over time, but that are not source code but rather specific to how the application should behave in relation to the platform where it's running or in relation to the tool used to develop it.

IDEs often use a configuration file or directory to hold specific configuration for an application, workspace or the whole tool. These files and directories are stored under the same control version system as where the source code lives, so that anytime a developer checks out that repository has control on how to configure that tool to work with that source code.

Git is another example of a tool that provides configuration, that can be related of every instance of any repository the user has or to a single instance. These are defined as global and local config. Git promoves the fact that local configuration is always more important than global configuration.

## Motivation

Provide a way to preserve conscisous decisions the developer has made that are orthogonal to the code but relevant to the tool/platform. Any user that continues working on an existing application don't need to know the decisions the original developer took, as these are preserved along with the source code and provides for better reproducibility of the development environment setup.

## Use Cases

1. As a user, I want to be able to preserve some configuration I've done about my current component along with my source code
1. As a user, I want to be able to use configuration stored with my source code to help me be more productive
1. As a user, I want to be able to manage explicit configuration that is saved with my source code
1. As a user, I want to be able to check out an application source code on different computers and have them producing the same result in relation to the development environment (deployment of the application on the platform)

## Design overview

We need to provide a configuration file `odo-config` in a configuration directory `.odo`, that will hold specific configuration that `odo` will use on execution. This configuration might modify how the CLI work or how the component will work on the platform.

### Scopes of configuration

`odo` CLI configuration needs to be preserved in configuration files. As a developer might work on multiple
components, there is a need to have different scopes for configuration so that a developer don't need to
explicitly configure each and every component independently.

#### Local configuration

Every component can be configured independently. We call this local configuration for the component as it only affects the component it references. This configuration will be stored in a file in `.odo` specific directory located at the root directory of the component's source code. In this way, the configuration file can be added to version control system and preserved along with the application. If any other developer checks out the same source code he will get the same configuration as the original developer.

We propose to have a `.odo` folder in the root of the application's source code directory to store every odo local configurations.

```bash
<app-dir>/.odo/odo-config
```

A good representation of this is a developer that while on the development process realizes his application requires 2 GB of memory to run appropriately, he might decide to save that information in this configuration file. At a later moment, any other developer (or even him in the future) that decides to work on the same application will have this configuration definition along with his source code. `odo` should take into account this configuration definition and materialize when appropriate. The type of configuration, format and applicability will be described later in this document.

#### Global configuration

When working with multiple components it might be tedious to configure each and every component, or a user wants to have a default configuration value applied to each and every component he works with when no explicit (local) configuration exists. For this purpose, there should be a file, stored in the user's home directory's odo specific folder, that should contain this configuration definitions. This configuration is global to every component the user works with.

We propose to have a `.odo` folder in the root of the users home directory to store every odo global configurations.

```bash
$HOME/.odo/odo-config
```

#### User explicit configuration

Often times a user wants to use different configuration values than the provided in the configuration file. We need to provide a way to allow a developer to specify this new values he want to use. For that, the user needs to explicitly provide the values as CLI arguments to the specific `odo` command.

An example of this would be:

```bash
odo create wildfly --memory 2GB
```

In this previous example, the user is explicitly instructing `odo` to use 2GB of memory when creating a component based on wildfly.

User explicit configuration will be stored in a local configuration file also in the odo`s specific local folder:

```bash
<app-dir>/.odo/odo-config.local
```

This user explicit configuration will be used in any subsequent command, avoiding the user the need to provide the flag every time. This file serves as configuration overrides. When explicitly setting additional configuration for a component, this configuration will be saved in the `.odo/odo-config.local`. Every time a flag is provided, it will overwrite the value in that file if it already existed, as new override value.

There is no specific command that will manage this file. If a user needs to change/delete any value, he will need to manually edit this file. A value will be set in this file when odo is used.

There's an additional purpose to this file an is for `odo` to automatically recognize the applied configuration to a component when the user changes directory into the application's source code dir, the same way as git does with .git directory and config.

**NOTE:** The automatic detection of `.odo` configuration when a user changes directories, and the corresponding `odo` behavior, will be described in a different proposal.

One caveat to this solution is that we need to guarantee that this file is not saved in version control system, as this is solely used as override for the specific user.



### <a name="configuration-precedence"></a> Configuration precedence

As configuration can exist both locally and globally or only on one of these scopes, it seems obvious that there needs to be some rule managing the precedence and applicability of the configuration.

The most explicit/concrete configuration wins over the less explicit one. In this case, providing a configuration value via a CLI argument always wins over any other existing configuration. If no explicit value is provided, then the configuration defined along with the component's source code wins over any other configuration. In case there is no explicit configuration or configuration for that single component then any global configuration (if existing) will be applied. In any other case, `odo`'s defaults will be used.

|global|local|CLI flag or override|applied value|
|--|--|--|--|
|--|--|--|odo defaults|
|Value A|--|--|Value A|
|Value A|--|Value X|Value X|
|Value A|Value B|--|Value B|
|Value A|Value B|Value X|Value X|
|--|Value B|--|Value B|
|--|Value B|Value X|Value X|
|--|-|Value X|Value X|

### Configuration typology

In the scope of this proposal, we refer to configuration in some different ways and it's worth noting what we mean in each case. What types of configuration will be supported and what use case they fullfil.

### `odo` CLI behavior configuration

Like many tools, there is configuration related to how the tool behaves itself. An example of this would be a timeout that the CLI should apply when issuing commands to the server, or the ability to enable/disable checks for new versions. This configuration will always be global, and stored relative to the user's odo directory in the user's home directory. By default this configuration file will live in `$HOME/.odo/odo-config`

The [format](#odo-file-format) of this file is described later in the proposal.

### Application related configuration

Sometimes what we want to configure is how the application/component will be deployed. Some components will
require special characteristics, like a specific amount of memory. Some other times, what we want is to instruct `odo` to use an specific value to name our component if none is explicitly provided. This configuration will be stored locally to the component, along with the source code, in an `odo-config` configuration file, in an `.odo` specific directory, that can be saved into the version control system used, and share by anyone using the same application code.

For some specific values, there will be the possibility to set them globally, as they will be used in that case for every component created, as defined in the [configuration precedence](#configuration-precedence) section. This values will be documented as being `global` configuration values.

### Developer friendly flags

One of the goals of `odo` is to be easy to use for developers. This forces to translate some configuration possible in the underlying technology (OpenShift/Kubernetes) into a friendlier name that developers could easily understand. An example of this would be the amount of memory a component needs. In OpenShift, this is reflected by `limits.memory` and `requests.memory`, and depending on how you set these values, influences the QoS of the deployment. In `odo` these should be simplified to `memory`, `min-memory` and `max-memory`. These are terms that any developer will most likely understand without looking at the documentation.

* When `memory` is set, memory requests and limits in k8s parlance and both have the same value.
* `min-memory` should be memory requests in k8s parlance.
* `max-memory` should be memory limits in k8s parlance.
* `memory` is fill the value of `min-memory` or `max-memory` if these are not set. Otherwise, when both `min-memory` and `max-memory` are set, `memory` will have no effect, possibly trying to set the 3 values should yield a warning.

#### Proposed flags

We need to define and delimit what information will be stored in this file. Configuration which can be local and global will be denoted. Otherwise, the configuration will only be local to the application's odo-config file.

### Deployment decoration

Configuration used to define how the deployment would be created:

* mem/memory (local/global)
  * min-mem/min-memory (local/global)
  * max-mem/max-memory (local/global)
* cpu (local/global)
  * min-cpu (local/global)
  * max-cpu (local/global)
* health (local)
  * live-health (local)
  * ready-health (local)
* build-image (local)
  * src-path (local)
  * artifacts-path (local)
* runtime-image (local)
  * src-path (local)
  * deployment-path (local)

### CLI behavior

Configuration used to define how the CLI should behave:

* app-name: (local) Preferred application name if none is defined. As this value is in the local configuration file, it will apply to every component created from the source code if not specific application name is provided, else, the one provided via command line will take precedence.
* component-name: (local) Preferred component name if none is defined. As this value is in the local configuration file, it will apply to the created component from the source code if not specific component name is provided, else, the one provided via command line will take precedence. In case there is already a component with this name, an error should be provided on component creation.
* odoignore: (local/global) list of files/patterns to ignore by default if no .odoignore file is provided, or when it's cerated. **This configuration can be global, as the developer might want to always ignore specific files (e.g. .git, .odo, ...) when pushing.**

#### <a name="odo-file-format"></a> `odo-config` file format

The format for both files should be human readable. Whether it's yaml, json or [toml](https://github.com/toml-lang/toml) is not the scope of this proposal to define, as it will be dictated by the engineering team implementing this feature. In any case, it should be possible to:

* Visually interpret the content of this file (vi, cat).
* Process this file in an easy way (grep, aqk, sed, jq|yq) for automation.
* Be stored (and visualized) in version control system as part of the application source code.
* Be independent and usable from any operating system (Linux, mac OS, windows)

Nevertheless, `odo` will provide a way to manipulate this file via a CLI command, as described in [managing configuration](#managing-configuration)

### <a name="managing-configuration"></a> Managing configuration

Developers will need to add configuration values to the `.odo/odo-config` config files. This configuration management will be managed by a new command in `odo` CLI. As configuration will be of different types, there will be needs to additional verbs (to the regular ones) for managing this configuration.

#### Creating configuration

There might be times that a developer would want to add/edit/remove values from the `.odo/odo-config` configuration file. There should be a command to manage this file in the `odo` CLI. That command should be the same used to touch configuration of the global `odo` config file located in the users's home directory.

Currently that command is:

```bash
odo utils config [create|update|delete|list|add|remove|get|describe] [KEY] [VALUES] [--local|--local]
```

In order to create configuration, we should call odo with the appropriate key and value or list of values to apply, as well as the scope for this configuration to be saved (local or global).

```bash
odo utils config set memory 2Gi               # Sets component memory to 2 GB local config
odo utils config add odoignore .odo --local   # Adds a value to odoignore list local config
odo utils config add odoignore .odo --global  # Adds a value to odoignore list global config
```

As can be seen, when a scope is omitted, the value will be applied to the local configuration store.

The verbs that add configuration are:

* *create*: Creates a configuration entry with the value provided. If the key exists, this command should error with the corresponding message.
* *add*: Adds a value to an existing config entry. The type of the entry should be a list. If the key does not exist, or is not a list, this command should error with the corresponding error message.

**NOTE**: When creating a component, no configuration file will be created. If a developer thinks a configuration value is worth being stored in the `.odo/odo-config` configuration file, he will need to use the previous command to create these files (whether global or local doesn't matter).

#### Listing configuration

When a user wants to know what specific configuration exists, all or just a specific key.

```bash
odo utils config list --global      # List all global config
odo utils config get memory         # List the value for memory config key
```

The verbs that add configuration are:

* *list*: Lists all configuration. No further argument.
* *get*: Lists configuration for a specific entry. If the key does not exist this command should error with the corresponding message.

#### Updating configuration

When a user wants to update specific configuration.

```bash
odo utils config update app-name example   # Updates app-name value
odo utils config add odoignore .eclipse    # Adds .eclipse to the list of ignored files
```

The verbs that add configuration are:

* *update*: Updates a configuration key. If it's a list, replaces all the values in the previous version with the ones provided. If the key does not exist, this command should error with the corresponding message.
* *add*: Adds a value to a configuration key entry that is a list of values. If the key does not exist or is not a list, this command will error with the corresponding message.

#### Deleting configuration

When a user wants to update specific configuration.

```bash
odo utils config delete memory --local                # Removes configuration for memory
odo utils config remove odoignore .git --local        # Removes .git from the list of values for odoignore
```

* *delete*: Deletes an entire configuration key with all it's values (in case of a list). If the key does not exist, this command should error with the corresponding message.
* *remove*: Removes a value from a configuration key entry that is a list of values.  If the key does not exist or is not a list or the value does not exist, this command will error with the corresponding message.

#### Describe configuration

To get information about a specific configuration key:

```bash
odo utils config describe memory        #Provides a description of the memory configuration key
```

* *describe*: Describe the meaning of a configuration key and the possible values and the types for the value.

## Future evolution

WIP
