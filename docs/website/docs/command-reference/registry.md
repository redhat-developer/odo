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
* `--filter <term>` to list the Devfile for which the term is found in the devfile name,  description or supported architectures

By default, the name, registry, description, supported architectures, and versions of the Devfile stacks are displayed on a table.

Note that Devfile stacks with no architectures are supposed to be compatible with **all** architectures.

The flags below let you change the content of the output:

* `--details` to display details about a specific Devfile stack (to be used only with `--devfile <name>`)
* `-o json` to output the information in a JSON format

## Running the command

Let us consider we have two registries in our preferences:

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
<details>
<summary>Example</summary>

```console
$ odo registry                                                                                                                                                                                 
 NAME                          REGISTRY                DESCRIPTION                                  ARCHITECTURES          VERSIONS                                                            
 dotnet50                      Staging                 .NET 5.0 application                                                1.0.3                                                               
 dotnet50                      DefaultDevfileRegistry  .NET 5.0 application                                                1.0.3                                                               
 dotnet60                      Staging                 .NET 6.0 application                                                1.0.2                                                               
 dotnet60                      DefaultDevfileRegistry  .NET 6.0 application                                                1.0.2                                                               
 dotnetcore31                  Staging                 .NET Core 3.1 application                                           1.0.3                                                               
 dotnetcore31                  DefaultDevfileRegistry  .NET Core 3.1 application                                           1.0.3                                                               
 go                            Staging                 Go is an open source programming languag...                         1.0.2, 1.1.0, 2.0.0, 2.1.0                                          
 go                            DefaultDevfileRegistry  Go is an open source programming languag...                         1.0.2, 1.1.0, 2.0.0, 2.1.0                                          
 java-maven                    Staging                 Java application based on Maven 3.6 and ...                         1.2.0                                                               
 java-maven                    DefaultDevfileRegistry  Java application based on Maven 3.6 and ...                         1.2.0                                                               
 java-openliberty              Staging                 Java application based on Java 11 and Ma...  amd64, ppc64le, s390x  0.9.0                                                               
 java-openliberty              DefaultDevfileRegistry  Java application based on Java 11 and Ma...  amd64, ppc64le, s390x  0.9.0 
 [...]
```
</details>


To list the Devfile stacks from a specific registry only:

```console
odo registry --devfile-registry <registry>
```

<details>
<summary>Example</summary>

```console
$ odo registry --devfile-registry Staging
 NAME                          REGISTRY  DESCRIPTION                                  ARCHITECTURES          VERSIONS                   
 dotnet50                      Staging   .NET 5.0 application                                                1.0.3                      
 dotnet60                      Staging   .NET 6.0 application                                                1.0.2                      
 dotnetcore31                  Staging   .NET Core 3.1 application                                           1.0.3                      
 go                            Staging   Go is an open source programming languag...                         1.0.2, 1.1.0, 2.0.0, 2.1.0 
 java-maven                    Staging   Java application based on Maven 3.6 and ...                         1.2.0                      
 java-openliberty              Staging   Java application based on Java 11 and Ma...  amd64, ppc64le, s390x  0.9.0
 [...]
```
</details>


To list the Devfile stacks related to a specific keyword:

```console
odo registry --filter <keyword>
```
<details>
<summary>Example</summary>

```console
$ odo registry --filter Maven
 NAME                       REGISTRY                DESCRIPTION                                  VERSIONS
 java-maven                 Staging                 Upstream Maven and OpenJDK 11                1.2.0
 java-maven                 DefaultDevfileRegistry  Upstream Maven and OpenJDK 11                1.2.0
 java-openliberty           Staging                 Java application Maven-built stack using...  0.9.0
 java-openliberty           DefaultDevfileRegistry  Java application Maven-built stack using...  0.9.0
 java-websphereliberty      Staging                 Java application Maven-built stack using...  0.9.0
 java-websphereliberty      DefaultDevfileRegistry  Java application Maven-built stack using...  0.9.0
 java-wildfly-bootable-jar  Staging                 Java stack with WildFly in bootable Jar ...  1.1.0
 java-wildfly-bootable-jar  DefaultDevfileRegistry  Java stack with WildFly in bootable Jar ...  1.1.0
```
</details>


To get the details of a specific Devfile from a specific registry:

```console
odo registry --devfile <devfile> --devfile-registry <registry> --details
```
<details>
<summary>Example</summary>

```console
$ odo registry list --devfile-registry Staging --devfile java-openliberty --details
Name: java-openliberty
Display Name: Open Liberty Maven
Registry: Staging
Registry URL: https://registry.stage.devfile.io
Version: 0.9.0
Description: Java application based on Java 11 and Maven 3.8, using the Open Liberty runtime 22.0.0.1 
Tags: Java, Maven
Project Type: Open Liberty
Language: Java
Starter Projects:
  - rest
Supported odo Features:
  - Dev: Y
  - Deploy: N
  - Debug: Y
Architectures:
  - amd64
  - ppc64le
  - s390x
Versions:
  - 0.9.0

```
</details>

