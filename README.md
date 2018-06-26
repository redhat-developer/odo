# Odo - OpenShift do

[![Build Status](https://travis-ci.org/redhat-developer/odo.svg?branch=master)](https://travis-ci.org/redhat-developer/odo) [![codecov](https://codecov.io/gh/redhat-developer/odo/branch/master/graph/badge.svg)](https://codecov.io/gh/redhat-developer/odo)

![Powered by OpenShift](/docs/img/powered_by_openshift.png)

## What's Odo?

Odo (OpenShift do...) is a CLI tool that provides developers with **fast** and **automated** source code deployments. Odo supplements iterative development by using the power of OpenShift's [Source-to-Image](https://github.com/openshift/source-to-image) with the stableness of [Kubernetes](https://github.com/kubernetes/kubernetes). Developers can immediately start coding while Odo builds, pushes and deploys the application in the background.

#### Features

  - **Multiple languages:** Odo supports Node.JS, Ruby, .Net Core, Perl, PHP and Python.
  - **Speed:** Building your source code *immediately* after saving and deployed to your cluster.
  - **Reproducible:** Allows for easy reproducibility by using tightly versioned Docker containers for your source code environment.
  - **Deployability:** Easily deploy a new version, or have Odo automatically build and re-deploy your code on each change.
  - **Support for multiple components and microservices:** Deploy only what you need. For example, having both a Ruby and a JavaScript application side-by-side.
  - **Serverless:** No requirement for running a server to automate tasks. Odo talks to OpenShift directly through an API.
  - **Instant feedback:** Deploy while making edits to files, showing direct and instant feedback.

### Documentation

Documentation can be found below:

  - [Installation](https://github.com/redhat-developer/odo/blob/master/docs/installation.md)
  - [Getting Started](https://github.com/redhat-developer/odo/blob/master/docs/getting-started.md)
  - [Development](https://github.com/redhat-developer/odo/blob/master/docs/development.md)

### Installation

#### Automated installation

The quickest way to install Odo is through our [bash script](./scripts/install.sh), which will automatically detect your operating system and install `odo` accordingly!

```sh
curl -L https://github.com/redhat-developer/odo/raw/master/scripts/install.sh | bash
```

#### macOS

```sh
# Binary installation
sudo curl -L https://github.com/redhat-developer/odo/releases/download/v0.0.7/odo-darwin-amd64 -o /usr/local/bin/odo && chmod +x /usr/local/bin/odo

# Alternative, compressed tarball installation
sudo sh -c 'curl -L https://github.com/redhat-developer/odo/releases/download/v0.0.7/odo-darwin-amd64.gz | gzip -d > /usr/local/bin/odo; chmod +x /usr/local/bin/odo'
```

#### Linux

```sh
# Binary installation
sudo curl -L https://github.com/redhat-developer/odo/releases/download/v0.0.7/odo-linux-amd64 -o /usr/local/bin/odo && chmod +x /usr/local/bin/odo

# Alternative, compressed tarball installation
sudo sh -c 'curl -L https://github.com/redhat-developer/odo/releases/download/v0.0.7/odo-linux-amd64.gz | gzip -d > /usr/local/bin/odo; chmod +x /usr/local/bin/odo'
```

#### Windows

1. Download the `odo-windows-amd64.exe.gz` file from the [GitHub releases page](https://github.com/redhat-developer/odo/releases).
2. Extract the file
3. Add the location of extracted binary to your PATH environment variable by following [this Wiki page](https://github.com/redhat-developer/odo/wiki/Setting-PATH-variable-on-Windows).

#### Other methods

For a list of other methods such as installing the latest mastery binary, or specific OS installations, visit our [installation page](/docs/installation.md).

## Getting started with Odo

Wanted to get started? Follow the instructions below or our [Katacoda tutorial](https://www.katacoda.com/mjelen/courses/introduction/developing-with-odo):

### Requirements

  - `minishift` or an OpenShift environment 3.9.0+, the best way to deploy a development environment is using [Minishift](https://github.com/minishift/minishift).
  - `oc` If you do not have it, there's an excellent guide on the [OpenShift site](https://docs.openshift.org/latest/cli_reference/get_started_cli.html#installing-the-cli) on how to install the latest client.

### Deploying a Node.js application using Odo

For a quick tutorial on how Odo works, follow the instructions below! Otherwise, we have an [excellent Katacoda tutorial](https://www.katacoda.com/mjelen/courses/introduction/developing-with-odo) or an [in-depth getting started guide](/docs/getting-started.md).

```sh
# Download the latest release!
$ curl -L https://github.com/redhat-developer/odo/raw/master/scripts/install.sh | bash

# Start your development environment
$ minishift start

# Download the Node.JS example directory
$ git clone https://github.com/openshift/nodejs-ex
$ cd nodejs-ex

# Create new nodejs component
$ odo create nodejs

# Now let's deploy your application!
$ odo push

# Last, we'll create a way to access the application
$ odo url create
nodejs - nodejs-myproject.192.168.42.147.nip.io

# Test it / visit the URL
$ curl nodejs-myproject.192.168.42.147.nip.io
```

## Glossary

- **Application:** Is, well, your application! It consists of multiple microservices or components, that work individually to build the entire application.
- **Component:** can be thought of as a microservice. Multiple components will make up an application. A component will have different attributes like storage, etc.
Multiple component types are currently supported, like nodejs, perl, php, python, ruby, etc.

## CLI Structure

```sh
odo --verbose : Odo (Openshift Do)
    app --short : Perform application operations
        create : Create an application
        delete --force : Delete the given application
        describe : Describe the given application
        get --short : Get the active application
        list : Lists all the applications
        set : Set application as active
    catalog : Catalog related operations
        list : List all available component types.
        search : Search component type in catalog
    completion : Output shell completion code
    component --short : Components of application.
        get --short : Get currently active component
        set : Set active component.
    create --binary --git --local : Create a new component
    delete --force : Delete an existing component
    describe : Describe the given component
    link --component : Link target component to source component
    list : List all components in the current application
    log : Retrieve the log for the given component.
    project --short : Perform project operations
        create : Create a new project
        get --short : Get the active project
        list : List all the projects
        set --short : Set the current active project
    push --local : Push source code to a component
    storage : Perform storage operations
        create --component --path --size : Create storage and mount to a component
        delete --force : Delete storage from component
        list --component : List storage attached to a component
        mount --component --path : mount storage to a component
        unmount --component : Unmount storage from the current component
    update --binary --git --local : Update the source code path of a component
    url : Expose component to the outside world
        create : Create a URL for a component
        delete --force : Delete a URL
        list --application --component : List URLs
    version : Print the client version information
    watch : Watch for changes, update component on change
```
*_autogenerated_
