<img src="./docs/img/openshift.png" width="180" align="right">

# OpenShift Do - Developer Focused CLI for OpenShift

[![Build Status](https://travis-ci.org/openshift/odo.svg?branch=master)](https://travis-ci.org/openshift/odo) [![codecov](https://codecov.io/gh/openshift/odo/branch/master/graph/badge.svg)](https://codecov.io/gh/openshift/odo) [![CircleCI](https://circleci.com/gh/openshift/odo/tree/master.svg?style=svg)](https://circleci.com/gh/openshift/odo/tree/master) [![mattermost](/docs/img/mattermost.svg)](https://chat.openshift.io/developers/channels/odo)

OpenShift Do (odo) is a fast, iterative, and opinionated CLI tool for developers who write, build, and deploy applications on OpenShift.

Existing tools such as `oc` are more operations-focused and require a deep-understanding of Kubernetes and OpenShift concepts. OpenShift Do abstracts away complex Kubernetes and OpenShift concepts, thus allowing developers to focus on what's most important to them: code.

## Key features
OpenShift Do is designed to be simple and concise with the following key features:

- Simple syntax and design centered around concepts familiar to developers, such as project, application, and component.
- Completely client based. No server is required within the OpenShift cluster for deployment.
- Supports multiple languages and frameworks such as Node.js, Java, Ruby, Perl, PHP, and Python.
- Detects changes to local code and deploys it to the cluster automatically, giving instant feedback to validate changes in real-time.
- Lists all available components and services from the OpenShift cluster.

## Installation

**Prerequisites:**
* OpenShift 3.10.0 and above is required.
* We recommend using [Minishift](https://github.com/minishift/minishift).

<details>
<summary> OS-independent automated install</summary>

#### Use this [bash script](./scripts/install.sh) to quickly install OpenShift Do. It will automatically detect your operating system and install `odo` accordingly.

```sh
curl -L https://github.com/openshift/odo/raw/master/scripts/install.sh | bash
```

</details>

<details>
<summary> MacOS</summary>

#### Binary installation:
```sh
sudo curl -L https://github.com/openshift/odo/releases/download/v0.0.20/odo-darwin-amd64 -o /usr/local/bin/odo && sudo chmod +x /usr/local/bin/odo
```

#### Tarball installation:
```sh
sudo sh -c 'curl -L https://github.com/openshift/odo/releases/download/v0.0.20/odo-darwin-amd64.gz | gzip -d > /usr/local/bin/odo; chmod +x /usr/local/bin/odo'
```

</details>

<details>
<summary> Linux</summary>

#### Binary installation:
```sh
sudo curl -L https://github.com/openshift/odo/releases/download/v0.0.20/odo-linux-amd64 -o /usr/local/bin/odo && sudo chmod +x /usr/local/bin/odo
```

#### Tarball installation:
```sh
sudo sh -c 'curl -L https://github.com/openshift/odo/releases/download/v0.0.20/odo-linux-amd64.gz | gzip -d > /usr/local/bin/odo; chmod +x /usr/local/bin/odo'
```

</details>

<details>
<summary> Windows</summary>

In order to correctly use OpenShift Do you must download it and add it to your PATH environment variable:

1. Download the `odo-windows-amd64.exe.gz` file from the [GitHub releases page](https://github.com/openshift/odo/releases).
2. Extract the file.
3. Add the location of extracted binary to your PATH environment variable by following [this Wiki page](https://github.com/openshift/odo/wiki/Setting-PATH-variable-on-Windows).

</details>

For a list of other methods such as installing the latest binary or specific OS installations, visit the [installation page](/docs/installation.md).

## Demonstration
The following demonstration provides an overview of OpenShift Do:

[![asciicast](https://asciinema.org/a/225717.svg)](https://asciinema.org/a/225717)

## Deploying an application using OpenShift Do

After installing OpenShift Do, follow these steps to build, push, and deploy a Node.js application. Examples for other supported languages and runtimes can be found [here](https://github.com/openshift/odo/blob/master/docs/examples.md).

1. Start a local OpenShift development cluster by running minishift.
```sh
$ minishift start
```

2. Log into the OpenShift cluster.
```sh
$ odo login -u developer -p developer`
```

3. Create an application. An application in OpenShift Do is an umbrella under which you add other components.
```sh
$ odo app create node-example-app`
```

4. Download the Node.js sample code and change directory to the location of the sample code.
```sh
$ git clone https://github.com/openshift/nodejs-ex
$ cd nodejs-ex
```

5. Add a component of type `nodejs` to your application.
```sh
$ odo create nodejs
```
6. Deploy your application.
```sh
$ odo push
```
7. Create a URL to access the application and visit it to test it.
```sh
$ odo url create
$ curl nodejs-myproject.192.168.42.147.nip.io
```

For more in-depth information and advanced use-cases such as adding storage to a component or linking components, see the [interactive tutorial](https://learn.openshift.com/introduction/developing-with-odo/) or the [getting started guide](/docs/getting-started.md).

## Additional documentation

Additional documentation can be found below:

  - [Detailed Installation Guide](https://github.com/openshift/odo/blob/master/docs/installation.md)
  - [Getting Started Guide](https://github.com/openshift/odo/blob/master/docs/getting-started.md)
  - [Usage Examples for Other Languages and Runtimes](https://github.com/openshift/odo/blob/master/docs/examples.md)
  - [CLI Reference](https://github.com/openshift/odo/blob/master/docs/cli-reference.md)
  - [Development Guide](https://github.com/openshift/odo/blob/master/docs/development.md)

## Community, discussion, contribution, and support

**Chat:** We have a public channel [#odo on chat.openshift.io](https://chat.openshift.io/developers/channels/odo).

**Issues:** If you have an issue with OpenShift Do, please [file it](https://github.com/openshift/odo/issues).

**Contributing:** Want to become a contributor and submit your own code? Have a look at our [development guide](https://github.com/openshift/odo/blob/master/docs/development.md).

## Glossary

**Application:** An application consists of multiple microservices or components that work individually to build the entire application.

**Component:** A component is similar to a microservice. Multiple components make up an application. A component has different attributes like storage. OpenShift Do supports multiple component types like nodejs, perl, php, python and ruby.

**Service:** Typically a service is a database or a service that a component links to or depends on. For example: MariaDB, Jenkins, MySQL. This comes from the OpenShift "Service Catalog" and must be enabled within your cluster.
