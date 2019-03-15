<img src="./docs/img/openshift.png" width="180" align="right">

# Odo - Developer focused CLI for OpenShift

[![Build Status](https://travis-ci.org/redhat-developer/odo.svg?branch=master)](https://travis-ci.org/redhat-developer/odo) [![codecov](https://codecov.io/gh/redhat-developer/odo/branch/master/graph/badge.svg)](https://codecov.io/gh/redhat-developer/odo) [![CircleCI](https://circleci.com/gh/redhat-developer/odo/tree/master.svg?style=svg)](https://circleci.com/gh/redhat-developer/odo/tree/master) [![mattermost](/docs/img/mattermost.svg)](https://chat.openshift.io/developers/channels/odo)

OpenShift Do (odo) is a  fast iterative CLI tool for developers who write, build, and deploy applications on OpenShift.
Existing tools such as `oc` are more operations-focused and require a deep-understanding of Kubernetes and OpenShift concepts.

odo is  an opinionated CLI tool that supports fast, iterative development and that abstracts away complex Kubernetes and OpenShift concepts, thus allowing developers to focus on what's most important to them: code.

odo improves the developer experience with OpenShift. It is designed to be simple and concise so that you can focus on coding rather than how to deploy your application. odo builds and deploys your code to your cluster as soon as you save your changes, giving you instant feedback to validate your changes in real-time. odo's syntax and design is centered around concepts already familiar to developers, such as: project, application and component.


## Key features

- Designed for fast, iterative development cycles
- Completely client based. No server is needed within your OpenShift cluster for deployment
- Supports multiple languages and frameworks such as Node.js, Java, Ruby, Perl, PHP and Python
- Detects changes to your local code and deploys automatically using `odo watch`
- Lists all available components and services from your OpenShift cluster

## Installation

.Prerequisites:
You need **OpenShift 3.9.0** and above.
*Note:* We recommend using [Minishift](https://github.com/minishift/minishift)to test and use OpenShift locally.

<details>
<summary> OS-independent automated install</summary>

#### Use this [bash script](./scripts/install.sh) to quickly install odo. It will automatically detect your operating system and install `odo` accordingly.

```sh
curl -L https://github.com/redhat-developer/odo/raw/master/scripts/install.sh | bash
```

</details>

<details>
<summary> MacOS</summary>

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
<summary> Linux</summary>

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
<summary> Windows</summary>

1. Download the `odo-windows-amd64.exe.gz` file from the [GitHub releases page](https://github.com/redhat-developer/odo/releases).
2. Extract the file.
3. Add the location of extracted binary to your PATH environment variable by following [this Wiki page](https://github.com/redhat-developer/odo/wiki/Setting-PATH-variable-on-Windows).

</details>

For a list of other methods such as installing the latest binary on specific OS installations, visit the [installation page](/docs/installation.md).

## odo overview

[![asciicast](https://asciinema.org/a/225717.svg)](https://asciinema.org/a/225717)

## Deploying an application using odo

After you install odo, follow these steps to build, push, and deploy a Node.js application using odo. Examples for other supported languages and runtimes can be found [here](https://github.com/redhat-developer/odo/blob/master/docs/examples.md).

1. Start a local OpenShift development cluster by running minishift
````
$ minishift start`
````

2. Log into the OpenShift cluster
```
$ odo login -u developer -p developer`
````

3. Create an application. An application in odo is an umbrella under which you add other components
````
$ odo app create node-example-app`
````

4. Download the Node.js sample code and change directory to the location of the sample code
````
$ git clone https://github.com/openshift/nodejs-ex
$ cd nodejs-ex
````

5. Add a component of type `nodejs` to your application
````
$ odo create nodejs
````
6. Deploy your application.
````
$ odo push
````
7. Create a URL to access the application and visit it to test it.
````
$ odo url create
$ curl nodejs-myproject.192.168.42.147.nip.io
````

For more in-depth information and advanced use-cases such as adding storage to a component or linking components, see the [interactive tutorial](https://learn.openshift.com/introduction/developing-with-odo/) or the [odo getting started guide](/docs/getting-started.md).

## Additional Documentation

Additional documentation can be found below:

  - [Detailed installation guide](https://github.com/redhat-developer/odo/blob/master/docs/installation.md)
  - [Odo getting started guide](https://github.com/redhat-developer/odo/blob/master/docs/getting-started.md)
  - [Usage examples for other languages and runtimes](https://github.com/redhat-developer/odo/blob/master/docs/examples.md)
  - [Odo CLI reference](https://github.com/redhat-developer/odo/blob/master/docs/cli-reference.md)
  - [Development guide](https://github.com/redhat-developer/odo/blob/master/docs/development.md)

## Community, Discussion, Contribution and Support

**Chat:** We have a public channel [#odo on chat.openshift.io](https://chat.openshift.io/developers/channels/odo).

**Issues:** If you have an issue with odo, please [file it](https://github.com/redhat-developer/odo/issues).

**Contributing:** Want to become a contributor and submit your own code? Have a look at our [development guide](https://github.com/redhat-developer/odo/blob/master/docs/development.md).

## Glossary

**Application:** An application consists of multiple microservices or components that work individually to build the entire application.

**Component:** A component is similar to a microservice. Multiple components make up an application. A component has different attributes like storage. odo supports multiple component types like nodejs, perl, php, python and ruby.

**Service:** Typically a service is a database or a service that a component links to or depends on. For example: MariaDB, Jenkins, MySQL. This comes from the OpenShift "Service Catalog" and must be enabled within your cluster.
