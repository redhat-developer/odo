---
title: Quickstart Guide
---

# Quickstart Guide

In this guide, we will be using `odo` to create a "Hello World" application, and then start a container-based development session using `odo`.

You have the option of developing and iterating locally against local [Podman](https://podman.io/) containers or any Kubernetes or OpenShift cluster.

A recommended way to get started with `odo` is to iterate on the application locally with Podman, as it does not require any additional clusters to be setup or available.
Later on, you can seamlessly run and iterate on the same application against a Kubernetes or OpenShift cluster.

This quickstart guide will show you how easy it can be get started with `odo`. You have the option of choosing from the following frameworks:
* [Node.js](nodejs)
* [.NET](dotnet)
* [Java (Spring Boot)](java)
* [Go](go)

A full list of example applications can be viewed with the `odo registry` command.
<details>
<summary>Example</summary>

```shell
$ odo registry                                                     
 NAME                          REGISTRY                DESCRIPTION                                  ARCHITECTURES          VERSIONS                   
 dotnet50                      DefaultDevfileRegistry  .NET 5.0 application                                                1.0.3                      
 dotnet60                      DefaultDevfileRegistry  .NET 6.0 application                                                1.0.2                      
 dotnetcore31                  DefaultDevfileRegistry  .NET Core 3.1 application                                           1.0.3                      
 go                            DefaultDevfileRegistry  Go is an open source programming languag...                         1.0.2, 1.1.0, 2.0.0, 2.1.0 
 java-maven                    DefaultDevfileRegistry  Java application based on Maven 3.6 and ...                         1.2.0                      
 java-openliberty              DefaultDevfileRegistry  Java application based on Java 11 and Ma...  amd64, ppc64le, s390x  0.9.0                      
 java-openliberty-gradle       DefaultDevfileRegistry  Java application based on Java 11, Gradl...  amd64, ppc64le, s390x  0.4.0                      
 java-quarkus                  DefaultDevfileRegistry  Java application using Quarkus and OpenJ...                         1.3.0                      
 java-springboot               DefaultDevfileRegistry  Spring Boot using Java                                              1.2.0, 2.0.0               
 java-vertx                    DefaultDevfileRegistry  Java application using Vert.x and OpenJD...                         1.2.0                      
 java-websphereliberty         DefaultDevfileRegistry  Java application based Java 11 and Maven...  amd64, ppc64le, s390x  0.9.0                      
 java-websphereliberty-gradle  DefaultDevfileRegistry  Java application based on Java 11 and Gr...  amd64, ppc64le, s390x  0.4.0                      
 java-wildfly                  DefaultDevfileRegistry  Java application based on Java 11, using...                         1.1.0                      
 java-wildfly-bootable-jar     DefaultDevfileRegistry  Java application using WildFly in bootab...                         1.1.0                      
 nodejs                        DefaultDevfileRegistry  Node.js application                                                 2.1.1, 2.2.0               
 nodejs-angular                DefaultDevfileRegistry  Angular is a development platform, built...                         2.0.2, 2.1.0, 2.2.0        
 nodejs-nextjs                 DefaultDevfileRegistry  Next.js gives you the best developer exp...                         1.0.3, 1.1.0, 1.2.0        
 nodejs-nuxtjs                 DefaultDevfileRegistry  Nuxt is the backbone of your Vue.js proj...                         1.0.3, 1.1.0, 1.2.0        
 nodejs-react                  DefaultDevfileRegistry  React is a free and open-source front-en...                         2.0.2, 2.1.0, 2.2.0        
 nodejs-svelte                 DefaultDevfileRegistry  Svelte is a radical new approach to buil...                         1.0.3, 1.1.0, 1.2.0        
 nodejs-vue                    DefaultDevfileRegistry  Vue is a JavaScript framework for buildi...                         1.0.2, 1.1.0, 1.2.0        
 php-laravel                   DefaultDevfileRegistry  Laravel is an open-source PHP framework,...                         1.0.1, 2.0.0               
 python                        DefaultDevfileRegistry  Python is an interpreted, object-oriente...                         2.1.0, 3.0.0               
 python-django                 DefaultDevfileRegistry  Django is a high-level Python web framew...                         2.1.0                      
 udi                           DefaultDevfileRegistry  Universal Developer Image provides vario...                         1.0.0

```
</details>
