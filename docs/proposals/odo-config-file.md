## Abstract

When developers work with applications on openshift there should be a way
to deal with configuration artifacts using odo.

Implement a way for users to deal with environment variables 
and configuration files for the applications deployed on openshift cluster 
using odo. User does not need to be bothered about complex k8s/openshift 
terminologies like configmaps or secrets,so abstracting 
the complexity is a prime thing to consider here.


## Motivation

Most applications requires environment variables or configuration files.
So while the developer works on openshift, developer should be able to
add/edit/delete environment variables as well as configuration files 
for the applications deployed on openshift.

## Use Cases 

1. As a developer I want to create environment variables for my application using odo.

2. As a developer I want to be able to upload configuration files for my application using odo.


## Design overview

This section requires undestanding of three openshift terminologies

1. Environment variables : https://docs.okd.io/latest/dev_guide/environment_variables.html
2. ConfigMap : https://docs.okd.io/latest/dev_guide/configmaps.html
3. Secrets : https://docs.okd.io/latest/dev_guide/secrets.html

#### Abstracting config map and secrets from the user 

By default values will come from configmaps or secrets, 
the name of configmap/secret keys will be same as environment variable names.
configmap name should be unique for each component, 
so using a combination of appname + component name. eg: <app_name>-<component_name>

*Problem with this way of implementing:*
If the user goes into configmap using oc or webUI and modify key names then it becomes messy.

###### for setting environment variables
```sh
$odo config set variable <variable name>=<Value>
Environment variable with <variable name> created successfully
```
This will create a Config map with `<variable name>` as key in it and define env to consume from config map in deployment config.

We can also add `--secret` as a flag which can be used for creating secrets instead of configmap. So the command will be
```sh
$odo config set variable <variable name>=<Value> --secret
Environment variable with <variable name> created successfully
```
###### for viewing environment variables in a particular ~application~ Component 
```sh
$odo config view variable <component name>
Environment Variable Name            Value
MM_DB_HOST                        postgres
```
OR for current component
```sh
$odo config view variable
```
This will show the name of environment variables and values which are fetched from configmap/Secrets.


###### for modifying already specified environment variable
```sh
$odo config set variable <variablename>=<value>
The variable with the name which you specified already exist.Do you want to override it?(Yes/No):
```

This pops up and ask:
The variable with the name which you specified already exist.Do you want to override it?(Yes/No)

In a similar way we can do for configuration files,

###### for pushing a local configuration file 
```sh
$odo config set file <path to file>:<path to mount>
File `<path to file>` has been mounted at <path to mount> and can be consumed by the component <component name>
```
This will embed the file which is specified into a configmap and push to server, set the mount point as specified in the command. 

###### for viewing configuration files 
```sh
$odo config view file <component name>
```
This will show the file name and mount path, also by using `--describe` flag for displaying the entire file into the terminal screen.

## Future evolution



