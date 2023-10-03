---
title: odo run
---

`odo run` is used to manually execute commands defined in a Devfile.

<details>
<summary>Example</summary>

A command `connect` is defined in the Devfile, executing the `bash` command in the `runtime` component.

```yaml
schemaVersion: 2.2.0
[...]
commands:
  - id: connect
    exec:
      component: runtime
      commandLine: bash
  [...]

```

```shell
$ odo run connect
bash-4.4$ 
```

</details>


For `Exec` commands, `odo dev` needs to be running, and `odo run` 
will execute commands in the containers deployed by the `odo dev` command. 

Standard input is redirected to the command running in the container, and the terminal is configured in Raw mode. For these reasons, any character will be redirected to the command in container, including the Ctrl-c character which can thus be used to interrupt the command in container.

The `odo run` command terminates when the command in container terminates, and the exit status of `odo run` will reflect the exit status of the distant command: it will be `0` if the command in container terminates with status `0` and will be `1` if the command in container terminates with any other status.

Resources deployed with `Apply` commands will be deployed in *Dev mode*, 
and these resources will be deleted when `odo dev` terminates.

