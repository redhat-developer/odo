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

By default, the name, registry, description and versions of the Devfile stacks are displayed on a table.

The flags below let you change the content of the output:

* `--details` to display details about the Devfile stacks
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
 NAME                          REGISTRY                DESCRIPTION                                  VERSIONS
 dotnet50                      Staging                 Stack with .NET 5.0                          1.0.3
 dotnet50                      DefaultDevfileRegistry  Stack with .NET 5.0                          1.0.3
 dotnet60                      Staging                 Stack with .NET 6.0                          1.0.2
 dotnet60                      DefaultDevfileRegistry  Stack with .NET 6.0                          1.0.2
 dotnetcore31                  Staging                 Stack with .NET Core 3.1                     1.0.3
 dotnetcore31                  DefaultDevfileRegistry  Stack with .NET Core 3.1                     1.0.3
 go                            Staging                 Go is an open source programming languag...  1.0.2, 2.0.0
 go                            DefaultDevfileRegistry  Go is an open source programming languag...  1.0.2, 2.0.0
 java-maven                    Staging                 Upstream Maven and OpenJDK 11                1.2.0
 java-maven                    DefaultDevfileRegistry  Upstream Maven and OpenJDK 11                1.2.0
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
 NAME                          REGISTRY  DESCRIPTION                                  VERSIONS
 dotnet50                      Staging   Stack with .NET 5.0                          1.0.3
 dotnet60                      Staging   Stack with .NET 6.0                          1.0.2
 dotnetcore31                  Staging   Stack with .NET Core 3.1                     1.0.3
 go                            Staging   Go is an open source programming languag...  1.0.2, 2.0.0
 java-maven                    Staging   Upstream Maven and OpenJDK 11                1.2.0
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
$ odo registry --devfile java-maven --devfile-registry Staging --details
Name: java-maven
Display Name: Maven Java
Registry: Staging
Registry URL: https://registry.stage.devfile.io
Version: 1.2.0
Description: Upstream Maven and OpenJDK 11 
Tags: Java, Maven
Project Type: Maven
Language: Java
Starter Projects:
  - springbootproject
Supported odo Features:
  - Dev: Y
  - Deploy: N
  - Debug: Y
Versions:
  - 1.2.0

```
</details>

