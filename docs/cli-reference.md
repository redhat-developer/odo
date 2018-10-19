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

Find more information at https://github.com/redhat-developer/odo

# Syntax

#### List of Commands

|          NAME           |                                 DESCRIPTION                                  |
|-------------------------|------------------------------------------------------------------------------|
| [app](#app)             | Perform application operations                                               |
| [catalog](#catalog)     | Catalog related operations                                                   |
| [component](#component) | Components of application.                                                   |
| [create](#create)       | Create a new component                                                       |
| [delete](#delete)       | Delete an existing component                                                 |
| [describe](#describe)   | Describe the given component                                                 |
| [link](#link)           | Link component to a service                                                  |
| [list](#list)           | List all components in the current application                               |
| [log](#log)             | Retrieve the log for the given component.                                    |
| [project](#project)     | Perform project operations                                                   |
| [push](#push)           | Push source code to a component                                              |
| [service](#service)     | Perform service catalog operations                                           |
| [storage](#storage)     | Perform storage operations                                                   |
| [update](#update)       | Update the source code path of a component                                   |
| [url](#url)             | Expose component to the outside world                                        |
| [utils](#utils)         | Utilities for completion, terminal commands and modifying Odo configurations |
| [version](#version)     | Print the client version information                                         |
| [watch](#watch)         | Watch for changes, update component on change                                |


#### CLI Structure

```sh
odo --alsologtostderr --log_backtrace_at --log_dir --logtostderr --skip-connection-check --stderrthreshold --v --vmodule : Odo (Openshift Do)
    app --short : Perform application operations
        create : Create an application
        delete --force : Delete the given application
        describe : Describe the given application
        get --short : Get the active application
        list --project : List all applications in the current project
        set : Set application as active
    catalog : Catalog related operations
        list : List all available component & service types.
            components : List all components available.
            services : Lists all available services
        search : Search available component & service types.
            component : Search component type in catalog
            service : Search service type in catalog
    component --short : Components of application.
        get --short : Get currently active component
        set : Set active component.
    create --binary --env --git --local --port : Create a new component
    delete --force : Delete an existing component
    describe : Describe the given component
    link --component : Link component to a service
    list : List all components in the current application
    log --follow : Retrieve the log for the given component.
    project --short : Perform project operations
        create : Create a new project
        delete --force : Delete a project
        get --short : Get the active project
        list : List all the projects
        set --short : Set the current active project
    push --local : Push source code to a component
    service : Perform service catalog operations
        create --parameters --plan : Create a new service
        delete --force : Delete an existing service
        list : List all services in the current application
    storage : Perform storage operations
        create --component --path --size : Create storage and mount to a component
        delete --force : Delete storage from component
        list --all --component : List storage attached to a component
        mount --component --path : mount storage to a component
        unmount --component : Unmount storage from the given path or identified by its name, from the current component
    update --binary --git --local : Update the source code path of a component
    url : Expose component to the outside world
        create --application --component --port : Create a URL for a component
        delete --component --force : Delete a URL
        list --application --component : List URLs
    utils : Utilities for completion, terminal commands and modifying Odo configurations
        completion : Output shell completion code
        config : Modifies configuration settings
            set : Set a value in odo config file
            view : View current configuration values
        terminal : Add Odo terminal support to your development environment
    version : Print the client version information
    watch --delay --ignore : Watch for changes, update component on change

```

## app

`app`

> Example using app

```sh
  # Create an application
  odo app create myapp
	
  # Get the currently active application
  odo app get
	
  # Delete the application
  odo app delete myapp
	
  # Describe webapp application,
  odo app describe webapp
	
  # List all applications in the current project 
  odo app list

  # List all applications in the specified project
  odo app list --project myproject
	
  # Set an application as active
  odo app set myapp
	
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
	
```


Catalog related operations

## component

`component`

> Example using component

```sh
  # Get the currently active component
  odo component get
	
  # Set component named 'frontend' as active
  odo set component frontend
  
```




## create

`create <component_type> [component_name] [flags]`

> Example using create

```sh
  # Create new Node.js component with the source in current directory.
  odo create nodejs

  # A specific image version may also be specified
  odo create nodejs:latest

  # Create new Node.js component named 'frontend' with the source in './frontend' directory
  odo create nodejs frontend --local ./frontend

  # Create new Node.js component with source from remote git repository.
  odo create nodejs --git https://github.com/openshift/nodejs-ex.git

  # Create a new Node.js component of version 6 from the 'openshift' namespace
  odo create openshift/nodejs:6 --local /nodejs-ex

  # Create new Wildfly component with binary named sample.war in './downloads' directory
  odo create wildfly wildly --binary ./downloads/sample.war

  # Create new Node.js component with the source in current directory and ports 8080-tcp,8100-tcp and 9100-udp exposed
  odo create nodejs --port 8080,8100/tcp,9100/udp

  # Create new Node.js component with the source in current directory and env variables key=value and key1=value1 exposed
  odo create nodejs --env key=value,key1=value1

  # For more examples, visit: https://github.com/redhat-developer/odo/blob/master/docs/examples.md
  odo create python --git https://github.com/openshift/django-ex.git
	
```


Create a new component to deploy on OpenShift.

If component name is not provided, component type value will be used for the name.

By default, builder images will be used from the current namespace. You can explicitly supply a namespace by using: odo create namespace/name:version
If version is not specified by default, latest wil be chosen as the version.

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

`link <service> --component [component]`

> Example using link

```sh
  # Link the current component to the 'my-postgresql' service
  odo link my-postgresql

  # Link component 'nodejs' to the 'my-postgresql' service
  odo link my-postgresql --component nodejs
	
```


Link component to a service

If source component is not provided, the link is created to the current active
component.

During the linking process, the secret that is created during the service creation (odo service create),
is injected into the component.


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


Retrieve the log for the given component.

## project

`project [options]`

> Example using project

```sh
  # Set the current active project
  odo project set myproject
	
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
  odo push my-component --local ~/mycode
	
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


 Perform service catalog operations, Limited to template service broker only.

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
  # Bash autocompletion support
  source <(odo utils completion bash)

  # Zsh autocompletion support
  source <(odo utils completion zsh)

  # Bash terminal PS1 support
  source <(odo utils terminal bash)

  # Zsh terminal PS1 support
  source <(odo utils terminal zsh)


   # Set a configuration value
   odo utils config set UpdateNotification false
	
   # For viewing the current configuration
   odo utils config view
```


Utilities for completion, terminal commands and modifying Odo configurations

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


