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

Odo (OpenShift Do) is a CLI tool for running OpenShift applications in a fast and automated matter. Odo reduces the complexity of deployment by adding iterative development without the worry of deploying your source code. Find more information at https://github.com/redhat-developer/odo

# Syntax

#### List of Commands

|          NAME           |                           DESCRIPTION                            |
|-------------------------|------------------------------------------------------------------|
| [app](#app)             | Perform application operations                                   |
| [catalog](#catalog)     | Catalog related operations                                       |
| [component](#component) | Components of application.                                       |
| [create](#create)       | Create a new component                                           |
| [delete](#delete)       | Delete an existing component                                     |
| [describe](#describe)   | Describe the given component                                     |
| [link](#link)           | Link component to a service or component                         |
| [list](#list)           | List all components in the current application                   |
| [log](#log)             | Retrieve the log for the given component.                        |
| [login](#login)         | Login to cluster                                                 |
| [logout](#logout)       | Log out of the current OpenShift session                         |
| [project](#project)     | Perform project operations                                       |
| [push](#push)           | Push source code to a component                                  |
| [service](#service)     | Perform service catalog operations                               |
| [storage](#storage)     | Perform storage operations                                       |
| [update](#update)       | Update the source code path of a component                       |
| [url](#url)             | Expose component to the outside world                            |
| [utils](#utils)         | Utilities for terminal commands and modifying Odo configurations |
| [version](#version)     | Print the client version information                             |
| [watch](#watch)         | Watch for changes, update component on change                    |


#### CLI Structure

```sh
odo --alsologtostderr --log_backtrace_at --log_dir --logtostderr --skip-connection-check --stderrthreshold --v --vmodule : Odo (OpenShift Do)
    app --short : Perform application operations
        create --project : Create an application
        delete --force --project : Delete the given application
        describe --project : Describe the given application
        get --project --short : Get the active application
        list --project : List all applications in the current project
        set --project : Set application as active
    catalog : Catalog related operations
        describe : Describe catalog item
            service : Describe a service
        list : List all available component & service types.
            components : List all components available.
            services : Lists all available services
        search : Search available component & service types.
            component : Search component type in catalog
            service : Search service type in catalog
    component --short : Components of application.
        get --app --project --short : Get currently active component
        set --app --project : Set active component.
    create --app --binary --cpu --env --git --local --max-cpu --max-memory --memory --min-cpu --min-memory --port --project : Create a new component
    delete --app --force --project : Delete an existing component
    describe --app --project : Describe the given component
    link --app --component --port --project --wait : Link component to a service or component
    list --app --project : List all components in the current application
    log --app --follow --project : Retrieve the log for the given component.
    login --certificate-authority --insecure-skip-tls-verify --password --token --username : Login to cluster
    logout : Log out of the current OpenShift session
    project --short : Perform project operations
        create : Create a new project
        delete --force --short : Delete a project
        get --short : Get the active project
        list : List all the projects
        set --short : Set the current active project
    push --app --local --project : Push source code to a component
    service : Perform service catalog operations
        create --app --parameters --plan --project : Create a new service
        delete --app --force --project : Delete an existing service
        list --app --project : List all services in the current application
    storage : Perform storage operations
        create --app --component --path --project --size : Create storage and mount to a component
        delete --app --component --force --project : Delete storage from component
        list --all --app --component --project : List storage attached to a component
        mount --app --component --path --project : mount storage to a component
        unmount --app --component --project : Unmount storage from the given path or identified by its name, from the current component
    update --app --binary --git --local --project : Update the source code path of a component
    url : Expose component to the outside world
        create --app --component --open --port --project : Create a URL for a component
        delete --app --component --force --project : Delete a URL
        list --app --component --project : List URLs
    utils : Utilities for terminal commands and modifying Odo configurations
        config : Modifies configuration settings
            set : Set a value in odo config file
            view : View current configuration values
        terminal : Add Odo terminal support to your development environment
    version --client : Print the client version information
    watch --app --delay --ignore --project : Watch for changes, update component on change

```

## app

`app`

> Example using app

```sh
  # Create an application
  odo app create myapp
  odo app create
	
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
	
  # Describe the given service
  odo catalog describe service mysql-persistent
	
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

  # Passing memory limits
  odo create nodejs:latest --memory 150Mi
  odo create nodejs:latest --min-memory 150Mi --max-memory 300 Mi

  # Passing cpu limits
  odo create nodejs:latest --cpu 2
  odo create nodejs:latest --min-cpu 0.25 --max-cpu 2
  odo create nodejs:latest --min-cpu 200m --max-cpu 2

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

If a component name is not provided, it'll be auto-generated.

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

We have created a frontend application called 'frontend':

    odo create nodejs frontend

We've also created a backend application called 'backend' with port 8080 exposed:

    odo create nodejs backend --port 8080

You can now link the two applications:

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


Retrieve the log for the given component.

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


Perform service catalog operations, limited to template service broker and OpenShift Ansible Broker only.

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
  # Bash terminal PS1 support
  source <(odo utils terminal bash)

  # Zsh terminal PS1 support
  source <(odo utils terminal zsh)


   # Set a configuration value
   odo utils config set UpdateNotification false
   odo utils config set NamePrefix "app"
   odo utils config set timeout 20
	
  # For viewing the current configuration
   odo utils config view
  
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


