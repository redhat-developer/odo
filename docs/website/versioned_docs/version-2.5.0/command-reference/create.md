---
title: odo create
sidebar_position: 3
---

odo uses the [_devfile_](https://devfile.io) to store the configuration of and describe the resources like storage, services, etc. of a component. The _odo create_ command allows you to generate this file.

## Creating a component

To create a _devfile_ for an existing project, you can execute `odo create` with the name and type of your component (for example, nodejs or go):

```
odo create nodejs mynodejs
```

Here `nodejs` is the type of the component and `mynodejs` is the name of the component odo creates for you.

> Note: for a list of all the supported component types, run `odo catalog list components`.

If your source code exists outside the current directory, the `--context` flag can be used to specify the path. For example, if the source for the nodejs component was in a folder called `node-backend` relative to the current working directory, you could run:

```
odo create nodejs mynodejs --context ./node-backend
```

Both relative and absolute paths are supported.

To specify the project or app of where your component will be deployed, you can use the `--project` and `--app` flags.

For example, to create a component that is a part of the `myapp` app inside the `backend` project:

```
odo create nodejs --app myapp --project backend
```

> Note: if these are not specified, they will default to the active app and project

## Starter projects

If you do not have existing source code but wish to get up and running quickly to experiment with devfiles and components, you could use the starter projects to get started. To use a starter project, include the `--starter` flag in your `odo create` command.

To get a list of available starter projects for a component type, you can use the `odo catalog describe component` command. For example, to get all available starter projects for the nodejs component type, run: 

```
odo catalog describe component nodejs
```

Then specify the desired project with the `--starter` flag: 

```
odo create nodejs --starter nodejs-starter
```

This will download the example template corresponding to the chosen component type (in the example above, `nodejs`) in your current directory (or the path provided with the `--context` flag).

If a starter project has its own devfile, then this devfile will be preserved.

## Using an existing devfile

If you want to create a new component from an existing devfile, you can do so by specifying the path to the devfile with the `--devfile` flag.

For example, the following command will create a component called `mynodejs`, based on the devfile from GitHub:

```
odo create mynodejs --devfile https://raw.githubusercontent.com/odo-devfiles/registry/master/devfiles/nodejs/devfile.yaml
```

## Interactive creation

The `odo create` command can also be run interactively. Execute `odo create`, which will guide you through a list of steps to create a component:

```sh
odo create

? Which devfile component type do you wish to create go
? What do you wish to name the new devfile component go-api
? What project do you want the devfile component to be created in default
Devfile Object Validation
 ✓  Checking devfile existence [164258ns]
 ✓  Creating a devfile component from registry: DefaultDevfileRegistry [246051ns]
Validation
 ✓  Validating if devfile name is correct [92255ns]
? Do you want to download a starter project Yes

Starter Project
 ✓  Downloading starter project go-starter from https://github.com/devfile-samples/devfile-stack-go.git [429ms]

Please use `odo push` command to create the component with source deployed
```

You will be prompted to choose the component type, name and the project for the component. You can also choose whether or not to download a starter project. Once finished, a new `devfile.yaml` file should be created in the working directory.
To deploy these resources to your cluster, run `odo push`.
