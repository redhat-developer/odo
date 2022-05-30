---
title: odo describe
---

`odo describe component` command is useful for getting information about a component managed by `odo`. 

There are 2 ways to describe a component:
- [Describe with access to Devfile](#describe-with-access-to-devfile)
- [Describe without access to Devfile](#describe-without-access-to-devfile)

## Describe with access to Devfile
```shell
odo describe component
```

This command returns information extracted from the Devfile:
- metadata (name, display name, project type, language, version, description and tags)
- supported odo features, indicating if the Devfile defines necessary information to run `odo dev`, `odo dev --debug` and `odo deploy`
- the list of container components,
- the list of Kubernetes components.

The command also displays if the component is currently running in the cluster on Dev and/or Deploy mode.

## Describe without access to Devfile
```shell
odo describe component --name <component_name> [--namespace <namespace>]
```

The command extracts information from the labels and annotations attached to the deployed component to display the known metadata of the Devfile used to deploy the component.

The command also displays if the component is currently running in the cluster on Dev and/or Deploy mode.
