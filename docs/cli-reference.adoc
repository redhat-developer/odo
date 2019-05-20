= Overview of the Odo (OpenShift Do) CLI Structure

___________________
Example application
___________________

[source,sh]
----
  # Creating and deploying a Node.js project
  git clone https://github.com/openshift/nodejs-ex && cd nodejs-ex
  odo create nodejs
  odo push
  
  # Accessing your Node.js component
  odo url create 
----

(OpenShift Do) odo is a CLI tool for running OpenShift applications in a fast and automated matter. Reducing the complexity of deployment, odo adds iterative development without the worry of deploying your source code. 

Find more information at https://github.com/openshift/odo

[[syntax]]
Syntax
------


.List of Commands
[width="100%",cols="21%,79%",options="header",]
|===
| Name | Description

| link:#app[app]
| Perform application operations (delete, describe, list)

| link:#catalog[catalog]
| Catalog related operations (describe, list, search)

| link:#component[component]
| Components of an application (create, delete, describe, link, list, log, push, unlink, update, watch)

| link:#config[config]
| Modifies configuration settings (set, unset, view)

| link:#create[create]
| Create a new component

| link:#delete[delete]
| Delete an existing component

| link:#describe[describe]
| Describe the given component

| link:#link[link]
| Link component to a service or component

| link:#list[list]
| List all components in the current application

| link:#log[log]
| Retrieve the log for the given component

| link:#login[login]
| Login to cluster

| link:#logout[logout]
| Log out of the current OpenShift session

| link:#preference[preference]
| Modifies preference settings (set, unset, view)

| link:#project[project]
| Perform project operations (create, delete, get, list, set)

| link:#push[push]
| Push source code to a component

| link:#service[service]
| Perform service catalog operations (create, delete, list)

| link:#storage[storage]
| Perform storage operations (create, delete, list)

| link:#unlink[unlink]
| Unlink component to a service or component

| link:#update[update]
| Update the source code path of a component

| link:#url[url]
| Expose component to the outside world (create, delete, list)

| link:#utils[utils]
| Utilities for terminal commands and modifying Odo configurations (terminal)

| link:#version[version]
| Print the client version information

| link:#watch[watch]
| Watch for changes, update component on change

|===

[[cli-structure]]
CLI Structure
+++++++++++++

[source,sh]
----
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
    component --app --context --project --short : Components of an application (create, delete, describe, link, list, log, push, unlink, update, watch)
        create --app --binary --context --cpu --env --git --max-cpu --max-memory --memory --min-cpu --min-memory --port --project --ref : Create a new component
        delete --all --app --context --force --project : Delete an existing component
        describe --app --context --output --project : Describe the given component
        get --app --context --project --short : Get currently active component
        link --app --component --context --port --project --wait --wait-for-target : Link component to a service or component
        list --app --context --output --project : List all components in the current application
        log --app --context --follow --project : Retrieve the log for the given component
        push --config --context --ignore --show-log --source : Push source code to a component
        unlink --app --component --context --port --project --wait : Unlink component to a service or component
        update --app --context --git --local --project --ref : Update the source code path of a component
        watch --app --context --delay --ignore --project --show-log : Watch for changes, update component on change
    config : Modifies configuration settings (set, unset, view)
        set --context --env --force : Set a value in odo config file
        unset --context --env --force : Unset a value in odo config file
        view --context : View current configuration values
    create --app --binary --context --cpu --env --git --max-cpu --max-memory --memory --min-cpu --min-memory --port --project --ref : Create a new component
    delete --all --app --context --force --project : Delete an existing component
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
    storage : Perform storage operations (create, delete, list)
        create --app --component --context --output --path --project --size : Create storage and mount to a component
        delete --app --component --context --force --project : Delete storage from component
        list --app --component --context --output --project : List storage attached to a component
    unlink --app --component --context --port --project --wait : Unlink component to a service or component
    update --app --context --git --local --project --ref : Update the source code path of a component
    url : Expose component to the outside world (create, delete, list)
        create --app --component --context --output --port --project : Create a URL for a component
        delete --app --component --context --force --project : Delete a URL
        list --app --component --context --output --project : List URLs
    utils : Utilities for terminal commands and modifying Odo configurations (terminal)
        terminal : Add Odo terminal support to your development environment
    version --client : Print the client version information
    watch --app --context --delay --ignore --project --show-log : Watch for changes, update component on change

