# `odo deploy`

[#5298](https://github.com/redhat-developer/odo/issues/5298)

The result of successful execution of `odo deploy` is Devfile component deployed in outer loop mode.

When there is no devfile in the current directory yet and no flags provided, command runs in  interactive mode to guide user through the devfile selection.

When some flags wre provided and there is no devfile in current directory, command exits with error:
```
No devfile.yaml in the current directory.
Use `odo init component` to get devfile.yaml for your application first.
```

When devfile exists in the current directory, deploy application in outer loop mode using the information from `devfile.yaml` in the current directory.

## Flags

- `-o` (string) output information in a specified format (json).

## Interactive mode
```
$ odo deploy
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
Deploying your component to cluster ...
⠏ Building container image locally ... DONE
⠏ Pushing image to container registry ... DONE
⠏ Waiting for Kubernetes resources ... DONE
Your component is running on cluster.
You can access it at https://example.com
```


Questions and their behavior is the same as in [`odo init`](odo-init.md) command executed in non-empty directory.
