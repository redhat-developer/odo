# Getting Started

This document walks you through the development and deployment of a Node.js application to an OpenShift cluster using odo. It guides you on how to push new changes to your application and on other useful features of odo.

## Developing and deploying an application to an OpenShift cluster
1. [Install the `odo` binary](/docs/installation.md).
2. Log into an OpenShift cluster. OpenShift Do automatically works with any OpenShift cluster you are currently logged into. If you lack a local development cluster, we recommend using [Minishift](https://docs.openshift.org/latest/minishift/getting-started/installing.html), the quickest and easiest way to deploy a development OpenShift cluster.

    If you use Minishift:
    1. Run Minishift:
    ```console
    $ minishift start
    ```
    2. Log into the OpenShift cluster:
    ```sh
    $ odo login -u developer -p developer
    ```

     **Note:** In order to make full use of odo functionality, it is recommended to enable the OpenShift [service catalog](https://docs.openshift.com/container-platform/3.11/architecture/service_catalog/index.html). Use `minishift` 1.30 or latest to enable this:
    ```sh
    $ MINISHIFT_ENABLE_EXPERIMENTAL=y minishift start --extra-clusterup-flags "--enable=*,service-catalog,automation-service-broker,template-service-broker"
    ```

3. An application is an umbrella that comprises all the components (microservices) you build.
Create an application:
```console
$ odo app create nodeapp
```

4. Create a component as follows:

    1. Download the sample application and change directory to the location of the application:
    ```console
    $ git clone https://github.com/openshift/nodejs-ex
    $ cd nodejs-ex
    ```

    2. Add a component of the type nodejs to the application:
    ```console
    $ odo create nodejs
    ```
    **Note:** By default, the latest image is used. You can also explicitly supply an image version by using `odo create openshift/nodejs:8`.

    3. Push the initial source code to the component:
    ```sh
    $ odo push
    ```
    Your component is now deployed to OpenShift

5. Access the component as follows:

    1. Create an OpenShift route:

    ```console
    $ odo url create
     ```
    2. Use the URL `nodejs-ex-nodejs-nnjf-nodeapp-myproject.192.168.42.90.nip.io` in the browser to view your deployed application.

6. Edit your code and push the changes to the component

    1. Edit one of the layout files within the Node.JS directory.
    ```sh
    $ vim views/index.html
    ```
    2. Push the changes:
    ```console
    $ odo push
    ```
    3. Refresh your application in the browser to see the changes.

    After each change, you can update your component using: `odo push nodejs`.



## Other key odo features
### Adding storage to the component

OpenShift Do enables you to persist data between restarts by making it easy to add storage to your component as follows:

```console
$ odo storage create nodestorage --path=/opt/app-root/src/storage/ --size=1Gi
```

This adds storage to your component with an allocated size of 1 Gb.

### Using command completion

**Note:** Currently command completion is only supported for bash, zsh and fish shells.

`odo` provides smart completion of command parameters based on user input. For this to work, `odo` needs to integrate with the
executing shell.

* To install command completion automatically, run `odo --complete` and press `y` when asked to install the completion hook.

* You can also install the completion hook manually by adding `complete -o nospace -C <full path to your odo binary> odo` to your shell configuration file (e.g. `.bashrc` for `bash`).

* To disable completion, run `odo --uncomplete`

After any modification to your shell configuration file, you will need to `source` it or restart your shell.

**NOTE**: If you either rename the odo executable or move it, the completion system stops working and you will need to re-enable it accordingly.

### Using the `.odoignore` and `.gitignore` files

The `.odoignore` file in the root directory of your application is used to ignore a list of files/patterns. This applies to both `odo push` and `odo watch`.

If the `.odoignore` file does *not* exist, the `.gitignore` file is used instead for ignoring specific files and folders.

For example, to ignore `.git` files, any files with the `.js` extension, and the folder `tests` add the following to either the `.odoignore` or the `.gitignore` file:

```sh
.git
*.js
/tests
```

The `.odoignore` file allows any [glob expressions](https://en.wikipedia.org/wiki/Glob_(programming) to be used, for example:

```sh
/openshift/**/*.json
```

### Using Service Catalog with odo

If you use `minishift` you need to use minishift version 1.22 or above.

In order to use the Service Catalog it must be enabled within your OpenShift cluster.

 1. Start an OpenShift cluster, version 3.10.0 and above
 2. Enable the Service Catalog:
 ```sh
  MINISHIFT_ENABLE_EXPERIMENTAL=y minishift start --extra-clusterup-flags "--enable=*,service-catalog,automation-service-broker"
 ```

3. After you enable or start `minishift` use:
 * `odo catalog list services` to list the services
  * `odo service <verb> <servicename>` to list service catalog related operations

### Adding a Custom Builder

OpenShift enables you to add a [custom image](https://docs.openshift.com/container-platform/3.7/creating_images/custom.html) to bridge the gap in the creation of custom images. A custom builder image usually includes the base image of [openshift/origin-custom-docker-builder](https://hub.docker.com/r/openshift/origin-custom-docker-builder/).

The following example demonstrates the successful import and use of [redhat-openjdk-18](registry.access.redhat.com/redhat-openjdk-18/openjdk18-openshift) image:

**Prerequisites:**
oc binary is [installed](https://docs.openshift.org/latest/cli_reference/get_started_cli.html#installing-the-cli) and present on the `$PATH`

**Procedure:**

 1. Import the image into OpenShift
```sh
oc import-image openjdk18 --from=registry.access.redhat.com/redhat-openjdk-18/openjdk18-openshift --confirm
```
 2. Tag the image to make it accessible to odo
```sh
oc annotate istag/openjdk18:latest tags=builder
```
 3. Deploy it with odo:
```sh
odo create openjdk18 --git https://github.com/openshift-evangelists/Wild-West-Backend
```