----

[[app]]
app
~~~

[source,sh]
----
app
----

_________________
Example using app
_________________

[source,sh]
----
  # Delete the application
  odo app delete myapp
  # Describe 'webapp' application,
  odo app describe webapp
  # List all applications in the current project
  odo app list
  
  # List all applications in the specified project
  odo app list --project myproject
----

Performs application operations related to your OpenShift project.

[[catalog]]
catalog
~~~~~~~

[source,sh]
----
catalog [options]
----

_________________
Example using catalog
_________________

[source,sh]
----
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
----

Catalog related operations

[[component]]
component
~~~~~~~~~

[source,sh]
----
component
----

_________________
Example using component
_________________

[source,sh]
----
odo component
create

  See sub-commands individually for more examples
----



[[config]]
config
~~~~~~

[source,sh]
----
config
----

_________________
Example using config
_________________

[source,sh]
----

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
----

Modifies odo specific configuration settings within the config file. 


Available Local Parameters:
Application - Application is the name of application the component needs to be part of
CPU - The minimum and maximum CPU a component can consume
Ignore - Consider the .odoignore file for push and watch
MaxCPU - The maximum CPU a component can consume
MaxMemory - The maximum memory a component can consume
Memory - The minimum and maximum memory a component can consume
MinCPU - The minimum CPU a component can consume
MinMemory - The minimum memory a component is provided
Name - The name of the component
Ports - Ports to be opened in the component
Project - Project is the name of the project the component is part of
Ref - Git ref to use for creating component from git source
SourceLocation - The path indicates the location of binary file or git source
SourceType - Type of component source - git/binary/local
Storage - Storage of the component
Type - The type of component
Url - URL to access the component


[[create]]
create
~~~~~~

[source,sh]
----
create <component_type> [component_name] [flags]
----

_________________
Example using create
_________________

[source,sh]
----
  # Create new Node.js component with the source in current directory.
  odo create nodejs
  
  # A specific image version may also be specified
  odo create nodejs:latest
  
  # Create new Node.js component named 'frontend' with the source in './frontend' directory
  odo create nodejs frontend --context ./frontend
  
  # Create a new Node.js component of version 6 from the 'openshift' namespace
  odo create openshift/nodejs:6 --context /nodejs-ex
  
  # Create new Wildfly component with binary named sample.war in './downloads' directory
  odo create wildfly wildfly --binary ./downloads/sample.war
  
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
----

Create a configuration describing a component to be deployed on OpenShift. 

If a component name is not provided, it'll be auto-generated. 

By default, builder images will be used from the current namespace. You can explicitly supply a namespace by using: odo create namespace/name:version If version is not specified by default, latest will be chosen as the version. 

A full list of component types that can be deployed is available using: 'odo catalog list'

[[delete]]
delete
~~~~~~

[source,sh]
----
delete <component_name>
----

_________________
Example using delete
_________________

[source,sh]
----
  # Delete component named 'frontend'.
  odo delete frontend
  odo delete frontend --all
----

Delete an existing component.

[[describe]]
describe
~~~~~~~~

[source,sh]
----
describe [component_name]
----

_________________
Example using describe
_________________

[source,sh]
----
  # Describe nodejs component,
  odo describe nodejs
----

Describe the given component.

[[link]]
link
~~~~

[source,sh]
----
link <service> --component [component] OR link <component> --component [component]
----

_________________
Example using link
_________________

[source,sh]
----
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
----

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

[[list]]
list
~~~~

[source,sh]
----
list
----

_________________
Example using list
_________________

[source,sh]
----
  # List all components in the application
  odo list
----

List all components in the current application.

[[log]]
log
~~~

[source,sh]
----
log [component_name]
----

_________________
Example using log
_________________

[source,sh]
----
  # Get the logs for the nodejs component
  odo log nodejs
----

Retrieve the log for the given component

[[login]]
login
~~~~~

[source,sh]
----
login
----

_________________
Example using login
_________________

