# Getting Started

This guide will get you started with developing your microservices iteratively on OpenShift using `odo`.

We will be developing a nodejs application in this guide, you can try along by getting the code for the application by running: `git clone https://github.com/kadel/nodejs-ex`

#### Running OpenShift

The easiest way to get a single node OpenShift cluster is by using [minishift](https://docs.openshift.org/latest/minishift/index.html), but `odo` will work with any other OpenShift instance you are logged in to.

- Install minishift using this [installation guide](https://docs.openshift.org/latest/minishift/getting-started/installing.html)
- The `oc` binary can be installed `minishift oc-env` command as described [here](https://docs.openshift.org/latest/minishift/getting-started/quickstart.html#starting-minishift)
- Install odo using this [installation guide](/README.md#installation)

Make sure that the commands exist by running -
- `minishift version`
- `oc version`
- `odo version`

The output should look something like -
```console
$ minishift version
minishift v1.13.1+75352e5

$ oc version
oc v3.7.1+ab0f056
kubernetes v1.7.6+a08f5eeb62
features: Basic-Auth GSSAPI Kerberos SPNEGO

error: server took too long to respond with version information.

$ odo version
v0.0.1 (HEAD)
```

Next, start a local OpenShift cluster using minishift -
```console
$ minishift start       
-- Starting profile 'minishift'                                
-- Checking if requested hypervisor 'kvm' is supported on this platform ... OK
-- Checking if KVM driver is installed ...                     
   Driver is available at /usr/local/bin/docker-machine-driver-kvm ... 
   Checking driver binary is executable ... OK                 
-- Checking if Libvirt is installed ... OK                     
-- Checking if Libvirt default network is present ... OK       
-- Checking if Libvirt default network is active ... OK        
-- Checking the ISO URL ... OK 
-- Starting local OpenShift cluster using 'kvm' hypervisor ... 
-- Starting Minishift VM ............... OK                    
-- Checking for IP address ... OK                              
-- Checking if external host is reachable from the Minishift VM ... 
   Pinging 8.8.8.8 ... OK      
-- Checking HTTP connectivity from the VM ...                  
   Retrieving http://minishift.io/index.html ... OK            
-- Checking if persistent storage volume is mounted ... OK     
-- Checking available disk space ... 19% used OK               
-- OpenShift cluster will be configured with ...               
   Version: v3.7.1             
-- Checking 'oc' support for startup flags ...                 
   routing-suffix ... OK       
   host-config-dir ... OK      
   host-data-dir ... OK        
   host-pv-dir ... OK          
   host-volumes-dir ... OK     
Starting OpenShift using openshift/origin:v3.7.1 ...           
OpenShift server started.

The server is accessible via web console at:
    https://192.168.42.147:8443
```

Now login to the OpenShift cluster using the server address that was in the output of `minishift start` -
```console
$ oc login -u developer -p developer https://192.168.42.147:8443
Login successful.

You have one project on this server: "myproject"

Using project "myproject".
```

Make sure you are logged in to the cluster by running `oc whoami` command.

Now we can move on to creating our application using `odo`.

#### Create an application

An application is an umbrella under which you will build all the components (microservices) of your entire application.

Let's create an application -

```console
$ odo application create nodeapp 
Creating application: nodeapp
Switched to application: nodeapp
```

#### Create a component

Now that you have created an application, now add a component of type _nodejs_ to the application, from the current directory where our code lies.

```console
$ odo create nodejs --local=.
--> Found image 2809a54 (3 weeks old) in image stream "openshift/nodejs" under tag "6" for "nodejs"
--> Creating resources with label app=nodeapp,app.kubernetes.io/component-name=nodejs,app.kubernetes.io/name=nodeapp ...
    imagestream "nodejs" created       
    buildconfig "nodejs" created       
    deploymentconfig "nodejs" created  
    service "nodejs" created           
--> Success        
    Build scheduled, use 'oc logs -f bc/nodejs' to track its progress.
    Application is not exposed. You can expose services to the outside world by executing one or more of the commands below:
     'oc expose svc/nodejs'            
    Run 'oc status' to view your app.  

please wait, building application...   
Uploading directory "." as binary input for the build ...
build "nodejs-2" started               
Pushing image 172.30.1.1:5000/myproject/nodejs:latest ...
Push successful
```

Great news! Your component has been deployed on OpenShift now. Let's quickly check how it looks!

##### Connect to the component

We need create the URL so we can connect to it our application.
```
$ odo url create
Adding URL to component: nodejs
URL created for component: nodejs

nodejs - nodejs-myproject.192.168.42.147.nip.io
```

Now just open the URL `nodejs-myproject.192.168.42.147.nip.io` in the browser and you will be able to view your deployed application.

#### Push new changes to the component

Well, your application looks great, but now you've made some changes in the code. Let's deploy these changes and see how it looks.

The current component is already set to nodejs, which can confirm from `odo component get`, so all we need to do is -

```console
$ odo push
pushing changes to component: nodejs   
changes successfully pushed to component: nodejs               
```

And now simply refresh your application in the browser, and you'll be able to see the changes.

Now you can repeat this cycle over and over again. Keep on making changes and keep pushing using `$ odo push nodejs`

#### Add storage to a component

You need to add storage to your component, `odo` makes it very easy for you to do this.

```console
$ odo storage add nodestorage --path=/opt/app-root/src/storage/ --size=1Gi 
Added storage nodestorage to nodejs
```
That just added 1Gi of storage to your nodejs component on the given path. Now your data will persist over application restarts.

That's all, folks!
