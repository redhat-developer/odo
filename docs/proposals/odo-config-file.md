# `odo` config file

## Abstract

When developers work on creating an application, there's 2 phases they go through. First phase is creating the application initial implementation from scratch, and then iterate development on an existing application. While the developer works on the application, he might take some decisions on how the application should behave that should be preserved over time, but that are not source code but rather specific to how the application should behave in relation to the platform where it's running or in relation to the tool used to develop it.

IDEs often use a configuration file to hold specific configuration for an application, workspace or the whole tool. These files are stored under the same control version system as where the source code lives, so that anytime a developer checks out that repository has control on how to configure that tool to work with that source code.

Git is another example of a tool that provides configuration, that can be related of every instance of any repository the user has or to a single instance. These are defined as global and local config. Git promoves the fact that local configuration is always more important than global configuration.

## Motivation

Provide a way to preserve conscisous decisions the developer has made that are orthogonal to the code but relevant to the tool/platform. Any user that continues working on an existing application don't need to know the decisions the original developer took, as these are preserved along with the source code and provides for better reproducibility of the development environment setup.

## Use Cases

1. As a user, I want to be able to preserve some configuration I've done about my current component along with my source code
1. As a user, I want to be able to use configuration stored with my source code to help me be more productive
1. As a user, I want to be able to manage explicit configuration that is saved with my source code
1. As a user, I want to be able to check out an application source code on different computers and have them producing the same result in relation to the development environment (deployment of the application on the platform)

## Design overview

We need to provide a configuration file, `.odo`, that will hold specific configuration that `odo` will use on execution. This configuration might modify how the CLI work or how the component will work on the platform.

### Scopes of configuration

`odo` CLI configuration needs to be preserved in configuration files. As a developer might work on multiple
components, there is a need to have different scopes for configuration so that a developer don't need to
explicitly configure each and every component independently.

#### Local configuration

Every component can be configured independently. We call this local configuration for the component as it only affects the component it references. This configuration will be stored in a file in the root directory of the component's source code. In this way, the configuration file can be added to version control system and preserved along with the application. If any other developer checks out the same source code he will get the same configuration as the original developer.

A good representation of this is a developer that while on the development process realizes his application requires 2 GB of memory to run appropriately, he might decide to save that information in this configuration file. At a later moment, any other developer (or even him in the future) that decides to work on the same application will have this configuration definition along with his source code. `odo` should take into account this configuration definition and materialize when appropriate. The type of configuration, format and applicability will be described later in this document.

#### Global configuration

When working with multiple components it might be tedious to configure each and every component, or a user wants to have a default configuration value applied to each and every component he works with when no explicit (local) configuration exists. For this purpose, there should be a file, stored in the user's home directory, that should contain this configuration definitions. This configuration is global to every component the user works with.

#### User explicit configuration

Often times a user wants to use different configuration values than the provided in the configuration file. We need to provide a way to allow a developer to specify this new values he want to use. For that, the user needs to explicitly provide he values as CLI arguments to the speficic `odo` command.

An example of this would be:

```bash
odo create wildfly --memory 2GB
```

In this previous example, the user is explicitly instructing `odo` to use 2GB of memory when creating a component based on wildfly.

### <a name="configuration-precedence"></a> Configuration precedence

As configuration can exist both localy and globaly or only on one of these scopes, it seems obvious that there needs to be some rule managing the precedence and applicability of the configuration.

The most explicit/concrete configuration wins over the less explicit one. In this case, providing a configuration value via a CLI argument always wins over any other existing configuration. If no explicit value is provided, then the configuration defined along with the component's source code wins over any other configuration. In case there is no explicit configuration or configuration for that single component then any global configuration (if existing) will be applied. In any other case, `odo`'s defaults will be used.

|global|local|CLI flag|applied value|
|--|--|--|-|
|--|--|--|odo defaults|
|Value A|--|--|Value A|
|Value A|--|Value X|Value X|
|Value A|Value B|--|Value B|
|Value A|Value B|Value X|Value X|
|--|Value B|--|Value B|
|--|Value B|Value X|Value X|
|--|-|Value X|Value X|

### Configuration typology

In the scop of this proposal, we refer to configuration in some different ways and it's worth noting what we meant in each case. What types of configuration will be supported and what use case they fullfil.

### `odo` CLI behavior configuration

Like every tool, there is configuration relative to how the tool behaves itself. An example of this would be a timeout that the CLI will apply when issuing commands to the server, or the ability to enable/disable checks for new versions. This configuration will always be global, and stored relative to the user's home directory. By default this configuration file will live in `$HOME/.kube/odo`

