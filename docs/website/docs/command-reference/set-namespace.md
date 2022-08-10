---
title: odo set namespace
---

`odo set namespace` lets you set a namespace/project as the current active one in your local `kubeconfig` configuration.

:::note
Executing this command inside a component directory will not update the namespace or project of the existing component.
:::

To set the current active namespace you can run `odo set namespace <name>`:
```console
odo set namespace mynamespace
```
```console
$ odo set namespace mynamespace
 ✓  Current active namespace set to "mynamespace"
```

Optionally, you can also use `project` as an alias to `namespace`.

To set the current active project you can run `odo set project <name>`:
```console
odo set project myproject
```
```console
$ odo set project myproject
  ✓  Current active project set to "myproject"
```

:::note
This command updates your current `kubeconfig` configuration, using either of the aliases.
So running either `odo set project` or `odo set namespace` performs the exact same operation in your configuration.
:::
