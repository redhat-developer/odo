---
title: odo logs
---

## Description

`odo logs` is used to display the logs for all the containers odo created for the component under current working 
directory.

## Running the command 

If you haven't already done so, you must [initialize](../command-reference/init) your source code with the `odo 
init` command. 

`odo logs` command can be used with the following flags:
* Use `odo logs --dev` to see the logs for the containers created by `odo dev` command.
* Use `odo logs --deploy` to see the logs for the containers created by `odo deploy` command.
* Use `odo logs` (without any flag) to see the logs of all the containers created by both `odo dev` and `odo deploy`.

Note that if multiple containers are named the same (for example, `main`), the `odo logs` output appends a number to 
container name to help differentiate between the containers. In the output, you will see containers named as `main`, 
`main[1]`, `main[2]`, so on and so forth.