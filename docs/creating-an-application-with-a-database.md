---
# Page settings
layout: default
keywords:
comments: false

# Hero section
title: Creating an application with a database
description: This example describes how to deploy and connect a database to a front-end application.

# Micro navigation
micro_nav: true
---
This example describes how to deploy and connect a database to a
front-end application.

  - `odo` is installed.

  - `oc` client is installed.

  - You have a running cluster. Developers can use [CodeReady Containers
    (CRC)](https://access.redhat.com/documentation/en-us/red_hat_codeready_containers/)
    to deploy a local cluster quickly.

  - The Service Catalog is installed and enabled on your cluster.
    
    > **Note**
    > 
    > Service Catalog is deprecated on OpenShift 4 and later.

# Creating a project

Create a project to keep your source code, tests, and libraries
organized in a separate single unit.

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

# Deploying the front-end component

To create and deploy a front-end component, download the Node.js
application and push the source code to your cluster with `odo`.

1.  Download the example front-end application:
    
    ``` terminal
    $ git clone https://github.com/openshift/nodejs-ex frontend
    ```

2.  Change the current directory to the front-end directory:
    
    ``` terminal
    $ cd frontend
    ```

3.  List the contents of the directory to see that the front end is a
    Node.js application.
    
    ``` terminal
    $ ls
    ```
    
    **Example
    output.**
    
    ``` terminal
    assets  bin  index.html  kwww-frontend.iml  package.json  package-lock.json  playfield.png  README.md  server.js
    ```
    
    > **Note**
    > 
    > The front-end component is written in an interpreted language
    > (Node.js); it does not need to be built.

4.  Create a component configuration of Node.js component-type named
    `frontend`:
    
    ``` terminal
    $ odo create nodejs frontend
    ```
    
    **Example output.**
    
    ``` terminal
     ✓  Validating component [5ms]
    Please use `odo push` command to create the component with source deployed
    ```

5.  Create a URL to access the frontend interface.
    
    ``` terminal
    $ odo url create myurl
    ```
    
    **Example output.**
    
    ``` terminal
     ✓  URL myurl created for component: nodejs-nodejs-ex-pmdp
    ```

6.  Push the component to the OpenShift cluster.
    
    ``` terminal
    $ odo push
    ```
    
    **Example output.**
    
    ``` terminal
    Validation
     ✓  Checking component [7ms]
    
     Configuration changes
     ✓  Initializing component
     ✓  Creating component [134ms]
    
     Applying URL changes
     ✓  URL myurl: http://myurl-app-myproject.192.168.42.79.nip.io created
    
     Pushing to component nodejs-nodejs-ex-mhbb of type local
     ✓  Checking files for pushing [657850ns]
     ✓  Waiting for component to start [6s]
     ✓  Syncing files to the component [408ms]
     ✓  Building component [7s]
     ✓  Changes successfully pushed to component
    ```

# Deploying a database in interactive mode

odo provides a command-line interactive mode which simplifies
deployment.

  - Run the interactive mode and answer the prompts:
    
    ``` terminal
    $ odo service create
    ```
    
    **Example output.**
    
    ``` terminal
    ? Which kind of service do you wish to create database
    ? Which database service class should we use mongodb-persistent
    ? Enter a value for string property DATABASE_SERVICE_NAME (Database Service Name): mongodb
    ? Enter a value for string property MEMORY_LIMIT (Memory Limit): 512Mi
    ? Enter a value for string property MONGODB_DATABASE (MongoDB Database Name): sampledb
    ? Enter a value for string property MONGODB_VERSION (Version of MongoDB Image): 3.2
    ? Enter a value for string property VOLUME_CAPACITY (Volume Capacity): 1Gi
    ? Provide values for non-required properties No
    ? How should we name your service  mongodb-persistent
    ? Output the non-interactive version of the selected options No
    ? Wait for the service to be ready No
     ✓  Creating service [32ms]
     ✓  Service 'mongodb-persistent' was created
    Progress of the provisioning will not be reported and might take a long time.
    You can see the current status by executing 'odo service list'
    ```

> **Note**
> 
> Your password or username will be passed to the front-end application
> as environment variables.

# Deploying a database manually

1.  List the available services:
    
    ``` terminal
    $ odo catalog list services
    ```
    
    **Example output.**
    
    ``` terminal
    NAME                         PLANS
    django-psql-persistent       default
    jenkins-ephemeral            default
    jenkins-pipeline-example     default
    mariadb-persistent           default
    mongodb-persistent           default
    mysql-persistent             default
    nodejs-mongo-persistent      default
    postgresql-persistent        default
    rails-pgsql-persistent       default
    ```

2.  Choose the `mongodb-persistent` type of service and see the required
    parameters:
    
    ``` terminal
    $ odo catalog describe service mongodb-persistent
    ```
    
    **Example
    output.**
    
    ``` terminal
      ***********************        | *****************************************************
      Name                           | default
      -----------------              | -----------------
      Display Name                   |
      -----------------              | -----------------
      Short Description              | Default plan
      -----------------              | -----------------
      Required Params without a      |
      default value                  |
      -----------------              | -----------------
      Required Params with a default | DATABASE_SERVICE_NAME
      value                          | (default: 'mongodb'),
                                     | MEMORY_LIMIT (default:
                                     | '512Mi'), MONGODB_VERSION
                                     | (default: '3.2'),
                                     | MONGODB_DATABASE (default:
                                     | 'sampledb'), VOLUME_CAPACITY
                                     | (default: '1Gi')
      -----------------              | -----------------
      Optional Params                | MONGODB_ADMIN_PASSWORD,
                                     | NAMESPACE, MONGODB_PASSWORD,
                                     | MONGODB_USER
    ```

3.  Pass the required parameters as flags and wait for the deployment of
    the
    database:
    
    ``` terminal
    $ odo service create mongodb-persistent --plan default --wait -p DATABASE_SERVICE_NAME=mongodb -p MEMORY_LIMIT=512Mi -p MONGODB_DATABASE=sampledb -p VOLUME_CAPACITY=1Gi
    ```

# Connecting the database to the front-end application

1.  Link the database to the front-end service:
    
    ``` terminal
    $ odo link mongodb-persistent
    ```
    
    **Example
    output.**
    
    ``` terminal
     ✓  Service mongodb-persistent has been successfully linked from the component nodejs-nodejs-ex-mhbb
    
    Following environment variables were added to nodejs-nodejs-ex-mhbb component:
    - database_name
    - password
    - uri
    - username
    - admin_password
    ```

2.  See the environment variables of the application and the database in
    the Pod:
    
    1.  Get the Pod name:
        
        ``` terminal
        $ oc get pods
        ```
        
        **Example
        output.**
        
        ``` terminal
        NAME                                READY     STATUS    RESTARTS   AGE
        mongodb-1-gsznc                     1/1       Running   0          28m
        nodejs-nodejs-ex-mhbb-app-4-vkn9l   1/1       Running   0          1m
        ```
    
    2.  Connect to the Pod:
        
        ``` terminal
        $ oc rsh nodejs-nodejs-ex-mhbb-app-4-vkn9l
        ```
    
    3.  Check the environment variables:
        
        ``` terminal
        sh-4.2$ env
        ```
        
        **Example output.**
        
        ``` terminal
        uri=mongodb://172.30.126.3:27017
        password=dHIOpYneSkX3rTLn
        database_name=sampledb
        username=user43U
        admin_password=NCn41tqmx7RIqmfv
        ```

3.  Open the URL in the browser and notice the database configuration in
    the bottom right:
    
    ``` terminal
    $ odo url list
    ```
    
    **Example output.**
    
    ``` terminal
    Request information
    Page view count: 24
    
    DB Connection Info:
    Type:   MongoDB
    URL:    mongodb://172.30.126.3:27017/sampledb
    ```

# Deleting an application

> **Important**
> 
> Deleting an application will delete all components associated with the
> application.

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

2.  List the components associated with the applications. These
    components will be deleted with the application:
    
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
    
    **Example
    output.**
    
    ``` terminal
        ? Are you sure you want to delete the application: <application_name> from project: <project_name>
    ```

4.  Confirm the deletion with `Y`. You can suppress the confirmation
    prompt using the `-f` flag.
