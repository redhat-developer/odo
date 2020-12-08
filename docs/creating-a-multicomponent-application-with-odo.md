---
# Page settings
layout: default
keywords:
comments: false

# Hero section
title: Creating a multicomponent application with odo
description: Deploy a multicomponent application

# Micro navigation
micro_nav: true
---
`odo` allows you to create a multicomponent application, modify it, and link its components in an easy and automated way.

This example describes how to deploy a multicomponent application - a shooter game. The application consists of a front-end Node.js component and a back-end Java component.

  - `odo` is installed.

  - You have a running cluster. Developers can use [CodeReady Containers (CRC)](https://access.redhat.com/documentation/en-us/red_hat_codeready_containers/) to deploy a local cluster quickly.

  - Maven is installed.

# Creating a project

Create a project to keep your source code, tests, and libraries organized in a separate single unit.

1.  Log in to an OpenShift cluster:
    
    ``` terminal
    $ odo login -u developer -p developer
    ```

2.  Create a project:
    
    ``` terminal
    $ odo project create myproject
    ```
    
    **Example output.**
    
    ``` terminal
     ✓  Project 'myproject' is ready for use
     ✓  New project created and now using project : myproject
    ```

# Deploying the back-end component

To create a Java component, import the Java builder image, download the Java application and push the source code to your cluster with `odo`.

1.  Import `openjdk18` into the cluster:
    
    ``` terminal
    $ oc import-image openjdk18 \
    --from=registry.access.redhat.com/redhat-openjdk-18/openjdk18-openshift --confirm
    ```

2.  Tag the image as `builder` to make it accessible for odo:
    
    ``` terminal
    $ oc annotate istag/openjdk18:latest tags=builder
    ```

3.  Run `odo catalog list components` to see the created image:
    
    ``` terminal
    $ odo catalog list components
    ```
    
    **Example output.**
    
    ``` terminal
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

4.  Create a directory for your components:
    
    ``` terminal
    $ mkdir my_components && cd my_components
    ```

5.  Download the example back-end application:
    
    ``` terminal
    $ git clone https://github.com/openshift-evangelists/Wild-West-Backend backend
    ```

6.  Change to the back-end source directory:
    
    ``` terminal
    $ cd backend
    ```

7.  Check that you have the correct files in the directory:
    
    ``` terminal
    $ ls
    ```
    
    **Example output.**
    
    ``` terminal
    debug.sh  pom.xml  src
    ```

8.  Build the back-end source files with Maven to create a JAR file:
    
    ``` terminal
    $ mvn package
    ```
    
    **Example output.**
    
    ``` terminal
    ...
    [INFO] --------------------------------------
    [INFO] BUILD SUCCESS
    [INFO] --------------------------------------
    [INFO] Total time: 2.635 s
    [INFO] Finished at: 2019-09-30T16:11:11-04:00
    [INFO] Final Memory: 30M/91M
    [INFO] --------------------------------------
    ```

9.  Create a component configuration of Java component-type named `backend`:
    
    ``` terminal
    $ odo create --s2i openjdk18 backend --binary target/wildwest-1.0.jar
    ```
    
    **Example output.**
    
    ``` terminal
     ✓  Validating component [1ms]
     Please use `odo push` command to create the component with source deployed
    ```
    
    Now the configuration file `config.yaml` is in the local directory of the back-end component that contains information about the component for deployment.

10. Check the configuration settings of the back-end component in the `config.yaml` file using:
    
    ``` terminal
    $ odo config view
    ```
    
    **Example output.**
    
    ``` terminal
    COMPONENT SETTINGS
    ------------------------------------------------
    PARAMETER         CURRENT_VALUE
    Type              openjdk18
    Application       app
    Project           myproject
    SourceType        binary
    Ref
    SourceLocation    target/wildwest-1.0.jar
    Ports             8080/TCP,8443/TCP,8778/TCP
    Name              backend
    MinMemory
    MaxMemory
    DebugPort
    Ignore
    MinCPU
    MaxCPU
    ```

11. Push the component to the OpenShift cluster.
    
    ``` terminal
    $ odo push
    ```
    
    **Example output.**
    
    ``` terminal
    Validation
     ✓  Checking component [6ms]
    
    Configuration changes
     ✓  Initializing component
     ✓  Creating component [124ms]
    
    Pushing to component backend of type binary
     ✓  Checking files for pushing [1ms]
     ✓  Waiting for component to start [48s]
     ✓  Syncing files to the component [811ms]
     ✓  Building component [3s]
    ```
    
    Using `odo push`, OpenShift creates a container to host the back-end component, deploys the container into a pod running on the OpenShift cluster, and starts the `backend` component.

12. Validate:
    
      - The status of the action in odo:
        
        ``` terminal
        $ odo log -f
        ```
        
        **Example output.**
        
        ``` terminal
        2019-09-30 20:14:19.738  INFO 444 --- [           main] c.o.wildwest.WildWestApplication         : Starting WildWestApplication v1.0 onbackend-app-1-9tnhc with PID 444 (/deployments/wildwest-1.0.jar started by jboss in /deployments)
        ```
    
      - The status of the back-end component:
        
        ``` terminal
        $ odo list
        ```
        
        **Example output.**
        
        ``` terminal
        APP     NAME        TYPE          SOURCE                             STATE
        app     backend     openjdk18     file://target/wildwest-1.0.jar     Pushed
        ```

# Deploying the front-end component

To create and deploy a front-end component, download the Node.js application and push the source code to your cluster with `odo`.

1.  Download the example front-end application:
    
    ``` terminal
    $ git clone https://github.com/openshift/nodejs-ex frontend
    ```

2.  Change the current directory to the front-end directory:
    
    ``` terminal
    $ cd frontend
    ```

3.  List the contents of the directory to see that the front end is a Node.js application.
    
    ``` terminal
    $ ls
    ```
    
    **Example output.**
    
    ``` terminal
    README.md       openshift       server.js       views
    helm            package.json    tests
    ```
    
    > **Note**
    > 
    > The front-end component is written in an interpreted language (Node.js); it does not need to be built.

4.  Create a component configuration of Node.js component-type named `frontend`:
    
    ``` terminal
    $ odo create --s2i nodejs frontend
    ```
    
    **Example output.**
    
    ``` terminal
     ✓  Validating component [5ms]
    Please use `odo push` command to create the component with source deployed
    ```

5.  Push the component to a running container.
    
    ``` terminal
    $ odo push
    ```
    
    **Example output.**
    
    ``` terminal
    Validation
     ✓  Checking component [8ms]
    
    Configuration changes
     ✓  Initializing component
     ✓  Creating component [83ms]
    
    Pushing to component frontend of type local
     ✓  Checking files for pushing [2ms]
     ✓  Waiting for component to start [45s]
     ✓  Syncing files to the component [3s]
     ✓  Building component [18s]
     ✓  Changes successfully pushed to component
    ```

# Linking both components

Components running on the cluster need to be connected in order to interact. OpenShift provides linking mechanisms to publish communication bindings from a program to its clients.

1.  List all the components that are running on the cluster:
    
    ``` terminal
    $ odo list
    ```
    
    **Example output.**
    
    ``` terminal
    OpenShift Components:
    APP     NAME         PROJECT     TYPE          SOURCETYPE     STATE
    app     backend      testpro     openjdk18     binary         Pushed
    app     frontend     testpro     nodejs        local          Pushed
    ```

2.  Link the current front-end component to the back end:
    
    ``` terminal
    $ odo link backend --port 8080
    ```
    
    **Example output.**
    
    ``` terminal
     ✓  Component backend has been successfully linked from the component frontend
    
    Following environment variables were added to frontend component:
    - COMPONENT_BACKEND_HOST
    - COMPONENT_BACKEND_PORT
    ```
    
    The configuration information of the back-end component is added to the front-end component and the front-end component restarts.

# Exposing components to the public

1.  Navigate to the `frontend` directory:
    
    ``` terminal
    $ cd frontend
    ```

2.  Create an external URL for the application:
    
    ``` terminal
    $ odo url create frontend --port 8080
    ```
    
    **Example output.**
    
    ``` terminal
     ✓  URL frontend created for component: frontend
    
    To create URL on the OpenShift  cluster, use `odo push`
    ```

3.  Apply the changes:
    
    ``` terminal
    $ odo push
    ```
    
    **Example output.**
    
    ``` terminal
    Validation
     ✓  Checking component [21ms]
    
    Configuration changes
     ✓  Retrieving component data [35ms]
     ✓  Applying configuration [29ms]
    
    Applying URL changes
     ✓  URL frontend: http://frontend-app-myproject.192.168.42.79.nip.io created
    
    Pushing to component frontend of type local
     ✓  Checking file changes for pushing [1ms]
     ✓  No file changes detected, skipping build. Use the '-f' flag to force the build.
    ```

4.  Open the URL in a browser to view the application.

> **Note**
> 
> If an application requires permissions to the active service account to access the OpenShift namespace and delete active pods, the following error may occur when looking at `odo log` from the back-end component:
> 
> `Message: Forbidden!Configured service account doesn’t have access. Service account may have been revoked`
> 
> To resolve this error, add permissions for the service account role:
> 
> ``` terminal
> $ oc policy add-role-to-group view system:serviceaccounts -n <project>
> ```
> 
> ``` terminal
> $ oc policy add-role-to-group edit system:serviceaccounts -n <project>
> ```
> 
> Do not do this on a production cluster.

# Modifying the running application

1.  Change the local directory to the front-end directory:
    
    ``` terminal
    $ cd frontend
    ```

2.  Monitor the changes on the file system using:
    
    ``` terminal
    $ odo watch
    ```

3.  Edit the `index.html` file to change the displayed name for the game.
    
    > **Note**
    > 
    > A slight delay is possible before odo recognizes the change.
    
    odo pushes the changes to the front-end component and prints its status to the terminal:
    
    ``` terminal
    File /root/frontend/index.html changed
    File  changed
    Pushing files...
     ✓  Waiting for component to start
     ✓  Copying files to component
     ✓  Building component
    ```

4.  Refresh the application page in the web browser. The new name is now displayed.

# Deleting an application

Use the `odo app delete` command to delete your application.

1.  List the applications in the current project:
    
    ``` terminal
    $ odo app list
    ```
    
    **Example output.**
    
    ``` terminal
        The project '<project_name>' has the following applications:
        NAME
        app
    ```

2.  List the components associated with the applications. These components will be deleted with the application:
    
    ``` terminal
    $ odo component list
    ```
    
    **Example output.**
    
    ``` terminal
        APP     NAME                      TYPE       SOURCE        STATE
        app     nodejs-nodejs-ex-elyf     nodejs     file://./     Pushed
    ```

3.  Delete the application:
    
    ``` terminal
    $ odo app delete <application_name>
    ```
    
    **Example output.**
    
    ``` terminal
        ? Are you sure you want to delete the application: <application_name> from project: <project_name>
    ```

4.  Confirm the deletion with `Y`. You can suppress the confirmation prompt using the `-f` flag.
