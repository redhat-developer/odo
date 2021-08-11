---
title: Create Component
sidebar_position: 1
---
# Creating components using odo

[Component](../getting-started/basics#component) is the most basic unit of operation for odo. And the way to create one is using `odo create` (short for `odo component create`) command.

In simplest terms, when you "create" an odo component, you populate your current working directory with the file `devfile.yaml`. A Devfile is a manifest file that contains information about various resources (URL, Storage, Services, etc.) that correspond to your component, and will be created on the Kubernetes cluster when you execute `odo push` command. Most odo commands will first modify (add or remove configuration from) this file, and then subsequent `odo push` will create or delete the resources from the Kubernetes cluster.

However, odo users are not expected to know how the `devfile.yaml` is organized; it is the odo commands that would create, update, or delete it.

One final thing to keep in mind - there can be only one odo component in a directory. Nesting odo components is not expected to work well. In other terms, if you have multiple parts (components), say frontend and backend, of your microservices application that you want to create odo components for, you should put them in separate directories and not try to nest them. Take a look at example structure below:
```shell
$ tree my-awesome-microservices-app 
my-awesome-microservices-app
├── backend
│   └── devfile.yaml
└── frontend
    └── devfile.yaml
```
In this guide, we are going to create a Spring Boot component to deploy the [Spring Boot petclinic](https://github.com/spring-projects/spring-petclinic) project to a Kubernetes cluster.

Let's clone the project first:
```shell
git clone https://github.com/spring-projects/spring-petclinic
cd spring-petclinic
```

Next, let's create a project <!-- add link to project command reference here --> on the Kubernetes cluster in which we will be creating our component:
```shell
odo project create myproject
```
Alternatively, you could also use one of the existing projects on the cluster:
```shell
odo project list
```
Now, set the project in which you want to create the component:
```shell
# replace <project-name> with a valid value from the list
odo project set <project-name>
```

odo supports interactive and non-interactive ways of creating a component.

## Creating a component interactively

To interactively create a component, `cd` into the cloned petclinic project (already done if you copy-pasted the command above) and execute:
```shell
odo create
```
You will be prompted with a few questions one after the another. Let's go through each one of them to create a component.

1. First question is about selecting the component type:
    ```shell
    $ odo create
    ? Which devfile component type do you wish to create  [Use arrows to move, enter to select, type to filter]
    > java-maven
    java-maven
    java-openliberty
    java-openliberty
    java-quarkus
    java-quarkus
    java-springboot
    ```
   By default, `java-maven` is selected for us. Since this is a Spring Boot application, we should be selecting `java-springboot`. 

    We can either scroll down to `java-springboot` using the arrow key, or start typing `spring` on the prompt. Typing `spring` will lead to odo filtering the component type based on your input.

2. Next, odo asks you to name the component:
    ```shell
    $ odo create                
    ? Which devfile component type do you wish to create java-springboot
    ? What do you wish to name the new devfile component (java-springboot) petclinic
    ```
    Let's name it `petclinic`.

3. Next, odo asks you for the project in which you would like to create the component. Use the project `myproject` that we created earlier or the one you had set using `odo project set` command
   ```shell
   $ odo create
   ? Which devfile component type do you wish to create java-springboot
   ? What do you wish to name the new devfile component java-springboot
   ? What project do you want the devfile component to be created in myproject
   ```
   Now you will have a `devfile.yaml` in your current working directory. But odo is just not done asking you questions yet.
4. Lastly, odo asks you if you would like to download a "starter project". Since we already cloned the petclinic project, we answer in No by typing `n` and hitting the return key. We discuss starter projects later in [this document](#starter-projects):
   ```shell
   $ odo create
   ? Which devfile component type do you wish to create java-springboot
   ? What do you wish to name the new devfile component java-springboot
   ? What project do you want the devfile component to be created in myproject
   Devfile Object Validation
   ✓  Checking devfile existence [66186ns]
   ✓  Creating a devfile component from registry: stage [92202ns]
   Validation
   ✓  Validating if devfile name is correct [99609ns]
   ? Do you want to download a starter project (y/N) n
   ```
   
Your Spring Boot component is now ready for use.

## Creating a component non-interactively

To non-interactively create the same component as we created in interactive mode, `cd` into the cloned petclinic project and execute:
```shell
odo create java-springboot petclinic -n myproject
```
Here `java-springboot` is the type of the component, `petclinic` is the name of the component, and `-n myproject` tells odo to use the project `myproject` for the mentioned `odo create` operation.

## Starter projects

Besides creating a component for an existing code, you could also use "starter project" when creating a component.

Starter projects are example projects developed by the community to showcase the usability of devfiles. An odo user can use these starter projects by running `odo create` command in an empty directory.

### Starer projects in interactive mode

To interactively create a Java Spring Boot component using the starter project, you can follow the below steps:
```shell
mkdir myOdoComponent && cd myOdoComponent
odo create
```
In the questions that odo asks you next, provide answers like below:
```shell
$ odo create
? Which devfile component type do you wish to create java-springboot
? What do you wish to name the new devfile component myFirstComponent
? What project do you want the devfile component to be created in myproject
Devfile Object Validation
 ✓  Checking devfile existence [60122ns]
 ✓  Creating a devfile component from registry: stage [91411ns]
Validation
 ✓  Validating if devfile name is correct [35749ns]
? Do you want to download a starter project Yes

Starter Project
 ✓  Downloading starter project springbootproject from https://github.com/odo-devfiles/springboot-ex.git [716ms]

Please use `odo push` command to create the component with source deployed
```

### Starter projects in non-interactive mode

To non-interactively create a Java Spring Boot component using the starter project, you can follow the below steps:
```shell
mkdir myOdoComponent && cd myOdoComponent
odo create java-springboot myFirstComponent --starter
```

## Push the component to Kubernetes

odo follows a "create & push" workflow for almost all the commands. Meaning, most odo commands won't create resources on Kubernetes cluster unless you run `odo push` command.

Among the various ways described above, irrespective of how you created the component, the next step to create the resources for our component on the cluster would be to run `odo push`.

Note that first run of `odo push` could take a long time for the Spring Boot petclinic project. This is because of the dependencies being pulled by maven for Spring Boot framework. Subsequent `odo push` runs shouldn't take longer when working on the same Kubernetes cluster. 