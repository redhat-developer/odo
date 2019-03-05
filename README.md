<img src="./docs/img/openshift.png" width="180" align="right">

# Odo - Developer Focused CLI for OpenShift

[![Build Status](https://travis-ci.org/redhat-developer/odo.svg?branch=master)](https://travis-ci.org/redhat-developer/odo) [![codecov](https://codecov.io/gh/redhat-developer/odo/branch/master/graph/badge.svg)](https://codecov.io/gh/redhat-developer/odo) [![CircleCI](https://circleci.com/gh/redhat-developer/odo/tree/master.svg?style=svg)](https://circleci.com/gh/redhat-developer/odo/tree/master) [![mattermost](/docs/img/mattermost.svg)](https://chat.openshift.io/developers/channels/odo)

A fast iterative tool for deploying your source code straight to OpenShift.

## Features

- Designed for fast, iterative development cycles
- 100% client based. No server required within your OpenShift cluster for deployment
- Supports multiple languages and frameworks such as Node.js, Java, Ruby, Perl, PHP and Python
- Detect changes to your local code and deploy automatically with `odo watch`
- List all available components and services from your OpenShift cluster

## Installation

> The only requirement is **OpenShift 3.9.0** and above. A recommended way of testing out and using OpenShift locally is [Minishift](https://github.com/minishift/minishift).

<details>
<summary> :package: :rocket: OS-independent automated install</summary>

#### The quickest way to install odo is via this [bash script](./scripts/install.sh), which will automatically detect your operating system and install `odo` accordingly.

```sh
curl -L https://github.com/redhat-developer/odo/raw/master/scripts/install.sh | bash
```

</details>

<details>
<summary> :package: :apple: MacOS</summary>

#### Binary installation:
```sh
sudo curl -L https://github.com/redhat-developer/odo/releases/download/v0.0.20/odo-darwin-amd64 -o /usr/local/bin/odo && sudo chmod +x /usr/local/bin/odo
```

#### Tarball installation:
```sh
sudo sh -c 'curl -L https://github.com/redhat-developer/odo/releases/download/v0.0.20/odo-darwin-amd64.gz | gzip -d > /usr/local/bin/odo; chmod +x /usr/local/bin/odo'
```

</details>

<details>
<summary> :package: :penguin: Linux</summary>

#### Binary installation:
```sh
sudo curl -L https://github.com/redhat-developer/odo/releases/download/v0.0.20/odo-linux-amd64 -o /usr/local/bin/odo && sudo chmod +x /usr/local/bin/odo
```

#### Tarball installation:
```sh
sudo sh -c 'curl -L https://github.com/redhat-developer/odo/releases/download/v0.0.20/odo-linux-amd64.gz | gzip -d > /usr/local/bin/odo; chmod +x /usr/local/bin/odo'
```

</details>

<details>
<summary> :package: :checkered_flag: Windows</summary>

#### In order to correctly use odo you must add it to your PATH environment variable

1. Download the `odo-windows-amd64.exe.gz` file from the [GitHub releases page](https://github.com/redhat-developer/odo/releases).
2. Extract the file
3. Add the location of extracted binary to your PATH environment variable by following [this Wiki page](https://github.com/redhat-developer/odo/wiki/Setting-PATH-variable-on-Windows).

</details>

#### For a list of other methods such as installing the latest mastery binary, or specific OS installations, visit the [installation page](/docs/installation.md).

## Purpose

OpenShift Do (odo) is a CLI tool for developers who are writing, building, and deploying applications on OpenShift. With odo, developers get an opinionated CLI tool that supports fast, iterative development which abstracts away Kubernetes and OpenShift concepts, thus allowing them to focus on what's most important to them: code.

Odo was created to improve the developer experience with OpenShift. Existing tools such as `oc` are more operations-focused and requires a deep-understanding of Kubernetes and OpenShift concepts. Odo is designed to be simple and concise so you may focus on coding rather than how to deploy your application. Since odo can build and deploy your code to your cluster immediately after you save you changes, you benefit from instant feedback and can thus validate your changes in real-time. Odo's syntax and design is centered around concepts already familiar to developers, such as: project, application and component.

## Demo

[![asciicast](https://asciinema.org/a/225717.svg)](https://asciinema.org/a/225717)

## Deploying an Application using odo

After you have odo installed, follow these steps to build, push, and deploy a Node.js application using odo. Examples for other supported languages and runtimes can be found [here](https://github.com/redhat-developer/odo/blob/master/docs/examples.md).

```sh

# Start a local OpenShift development cluster by running minishift
$ minishift start

# Log into the OpenShift cluster
$ odo login -u developer -p developer

# Create an application. An application in odo is an umbrella under which you add other components
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

For more in-depth information and advanced use-cases such as adding storage to a component or linking components, see the [interactive tutorial](https://learn.openshift.com/introduction/developing-with-odo/) or the [odo user guide](/docs/getting-started.md).

## Additional Documentation

Additional documentation can be found below:

  - [Detailed Installation Guide](https://github.com/redhat-developer/odo/blob/master/docs/installation.md)
  - [Odo User Guide](https://github.com/redhat-developer/odo/blob/master/docs/getting-started.md)
  - [Usage Examples for Other Languages and Runtimes](https://github.com/redhat-developer/odo/blob/master/docs/examples.md)
  - [Odo CLI Reference](https://github.com/redhat-developer/odo/blob/master/docs/cli-reference.md)
  - [Development](https://github.com/redhat-developer/odo/blob/master/docs/development.md)

## Community, Discussion, Contribution and Support

**Chat:** We have a public channel [#odo on chat.openshift.io](https://chat.openshift.io/developers/channels/odo).

**Issues:** If you have an issue with odo, please [file it](https://github.com/redhat-developer/odo/issues).

**Contributing:** Want to become a contributor and submit your own code? Have a look at our [development guide](https://github.com/redhat-developer/odo/blob/master/docs/development.md).

## Glossary

**Application:** Is, well, your application! It consists of multiple microservices or components, that work individually to build the entire application.

**Component:** Can be thought of as a microservice. Multiple components will make up an application. A component will have different attributes like storage, etc.

Multiple component types are currently supported, like nodejs, perl, php, python, ruby, etc.
**Service:** A service will typically be a database or a "service" a component links / depends on. For example: MariaDB, Jenkins, MySQL. This comes from the OpenShift "Service Catalog" and must be enabled within your cluster.