The [format](#odo-file-format) of this file is described later in the proposal.

### Application related configuration

Some times what we want to configure is how the application/component will be deployed. Some components will
require special characteristics, like a specific amount of memory. Some other times, what we want is to instruct `odo` to use an specific value to name our component if none is explicitly provided. This configuration will be stored locally to the component, along with the source code, in an `.odo` configuration file that can be saved into the version control system used, and share by anyone using the same application code.

For some specific values, there will be the possibility to set them globally, as they will be used in that case for every component created, as defined in the [configuration precedence](#configuration-precedence) section. This values will be documented as being `global` configuration values.

### Developer friendly flags

One of the goals of `odo` is to be easy to use for developers. This forces to translate some configuration possible in the underlying technology (OpenShift/Kubernetes) into a friendlier name that developers could easily understand. An example of this would be the amount of memory a component needs. In OpenShift, this is reflected by limits.memory and requests.memory, and dependening how you set these values influence the QoS of the deployment. In `odo` these should be simplified to `memory`, `min-memory` and `max-memory`. These are terms that any developer will most likely understand without looking at the documentation.

* When `memory` is set, memory requests and limits are the same.
* `min-memory` should be memory requests.
* `max-memory` should be memory limits.
* `memory` is exclusive with `min-memory` or `max-memory`. If set, min-mem and max-mem have the same value as `memory`.

#### Proposed flags

We need to define and delimit what information will be stored in this file. Configuration which can be local and global will be denoted. Otherwise, the configuration will only be local to the application .odo file.

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

* app-name: (local) Preferred application name if none is defined
* component-name: (local) Preferred component name if none is defined
* odoignore: (local/global) list of files/patterns to ignore by default if no .odoignore file is provided, or when it's cerated. **This configuration can be global, as the developer might want to always ignore specific files (e.g. .git, .odo, ...) when pushing.**

#### <a name="odo-file-format"></a> `odo` file format

The format for both files should be human readable. Whether it's yaml, json or [toml](https://github.com/toml-lang/toml) is not the scope of this proposal to define, as it will be dictated by the engineering team implementing this feature. In any case, it should be possible to:

* Visually interpret the content of a file (vi, cat)
* Process the file in an easy way (grep, aqk, sed, jq|yq)
* Be stored (and visualized) in version control system
* Be used in any operating system (Linux, mac OS, windows)

Nevertheless, `odo` will provide a way to manipulate this file via a CLI command, as described in [managing configuration](#managing-configuration)

### <a name="managing-configuration"></a> Managing configuration

Developers will need to add configuration values to the `.odo` config files. This configuration management will be managed by a new command in `odo` CLI. As configuration will be of different types, there will be needs to additional verbs (to the regular ones) for managing this configuration.

#### Creating configuration

There might be times that a developer would want to add/edit/remove values from the .odo configuration file. There should be a command to manage this file in the `odo` CLI. That command should be the same used to touch configuration of the global `odo` config file located in the users's home directory.

Currently that command is:

```bash
odo utils config [create|update|delete|list|add|remove|get|describe] [KEY] [VALUES] [--local|--local]
```

In order to create configuration, we should call odo with the appropriate key and value or list of values to apply, as well as the scope for this configuration to be saved (local or global).

```bash
odo utils config set memory 2Gi               # Set's component memory to 2 GB local config
odo utils config add odoignore .odo --local   # Adds a value to odoignore list local config
odo utils config add odoignore .odo --global  # Adds a value to odoignore list global config
```

As can be seen, when a scope is ommited, the value will be applied to the local configuration store.

The verbs that add configuration are:

* *create*: Creates a configuration entry with the value provided. If the key exists, this command should error with the corresponding message.
* *add*: Adds a value to an existing config entry. The type of the entry should be a list. If the key does not exist, or is not a list, this command should error with the corresponding error message.

*NOTE*: When creating a component, no configuration file will be created. If a developer thinks a configuration value is worth being stored in the `.odo` configuration file, he will need to use the previous command to create these files (whether global or local doesn't matter).

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
odo utils config list --global                # List all global config
```

The verbs that add configuration are:

* *update*: Updates a configuration key. If it's a list, replaces all the values in the previous version with the ones provided. If the key does not exist, this command should error with the corresponding message.
* *add*: Adds a value to a configuration key entry that is a list of values. If the key does not exist or is not a list, this command will error with the corresponding message.

#### Deleting configuration

When a user wants to update specific configuration.

```bash
odo utils config delete memory --local                # Removes configuration for memory
odo utils config remove odoignore .git --local        # Removes .git from the list of values for odoignore
```

* *delete*: Deletes an entire configuration key with all it's values (in case of a list). If the key does not exist, this command should error with the corresponding message.
* *remove*: Removes a value from a configuration key entry that is a list of values.  If the key does not exist or is not a list or the value does not exist, this command will error with the corresponding message.

#### Describe configuration

To get information about a specific configuration key:

```bash
odo utils config describe memory        #Provides a description of the memory configuration key
```

* *describe*: Describe the meaning of a configuration key and the possible values and the types for the value.

## Future evolution

WIP