# ODO - OpenShift Developer Optimization

[![Build Status](https://travis-ci.org/redhat-developer/odo.svg?branch=master)](https://travis-ci.org/redhat-developer/odo) [![codecov](https://codecov.io/gh/redhat-developer/odo/branch/master/graph/badge.svg)](https://codecov.io/gh/redhat-developer/odo)

## What is odo?

ODO (OpenShift Developer Optimization) is an OpenShift tool to deploy applications and components in a *fast* and *optimized* manner.

![Powered by OpenShift](/docs/img/powered_by_openshift.png)

## Pre-requisites
- OpenShift version 3.7.0 and up

To use odo you need access to an OpenShift instance and have OpenShift CLI installed on your local machine (`oc` should be in your $PATH).

### OpenShift instance
You can use [Minishift](https://docs.openshift.org/latest/minishift/index.html) to get a local instance of OpenShift. However odo can be used with any instance of OpenShift.

### OpenShift CLI
There are different ways to install OpenShift CLI. 
Please follow [OpenShift documentation](https://docs.openshift.org/latest/cli_reference/get_started_cli.html#installing-the-cli).

## Installation
To install `odo` on your system, you can use the fully automated [install.sh](./scripts/install.sh) script.
This script will enable odo repository on your system and install odo using package manager depending on your system.
Supported systems are Debian, Ubuntu, Fedora, CentOS and macOS. You can find more information about package repositories in 
[Advanced installation guide](./docs/advanced-installation-guide.md)

```
curl -L https://github.com/redhat-developer/odo/raw/master/scripts/install.sh | bash
```


If you don't want to add extra package repositories to your system you can just extract  `odo` binary from [GitHub releases page](https://github.com/redhat-developer/odo/releases) to one of the directories that are in your `$PATH`.

For macOS:

```
sudo curl -L  "https://github.com/redhat-developer/odo/releases/download/v0.0.4/odo-darwin-amd64.gz" | gzip -d > /usr/local/bin/odo; chmod +x /usr/local/bin/odo
```

For Linux:
```
sudo curl -L  "https://github.com/redhat-developer/odo/releases/download/v0.0.4/odo-linux-amd64.gz" | gzip -d > /usr/local/bin/odo; chmod +x /usr/local/bin/odo
```

For Windows:

Download the `odo-windows-amd64.exe.gz` file from the [GitHub releases page](https://github.com/redhat-developer/odo/releases), extract it into a directory 
and add the location of extracted binary to your PATH environment variable by following [this wiki page.](https://github.com/redhat-developer/odo/wiki/Setting-PATH-variable-on-Windows)

You can also download latest master builds from [Bintray](https://dl.bintray.com/odo/odo/latest/). This is updated every time there is a change in master git branch.



## Concepts
- An **_application_** is, well, your application! It consists of multiple microservices or components, that work individually to build the entire application.
- A **_component_** can be thought of as a microservice. Multiple components will make up an application. A component will have different attributes like storage, etc.
Multiple component types are currently supported, like nodejs, perl, php, python, ruby, etc.

## Getting Started
Developing applications using odo is as simple as -
- `odo app create <name>`
- `odo create <name>`
- `odo push`

Check out our [Getting Started](docs/getting-started.md) guide and get going!

## CLI Structure
```
odo --verbose : OpenShift CLI for Developers
    app --short : Perform application operations
        create : create an application
        delete --force : delete the given application
        get --short : get the active application
        list : lists all the applications
        set : Set application as active.
    catalog : Catalog related operations
        list : List all available component types.
        search : Search component type in catalog
    completion : Output shell completion code
    component --short : Components of application.
        get --short : Get currently active component
        set : Set active component.
    create --binary --git --local : Create new component
    delete --force : Delete existing component
    describe : Describe the given component
    link --component : Link target component to source component
    list : List all components in the current application
    project --short : Perform project operations
        create : create a new project
        get --short : get the active project
        list : list all the projects
        set --short : set the current active project
    push --local : Push source code to component
    storage --component : Perform storage operations
        create --path --size : create storage and mount to component
        list : list storage attached to a component
        remove : remove storage from component
    update --binary --git --local : Change the source of a component
    url : Expose component to the outside world
        create : Create a URL for a component
        delete : Delete a URL
        list --application --component : List URLs
    version : Print the version of odo
    watch --local : Watch for changes, update component on change
```
*_autogenerated_
