---
title: odo registry
---

The `odo registry` command lists all the Devfile stacks from Devfile registries.

The Devfile registries that are taken into account are the registries added with the command
`odo preference add registry`.

## Available Flags

By default, `odo registry` lists all the Devfile stacks from all the Devfile registries.

These flags let you filter the listed Devfile stacks:

* `--devfile <name>` to list the Devfile stacks with this exact name
* `--devfile-registry <name>` to list the Devfile stack of this registry (this is the `name` used
when adding the registry to the preferences with `odo preference add registry <name> <url>`)
* `--filter <term>` to list the Devfile for which the term is found in the devfile name or description

By default, the name, registry and description 
of the Devfile stacks are displayed on a table.

This flag lets you change the content of the output:

* `--details` to display details about the Devfile stacks
* `-o json` to output the information in a JSON format

## Running the command

For these examples, we consider we have two registries in our preferences:

```console
$ odo preference view
[...]

Devfile registries:
 NAME                       URL                                   SECURE
 Staging                    https://registry.stage.devfile.io     No
 DefaultDevfileRegistry     https://registry.devfile.io           No
 ```

To get the complete list of accessible Devfile stacks:

```console
odo registry
```
```console
$ odo registry
 NAME                          REGISTRY                DESCRIPTION                                 
 dotnet50                      Staging                 Stack with .NET 5.0                         
 dotnet50                      DefaultDevfileRegistry  Stack with .NET 5.0                         
 dotnet60                      Staging                 Stack with .NET 6.0                         
 dotnet60                      DefaultDevfileRegistry  Stack with .NET 6.0                         
 dotnetcore31                  Staging                 Stack with .NET Core 3.1                    
 dotnetcore31                  DefaultDevfileRegistry  Stack with .NET Core 3.1                    
 go                            Staging                 Stack with the latest Go version            
 go                            DefaultDevfileRegistry  Stack with the latest Go version            
 java-maven                    Staging                 Upstream Maven and OpenJDK 11               
 java-maven                    DefaultDevfileRegistry  Upstream Maven and OpenJDK 11               
[...]
```

To list the Devfile stacks from the Staging registry only:

```console
odo registry --devfile-registry Staging
```
```console
$ odo registry --devfile-registry Staging
 NAME                          REGISTRY                DESCRIPTION                                 
 dotnet50                      Staging                 Stack with .NET 5.0                         
 dotnet60                      Staging                 Stack with .NET 6.0                         
 dotnetcore31                  Staging                 Stack with .NET Core 3.1                    
 go                            Staging                 Stack with the latest Go version            
 java-maven                    Staging                 Upstream Maven and OpenJDK 11               
[...]
```

To list the Devfile stacks related to Maven:

```console
odo registry --filter Maven
```
```console
$ odo registry --filter Maven
 NAME                       REGISTRY                DESCRIPTION                                 
 java-maven                 Staging                 Upstream Maven and OpenJDK 11               
 java-maven                 DefaultDevfileRegistry  Upstream Maven and OpenJDK 11               
 java-openliberty           Staging                 Java application Maven-built stack using... 
 java-openliberty           DefaultDevfileRegistry  Java application Maven-built stack using... 
 java-websphereliberty      Staging                 Java application Maven-built stack using... 
 java-websphereliberty      DefaultDevfileRegistry  Java application Maven-built stack using... 
 java-wildfly-bootable-jar  Staging                 Java stack with WildFly in bootable Jar ... 
 java-wildfly-bootable-jar  DefaultDevfileRegistry  Java stack with WildFly in bootable Jar ... 
```

To get the details of the `java-maven` Devfile in the Staging registry:

```console
odo registry --devfile java-maven --devfile-registry Staging --details
```
```console
$ odo registry --devfile java-maven --devfile-registry Staging --details
Name: java-maven
Display Name: Maven Java
Registry: Staging
Registry URL: https://registry.stage.devfile.io
Version: 1.1.0
Description: Upstream Maven and OpenJDK 11 
Tags: Java, Maven
Project Type: maven
Language: java
Starter Projects:
  - springbootproject
Supported odo Features:
  - Dev: Y
  - Deploy: N
  - Debug: Y
```
