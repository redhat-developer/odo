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
- Run the following command to install the latest ocdev release on Linux or macOS -

`curl -L https://github.com/redhat-developer/ocdev/raw/master/scripts/install.sh | sh`

You can download latest master builds from [Bintray](https://dl.bintray.com/ocdev/ocdev/latest/) or 
builds for released versions from [GitHub releases page](https://github.com/redhat-developer/ocdev/releases).

### macOS
1. First you need enable `kadel/ocdev` Homebrew Tap:
    ```sh
    brew tap kadel/ocdev
    ```
2. 
    - If you want to install latest master build
    ```sh
    brew install kadel/ocdev/ocdev -- HEAD
    ```
    - If you want to install latest released version
    ```sh
    brew install kadel/ocdev/ocdev
    ```

### Linux
#### Debian/Ubuntu and other distributions using deb
1. First you need to add gpg [public key](https://bintray.com/user/downloadSubjectPublicKey?username=bintray) used to sign repositories.
    ```sh
    curl -L https://bintray.com/user/downloadSubjectPublicKey?username=bintray | apt-key add -
    ```
2. Add ocdev repository to your `/etc/apt/sources.list`
    - If you want to use latest master builds add  `deb https://dl.bintray.com/ocdev/ocdev-deb-dev stretch main` repository.
      ```sh
      echo "deb https://dl.bintray.com/ocdev/ocdev-deb-dev stretch main" | sudo tee -a /etc/apt/sources.list
      ```
    - If you want to use latest released version add  `deb https://dl.bintray.com/ocdev/ocdev-deb-releases stretch main` repository.
      ```sh
      echo "deb https://dl.bintray.com/ocdev/ocdev-deb-releases stretch main" | sudo tee -a /etc/apt/sources.list
      ```
3. Now you can install `ocdev` and you would install any other package.
   ```sh
   apt-get update
   apt-get install ocdev
   ```


#### Fedora/Centos/RHEL and other distribution using rpm
1. Add ocdev repository to your `/etc/yum.repos.d/`
    - If you want to use latest master builds save following text to `/etc/yum.repos.d/bintray-ocdev-ocdev-rpm-dev.repo`
        ```
        # /etc/yum.repos.d/bintray-ocdev-ocdev-rpm-dev.repo
        [bintraybintray-ocdev-ocdev-rpm-dev]
        name=bintray-ocdev-ocdev-rpm-dev
        baseurl=https://dl.bintray.com/ocdev/ocdev-rpm-dev
        gpgcheck=0
        repo_gpgcheck=0
        enabled=1
        ```
        Or you can download it using following command:
        ```sh
        sudo curl -L https://bintray.com/ocdev/ocdev-rpm-dev/rpm -o /etc/yum.repos.d/bintray-ocdev-ocdev-rpm-dev.repo
        ```
    - If you want to use latest released version save following text to `/etc/yum.repos.d/bintray-ocdev-ocdev-rpm-releases.repo`
        ```
        # /etc/yum.repos.d/bintray-ocdev-ocdev-rpm-releases.repo
        [bintraybintray-ocdev-ocdev-rpm-releases]
        name=bintray-ocdev-ocdev-rpm-releases
        baseurl=https://dl.bintray.com/ocdev/ocdev-rpm-releases
        gpgcheck=0
        repo_gpgcheck=0
        enabled=1
        ```
        Or you can download it using following command:
        ```sh
        sudo curl -L https://bintray.com/ocdev/ocdev-rpm-releases/rpm -o /etc/yum.repos.d/bintray-ocdev-ocdev-rpm-releases.repo
        ```
3. Now you can install `ocdev` and you would install any other package.
   ```sh
   yum install ocdev
   # or 'dnf install ocdev'
   ```

### Windows
Download latest master builds from Bintray [ocdev.exe](https://dl.bintray.com/ocdev/ocdev/latest/windows-amd64/:ocdev.exe) or 
builds for released versions from [GitHub releases page](https://github.com/kadel/ocdev/releases).

## Concepts

- An **_application_** is, well, your application! It consists of multiple microservices or components, that work individually to build the entire application.
- A **_component_** can be thought of as a microservice. Multiple components will make up an application. A component will have different attributes like storage, etc.
Multiple component types are currently supported, like nodejs, perl, php, python, ruby, etc.

## Getting Started

1. Create and start working on a new application:
`ocdev application create <application name>`

2. Deploy code in current directory:
`ocdev component create <component type> --dir=.`

3. Test your application and make changes to the code

4. Deploy changes to the application:
`ocdev component push`

5. Go back to 3.

## CLI Structure
```
ocdev --verbose : OpenShift CLI for Developers
    application : application
        create : create an application
        delete : delete the given application
        get --short : get the active application
        list : lists all the applications
    completion : Output shell completion code
    component : components of application
        create --binary --dir --git : component create <component_type> [component_name]
        delete : component delete <component_name>
        get --short : component get
        push --dir : component push
    storage --component : storage
        add --path --size : create storage and mount to component
        list : list storage attached to a component
        remove : remove storage from component
    version : Print the version of ocdev
```
*_autogenerated_
