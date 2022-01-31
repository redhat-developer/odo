# `odo dev`

[#5299](https://github.com/redhat-developer/odo/issues/5298)

Running the application on the cluster for **development** (inner loop)

The result of successful execution of `odo dev` is Devfile component deployed in inner loop mode.
**By default command deletes all resources that it created before it exits.**

When users runs command with `--cleanup=false` there will be no cleanup before existing. The message "Press Ctrl+c to exit and clean up resources from cluster." will be only "Press Ctrl+c to exit."

If user executed `odo dev --cleanup=false` and then run this command again. The first line of the output should display warning: "Reusing already existing resources".

When there is no devfile in the current directory and no flags were provided command starts in interactive mode to guide user through the devfile selection.

When some flags were provided and there is no devfile in the current directory command exits with error:
```
No devfile.yaml in the current directory. Use `odo init component` to get devfile.yaml for your application."
```

When devfile exists in the current directory deploy application in inner loop mode using the information form `devfile.yaml` in the current directory.


## Flags

- `-o` (string) output information in a specified format (json).
- `--watch` (boolean) Run command in watch mode. In this mode command is watching the current directory for file changes and automatically sync change to the container where it rebuilds and reload the application.
  By default, this is `true` (`--watch=true`). You can disable watch using `--watch=false`
- `--cleanup` (boolean). default is `true` when user presses `ctrl+c` it deletes all resource that it created on the cluster.

**`--watch` and `--cleanup` flags will be added later. v3.0.0-alpha1 won't have those flags.**


## Interactive mode
```
$ odo dev
There is no devfile.yaml in the current directory.

Based on the files in the current directory odo detected
Language: Java
Project type: SpringBoot

? Is this correct? Yes

Current component configuration:
Opened ports:
- 8080
- 8084
Environment variables:
- FOO = bar
- FOO1 = bar1

? What configuration do you want change?  [Use arrows to move, type to filter]
> NOTHING - configuration is correct
  Delete port "8080"
  Delete port "8084"
  Add new port
  Delete environment variable "FOO"
  Delete environment variable "FOO1"
  Add new environment variable


⠏ Downloading "java-quarkus". DONE
Starting your application on cluster in developer mode ...
⠏ Waiting for Kubernetes resources ... DONE
⠏ Syncing files into the container ... DONE
⠏ Building your application in container on cluster ... DONE
⠏ Execting the application ... DONE
Your application is running on cluster.
You can access it at https://example.com

⠏ Watching for changes in the current directory ... DONE
Change in main.java detected.
⠏ Syncing files into the container ... DONE
⠏ Reloading application ... DONE
⠏ Watching for changes in the current directory ...

Press Ctrl+c to exit and clean up resources from cluster.

<ctrl+c>

⠏ Cleaning up ... DONE
```

Questions and their behavior is the same as for [`odo deploy`](odo-deploy.md) command
