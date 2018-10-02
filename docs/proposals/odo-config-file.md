# odo config file

## Abstract
When developers work on creating an application, there's 2 phases they go though. First is creating the application initial implementation from scratch, and then iterate development on an existing application. While the developer works on the application, he might take some decisions on how the application should behave that should be preserved over time, but that are not source code but specific to how the application should behave in relation to the platform where it's running or in relation to the tool used to develop it.

IDEs often use a configuration file to hold specific configuration for an application, workspace or the whole tool. These files are stored under the same control version system as where the source code lives so that anytime a developer checks out that repository has control on how to configure that tool to work with that source code.

## Motivation
Provide a way to preserve conscisous decisions the developer has made that are orthogonal to the code but relevant to the tool/platform. Any user that continues working on an existing application don't need to know the decisions the original developer took, as these are preserved along with the source code.

## Use Cases

1. As a user, I want to be able to preserve some configuration I've done about my current component along with my source code
1. As a user, I want to be able to use configuration stored with my source code to help me be more productive
1. As a user, I want to be able to manage explicit configuration that is saved with my source code

## Design overview
We need to provide a local file, `.odo` (name subject to change) that will hold specific configuration for the source code to instruct how a component should be configured.

### Save user provided explicit flags
The user should not need to explicitly save the configuration in this file, but the odo CLI should keep the values explicitly used by a user into this file.

- If the file does not exist, it should create it.
- If the value is already present in the file, it should overwrite it.

```
odo create wildfly example --memory=1GB

# This should set value in .odo config file   (memory=2GB)
```

### Use configuration to "decorate" the deployment
On component creation, via `odo create ...` the CLI will take information stored in this configuration file to create an actual deployment on OpenShift.

```
# Value in .odo config file   (memory=2GB)
odo create wildfly example
```

The deployment that should be created, would have the following configuration:

```
apiVersion: v1
kind: Pod
metadata:
spec:
  containers:
  - image: xxx
    resources:
      limits:
        memory: 2Gi
      requests:
        memory: 2Gi
```

### Use configuration to "decorate" the CLI behavior
On component creation, via `odo create ...` the CLI will take information stored in this configuration file to define some values that would typically be defined by the CLI and that are relative to how this works..

```
# Value in .odo config file   (app-name=myapp, component-name=mycomp)
odo create wildfly
# Would create an application called myapp and component named mycomp using wildfly
```


### Developer friendly flags
Configuration options that will decorate the deployment should try to not use (as much as possible) terminology used in k8s. An example are memory resource requests and limits. Instead one could use terms like max-mem, min-mem and mem, that could work like this:

* When `mem` is set, memory requests and limits are the same.
* `min-mem` should be memory requests.
* `max-mem` should be memory limits. 
* `mem` is exclusive with `min-mem` or `max-mem`

### Configuration precedence
If the CLI provides some configuration values as flags for configuration that is store in the file, the ones provided explicitly from the user should take precedence.

```
# Value in .odo config file   (memory=2GB)
odo create wildfly example --memory=1GB

# The value that should be set for memory is 1 GB
```

### Explicitly modify the .odo config file
There might be times that a developer would want to add/edit/remove values from the .odo configuration file. There should be a command to manage this file in the odo CLI. That command should be the same used to touch configuration of the global `.odo` config file located in the users's home directory.

Currently that command is:

```
odo utils config [set|get|delete|add] KEY [VALUE] --local
```

Some examples of this command could be:

```
odo utils config set memory 2Gi --local      # Set's component memory to 2 GB
odo utils config delete memory --local       # Removes configuration for memory
odo utils config add odoignore .odo --local  # Adds a value to odoignore list
odo utils config list --local                # List all config
```

## Proposed information to be stored
We need to define and delimit what information will be stored in this file. 

### Deployment decoration
Configuration used to define how the deployment would be created:

- mem
  - min-mem
  - max-mem
- cpu
  - min-cpu
  - max-cpu
- health
  - live-health
  - ready-health
- build-image
  - src-path
  - artifacts-path
- runtime-image
  - src-path
  - deployment-path

### odo behavior
Configuration used to define how the CLI should behave:

- app-name: Preferred application name if none is defined
- component-name: Preferred component name if none is defined
- odoignore: list of files/patterns to ignore by default if no .odoignore file is provided, or when it's cerated

## Future evolution
WIP