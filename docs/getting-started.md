# Getting Started

We will be developing and deploying a Node.JS application to an OpenShift cluster in this guide.

We'll be going over the following steps:

1. Running OpenShift
2. Creating an application
3. Creating a component
4. Accessing the component
5. Pushing new changes to the component
6. Adding storage to the component


### 1. Running OpenShift

A few requirements before we proceed!

 - A running OpenShift cluster (we recommend using [`minishift`](https://docs.openshift.org/latest/minishift/getting-started/installing.html))
 - `oc` binary ([installation guide here](https://docs.openshift.org/latest/minishift/getting-started/quickstart.html#starting-minishift))
 - `odo` binary ([installation guide here](/README.md#installation))

The quickest way to deploy a development OpenShift cluster is by using [minishift](https://docs.openshift.org/latest/minishift/index.html). Alternatively, `odo` will automatically work with any OpenShift cluster you're currently logged into.


#### Running Minishift

If you lack a local development cluster, Minishift provides the easiest way of getting started:

```console
$ minishift start       
-- Starting profile 'minishift'
-- Checking if https://github.com is reachable ... OK
-- Checking if requested OpenShift version 'v3.9.0' is valid ... OK
-- Checking if requested OpenShift version 'v3.9.0' is supported ... OK
-- Checking if requested hypervisor 'kvm' is supported on this platform ... OK
-- Checking if KVM driver is installed ...
   Driver is available at /usr/local/bin/docker-machine-driver-kvm ...
   Checking driver binary is executable ... OK
-- Checking if Libvirt is installed ... OK
-- Checking if Libvirt default network is present ... OK
-- Checking if Libvirt default network is active ... OK
-- Checking the ISO URL ... OK
-- Checking if provided oc flags are supported ... OK
-- Starting local OpenShift cluster using 'kvm' hypervisor ...
-- Starting Minishift VM .................. OK
-- Checking for IP address ... OK
-- Checking for nameservers ... OK
-- Checking if external host is reachable from the Minishift VM ...
   Pinging 8.8.8.8 ... OK
-- Checking HTTP connectivity from the VM ...
   Retrieving http://minishift.io/index.html ... FAIL
   VM cannot connect to external URL with HTTP
-- Checking if persistent storage volume is mounted ... OK
-- Checking available disk space ... 11% used OK
-- OpenShift cluster will be configured with ...
   Version: v3.9.0
-- Copying oc binary from the OpenShift container image to VM .... OK
-- Starting OpenShift cluster .................
Deleted existing OpenShift container
Using Docker shared volumes for OpenShift volumes
Using public hostname IP 192.168.42.10 as the host IP
Using 192.168.42.10 as the server IP
Starting OpenShift using openshift/origin:v3.9.0 ...
OpenShift server started.

The server is accessible via web console at:
    https://192.168.42.10:8443
```

Now log into the OpenShift cluster:

```sh
$ oc login -u developer -p developer
Login successful.

You have one project on this server: "myproject"

Using project "myproject".
```

Now we can move on to creating our application using `odo`.

### 2. Create an application

An application is an umbrella that will comprise all the components (microservices) you will build.

Let's create an application:

```console
$ odo app create nodeapp
Creating application: nodeapp
Switched to application: nodeapp
```

### 3. Create a component

First, we'll download the our test application: 

```console
$ git clone https://github.com/openshift/nodejs-ex
Cloning into 'nodejs-ex'...
remote: Counting objects: 568, done.
remote: Total 568 (delta 0), reused 0 (delta 0), pack-reused 568
Receiving objects: 100% (568/568), 174.63 KiB | 1.53 MiB/s, done.
Resolving deltas: 100% (224/224), done.

$ cd nodejs-ex 
~/nodejs-ex  master
```

Now that you've created an application, add a component of type _nodejs_ to the application, from the current directory where our code lies:

```console
$ odo create nodejs --local=.
Please wait, creating nodejs component ...
Component 'nodejs' was created.
To push source code to the component run 'odo push'

Component 'nodejs' is now set as active component.
```

Note:

By default, builder images will be used from the default OpenShift namespace.
You can explicitly supply a namespace by using: `odo create opesnshift/nodejs:8`
For more details please refer https://github.com/redhat-developer/odo/blob/master/docs/cli-reference.md#create

Now that a component is running we'll go ahead and push our initial source code!

```sh
$ odo push
Pushing changes to component: nodejs
sending incremental file list
...
---> Installing application source ...
---> Installing all dependencies
...
---> Building in production mode
---> Pruning the development dependencies
/opt/app-root/src/.npm is not a mountpoint
---> Cleaning the npm cache /opt/app-root/src/.npm
/tmp is not a mountpoint
---> Cleaning the /tmp/npm-*
run: ERROR (not running)
run: started
changes successfully pushed to component: nodejs
```

Great news! Your component has been deployed to OpenShift! Now we'll connect to the component.

### 4. Accessing the component

To access the component, we'll need to create an OpenShift route:

```console
$ odo url create
Adding URL to component: nodejs
URL created for component: nodejs

nodejs - nodejs-myproject.192.168.42.147.nip.io
```

Now simply access the URL `nodejs-myproject.192.168.42.147.nip.io` in the browser and you will be able to view your deployed application.

### 5. Pushing new changes to the component

Let's make some changes to the code and push them.

Edit one of the layout files within the Node.JS directory.

```sh
$ vim views/index.html
```

Now let's push the changes:

```console
$ odo push
Pushing changes to component: nodejs
sending incremental file list
...
changes successfully pushed to component: nodejs
```

Refresh your application in the browser, and you'll be able to see the changes.

After each change, you can continue updating your component by using: `odo push nodejs`.

### 6. Adding storage to a component

Now that you've got your component running, how do you add persistent any data between restarts?

If you wish to add storage to your component, `odo` makes it very easy for you to do this:

```console
$ odo storage create nodestorage --path=/opt/app-root/src/storage/ --size=1Gi 
Added storage nodestorage to nodejs
```

That's it! Storage has been added your component with an allocated size of 1 Gb.

## Extra OpenShift notes

These are some extra installation / getting started instructions for your local OpenShift cluster.

### Service Catalog

In order to use the Service Catalog it must be enabled within your OpenShift cluster.

Requirements:
  - `minishift` version 1.22+

If you are using `minishift` you'll need to start an OpenShift cluster with version 3.10.0+ and Service Catalog explicitly enabled.

```sh
# Deploy minishift
MINISHIFT_ENABLE_EXPERIMENTAL=y minishift start --extra-clusterup-flags "--enable=*,service-catalog,automation-service-broker"
```

After you've enabled / started `minishift`, you'll be able to list the services via `odo catalog list services` and service catalog related operations via `odo service <verb> <servicename>`.
