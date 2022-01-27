
# `odo create component`



Create new Devfile component in the current directory.
Some users might prefer creating components separately before running `odo dev` or `odo deploy`.
But this command is mainly intended to be used by IDE plugins or scripts.





When devfile exists in the current directory command should exit with error:
```
Unable to create new component in the current directory.
The current directory already contains `devfile.yaml` file.
```

When there is no devfile in the current directory download devfile.yaml based on the input provided by user and place it into the current directory.

## Flags

- `--name` - name of the component (required)
- `-o` output information in a specified format (json).
- `--devfile` - name of the devfile in devfile registry (required if `--devfile-path` is not defined)
- `--devfile-registry` - name of the devfile registry (as configured in `odo registry`). It can be used in combination with `--devfile`, but not with `--devfile-path` (optional)
- `--devfile-path` - path to a devfile. This is alternative to using devfile from Devfile registry. It can be local file system path or http(s) URL (required if `--devfile` is not defined)



## Interactive mode
```
$ odo create

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

Your component is ready in the current directory.
To deploy it to the cluster you can run `odo deploy`.
```

The questions for interactive mode are identical to questions in `odo dev` and `odo deploy` command.

