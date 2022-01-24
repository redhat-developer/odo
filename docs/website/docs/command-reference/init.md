---
title: odo init
sidebar_position: 6
---

The `odo init` command is the first command to be executed when you want to bootstrap a new development, using `odo`. If sources already exist,
the command `odo dev` should be considered instead.

This command must be executed from an empty directory, and as a result, the command will download a `devfile.yaml` file and, optionally, a starter project.

The command can be exectued in two flavors, either interactive or non-interactive.

## Interactive mode

In interactive mode, you will be guided to choose a devfile from the list of devfiles present in the registry or registries referenced (using the `odo registry` command) and a starter project referenced by the selected devfile, and you will be asked for a name for the component present in the devfile.

## Non-interactive mode

In non-interactive mode, you will have to specify from the command-line the information needed to get a devfile.

If you want to download a devfile from a registry, you must specify the devfile name with the `--devfile` flag. The devfile with the specified name will be searched into the registries referenced (using `odo registry`), and the first one matching will be downloaded. If you want to download the devfile from a specific registry in the list or referenced registries, you can use the `--devfile-registry` flag to specify the name of this registry.

If you prefer to download a devfile from an URL or from the local filesystem, you can use the `--devfile-path` instead.

The `--starter` flag indicates the name of the starter project (as referenced in the selected devfile), that you want to use to start your development.

The required `--name` flag indicates how will be named the component present in the devfile.

## Examples

### Interactive mode

```
$ odo init
? Select language: go
? Select project type: Go Runtime (go, registry: DefaultDevfileRegistry)
? Which starter project do you want to use? go-starter
? Enter component name: my-go-app
 ✓  Downloading devfile "go" from registry "DefaultDevfileRegistry" [944ms]
 ✓  Downloading starter project "go-starter" [622ms]

Your new component "my-go-app" is ready in the current directory.
To start editing your component, use "odo dev" and open this folder in your favorite IDE.
Changes will be directly reflected on the cluster.
To deploy your component to a cluster use "odo deploy".
```

### Non-interactive mode from any registry of the list

In this example, the devfile will be downloaded from the **Staging** registry, which is the first one in the list containing the `go` devfile.

```
$ odo registry list
NAME                       URL                                   SECURE
Staging                    https://registry.stage.devfile.io     No
DefaultDevfileRegistry     https://registry.devfile.io           No

$ odo init --name my-go-app --devfile go --starter go-starter
 ✓  Downloading devfile "go" [948ms]
 ✓  Downloading starter project "go-starter" [408ms]

Your new component "my-go-app" is ready in the current directory.
To start editing your component, use "odo dev" and open this folder in your favorite IDE.
Changes will be directly reflected on the cluster.
To deploy your component to a cluster use "odo deploy".
```

### Non-interactive mode from a specific registry of the list

In this example, the devfile will be downloaded from the **DefaultDevfileRegistry** registry, as explicitely indicated by the `--devfile-registry` flag.

```
$ odo registry list
NAME                       URL                                   SECURE
Staging                    https://registry.stage.devfile.io     No
DefaultDevfileRegistry     https://registry.devfile.io           No

$ odo init --name my-go-app --devfile go --devfile-registry DefaultDevfileRegistry --starter go-starter
 ✓  Downloading devfile "go" from registry "DefaultDevfileRegistry" [1s]
 ✓  Downloading starter project "go-starter" [405ms]

Your new component "my-go-app" is ready in the current directory.
To start editing your component, use "odo dev" and open this folder in your favorite IDE.
Changes will be directly reflected on the cluster.
To deploy your component to a cluster use "odo deploy".
```

### Non-interactive mode from a URL

```
$ odo init --devfile-path https://registry.devfile.io/devfiles/nodejs-angular --name my-nodejs-app --starter nodejs-angular-starter
 ✓  Downloading devfile from "https://registry.devfile.io/devfiles/nodejs-angular" [415ms]
 ✓  Downloading starter project "nodejs-angular-starter" [484ms]

Your new component "my-nodejs-app" is ready in the current directory.
To start editing your component, use "odo dev" and open this folder in your favorite IDE.
Changes will be directly reflected on the cluster.
To deploy your component to a cluster use "odo deploy".
```
