# `odo deploy`

[#5298](https://github.com/redhat-developer/odo/issues/5298)

The result of successful execution of `odo deploy` is Devfile component deployed in outer loop mode.

When there is no devfile in the current directory yet and no flags provided, command runs in  interactive mode to guide user through the devfile selection.

When some flags wre provided and there is no devfile in current directory, command exits with error:
```
No devfile.yaml in the current directory.
Use `odo create component` to get devfile.yaml for your application first.
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

If there is no devfile in the current directory `odo` should use ([Alizer](https://github.com/redhat-developer/alizer/pull/55/)) to get corresponding Devfile for code base in the current directory.
After the successful detection it will show  Language and Project Type information to users and ask for confirmation.
If user answers that the information is not correct, odo should ask "Select language" and "Select project type" questions (see: [`odo init`](odo-init.md) interactive mode).

The configuration part helps users to modify most common configurations done on Devfiles.
"? What configuration do you want change? " question is repeated over and over again until user confirms that the configuration is done and there is nothing else to change.

