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

---
# Introduction to devfile

What is a devfile?

A [devfile](https://redhat-developer.github.io/devfile/) is a portable
file that describes your development environment. It allows for a
*portable* developmental environment without the need of
reconfiguration.

With a devfile you can describe:

  - The source code being used

  - Development components such as IDE tools (VSCode) and application
    runtimes (Yarn / NPM)

  - A list of pre-defined commands that can be run

  - Projects to initially clone

Odo takes this devfile and transforms it into a workspace of multiple
containers running on OpenShift, Kubernetes or Docker.

Devfiles are YAML files with a defined
[schema](https://github.com/redhat-developer/devfile/blob/master/docs/devfile.md).

# Odo and devfile

Odo can now create components from devfiles as recorded in registries.
Odo automatically consults the default
[devfile](https://github.com/elsony/devfile-registry)
[registries](https://github.com/eclipse/che-devfile-registry/) but users
can also add their own registries. Devfiles contribute new component
types that users can pull to begin development immediately.

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

**Prerequisites**

  - Before proceeding, you must know your ingress domain cluster name.
    For example: `apps-crc.testing` is the cluster domain name for [Red
    Hat CodeReady Containers](https://github.com/code-ready/crc)

  - Enable experimental mode for odo. This can be done by: `odo
    preference set experimental true`

# Creating a project

Create a project to keep your source code, tests, and libraries
organized in a separate single unit.

1.  Log in to a OpenShift cluster:
    
    ``` sh
      $ odo login -u developer -p developer
    ```

2.  Create a project:
    
    ``` sh
      $ odo project create myproject
       ✓  Project 'myproject' is ready for use
       ✓  New project created and now using project : myproject
    ```
    
    \== Listing all available devfile components
    
      - Before deploying your first component, have a look at what is
        available:
        
        ``` sh
          $ odo catalog list components
          Odo OpenShift Components:
          NAME              PROJECT       TAGS                        SUPPORTED
          java              openshift     11,8,latest                 YES
          nodejs            openshift     10-SCL,8,8-RHOAR,latest     YES
          dotnet            openshift     2.1,2.2,3.0,latest          NO
          golang            openshift     1.11.5,latest               NO
          httpd             openshift     2.4,latest                  NO
          modern-webapp     openshift     10.x,latest                 NO
          nginx             openshift     1.10,1.12,latest            NO
          perl              openshift     5.24,5.26,latest            NO
          php               openshift     7.0,7.1,7.2,latest          NO
          python            openshift     2.7,3.6,latest              NO
          ruby              openshift     2.4,2.5,latest              NO
        
          Odo Devfile Components:
          NAME                 DESCRIPTION                           SUPPORTED
          maven                Upstream Maven and OpenJDK 11         YES
          nodejs               Stack with NodeJS 10                  YES
          openLiberty          Open Liberty microservice in Java     YES
          java-spring-boot     Spring Boot® using Java               YES
        ```

In our example, we will be using `java-spring-boot` to deploy a sample
[Springboot](https://spring.io/projects/spring-boot) component.

# Deploying a Java Spring Boot® component to an OpenShift cluster

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

3.  Create a component configuration using the `java-spring-boot`
    component-type named `myspring`:
    
    ``` sh
       $ odo create java-spring-boot myspring
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
     $ odo url create --host apps-crc.testing
      ✓  URL myspring-8080.apps-crc.testing created for component: myspring
    
     To apply the URL configuration changes, please use odo push
    ```
    
    > **Note**
    > 
    > You must use your cluster host domain name when creating your URL.

6.  Push the component to the cluster:
    
    ``` sh
      $ odo push
    
      Validation
       ✓  Validating the devfile [81808ns]
    
      Creating Kubernetes resources for component myspring
       ✓  Waiting for component to start [5s]
    
      Applying URL changes
       ✓  URL myspring-8080: http://myspring-8080.apps-crc.testing created
    
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
     myspring-8080     http://myspring-8080.apps-crc.testing     8080     false
    ```

8.  View your deployed application using the generated URL:
    
    ``` sh
      $ curl http://myspring-8080.apps-crc.testing
    ```

9.  To delete your deployed application:
    
    ``` sh
      $ odo delete
      ? Are you sure you want to delete the devfile component: myspring? Yes
       ✓  Deleting devfile component myspring [152ms]
       ✓  Successfully deleted component
    ```

# Deploying a Node.js® component to an OpenShift cluster

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
     app  LICENSE  package.json  package-lock.json  README.md
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
     $ odo url create --host apps-crc.testing
      ✓  URL mynodejs-8080.apps-crc.testing created for component: mynodejs
    
     To apply the URL configuration changes, please use odo push
    ```
    
    > **Note**
    > 
    > You must use your cluster host domain name when creating your URL.

6.  Push the component to the cluster:
    
    ``` sh
      $ odo push
    
      Validation
       ✓  Validating the devfile [89380ns]
    
      Creating Kubernetes resources for component mynodejs
       ✓  Waiting for component to start [3s]
    
      Applying URL changes
       ✓  URL mynodejs-3000: http://mynodejs-3000.apps-crc.testing created
    
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
         mynodejs-8080     http://mynodejs-8080.apps-crc.testing     8080     false
    ```

8.  View your deployed application using the generated URL:
    
    ``` sh
       $ curl http://mynodejs-8080.apps-crc.testing
    ```

9.  To delete your deployed application:
    
    ``` sh
       $ odo delete
       ? Are you sure you want to delete the devfile component: mynodejs? Yes
        ✓  Deleting devfile component mynodejs [139ms]
        ✓  Successfully deleted component
    ```

# Deploying a Java Spring Boot® component locally to Docker

In this example, we will be deploying the same Java Spring Boot®
component we did earlier, but to a locally running Docker instance.

**Prerequisites:** Docker `17.05` or higher installed

1.  Enabling a separate push target, using the `pushtarget` preference:
    
    ``` sh
     $ odo preference set pushtarget docker
     Global preference was successfully updated
    ```

2.  Download the example Spring Boot® component.
    
    ``` sh
     $ git clone https://github.com/odo-devfiles/springboot-ex
    ```
    
    Alternatively, you can pass in `--starter` to `odo create` to have
    odo download a project specified in the devfile.

3.  Change the current directory to the component directory:
    
    ``` sh
     $ cd <directory-name>
    ```

4.  Create a component configuration using the `java-spring-boot`
    component-type named `mydockerspringboot`:
    
    ``` sh
       $ odo create java-spring-boot mydockerspringboot
       Experimental mode is enabled, use at your own risk
    
       Validation
        ✓  Checking devfile compatibility [195728ns]
        ✓  Creating a devfile component from registry: DefaultDevfileRegistry [170275ns]
        ✓  Validating devfile component [281940ns]
    
        Please use odo push command to create the component with source deployed
    ```

5.  Create a URL in order to access the deployed component:
    
    ``` sh
     $ odo url create --port 8080
      ✓  URL java-spring-boot-8080 created for component: java-spring-boot with exposed port: 59382
    
     To apply the URL configuration changes, please use odo push
    ```
    
    In order to access the docker application, exposed ports are
    required and automatically generated by odo.

6.  Deploy the Spring Boot® devfile component to Docker:
    
    ``` sh
      $ odo push
    
      Validation
       ✓  Validating the devfile [52685ns]
    
      Creating Docker resources for component java-spring-boot
       ✓  Pulling image maysunfaisal/springbootbuild [879ms]
       ✓  Starting container for maysunfaisal/springbootbuild [397ms]
       ✓  Pulling image maysunfaisal/springbootruntime [1s]
       ✓  URL 127.0.0.1:59382 created
       ✓  Starting container for maysunfaisal/springbootruntime [380ms]
    
      Syncing to component java-spring-boot
       ✓  Checking files for pushing [2ms]
       ✓  Syncing files to the component [231ms]
    
      Executing devfile commands for component java-spring-boot
       ✓  Executing devbuild command "/artifacts/bin/build-container-full.sh" [1m]
       ✓  Executing devrun command "/artifacts/bin/start-server.sh" [1s]
    
      Pushing devfile component java-spring-boot
       ✓  Changes successfully pushed to component
    ```
    
    When odo deploys a devfile component, it pulls the images for each
    `dockercontainer` in `devfile.yaml` and deploys them.
    
    Each docker container that is deployed is labeled with the name of
    the odo component.
    
    Docker volumes are created for the project source, and any other
    volumes defined in the devfile and mounted to the necessary
    containers.

7.  View your deployed application using the generated URL:
    
    ``` sh
       $ curl http://127.0.0.1:59382
    ```

8.  To delete your deployed application:
    
    ``` sh
       $ odo delete
       ? Are you sure you want to delete the devfile component: java-spring-boot? Yes
        ✓  Deleting devfile component java-spring-boot [139ms]
        ✓  Successfully deleted component
    ```
