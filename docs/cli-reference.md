# Overview of the Odo (OpenShift Do) CLI Structure

> Example application

```sh
  # Creating and deploying a Node.js project
  git clone https://github.com/openshift/nodejs-ex && cd nodejs-ex
  odo create nodejs
  odo push
  
  # Accessing your Node.js component
  odo url create
``` 

Odo (OpenShift Do) is a CLI tool for running OpenShift applications in a fast and automated matter. Odo reduces the complexity of deployment by adding iterative development without the worry of deploying your source code. 

Find more information at https://github.com/openshift/odo

# Syntax

#### List of Commands

|           NAME            |                                                DESCRIPTION                                                 |
|---------------------------|------------------------------------------------------------------------------------------------------------|
| [app](#app)               | Perform application operations (delete, describe, list)                                                    |
| [catalog](#catalog)       | Catalog related operations (describe, list, search)                                                        |
| [component](#component)   | Components of an application (create, delete, describe, get, link, list, log, push, unlink, update, watch) |
| [config](#config)         | Modifies configuration settings (set, unset, view)                                                         |
| [create](#create)         | Create a new component                                                                                     |
| [delete](#delete)         | Delete an existing component                                                                               |
| [describe](#describe)     | Describe the given component                                                                               |
| [link](#link)             | Link component to a service or component                                                                   |
| [list](#list)             | List all components in the current application                                                             |
| [log](#log)               | Retrieve the log for the given component                                                                   |
| [login](#login)           | Login to cluster                                                                                           |
| [logout](#logout)         | Log out of the current OpenShift session                                                                   |
| [preference](#preference) | Modifies preference settings (set, unset, view)                                                            |
| [project](#project)       | Perform project operations (create, delete, get, list, set)                                                |
| [push](#push)             | Push source code to a component                                                                            |
| [service](#service)       | Perform service catalog operations (create, delete, list)                                                  |
| [storage](#storage)       | Perform storage operations (create, delete, list, mount, unmount)                                          |
| [unlink](#unlink)         | Unlink component to a service or component                                                                 |
| [update](#update)         | Update the source code path of a component                                                                 |
| [url](#url)               | Expose component to the outside world (create, delete, list)                                               |
| [utils](#utils)           | Utilities for terminal commands and modifying Odo configurations (terminal)                                |
| [version](#version)       | Print the client version information                                                                       |
| [watch](#watch)           | Watch for changes, update component on change                                                              |


#### CLI Structure

```sh
odo --alsologtostderr --log_backtrace_at --log_dir --logtostderr --skip-connection-check --stderrthreshold --v --vmodule : Odo (OpenShift Do) (app, catalog, component, config, create, delete, describe, link, list, log, login, logout, preference, project, push, service, storage, unlink, update, url, utils, version, watch)
    app : Perform application operations (delete, describe, list)
        delete --force --output --project : Delete the given application
        describe --output --project : Describe the given application
        list --output --project : List all applications in the current project
    catalog : Catalog related operations (describe, list, search)
        describe : Describe catalog item (service)
            service : Describe a service
        list : List all available component & service types. (components, services)
            components : List all components available.
            services : Lists all available services
        search : Search available component & service types. (component, service)
            component : Search component type in catalog
            service : Search service type in catalog
    component --app --context --project --short : Components of an application (create, delete, describe, get, link, list, log, push, unlink, update, watch)
        create --app --binary --context --cpu --env --git --max-cpu --max-memory --memory --min-cpu --min-memory --port --project --ref : Create a new component
        delete --app --context --force --project : Delete an existing component
        describe --app --context --output --project : Describe the given component
        get --app --context --project --short : Get currently active component
        link --app --component --context --port --project --wait --wait-for-target : Link component to a service or component
        list --app --context --output --project : List all components in the current application
        log --app --context --follow --project : Retrieve the log for the given component
        push --config --context --ignore --show-log --source : Push source code to a component
        unlink --app --component --context --port --project --wait : Unlink component to a service or component
        update --app --context --git --local --project --ref : Update the source code path of a component
        watch --app --context --delay --ignore --project : Watch for changes, update component on change
    config : Modifies configuration settings (set, unset, view)
        set --context --env --force : Set a value in odo config file
        unset --context --env --force : Unset a value in odo config file
        view --context : View current configuration values
    create --app --binary --context --cpu --env --git --max-cpu --max-memory --memory --min-cpu --min-memory --port --project --ref : Create a new component
    delete --app --context --force --project : Delete an existing component
    describe --app --context --output --project : Describe the given component
    link --app --component --context --port --project --wait --wait-for-target : Link component to a service or component
    list --app --context --output --project : List all components in the current application
    log --app --context --follow --project : Retrieve the log for the given component
    login --certificate-authority --insecure-skip-tls-verify --password --token --username : Login to cluster
    logout : Log out of the current OpenShift session
    preference : Modifies preference settings (set, unset, view)
        set --force : Set a value in odo config file
        unset --force : Unset a value in odo preference file
        view : View current preference values
    project --short : Perform project operations (create, delete, get, list, set)
        create --wait : Create a new project
        delete --force : Delete a project
        get --short : Get the active project
        list --output : List all the projects
        set --short : Set the current active project
    push --config --context --ignore --show-log --source : Push source code to a component
    service : Perform service catalog operations (create, delete, list)
        create --app --parameters --plan --project --wait : Create a new service from service catalog using the plan defined and deploy it on OpenShift.
        delete --app --force --project : Delete an existing service
        list --app --project : List all services in the current application
    storage : Perform storage operations (create, delete, list, mount, unmount)
        create --app --component --output --path --project --size : Create storage and mount to a component
        delete --app --component --force --project : Delete storage from component
        list --all --app --component --output --project : List storage attached to a component
        mount --app --component --path --project : mount storage to a component
        unmount --app --component --project : Unmount storage from the given path or identified by its name, from the current component
    unlink --app --component --context --port --project --wait : Unlink component to a service or component
    update --app --context --git --local --project --ref : Update the source code path of a component
    url : Expose component to the outside world (create, delete, list)
        create --app --component --context --output --port --project : Create a URL for a component
        delete --app --component --context --force --project : Delete a URL
        list --app --component --context --output --project : List URLs
    utils : Utilities for terminal commands and modifying Odo configurations (terminal)
        terminal : Add Odo terminal support to your development environment
    version --client : Print the client version information
    watch --app --context --delay --ignore --project : Watch for changes, update component on change

```

## app

`app`

> Example using app

```sh
  # Delete the application
  odo app delete myapp
  # Describe 'webapp' application,
  odo app describe webapp
  # List all applications in the current project
  odo app list
  
  # List all applications in the specified project
  odo app list --project myproject
```


Performs application operations related to your OpenShift project.

## catalog

`catalog [options]`

> Example using catalog

```sh
  # Get the supported components
  odo catalog list components

  # Get the supported services from service catalog
  odo catalog list services

  # Search for a component
  odo catalog search component python

  # Search for a service
  odo catalog search service mysql

  # Describe a service
  odo catalog describe service mysql-persistent
```


Catalog related operations

## component

`component`

> Example using component

```sh
odo component
create

  See sub-commands individually for more examples
```




## config

`config`

> Example using config

```sh

  # For viewing the current local configuration
  odo config view

  # Set a configuration value in the local config
  odo config set Type java
  odo config set Name test
  odo config set MinMemory 50M
  odo config set MaxMemory 500M
  odo config set Memory 250M
  odo config set Ignore false
  odo config set MinCPU 0.5
  odo config set MaxCPU 2
  odo config set CPU 1
  
  # Set a env variable in the local config
  odo config set --env KAFKA_HOST=kafka --env KAFKA_PORT=6639

  # Unset a configuration value in the local config
  odo config unset Type
  odo config unset Name
  odo config unset MinMemory
  odo config unset MaxMemory
  odo config unset Memory
  odo config unset Ignore
  odo config unset MinCPU
  odo config unset MaxCPU
  odo config unset CPU
  
  # Unset a env variable in the local config
  odo config unset --env KAFKA_HOST --env KAFKA_PORT
```


Modifies Odo specific configuration settings within the config file. 


Available Local Parameters:
Application - Application is the name of application the component needs to be part of
CPU - The minimum and maximum CPU a component can consume
Ignore - Consider the .odoignore file for push and watch
MaxCPU - The maximum cpu a component can consume
MaxMemory - The maximum memory a component can consume
Memory - The minimum and maximum Memory a component can consume
MinCPU - The minimum cpu a component can consume
MinMemory - The minimum memory a component is provided
Name - The name of the component
Ports - Ports to be opened in the component
Project - Project is the name of the project the component is part of
Ref - Git ref to use for creating component from git source
SourceLocation - The path indicates the location of binary file or git source
SourceType - Type of component source - git/binary/local
Type - The type of component
Url - Url to access the compoent


## create

`create <component_type> [component_name] [flags]`

> Example using create

```sh
  # Create new Node.js component with the source in current directory.
  odo create nodejs
  
  # A specific image version may also be specified
  odo create nodejs:latest
  
  # Create new Node.js component named 'frontend' with the source in './frontend' directory
  odo create nodejs frontend --context ./frontend
  
  # Create a new Node.js component of version 6 from the 'openshift' namespace
  odo create openshift/nodejs:6 --context /nodejs-ex
  
  # Create new Wildfly component with binary named sample.war in './downloads' directory
  odo create wildfly wildly --binary ./downloads/sample.war
  
  # Create new Node.js component with source from remote git repository
  odo create nodejs --git https://github.com/openshift/nodejs-ex.git
  
  # Create new Node.js git component while specifying a branch, tag or commit ref
  odo create nodejs --git https://github.com/openshift/nodejs-ex.git --ref master
  
  # Create new Node.js git component while specifying a tag
  odo create nodejs --git https://github.com/openshift/nodejs-ex.git --ref v1.0.1
  
  # Create new Node.js component with the source in current directory and ports 8080-tcp,8100-tcp and 9100-udp exposed
  odo create nodejs --port 8080,8100/tcp,9100/udp
  
  # Create new Node.js component with the source in current directory and env variables key=value and key1=value1 exposed
  odo create nodejs --env key=value,key1=value1
  
  # For more examples, visit: https://github.com/openshift/odo/blob/master/docs/examples.md
  odo create python --git https://github.com/openshift/django-ex.git
  
  # Passing memory limits
  odo create nodejs --memory 150Mi
  odo create nodejs --min-memory 150Mi --max-memory 300 Mi
  
  # Passing cpu limits
  odo create nodejs --cpu 2
  odo create nodejs --min-cpu 200m --max-cpu 2
```


Create a configuration describing a component to be deployed on OpenShift. 

If a component name is not provided, it'll be auto-generated. 

By default, builder images will be used from the current namespace. You can explicitly supply a namespace by using: odo create namespace/name:version If version is not specified by default, latest will be chosen as the version. 

A full list of component types that can be deployed is available using: 'odo catalog list'

## delete

`delete <component_name>`

> Example using delete

```sh
  # Delete component named 'frontend'.
  odo delete frontend
```


Delete an existing component.

## describe

`describe [component_name]`

> Example using describe

```sh
  # Describe nodejs component,
  odo describe nodejs
```


Describe the given component.

## link

`link <service> --component [component] OR link <component> --component [component]`

> Example using link

```sh
  # Link the current component to the 'my-postgresql' service
  odo link my-postgresql
  
  # Link component 'nodejs' to the 'my-postgresql' service
  odo link my-postgresql --component nodejs
  
  # Link current component to the 'backend' component (backend must have a single exposed port)
  odo link backend
  
  # Link component 'nodejs' to the 'backend' component
  odo link backend --component nodejs
  
  # Link current component to port 8080 of the 'backend' component (backend must have port 8080 exposed)
  odo link backend --port 8080
```


Link component to a service or component

If the source component is not provided, the current active component is assumed.
In both use cases, link adds the appropriate secret to the environment of the source component. 
The source component can then consume the entries of the secret as environment variables.

For example:

We have created a frontend application called 'frontend' using:
odo create nodejs frontend

We've also created a backend application called 'backend' with port 8080 exposed:
odo create nodejs backend --port 8080

We can now link the two applications:
odo link backend --component frontend

Now the frontend has 2 ENV variables it can use:
COMPONENT_BACKEND_HOST=backend-app
COMPONENT_BACKEND_PORT=8080

If you wish to use a database, we can use the Service Catalog and link it to our backend:
odo service create dh-postgresql-apb --plan dev -p postgresql_user=luke -p postgresql_password=secret
odo link dh-postgresql-apb

Now backend has 2 ENV variables it can use:
DB_USER=luke
DB_PASSWORD=secret

## list

`list`

> Example using list

```sh
  # List all components in the application
  odo list
```


List all components in the current application.

## log

`log [component_name]`

> Example using log

```sh
  # Get the logs for the nodejs component
  odo log nodejs
```


Retrieve the log for the given component

## login

`login`

> Example using login

```sh
  # Log in interactively
  odo login
  
  # Log in to the given server with the given certificate authority file
  odo login localhost:8443 --certificate-authority=/path/to/cert.crt
  
  # Log in to the given server with the given credentials (basic auth)
  odo login localhost:8443 --username=myuser --password=mypass
  
  # Log in to the given server with the given credentials (token)
  odo login localhost:8443 --token=xxxxxxxxxxxxxxxxxxxxxxx
```


Login to cluster

## logout

`logout`

> Example using logout

```sh
  # Logout
  odo logout
```


Log out of the current OpenShift session

## preference

`preference`

> Example using preference

```sh

  # For viewing the current local preference
  odo preference view
  
  # For viewing the current global preference
  odo preference view

  # Set a preference value in the global preference
  odo preference set UpdateNotification false
  odo preference set NamePrefix "app"
  odo preference set Timeout 20

  # Unset a preference value in the global preference
  odo preference unset  UpdateNotification
  odo preference unset  NamePrefix
  odo preference unset  Timeout
```


Modifies Odo specific configuration settings within the global preference file. 


Available Parameters:
NamePrefix - Default prefix is the current directory name. Use this value to set a default name prefix
Timeout - Timeout (in seconds) for OpenShift server connection check
UpdateNotification - Controls if an update notification is shown or not (true or false)


## project

`project [options]`

> Example using project

```sh
  # Set the active project
  odo project set

  # Create a new project
  odo project create myproject

  # List all the projects
  odo project list

  # Delete a project
  odo project delete myproject

  # Get the active project
  odo project get
```


Perform project operations

## push

`push [component name]`

> Example using push

```sh
  # Push source code to the current component
  odo push
  
  # Push data to the current component from the original source.
  odo push
  
  # Push source code in ~/mycode to component called my-component
  odo push my-component --context ~/mycode
```


Push source code to a component.

## service

`service`

> Example using service

```sh
  # Create new postgresql service from service catalog using dev plan and name my-postgresql-db.
  odo service create dh-postgresql-apb my-postgresql-db --plan dev -p postgresql_user=luke -p postgresql_password=secret

  # Delete the service named 'mysql-persistent'
  odo service delete mysql-persistent

  # List all services in the application
  odo service list
```


Perform service catalog operations

## storage

`storage`

> Example using storage

```sh
  # Create storage of size 1Gb to a component
  odo storage create mystorage --path=/opt/app-root/src/storage/ --size=1Gi
  # Delete storage mystorage from the currently active component
  odo storage delete mystorage
  
  # Delete storage mystorage from component 'mongodb'
  odo storage delete mystorage --component mongodb
  # Unmount storage 'dbstorage' from current component
  odo storage unmount dbstorage
  
  # Unmount storage 'database' from component 'mongodb'
  odo storage unmount database --component mongodb
  
  # Unmount storage mounted to path '/data' from current component
  odo storage unmount /data
  
  # Unmount storage mounted to path '/data' from component 'mongodb'
  odo storage unmount /data --component mongodb
  # List all storage attached or mounted to the current component and
  # all unattached or unmounted storage in the current application
  odo storage list
```


Perform storage operations

## unlink

`unlink <service> --component [component] OR unlink <component> --component [component]`

> Example using unlink

```sh
  # Unlink the 'my-postgresql' service from the current component
  odo unlink my-postgresql
  
  # Unlink the 'my-postgresql' service  from the 'nodejs' component
  odo unlink my-postgresql --component nodejs
  
  # Unlink the 'backend' component from the current component (backend must have a single exposed port)
  odo unlink backend
  
  # Unlink the 'backend' service  from the 'nodejs' component
  odo unlink backend --component nodejs
  
  # Unlink the backend's 8080 port from the current component
  odo unlink backend --port 8080
```


Unlink component or service from a component. 
For this command to be successful, the service or component needs to have been linked prior to the invocation using 'odo link'

## update

`update`

> Example using update

```sh
  # Change the source code path of a currently active component to local (use the current directory as a source)
  odo update --local
  
  # Change the source code path of the frontend component to local with source in ./frontend directory
  odo update frontend --local ./frontend
  
  # Change the source code path of a currently active component to git
  odo update --git https://github.com/openshift/nodejs-ex.git
  
  # Change the source code path of the component named node-ex to git
  odo update node-ex --git https://github.com/openshift/nodejs-ex.git
  
  # Change the source code path of the component named wildfly to a binary named sample.war in ./downloads directory
  odo update wildfly --binary ./downloads/sample.war
```


Update the source code path of a component

## url

`url`

> Example using url

```sh
  # Create a URL for the current component with a specific port
  odo url create --port 8080
  
  # Create a URL with a specific name and port
  odo url create example --port 8080
  
  # Create a URL with a specific name by automatic detection of port (only for components which expose only one service port)
  odo url create example
  
  # Create a URL with a specific name and port for component frontend
  odo url create example --port 8080 --component frontend
  # Delete a URL to a component
  odo url delete myurl
  # List the available URLs
  odo url list
```


Expose component to the outside world. 

The URLs that are generated using this command, can be used to access the deployed components from outside the cluster.

## utils

`utils`

> Example using utils

```sh
  # Bash terminal PS1 support
  source <(odo utils terminal bash)
  
  # Zsh terminal PS1 support
  source <(odo utils terminal zsh)

```


Utilities for terminal commands and modifying Odo configurations

## version

`version`

> Example using version

```sh
  # Print the client version of Odo
  odo version
```


Print the client version information

## watch

`watch [component name]`

> Example using watch

```sh
  # Watch for changes in directory for current component
  odo watch
  
  # Watch for changes in directory for component called frontend
  odo watch frontend
```


Watch for changes, update component on change.