[source,sh]
----
  # Log in interactively
  odo login
  
  # Log in to the given server with the given certificate authority file
  odo login localhost:8443 --certificate-authority=/path/to/cert.crt
  
  # Log in to the given server with the given credentials (basic auth)
  odo login localhost:8443 --username=myuser --password=mypass
  
  # Log in to the given server with the given credentials (token)
  odo login localhost:8443 --token=xxxxxxxxxxxxxxxxxxxxxxx
----

Login to cluster

[[logout]]
logout
~~~~~~

[source,sh]
----
logout
----

_________________
Example using logout
_________________

[source,sh]
----
  # Logout
  odo logout
----

Log out of the current OpenShift session

[[preference]]
preference
~~~~~~~~~~

[source,sh]
----
preference
----

_________________
Example using preference
_________________

[source,sh]
----

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
----

Modifies Odo specific configuration settings within the global preference file. 


Available Parameters:
NamePrefix - Default prefix is the current directory name. Use this value to set a default name prefix
Timeout - Timeout (in seconds) for OpenShift server connection check
UpdateNotification - Controls if an update notification is shown or not (true or false)


[[project]]
project
~~~~~~~

[source,sh]
----
project [options]
----

_________________
Example using project
_________________

[source,sh]
----
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
----

Perform project operations

[[push]]
push
~~~~

[source,sh]
----
push [component name]
----

_________________
Example using push
_________________

[source,sh]
----
  # Push source code to the current component
  odo push
  
  # Push data to the current component from the original source.
  odo push
  
  # Push source code in ~/mycode to component called my-component
  odo push my-component --context ~/mycode
----

Push source code to a component.

[[service]]
service
~~~~~~~

[source,sh]
----
service
----

_________________
Example using service
_________________

[source,sh]
----
  # Create new postgresql service from service catalog using dev plan and name my-postgresql-db.
  odo service create dh-postgresql-apb my-postgresql-db --plan dev -p postgresql_user=luke -p postgresql_password=secret

  # Delete the service named 'mysql-persistent'
  odo service delete mysql-persistent

  # List all services in the application
  odo service list
----

Perform service catalog operations

[[storage]]
storage
~~~~~~~

[source,sh]
----
storage
----

_________________
Example using storage
_________________

[source,sh]
----
  # Create storage of size 1Gb to a component
  odo storage create mystorage --path=/opt/app-root/src/storage/ --size=1Gi
  # Delete storage mystorage from the currently active component
  odo storage delete mystorage
  
  # Delete storage mystorage from component 'mongodb'
  odo storage delete mystorage --component mongodb
  # List all storage attached or mounted to the current component and
  # all unattached or unmounted storage in the current application
  odo storage list
----

Perform storage operations

[[unlink]]
unlink
~~~~~~

[source,sh]
----
unlink <service> --component [component] OR unlink <component> --component [component]
----

_________________
Example using unlink
_________________

[source,sh]
----
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
----

Unlink component or service from a component. 
For this command to be successful, the service or component needs to have been linked prior to the invocation using 'odo link'

[[update]]
update
~~~~~~

[source,sh]
----
update
----

_________________
Example using update
_________________

[source,sh]
----
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
----

Update the source code path of a component

[[url]]
url
~~~

[source,sh]
----
url
----

_________________
Example using url
_________________

[source,sh]
----
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
----

Expose component to the outside world. 

The URLs that are generated using this command, can be used to access the deployed components from outside the cluster.

[[utils]]
utils
~~~~~

[source,sh]
----
utils
----

_________________
Example using utils
_________________

[source,sh]
----
  # Bash terminal PS1 support
  source <(odo utils terminal bash)
  
  # Zsh terminal PS1 support
  source <(odo utils terminal zsh)

----

Utilities for terminal commands and modifying Odo configurations

[[version]]
version
~~~~~~~

[source,sh]
----
version
----

_________________
Example using version
_________________

[source,sh]
----
  # Print the client version of Odo
  odo version
----

Print the client version information

[[watch]]
watch
~~~~~

[source,sh]
----
watch [component name]
----

_________________
Example using watch
_________________

[source,sh]
----
  # Watch for changes in directory for current component
  odo watch
  
  # Watch for changes in directory for component called frontend
  odo watch frontend
----

Watch for changes, update component on change.


