# Odo - OpenShift Do

[![Build Status](https://travis-ci.org/redhat-developer/odo.svg?branch=master)](https://travis-ci.org/redhat-developer/odo) [![codecov](https://codecov.io/gh/redhat-developer/odo/branch/master/graph/badge.svg)](https://codecov.io/gh/redhat-developer/odo)

![Powered by OpenShift](/docs/img/powered_by_openshift.png)

- [What is Odo?](#what-is-odo)
- [Why did we create Odo?](#why-did-we-create-odo)
- [Who is Odo for?](#who-is-odo-for)
- [Setup and installation](#setup-and-installation)
- [Deploying an application using Odo](#deploying-a-nodejs-application-using-odo)
- [Additional documentation](#additional-documentation)
- [Community, Discussion, Contribution and Support](#community-discussion-contribution-and-support)
- [Glossary](#glossary)
- [CLI Structure](#cli-structure)


## What is Odo?

OpenShift Do (Odo) is a CLI tool for developers who are writing, building, and deploying applications on OpenShift. With Odo, developers get an opinionated CLI tool that supports fast, iterative development which abstracts away Kubernetes and OpenShift concepts, thus allowing them to focus on what's most important to them: code.

### Demo

![demo](/docs/img/example.gif)

## Why Did We Create Odo?

Odo was created to improve the developer experience with OpenShift. We understand that, as developers, you want tools that help you be productive. What do you don't want, is to have to change the way you work or have to become an expert in Kubernetes or OpenShift just to get your work done. 

Existing tools for interacting with OpenShift, such as the `oc` CLI tool, are more operations-focused and require an understanding of Kubernetes and OpenShift concepts such as pods, services, deployment configurations, and routes. The Odo CLI tool, on the other hand, was designed to use syntax and concepts that are familiar to developers, such as project, application, and component. Odo enables a fast, iterative development cycle allowing you to immediately start coding, while Odo builds, pushes, and deploys your application in the background. Since Odo can build and deploy your code to your cluster immediately after you save you changes, you benefit from instant feedback and can thus validate your changes in real-time. 

## Who is Odo For?

Odo is for developers! If you are a developer who is creating, building, or maintaining applications on OpenShift, Odo was designed for you. Odo supports multiple languages and frameworks, so whether you're using Node.js, Java, Ruby, Perl, PHP, or Python, the goal of Odo is to improve your development experience and to support fast, iterative development of applications for OpenShift. 

Odo is *not* a replacement for the standard `oc` CLI tool, and it is not intended to be used by operations, security, or release management teams.

## Setup and installation

Ready to get started? Follow the instructions below to set up Odo in your environment or give it a try in our [interactive tutorial](https://learn.openshift.com/introduction/developing-with-odo/):

### Requirements

  - `minishift` or an OpenShift environment 3.9.0+. The best way to deploy a development environment for OpenShift is using [Minishift](https://github.com/minishift/minishift).
  - `oc`, the OpenShift command line tool. Instructions for installing `oc` can be found [here](https://docs.openshift.org/latest/cli_reference/get_started_cli.html#installing-the-cli).

### Installing Odo

#### Automated installation

The quickest way to install Odo is via this [bash script](./scripts/install.sh), which will automatically detect your operating system and install `odo` accordingly.

```sh
curl -L https://github.com/redhat-developer/odo/raw/master/scripts/install.sh | bash
```

#### OS-specific installation methods

#### macOS

```sh
# Binary installation
sudo curl -L https://github.com/redhat-developer/odo/releases/download/v0.0.14/odo-darwin-amd64 -o /usr/local/bin/odo && sudo chmod +x /usr/local/bin/odo

# Alternative, compressed tarball installation
sudo sh -c 'curl -L https://github.com/redhat-developer/odo/releases/download/v0.0.14/odo-darwin-amd64.gz | gzip -d > /usr/local/bin/odo; chmod +x /usr/local/bin/odo'
```

#### Linux

```sh
# Binary installation
sudo curl -L https://github.com/redhat-developer/odo/releases/download/v0.0.14/odo-linux-amd64 -o /usr/local/bin/odo && sudo chmod +x /usr/local/bin/odo

# Alternative, compressed tarball installation
sudo sh -c 'curl -L https://github.com/redhat-developer/odo/releases/download/v0.0.14/odo-linux-amd64.gz | gzip -d > /usr/local/bin/odo; chmod +x /usr/local/bin/odo'
```

#### Windows

1. Download the `odo-windows-amd64.exe.gz` file from the [GitHub releases page](https://github.com/redhat-developer/odo/releases).
2. Extract the file
3. Add the location of extracted binary to your PATH environment variable by following [this Wiki page](https://github.com/redhat-developer/odo/wiki/Setting-PATH-variable-on-Windows).

#### Other methods

For a list of other methods such as installing the latest mastery binary, or specific OS installations, visit the [installation page](/docs/installation.md).

## Deploying a Node.js application using Odo

Now that you have Odo installed, follow these steps to build, push, and deploy a Node.js application using Odo. Examples for other supported languages and runtimes can be found [here](https://github.com/redhat-developer/odo/blob/master/docs/examples.md).

```sh

# Start a local OpenShift development cluster by running minishift
$ minishift start

# Log into the OpenShift cluster
$ oc login -u developer -p developer

# Create an application. An application in Odo is an umbrella under which you add other components
$ odo app create node-example-app

# Download the Node.js sample code
$ git clone https://github.com/openshift/nodejs-ex && cd nodejs-ex

# From the directory where the sample code is located, add a component of type nodejs to your application 
$ odo create nodejs

# Now let's deploy your application!
$ odo push

# Last, we'll create a way to access the application
$ odo url create

# Test it / visit the URL
$ curl nodejs-myproject.192.168.42.147.nip.io
```

For more in-depth information and advanced use-cases such as adding storage to a component or linking components, see the [interactive tutorial](https://learn.openshift.com/introduction/developing-with-odo/) or the [Odo user guide](/docs/getting-started.md).

## Additional Documentation

Additional documentation can be found below:

  - [Detailed Installation Guide](https://github.com/redhat-developer/odo/blob/master/docs/installation.md)
  - [Odo User Guide](https://github.com/redhat-developer/odo/blob/master/docs/getting-started.md)
  - [Usage Examples for Other Languages and Runtimes](https://github.com/redhat-developer/odo/blob/master/docs/examples.md)
  - [Odo CLI Reference](https://github.com/redhat-developer/odo/blob/master/docs/cli-reference.md)
  - [Development](https://github.com/redhat-developer/odo/blob/master/docs/development.md)

## Community, Discussion, Contribution and Support

**Chat:** We have a public channel [#Odo on chat.openshift.io](https://chat.openshift.io/developers/channels/odo).

**Issues:** If you have an issue with Odo, please [file it](https://github.com/redhat-developer/odo/issues).

**Contributing:** Want to become a contributor and submit your own code? Have a look at our [development guide](https://github.com/redhat-developer/odo/blob/master/docs/development.md).

## Glossary

- **Application:** Is, well, your application! It consists of multiple microservices or components, that work individually to build the entire application.
- **Component:** Can be thought of as a microservice. Multiple components will make up an application. A component will have different attributes like storage, etc.
Multiple component types are currently supported, like nodejs, perl, php, python, ruby, etc.
- **Service:** A service will typically be a database or a "service" a component links / depends on. For example: MariaDB, Jenkins, MySQL. This comes from the OpenShift "Service Catalog" and must be enabled within your cluster.
