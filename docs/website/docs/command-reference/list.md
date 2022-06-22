---
title: odo list
---

`odo list` command is useful for getting information about components running on a specific namespace.

If the command is executed from a directory containing a Devfile, it also displays the command
defined in the Devfile as part of the list, prefixed with a star(*).

For each component, the command displays:
- its name,
- its project type,
- on which mode it is running (None, Dev, Deploy, or both), not that None is only applicable to the component 
defined in the local Devfile,
- by which application the component has been deployed.

## Available flags

* `--namespace` - Namespace to list the components from (optional). By default, the current namespace defined in kubeconfig is used
* `-o json` - Outputs the list in JSON format. See [JSON output](json-output.md) for more information
