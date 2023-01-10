---
title: odo delete component
toc_min_heading_level: 2
toc_max_heading_level: 4
---

`odo delete component` command is useful for deleting resources that are managed by `odo`.
By default, it deletes the component and its related inner-loop, and outer-loop resources from the cluster.
But the `running-in` flag allows to be more specific about which resources (either inner-loop or outer-loop) to delete.

## Running the command
There are 2 ways to delete a component:
- [Delete with access to Devfile](#delete-with-access-to-devfile)
- [Delete without access to Devfile](#delete-without-access-to-devfile)

### Delete with access to Devfile
```shell
odo delete component [--force] [--wait]
```
<details>
<summary>Example</summary>

import DeleteWithAccessToDevfileOutput from './docs-mdx/delete-component/delete_with_access_to_devfile.mdx'

<DeleteWithAccessToDevfileOutput />

</details>

`odo` looks into the Devfile _present in the current directory_ for the component resources for the innerloop, and outerloop.
If these resources have been deployed on the cluster, then `odo` will delete them after user confirmation.
Otherwise, `odo` will exit with a message stating that it could not find the resources on the cluster.

:::note
If some resources attached to the component are present on the cluster, but not in the Devfile, then they will not be deleted.
You can delete these resources by running the command in the [next section](#delete-without-access-to-devfile).
:::

#### Filtering resources to delete
You can specify the type of resources candidate for deletion via the `--running-in` flag.
Acceptable values are `dev` (for inner-loop resources) or `deploy` (for outer-loop resources).

<details>
<summary>Example</summary>

import DeleteRunningInWithAccessToDevfileOutput from './docs-mdx/delete-component/delete_running-in_with_access_to_devfile.mdx'

<DeleteRunningInWithAccessToDevfileOutput />

</details>

#### Deleting local files with `--files`

By default, `odo` does not delete the Devfile, the `odo` configuration files, or the source code.
But when `--files` is passed, `odo` attempts to delete files or directories it initially created locally.

This will delete the following files or directories:
- the `.odo` directory in the current directory
- optionally, the Devfile only if it was initially created via `odo` (initialization via any of the `odo init`, `odo dev` or `odo deploy` commands).

Note that `odo dev` might generate a `.gitignore` file if it does not exist in the current directory,
but this file will not be removed when `--files` is passed to `odo delete component`.

:::caution
Use this flag with caution because this permanently deletes the files mentioned above.
This operation is not reversible, unless your files are backed up or under version control.
:::

```shell
odo delete component --files [--force] [--wait]
```
<details>
<summary>Example</summary>

import DeleteWithFilesAndAccessToDevfileOutput from './docs-mdx/delete-component/delete_with_files_and_access_to_devfile.mdx'

<DeleteWithFilesAndAccessToDevfileOutput />

</details>

### Delete without access to Devfile
```shell
odo delete component --name <component_name> [--namespace <namespace>] [--force] [--wait]
```
<details>
<summary>Example</summary>

import DeleteNamedComponentOutput from './docs-mdx/delete-component/delete_named_component.mdx'

<DeleteNamedComponentOutput />

</details>


`odo` searches for resources attached to the given component in the given namespace on the cluster.
If `odo` finds the resources, it will delete them after user confirmation.
Otherwise, `odo` will exit with a message stating that it could not find the resources on the cluster.

`--namespace` is optional, if not provided, `odo` will use the current active namespace.

#### Filtering resources to delete
You can specify the type of resources candidate for deletion via the `--running-in` flag.
Acceptable values are `dev` (for inner-loop resources) or `deploy` (for outer-loop resources).

<details>
<summary>Example</summary>

import DeleteNamedComponentRunningInOutput from './docs-mdx/delete-component/delete_named_component_running-in.mdx'

<DeleteNamedComponentRunningInOutput />

</details>
