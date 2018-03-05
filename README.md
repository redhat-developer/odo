# ocdev
[![Build Status](https://travis-ci.org/redhat-developer/ocdev.svg?branch=master)](https://travis-ci.org/redhat-developer/ocdev) [![codecov](https://codecov.io/gh/redhat-developer/ocdev/branch/master/graph/badge.svg)](https://codecov.io/gh/redhat-developer/ocdev)

## What is ocdev?
OpenShift Command line for Developers

## Pre-requisites
- OpenShift version 3.7.0 and up

To use ocdev you need access to an OpenShift instance and have OpenShift CLI installed on your local machine (`oc` should be in your $PATH).

### OpenShift instance
You can use [Minishift](https://docs.openshift.org/latest/minishift/index.html) to get a local instance of OpenShift. However ocdev can be used with any instance of OpenShift.

### OpenShift CLI
There are different ways to install OpenShift CLI. 
Please follow [OpenShift documentation](https://docs.openshift.org/latest/cli_reference/get_started_cli.html#installing-the-cli).

## Installation
To install `ocdev` on your system, you cat use the fully automated [install.sh](./scripts/install.sh) script.
This script will enable ocdev repository on your system and install ocdev using package manager depending on your system.
Supported systems are Debian, Ubuntu, Fedora, CentOS and macOS. You can find more information about package repositories in 
[Advanced installation guide](./docs/advanced-installation-guide.md)

```
curl -L https://github.com/redhat-developer/ocdev/raw/master/scripts/install.sh | bash
```


If you don't want to add extra package repositories to your system you can just extract  `ocdev` binary from [GitHub releases page](https://github.com/redhat-developer/ocdev/releases) to one of the directories that are in your `$PATH`.

For macOS:

```
sudo curl -L  "https://github.com/redhat-developer/ocdev/releases/download/v0.0.1/ocdev-darwin-amd64.gz" | gzip -d > /usr/local/bin/ocdev; chmod +x /usr/local/bin/ocdev
```

For Linux:
```
sudo curl -L  "https://github.com/redhat-developer/ocdev/releases/download/v0.0.1/ocdev-linux-amd64.gz" | gzip -d > /usr/local/bin/ocdev; chmod +x /usr/local/bin/ocdev
```

You can also download latest master builds from [Bintray](https://dl.bintray.com/ocdev/ocdev/latest/). This is updated every time there is a change in master git branch.



## Concepts
- An **_application_** is, well, your application! It consists of multiple microservices or components, that work individually to build the entire application.
- A **_component_** can be thought of as a microservice. Multiple components will make up an application. A component will have different attributes like storage, etc.
Multiple component types are currently supported, like nodejs, perl, php, python, ruby, etc.

## Getting Started
This will show you how easy is to run your first application on OpenShift using ocdev and minishift.

1. You should have minishift and ocdev already installed on your system.
    If you don't, follow [minishift installation guide](https://docs.openshift.org/latest/minishift/getting-started/installing.html) and [ocdev installation](/README.md#Installation)
    Verify that you have both commands installed.
    ```
    $ ocdev version
    v0.0.1 (2168609)

    $ minishift version
    minishift v1.13.1+75352e5

    $ oc version
    oc v3.7.1+ab0f056
    kubernetes v1.7.6+a08f5eeb62
    features: Basic-Auth

    error: server took too long to respond with version information.
    ```
    Error message that you get when running `oc version` is ok. It just means that you you are currently not connected to OpenShift cluster.
1. Start local OpenShift cluster using minishift.
    ```
    $ minishift start
    -- Starting profile 'minishift'
    -- Checking if requested hypervisor 'virtualbox' is supported on this platform ... OK
    -- Checking if VirtualBox is installed ... OK
    -- Checking the ISO URL ... OK
    -- Starting local OpenShift cluster using 'virtualbox' hypervisor ...
    -- Minishift VM will be configured with ...
    Memory:    2 GB
    vCPUs :    2
    Disk size: 20 GB
    -- Starting Minishift VM ............................ OK
    -- Checking for IP address ... OK
    -- Checking if external host is reachable from the Minishift VM ...
    Pinging 8.8.8.8 ... OK
    -- Checking HTTP connectivity from the VM ...
    Retrieving http://minishift.io/index.html ... OK
    -- Checking if persistent storage volume is mounted ... OK
    -- Checking available disk space ... 0% used OK
    Importing 'openshift/origin:v3.7.1' ............. OK
    Importing 'openshift/origin-docker-registry:v3.7.1' ... OK
    Importing 'openshift/origin-haproxy-router:v3.7.1' . CACHE MISS
    -- OpenShift cluster will be configured with ...
    Version: v3.7.1
    -- Checking 'oc' support for startup flags ...
    host-config-dir ... OK
    host-data-dir ... OK
    host-pv-dir ... OK
    host-volumes-dir ... OK
    routing-suffix ... OK
    Starting OpenShift using openshift/origin:v3.7.1 ...
    OpenShift server started.

    The server is accessible via web console at:
        https://192.168.99.100:8443

    You are logged in as:
        User:     developer
        Password: <any value>

    To login as administrator:
        oc login -u system:admin

    -- Exporting of OpenShift images is occuring in background process with pid 9314.

    ```
1. Make sure that you are `oc` is authenticated to access OpenShift cluster.
   When you are running OpenShift via minishift any combination of user and password will be valid.
   I'll be using `developer` user in this tutorial.
   ```
    $ oc login -u developer https://192.168.99.100:8443
    The server uses a certificate signed by an unknown authority.
    You can bypass the certificate check, but any data you send to the server could be intercepted by others.
    Use insecure connections? (y/n): y

    Authentication required for https://192.168.99.100:8443 (openshift)
    Username: developer
    Password:
    Login successful.

    You don't have any projects. You can try to create a new project, by running

        oc new-project <projectname>

    Welcome! See 'oc help' to get started.
   ```
1. Now lets create our first application using ocdev
    ```
    $ ocdev application create my-first-app
    Creating application: my-first-app
    Switched to application: my-first-app
    ```
1. Every ocdev application is composed of one or multiple components.
    `ocdev component create` command is used to create new component.
    This command has two arguments. Fist one is type of the component and this argument is required. 
    The second one is name of the component, if you omit second argument name of the type will be used also as name of the component.
    We will be deploying sample SpringBoot application running in WildFly application server so we need to use `wildfly` component.
    ``` bash
    $ ocdev component create wildfly
    --> Found image fe56d0d (14 hours old) in image stream "openshift/wildfly" under tag "10.1" for "wildfly"

    WildFly 10.1.0.Final
    --------------------
    Platform for building and running JEE applications on WildFly 10.1.0.Final

    Tags: builder, wildfly, wildfly10

    * A source build using binary input will be created
      * The resulting image will be pushed to image stream "wildfly:latest"
      * A binary build was created, use 'start-build --from-dir' to trigger a new build
    * This image will be deployed in deployment config "wildfly"
    * Port 8080/tcp will be load balanced by service "wildfly"
      * Other containers can access this service through the hostname "wildfly"

--> Creating resources with label app=my-first-app,app.kubernetes.io/component-name=wildfly,app.kubernetes.io/name=my-first-app ...
    imagestream "wildfly" created
    buildconfig "wildfly" created
    deploymentconfig "wildfly" created
    service "wildfly" created
--> Success
    Build scheduled, use 'oc logs -f bc/wildfly' to track its progress.
    Application is not exposed. You can expose services to the outside world by executing one or more of the commands below:
     'oc expose svc/wildfly'
    Run 'oc status' to view your app.
    ```

1. Now we will deploy SpringBoot application.
    First get the source code of the app.
    ```
    $ git clone https://github.com/gshipley/bootwildfly
    Cloning into 'bootwildfly'...
    remote: Counting objects: 60, done.
    remote: Total 60 (delta 0), reused 0 (delta 0), pack-reused 60
    Unpacking objects: 100% (60/60), done.
    ```

    Go to the directory with your application.
    ```
    $ cd bootwildfly
    ```

    Now we will push source code to our wildfly component on the OpenShift cluster.
    ```
    $ ocdev component push
    pushing changes to component: wildfly
    changes successfully pushed to component: wildfly
    ```
    This command pushes current directory to current component. Run `ocdev component` to check what current component is.
1. expose (ocdev currently doesn't support creating routes) 
    ```
    $ oc expose svc wildfly
    route "wildfly" exposed

    $ oc get route wildfly
    NAME      HOST/PORT                                 PATH      SERVICES   PORT       TERMINATION   WILDCARD
    wildfly   wildfly-myproject.192.168.99.100.nip.io             wildfly    8080-tcp                 None

    ```