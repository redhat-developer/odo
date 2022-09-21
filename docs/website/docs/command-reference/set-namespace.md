---
title: odo set namespace
---

`odo set namespace` lets you set a namespace/project as the current active one in your local `kubeconfig` configuration.

## Running the command
To set the current active namespace you can run `odo set namespace <name>`:
```console
odo set namespace mynamespace
```

<details>
<summary>Example</summary>

```console
$ odo set namespace mynamespace
 ✓  Current active namespace set to "mynamespace"
```
</details>

Optionally, you can also use `project` as an alias to `namespace`.

To set the current active project you can run `odo set project <name>`:
```console
odo set project myproject
```

<details>
<summary>Example</summary>

```console
$ odo set project myproject
  ✓  Current active project set to "myproject"
```
</details>

:::tip
This command updates your current `kubeconfig` configuration, using either of the aliases.
So running either `odo set project` or `odo set namespace` performs the exact same operation in your configuration.
:::
