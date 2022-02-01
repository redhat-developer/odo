
# `odo init`

[#5297](https://github.com/redhat-developer/odo/issues/5297)

The result of running `odo init` command should be local devfile.yaml saved in the current directory, and starter project extracted in the current directory (if user picked one)


When devfile.yaml exists in the current directory command should exit with error:
```
Unable to create new component in the current directory.
The current directory already contains `devfile.yaml` file.
You can use `odo dev` or `odo deploy` to start your application on a cluster.
```

Command should use registries as configured in `odo registry` command. If there is multiple registries configured it should use all of them.

## Flags

- `--name` (string) - name of the component (required)
- `--devfile` (string) - name of the devfile in devfile registry (required if `--devfile-path` is not defined)
- `--registry` - (string) name of the devfile registry (as configured in `odo registry`). It can be used in combination with `--devfile`, but not with `--devfile-path` (optional)
- `--starter`  (string) - name of the  starter project (optional)
- `--devfile-path` (string) - path to a devfile. This is alternative to using devfile from Devfile registry. It can be local filesystem path or http(s) URL (required if `--devfile` is not defined)
- `-o` (string) output information in a specified format (json).

If no flag is specified it should enter interactive mode.
If even a single optional flag is specified then run in non-interactive mode and requires all required flags.

## Interactive mode
### when executed in empty directory

```
$ odo init
TODO: Intro text (Include  goal as well as the steps that they are going to take ( including terminology ))

? Select language:  [Use arrows to move, type to filter]
> dotnet
  go
  java
  javascript
  typescript
  php
  python

? Select project type:  [Use arrows to move, type to filter]
  .NET 5.0
> .NET 6.0
  .NET Core 3.1
  ** GO BACK ** (not implemented)

? Which starter project do you want to use?  [Use arrows to move, type to filter]
> starter1
  starter2

? Enter component name: mydotnetapp

⠏ Downloading "dotnet60". DONE
⠏ Downloading starter project "starter1" ... DONE
Your new component "mydotnetapp" is ready in the current directory.
To start editing your component, use “odo dev” and open this folder in your favorite IDE.
Changes will be directly reflected on the cluster.
To deploy your component to a cluster use “odo deploy”.
```

1. **"Select language:"**

   Shows list of all values of `metadata.language` fields from all devfiles in the current active Devfile registry. (every language only once)

2. **"Select project type:"**

   Select all possible values of `metadata.projectType` fields from all Devfiles that have selected language.
    If there is a Devfile that doesn't have `metadata.projectType` it should display its `metadata.name`.

    If there there is more than one devfile with the same projectType the list item should include the `metadata.name` and registry name. For example  if there are the same devfiles in multiple registries

    ```
    SpringBoot (java-springboot, registry: DefaultRegistry)
    SpringBoot (java-springboot, registry: MyRegistry)
    ```

    or if there is the same projectType in mulitple Devfiles

    ```
    SpringBoot (java-maven-springboot, registry: MyRegistry)
    SpringBoot (java-gradle-springboot, registry: MyRegistry)
    ```

3. **"Which starter project do you want to use:"**

    At this point, the previous answers should be enough to uniquely select one Devfile from registry.
    List of all starter projects defined in selected devfile.

4. **"Enter component name:"**
    Name of the component. This should be saved in the local `devfile.yaml` as a value for `metadata.name` field.

### When executed in non-empty directory

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


If there is no devfile in the current directory `odo` should use ([Alizer](https://github.com/redhat-developer/alizer/pull/55/)) to get corresponding Devfile for code base in the current directory.
After the successful detection it will show  Language and Project Type information to users and ask for confirmation.
If user answers that the information is not correct, odo should ask "Select language" and "Select project type" questions (see: [`odo init`](odo-init.md) interactive mode).

The configuration part helps users to modify most common configurations done on Devfiles.
"? What configuration do you want change? " question is repeated over and over again until user confirms that the configuration is done and there is nothing else to change.
You can find a naive approach trying to implemented this in odo v3 mockup https://github.com/kadel/odo-v3-prototype/blob/1614ef74a6afdd056d0f87e1e1fafb8275a08a27/cmd/utils.go#L117-L203 