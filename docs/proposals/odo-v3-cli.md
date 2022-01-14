# odo v3 CLI
- [odo v3 CLI](#odo-v3-cli)
  - [odo v2 commands that should be removed](#odo-v2-commands-that-should-be-removed)
  - [Command that won't change much in v3](#command-that-wont-change-much-in-v3)
  - [Commands that should be present in v3.0.0-alpha1](#commands-that-should-be-present-in-v300-alpha1)
    - [general rules for odo cli](#general-rules-for-odo-cli)
    - [`odo login`](#odo-login)
    - [`odo logout`](#odo-logout)
    - [`odo init`](#odo-init)
      - [Interactive mode](#interactive-mode)
        - [Example](#example)
      - [Flags](#flags)
      - [Command behavior and error states](#command-behavior-and-error-states)
    - [`odo deploy`](#odo-deploy)
      - [Interactive mode](#interactive-mode-1)
      - [Flags](#flags-1)
      - [Command behavior and error states](#command-behavior-and-error-states-1)
    - [`odo build-images`](#odo-build-images)
    - [`odo dev`](#odo-dev)
    - [`odo list`](#odo-list)
    - [`odo preference`](#odo-preference)
    - [`odo project`](#odo-project)
    - [`odo registry`](#odo-registry)
    - [`odo describe`](#odo-describe)
    - [`odo exec`](#odo-exec)
    - [`odo status`](#odo-status)
  - [Commands that will be added in v3.0.0-alpha2](#commands-that-will-be-added-in-v300-alpha2)
    - [`odo create`](#odo-create)
      - [`odo create binding`](#odo-create-binding)
      - [`odo create service`](#odo-create-service)
      - [`odo create endpoint`](#odo-create-endpoint)
      - [`odo create storage`](#odo-create-storage)
    - [`odo create url`](#odo-create-url)
    - [`odo delete`](#odo-delete)
      - [`odo delete binding`](#odo-delete-binding)
      - [`odo delete service`](#odo-delete-service)
      - [`odo delete endpoint`](#odo-delete-endpoint)
      - [`odo delete storage`](#odo-delete-storage)
    - [`odo list`](#odo-list-1)
      - [`odo list binding`](#odo-list-binding)
      - [`odo list service`](#odo-list-service)
      - [`odo list endpoint`](#odo-list-endpoint)
      - [`odo list storage`](#odo-list-storage)
    - [`odo app`](#odo-app)
    - [`odo catalog`](#odo-catalog)
    - [`odo config`](#odo-config)
    - [`odo build-images`](#odo-build-images-1)
    - [`odo utils`](#odo-utils)
    - [`odo version`](#odo-version)

Over the years of odo development we picked up a lot of commands.
Odo v3 will introduce new commands (`odo dev` #5299, `odo init` #5297).
This will change the command flow from what we currently have in v2. We need to make sure that the whole odo CLI is consistent, and all commands follow the same pattern.

There are also some commands that are there since the original odo v1 and were originally designed for s2i approach only, those commands or flags should be removed, or reworked to better fit Devfile

## odo v2 commands that should be removed

- `odo component` every command from “odo component *” already exists as a root command.
- `odo watch` will be replaced with `odo dev`
- `odo push` will be replaced with `odo dev`
- `odo unlink` will be replaced with `odo delete binding`
- `odo link` will be replaced with `odo create binding`
- `odo url` will be replaced with `odo create endpoint`
- `odo test` will be replaced `odo run test` ?
- `odo service` will be replaced with `odo create service`
- `odo storage` if needed it will be replaced with `odo create storage`
- `odo env` should be integrated with `odo config`
- `odo debug` will be replaced with `odo run debug`
- `odo registry` will be replaced with `odo preference ....`?

## Command that won't change much in v3
- `odo login`
- `odo logout`
- `odo preference` (maybe renamed to `odo config`? TODO: need to figure out right naming for configuring devfile component and configuring odo)
- `odo build-images`
- `odo deploy`
- `odo exec`


## Commands that should be present in v3.0.0-alpha1

### general rules for odo cli

- Once even a single flag is provided it ruins in non-interactive mode. All required information needs to be provided via flags
- If command is executed without flags it can enter interactive mode


### `odo login`

todo

### `odo logout`

todo

### `odo init`
[#5297](https://github.com/redhat-developer/odo/issues/5297)

#### Interactive mode

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

##### Example
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

Proof of concept was implemented in https://github.com/kadel/odo-v3-prototype/blob/main/cmd/init.go



#### Flags
- `--name` - name of the component (required)
- `--devfile` - name of the devfile in devfile registry (required if `--devfile-path` is not defined)
- `--devfile-registry` - name of the devfile registry (as configured in `odo registry`). It can be used in combination with `--devfile`, but not with `--devfile-path` (optional)
- `--starter` - name of the  starter project (optional)
- `--devfile-path` - path to a devfile. This is alternative to using devfile from Devfile registry. It can be local filesystem path or http(s) URL (required if `--devfile` is not defined)


If no flag is specified it should enter interactive mode.
If even a single optional flag is specified then run in non-interactive mode and requires all required flags.


#### Command behavior and error states

- the result of running `odo init` command should be local devfile.yaml saved in the current directory, and starter project extracted in the current directory (if user picked one)
- running `odo init` in non-empty directory exits with error
  
  ```
  The current directory is not empty. You can bootstrap new component only in empty directory.
  If you have existing code that you want to deploy use `odo deploy` or use `odo dev` command to quickly iterate on your component.
  ```

- command should use registry as configured in `odo registry` command. If there is multiple registries configured it should use all of them.




### `odo deploy`

Deploying application (outer loop)
[#5298](https://github.com/redhat-developer/odo/issues/5298)

#### Interactive mode
#### Flags
#### Command behavior and error states

### `odo build-images`
mostly as it is in v2

### `odo dev`

Running the application on the cluster for **development** (inner loop)

[#5299](https://github.com/redhat-developer/odo/issues/5298)


### `odo list`

Listing "stuff" created by odo.

- without arguments lists components in local Devfile and in cluster
- `odo list services` list services defined in local devfile and in cluster
- `odo list endpoints` list endpoints defined in local devfile and in cluster
- `odo list bindings` list bindings defined in local devfile and in cluster 


### `odo preference`

Configures odo behavior like timeouts.
- UpdateNotification  - keep
- NamePrefix          - remove
- Timeout             - add explanations  
- BuildTimeout        - remove
- PushTimeout         - add explanations
- Ephemeral           - keep
- ConsentTelemetry    - keep 



### `odo project`
mostly as it is

### `odo registry`
mostly as it is just drop  support "github"  registry

### `odo describe`
TODO

### `odo exec`
mostly as it is

### `odo status`
mostly as it is

## Commands that will be added in v3.0.0-alpha2

### `odo create`

#### `odo create binding`

#### `odo create service`

#### `odo create endpoint`

#### `odo create storage`

### `odo create url`

### `odo delete`

#### `odo delete binding`

#### `odo delete service`

#### `odo delete endpoint`

#### `odo delete storage`

### `odo list`

#### `odo list binding`

#### `odo list service`

#### `odo list endpoint`

#### `odo list storage`

### `odo app`
TODO

### `odo catalog`
TODO

### `odo config`
TODO

### `odo build-images`
mostly as it is



### `odo utils`
mostly as it is

### `odo version`
mostly as it is

