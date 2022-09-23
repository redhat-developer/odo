---
title: odo logs
---

`odo logs` is used to display the logs for all the containers odo created for the component under current working 
directory.

## Running the command 

If you haven't already done so, you must [initialize](../command-reference/init) your source code with the `odo 
init` command. 

```shell
odo logs [--follow] [--dev | --deploy]
```
<details>
<summary>Example</summary>

```shell
$ odo logs
runtime: npm WARN nodejs-starter@1.0.0 No repository field.
runtime:
runtime: added 64 packages from 57 contributors and audited 64 packages in 7.761s
runtime: found 0 vulnerabilities
runtime:
runtime:
runtime: > nodejs-starter@1.0.0 start /projects
runtime: > node server.js
runtime:
runtime: App started on PORT 3000
main: Wed Sep 21 08:26:27 UTC 2022 - this is infinite while loop
main: Wed Sep 21 08:26:32 UTC 2022 - this is infinite while loop
main: Wed Sep 21 08:26:37 UTC 2022 - this is infinite while loop
main: Wed Sep 21 08:26:42 UTC 2022 - this is infinite while loop
main: Wed Sep 21 08:26:47 UTC 2022 - this is infinite while loop
main: Wed Sep 21 08:26:52 UTC 2022 - this is infinite while loop
main: Wed Sep 21 08:26:57 UTC 2022 - this is infinite while loop
main: Wed Sep 21 08:27:02 UTC 2022 - this is infinite while loop
main: Wed Sep 21 08:27:07 UTC 2022 - this is infinite while loop
main: Wed Sep 21 08:27:12 UTC 2022 - this is infinite while loop
main: Wed Sep 21 08:27:17 UTC 2022 - this is infinite while loop
main: Wed Sep 21 08:27:22 UTC 2022 - this is infinite while loop
```
</details>

`odo logs` command can be used with the following flags:
* Use `odo logs --dev` to see the logs for the containers created by `odo dev` command.
* Use `odo logs --deploy` to see the logs for the containers created by `odo deploy` command.
* Use `odo logs` (without any flag) to see the logs of all the containers created by both `odo dev` and `odo deploy`.

Note that if multiple containers are named the same (for example, `main`), the `odo logs` output appends a number to 
container name to help differentiate between the containers. In the output, you will see containers named as `main`, 
`main[1]`, `main[2]`, so on and so forth.

It also supports `--follow` flag which allows you to follow/tail/stream the logs of the containers. It works by using 
the same commands as above albeit, with a `--follow` flag:
* Use `odo logs --dev --follow` to follow the logs for the containers created by `odo dev` command.
* Use `odo logs --deploy --follow` to follow the logs for the containers created by `odo deploy` command.
* Use `odo logs --follow` (without `--dev` or `--deploy`) to follow the logs of all the containers created by both `odo 
  dev` and `odo deploy`.
