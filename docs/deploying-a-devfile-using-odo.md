---
# Page settings
layout: default
keywords:
comments: false

# Hero section
title: Deploying a devfile using odo
description: Deploying a portable devfile that describes your development environment

# Micro navigation
micro_nav: true

# Page navigation
page_nav:
    prev:
        content: Understanding odo
        url: '/docs/understanding-odo'
    next:
        content: Devfile file reference
        url: '/file-reference/'
---
# Introduction to devfile

What is a devfile?

A [devfile](https://redhat-developer.github.io/devfile/) is a portable
file that describes your development environment. It allows reproducing
a *portable* developmental environment without the need of
reconfiguration.

With a devfile you can describe:

  - Development components such as container definition for build and
    application runtimes

  - A list of pre-defined commands that can be run

  - Projects to initially clone

odo takes this devfile and transforms it into a workspace of multiple
containers running on OpenShift, Kubernetes or Docker.

Devfiles are YAML files with a defined
[schema](https://devfile.github.io/devfile/_attachments/api-reference.html).

# odo and devfile

odo can now create components from devfiles as recorded in registries.
odo automatically consults the [default
registry](https://github.com/odo-devfiles/registry) but users can also
add their own registries. Devfiles contribute new component types that
users can pull to begin development immediately.

An example deployment scenario:

1.  `odo create` will consult the recorded devfile registries to offer
    the user a selection of available component types and pull down the
    associated `devfile.yaml` file

2.  `odo push` parses and then deploys the component in the following
    order:
    
    1.  Parses and validates the YAML file
    
    2.  Deploys the development environment to your OpenShift cluster
    
    3.  Synchronizes your source code to the containers
    
    4.  Executes any prerequisite commands

# Deploying your first devfile

**Prerequisites for an OpenShift Cluster**

  - Create a project to keep your source code, tests, and libraries
    organized in a separate single unit.
    
    1.  Log in to an OpenShift cluster:
        
        ``` sh
          $ odo login -u developer -p developer
        ```
    
    2.  Create a project:
        
        ``` sh
          $ odo project create myproject
           ✓  Project 'myproject' is ready for use
           ✓  New project created and now using project : myproject
        ```

## Prerequisites for a Kubernetes Cluster

  - Before proceeding, you must know your ingress domain name or ingress
    IP to specify `--host` for `odo url create`.
    
    Ingress IP is usually the external IP of ingress controller service,
    for Minikube or CRC clusters running in a virtual machine you can
    get it by `minikube ip` or `crc ip`. Checkout this
    [document](https://kubernetes.io/docs/concepts/services-networking/ingress/)
    to know more about ingress.

# Listing all available devfile components

  - Before deploying your first component, have a look at what is
    available:
    
    ``` sh
      $ odo catalog list components
      Odo Devfile Components:
      NAME                 DESCRIPTION                            REGISTRY
      java-maven           Upstream Maven and OpenJDK 11          DefaultDevfileRegistry
      java-openliberty     Open Liberty microservice in Java      DefaultDevfileRegistry
      java-quarkus         Upstream Quarkus with Java+GraalVM     DefaultDevfileRegistry
      java-springboot      Spring Boot® using Java                DefaultDevfileRegistry
      nodejs               Stack with NodeJS 12                   DefaultDevfileRegistry
    
      Odo OpenShift Components:
      NAME        PROJECT       TAGS                                                                           SUPPORTED
      java        openshift     11,8,latest                                                                    YES
      dotnet      openshift     2.1,3.1,latest                                                                 NO
      golang      openshift     1.13.4-ubi7,1.13.4-ubi8,latest                                                 NO
      httpd       openshift     2.4-el7,2.4-el8,latest                                                         NO
      nginx       openshift     1.14-el7,1.14-el8,1.16-el7,1.16-el8,latest                                     NO
      nodejs      openshift     10-ubi7,10-ubi8,12-ubi7,12-ubi8,latest                                         NO
      perl        openshift     5.26-el7,5.26-ubi8,5.30-el7,latest                                             NO
      php         openshift     7.2-ubi7,7.2-ubi8,7.3-ubi7,7.3-ubi8,latest                                     NO
      python      openshift     2.7-ubi7,2.7-ubi8,3.6-ubi7,3.6-ubi8,3.8-ubi7,3.8-ubi8,latest                   NO
      ruby        openshift     2.5-ubi7,2.5-ubi8,2.6-ubi7,2.6-ubi8,2.7-ubi7,latest                            NO
      wildfly     openshift     10.0,10.1,11.0,12.0,13.0,14.0,15.0,16.0,17.0,18.0,19.0,20.0,8.1,9.0,latest     NO
    ```

In our example, we will be using `java-springboot` to deploy a sample
[Springboot](https://spring.io/projects/spring-boot)
component.

# Deploying a Java Spring Boot® component to an OpenShift / Kubernetes cluster

In this example we will be deploying an [example Spring Boot®
component](https://github.com/odo-devfiles/springboot-ex) that uses
[Maven](https://maven.apache.org/install.html) and Java 8 JDK.

1.  Download the example Spring Boot® component.
    
    ``` sh
     $ git clone https://github.com/odo-devfiles/springboot-ex
    ```
    
    Alternatively, you can pass in `--starter` to `odo create` to have
    odo download a project specified in the devfile.

2.  Change the current directory to the component directory:
    
    ``` sh
     $ cd <directory-name>
    ```

3.  Create a component configuration using the `java-springboot`
    component-type named `myspring`:
    
    ``` sh
       $ odo create java-springboot myspring
       Experimental mode is enabled, use at your own risk
    
       Validation
        ✓  Checking devfile compatibility [195728ns]
        ✓  Creating a devfile component from registry: DefaultDevfileRegistry [170275ns]
        ✓  Validating devfile component [281940ns]
    
        Please use odo push command to create the component with source deployed
    ```

4.  List the contents of the directory to see the devfile and sample
    Java application source code:
    
    ``` sh
      $ ls
      README.md devfile.yaml    pom.xml     src
    ```

5.  Create a URL in order to access the deployed component:
    
    ``` sh
     $ odo url create  --host example.com
      ✓  URL myspring-8080.example.com created for component: myspring
    
     To apply the URL configuration changes, please use odo push
    ```
    
    > **Note**
    > 
    > If deploying on Kubernetes, you need to pass ingress domain name
    > via `--host` flag.

6.  Push the component to the cluster:
    
    ``` sh
      $ odo push
    
      Validation
       ✓  Validating the devfile [81808ns]
    
      Creating Kubernetes resources for component myspring
       ✓  Waiting for component to start [5s]
    
      Applying URL changes
       ✓  URL myspring-8080: http://myspring-8080.example.com created
    
      Syncing to component myspring
       ✓  Checking files for pushing [2ms]
       ✓  Syncing files to the component [1s]
    
      Executing devfile commands for component myspring
       ✓  Executing devbuild command "/artifacts/bin/build-container-full.sh" [1m]
       ✓  Executing devrun command "/artifacts/bin/start-server.sh" [2s]
    
      Pushing devfile component myspring
       ✓  Changes successfully pushed to component
    ```

7.  List the URLs of the component:
    
    ``` sh
     $ odo url list
     Found the following URLs for component myspring
     NAME              URL                                       PORT     SECURE
     myspring-8080     http://myspring-8080.example.com     8080     false
    ```

8.  View your deployed application using the generated URL:
    
    ``` sh
      $ curl http://myspring-8080.example.com
    ```

9.  To delete your deployed application:
    
    ``` sh
      $ odo delete
      ? Are you sure you want to delete the devfile component: myspring? Yes
       ✓  Deleting devfile component myspring [152ms]
       ✓  Successfully deleted component
    ```

# Deploying a Node.js® component to an OpenShift / Kubernetes cluster

In this example we will be deploying an [example Node.js®
component](https://github.com/odo-devfiles/nodejs-ex) that uses
[NPM](https://www.npmjs.com/).

1.  Download the example Node.js® component
    
    ``` sh
     $ git clone https://github.com/odo-devfiles/nodejs-ex
    ```

2.  Change the current directory to the component directory:
    
    ``` sh
     $ cd <directory-name>
    ```

3.  List the contents of the directory to confirm that the application
    is indeed a Node.js® application:
    
    ``` sh
     $ ls
     LICENSE  package.json  package-lock.json  README.md  server.js  test
    ```

4.  Create a component configuration using the `nodejs` component-type
    named `mynodejs`:
    
    ``` sh
     $ odo create nodejs mynodejs
     Experimental mode is enabled, use at your own risk
    
     Validation
      ✓  Checking devfile compatibility [111738ns]
      ✓  Creating a devfile component from registry: DefaultDevfileRegistry [89567ns]
      ✓  Validating devfile component [186982ns]
    
     Please use odo push command to create the component with source deployed
    ```

5.  Create a URL in order to access the deployed component:
    
    ``` sh
     $ odo url create --host example.com
      ✓  URL mynodejs-8080.example.com created for component: mynodejs
    
     To apply the URL configuration changes, please use odo push
    ```
    
    > **Note**
    > 
    > If deploying on Kubernetes, you need to pass ingress domain name
    > via `--host` flag.

6.  Push the component to the cluster:
    
    ``` sh
      $ odo push
    
      Validation
       ✓  Validating the devfile [89380ns]
    
      Creating Kubernetes resources for component mynodejs
       ✓  Waiting for component to start [3s]
    
      Applying URL changes
       ✓  URL mynodejs-3000: http://mynodejs-3000.example.com created
    
      Syncing to component mynodejs
       ✓  Checking files for pushing [2ms]
       ✓  Syncing files to the component [1s]
    
      Executing devfile commands for component mynodejs
       ✓  Executing devbuild command "npm install" [3s]
       ✓  Executing devrun command "nodemon app.js" [2s]
    
      Pushing devfile component mynodejs
       ✓  Changes successfully pushed to component
    ```

7.  List the URLs of the component:
    
    ``` sh
     $ odo url list
         Found the following URLs for component mynodejs
         NAME              URL                                       PORT     SECURE
         mynodejs-8080     http://mynodejs-8080.example.com     8080     false
    ```

8.  View your deployed application using the generated URL:
    
    ``` sh
       $ curl http://mynodejs-8080.example.com
    ```

9.  To delete your deployed application:
    
    ``` sh
       $ odo delete
       ? Are you sure you want to delete the devfile component: mynodejs? Yes
        ✓  Deleting devfile component mynodejs [139ms]
        ✓  Successfully deleted component
    ```

# Deploying a Quarkus Application to an OpenShift / Kubernetes cluster

In this example we will be deploying a [Quarkus
component](https://github.com/odo-devfiles/quarkus-ex) that uses GraalVM
and JDK1.8+.

1.  Download the example Quarkus
    component
    
    ``` sh
     $ git clone https://github.com/odo-devfiles/quarkus-ex && cd quarkus-ex
    ```

2.  Create a Quarkus odo component
    
    ``` sh
       $ odo create java-quarkus myquarkus
       Experimental mode is enabled, use at your own risk
    
       Validation
        ✓  Checking devfile compatibility [195728ns]
        ✓  Creating a devfile component from registry: DefaultDevfileRegistry [170275ns]
        ✓  Validating devfile component [281940ns]
    
        Please use odo push command to create the component with source deployed
    ```

3.  Create a URL in order to access the deployed component:
    
    ``` sh
     $ odo url create  --host example.com
      ✓  URL myquarkus-8080.example.com created for component: myquarkus
    
     To apply the URL configuration changes, please use odo push
    ```
    
    > **Note**
    > 
    > If deploying on Kubernetes, you need to pass ingress domain name
    > via `--host` flag.

4.  Push the component to the cluster:
    
    ``` sh
      $ odo push
    
    Validation
     ✓  Validating the devfile [44008ns]
    
    Creating Kubernetes resources for component myquarkus
     ✓  Waiting for component to start [10s]
    
    Applying URL changes
     ✓  URLs are synced with the cluster, no changes are required.
    
    Syncing to component myquarkus
     ✓  Checking files for pushing [951138ns]
     ✓  Syncing files to the component [204ms]
    
    Executing devfile commands for component myquarkus
     ✓  Executing init-compile command "mvn compile" [3m]
     ✓  Executing dev-run command "mvn quarkus:dev" [1s]
    
    Pushing devfile component myquarkus
     ✓  Changes successfully pushed to component
    ```

5.  View your deployed application in a browser using the generated url
    
    ``` sh
     $ odo url list
     Found the following URLs for component myspring
     NAME              URL                                       PORT     SECURE
     myquarkus-8080     http://myquarkus-8080.example.com     8080     false
    ```

You can now continue developing your application. Just run `odo push`
and refresh your browser to view the latest changes.

You can also run `odo watch` to watch changes in the source code. Just
refreshing the browser will render the source code changes.

Run `odo delete` to delete the application from cluster.

1.  To delete your deployed application:
    
    ``` sh
       $ odo delete
       ? Are you sure you want to delete the devfile component: java-springboot? Yes
        ✓  Deleting devfile component java-springboot [139ms]
        ✓  Successfully deleted component
    ```
