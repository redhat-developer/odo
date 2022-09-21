---
title: odo init
---

The `odo init` command is the first command to be executed when you want to bootstrap a new component, using `odo`. If sources already exist,
the command `odo dev` should be considered instead.

This command must be executed from an empty directory, and as a result, the command will download a `devfile.yaml` file and, optionally, a starter project.

The command can be executed in two flavors, either interactive or non-interactive.

## Running the command
### Interactive mode

In interactive mode, you will be guided to:
- choose a devfile from the list of devfiles present in the registry or registries referenced (using the `odo registry` command),
- configure the devfile if there is an existing project
- choose a starter project referenced by the selected devfile,
- choose a name for the component present in the devfile; this name must follow the [Kubernetes naming convention](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#dns-label-names) and not be all-numeric.

```console
odo init
```

<details>
<summary>Example</summary>

```console
$ odo init
? Select language: java
? Select project type: Maven Java (java-maven, registry: DefaultDevfileRegistry)
? Which starter project do you want to use? springbootproject
? Enter component name: my-java-maven-app
 ✓  Downloading devfile "java-maven" from registry "DefaultDevfileRegistry" [949ms]
 ✓  Downloading starter project "springbootproject" [430ms]

Your new component "my-java-maven-app" is ready in the current directory.
To start editing your component, use "odo dev" and open this folder in your favorite IDE.
Changes will be directly reflected on the cluster.
To deploy your component to a cluster use "odo deploy".
```
</details>


### Non-interactive mode

In non-interactive mode, you will have to specify from the command-line the information needed to get a devfile.

If you want to download a devfile from a registry, you must specify the devfile name with the `--devfile` flag. The devfile with the specified name will be searched in the registries referenced (using `odo preference view`), and the first one matching will be downloaded. If you want to download the devfile from a specific registry in the list or referenced registries, you can use the `--devfile-registry` flag to specify the name of this registry. By default odo uses official devfile registry [registry.devfile.io](https://registry.devfile.io). You can use registry's [web interface](https://registry.devfile.io/viewer) to view its content.

If you prefer to download a devfile from an URL or from the local filesystem, you can use the `--devfile-path` instead.

The `--starter` flag indicates the name of the starter project (as referenced in the selected devfile), that you want to use to start your development. To see the available starter projects for devfile stacks in the official devfile registry use its [web interface](https://registry.devfile.io/viewer) to view its content.  

The required `--name` flag indicates how the component initialized by this command should be named. The name must follow the [Kubernetes naming convention](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#dns-label-names) and not be all-numeric.

#### Non-interactive mode from any registry of the list

In this example, the devfile will be downloaded from the **Staging** registry, which is the first one in the list containing the `nodejs-react` devfile.
```shell
odo init --name <component-name> --devfile <devfile> [--starter STARTER]
```
<details>
<summary>Example</summary>

```console
$ odo preference view
[...]

Devfile registries:
 NAME                       URL                                   SECURE
 Staging                    https://registry.stage.devfile.io     No
 DefaultDevfileRegistry     https://registry.devfile.io           No
```
</details>


#### Non-interactive mode from a specific registry of the list

In this example, the devfile will be downloaded from the **DefaultDevfileRegistry** registry, as explicitly indicated by the `--devfile-registry` flag.
<details>
<summary>Example</summary>

```console
$ odo preference view
[...]

Devfile registries:
 NAME                       URL                                   SECURE
 Staging                    https://registry.stage.devfile.io     No
 DefaultDevfileRegistry     https://registry.devfile.io           No
```

```shell
$ odo init --name my-spring-app --devfile java-springboot --devfile-registry DefaultDevfileRegistry --starter springbootproject
 ✓  Downloading devfile "java-springboot" from registry "DefaultDevfileRegistry" [980ms]
 ✓  Downloading starter project "springbootproject" [399ms]

Your new component "my-spring-app" is ready in the current directory.
To start editing your component, use "odo dev" and open this folder in your favorite IDE.
Changes will be directly reflected on the cluster.
To deploy your component to a cluster use "odo deploy".
```
</details>


#### Non-interactive mode from a URL

```console
odo init --devfile-path <URL> --name <component-name> [--starter STARTER]
```
<details>
<summary>Example</summary>

```console
$ odo init --devfile-path https://registry.devfile.io/devfiles/nodejs-angular --name my-nodejs-app --starter nodejs-angular-starter
 ✓  Downloading devfile from "https://registry.devfile.io/devfiles/nodejs-angular" [415ms]
 ✓  Downloading starter project "nodejs-angular-starter" [484ms]

Your new component "my-nodejs-app" is ready in the current directory.
To start editing your component, use "odo dev" and open this folder in your favorite IDE.
Changes will be directly reflected on the cluster.
To deploy your component to a cluster use "odo deploy".
```
</details>

