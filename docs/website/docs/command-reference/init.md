---
title: odo init
---

The `odo init` command is the first command to be executed when you want to bootstrap a new component, using `odo`. If sources already exist,
the command `odo dev` should be considered instead.

This command must be executed from a directory with no `devfile.yaml` file.

The command can be executed in two flavors, either interactive or non-interactive.

## Running the command
### Interactive mode

In interactive mode, the behavior of `odo init` depends on whether the current directory already contains source code or not.

#### Empty directory

If the directory is empty, you will be guided to:
- choose a devfile from the list of devfiles present in the registry or registries referenced (using the `odo registry` command),
- configure the devfile
- choose a starter project referenced by the selected devfile,
- choose a name for the component present in the devfile; this name must follow the [Kubernetes naming convention](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#dns-label-names) and not be all-numeric.

```console
odo init
```

<details>
<summary>Example</summary>

import EmptyDirOutput from './docs-mdx/init/interactive_mode_empty_directory_output.mdx';

<EmptyDirOutput />

</details>

#### Directory with sources

If the current directory is not empty, `odo init` will make its best to autodetect the type of application and propose you a Devfile that should suit your project.
It will try to detect the following, based on the files in the current directory:
- Language
- Project Type
- Ports used in your application
- A Devfile that should help you start with `odo` 

If the information detected does not seem correct to you, you are able to select a different Devfile.

In all cases, you will be guided to:
- configure the devfile
- choose a name for the component present in the devfile; this name must follow the [Kubernetes naming convention](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#dns-label-names) and not be all-numeric.

```console
odo init
```

<details>
<summary>Example</summary>

import NonEmptyDirectoryOutput from './docs-mdx/init/interactive_mode_directory_with_sources_output.mdx'

<NonEmptyDirectoryOutput />
</details>

### Non-interactive mode

In non-interactive mode, you will have to specify from the command-line the information needed to get a devfile.

If you want to download a devfile from a registry, you must specify the devfile name with the `--devfile` flag. The devfile with the specified name will be searched in the registries referenced (using `odo preference view`), and the first one matching will be downloaded.
If you want to download the devfile from a specific registry in the list or referenced registries, you can use the `--devfile-registry` flag to specify the name of this registry. By default odo uses official devfile registry [registry.devfile.io](https://registry.devfile.io). You can use registry's [web interface](https://registry.devfile.io/viewer) to view its content.
If you want to download a version devfile, you must specify the version with `--devfile-version` flag.

If you prefer to download a devfile from an URL or from the local filesystem, you can use the `--devfile-path` instead.

The `--starter` flag indicates the name of the starter project (as referenced in the selected devfile), that you want to use to start your development. To see the available starter projects for devfile stacks in the official devfile registry use its [web interface](https://registry.devfile.io/viewer) to view its content.  

The required `--name` flag indicates how the component initialized by this command should be named. The name must follow the [Kubernetes naming convention](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#dns-label-names) and not be all-numeric.

#### Fetch Devfile from any registry of the list

In this example, the devfile will be downloaded from the **StagingRegistry** registry, which is the first one in the list containing the `nodejs-react` devfile.
```shell
odo init --name <component-name> --devfile <devfile> [--starter STARTER]
```
<details>
<summary>Example</summary>

<RegistryOutput />

import RegistryListOutput from './docs-mdx/init/registry_list_output.mdx'

<RegistryListOutput />

import DevfileFromAnyRegistryOutput from './docs-mdx/init/devfile_from_any_registry_output.mdx'

<DevfileFromAnyRegistryOutput />

</details>


#### Fetch Devfile from a specific registry of the list

In this example, the devfile will be downloaded from the **DefaultDevfileRegistry** registry, as explicitly indicated by the `--devfile-registry` flag.
<details>
<summary>Example</summary>

import RegistryOutput from './docs-mdx/init/registry_output.mdx'

<RegistryOutput />

import DevfileFromSpecificRegistryOutput from './docs-mdx/init/devfile_from_specific_registry_output.mdx';

<DevfileFromSpecificRegistryOutput />

</details>


#### Fetch Devfile from a URL

```console
odo init --devfile-path <URL> --name <component-name> [--starter STARTER]
```
<details>
<summary>Example</summary>

import DevfileFromURLOutput from './docs-mdx/init/devfile_from_url_output.mdx';

<DevfileFromURLOutput />

</details>

#### Fetch Devfile of a specific version

```console
odo init --devfile <devfile-name> --devfile-version <devfile-version> --name <component-name> [--starter STARTER]
```

<details>
<summary>Examples</summary>

import VersionedOutput from './docs-mdx/init/versioned_devfile_output.mdx';

<VersionedOutput />

</details>

import LatestVersionedOutput from './docs-mdx/init/latest_versioned_devfile_output.mdx';

:::note
Use "latest" as the version name to fetch the latest version of a given Devfile.

<details>
<summary>Example</summary>

<LatestVersionedOutput />

</details>
:::
