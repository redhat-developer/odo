---
title: Quickstart Guide
---

# Quickstart Guide

In this guide, we will be using `odo` to create a "Hello World" application.

You have the option of choosing from the following frameworks for the quickstart guide:
* [Node.js](nodejs)
* [.NET](dotnet)
* [Java (Spring Boot)](java)
* [Go](go)

A full list of example applications can be viewed with the `odo registry` command.
<details>
<summary>Example</summary>

```shell
$ odo registry
 NAME                          REGISTRY                DESCRIPTION                                  VERSIONS
 dotnet50                      DefaultDevfileRegistry  Stack with .NET 5.0                          1.0.3
 dotnet60                      DefaultDevfileRegistry  Stack with .NET 6.0                          1.0.2
 dotnetcore31                  DefaultDevfileRegistry  Stack with .NET Core 3.1                     1.0.3
 go                            DefaultDevfileRegistry  Go is an open source programming languag...  1.0.2, 2.0.0
 java-maven                    DefaultDevfileRegistry  Upstream Maven and OpenJDK 11                1.2.0
 java-openliberty              DefaultDevfileRegistry  Java application Maven-built stack using...  0.9.0
 java-openliberty-gradle       DefaultDevfileRegistry  Java application Gradle-built stack usin...  0.4.0
 java-quarkus                  DefaultDevfileRegistry  Quarkus with Java                            1.3.0
 java-springboot               DefaultDevfileRegistry  Spring Boot using Java                       1.2.0, 2.0.0
 java-vertx                    DefaultDevfileRegistry  Upstream Vert.x using Java                   1.2.0
 java-websphereliberty         DefaultDevfileRegistry  Java application Maven-built stack using...  0.9.0
 java-websphereliberty-gradle  DefaultDevfileRegistry  Java application Gradle-built stack usin...  0.4.0
 java-wildfly                  DefaultDevfileRegistry  Upstream WildFly                             1.1.0
 java-wildfly-bootable-jar     DefaultDevfileRegistry  Java stack with WildFly in bootable Jar ...  1.1.0
 nodejs                        DefaultDevfileRegistry  Stack with Node.js 16                        2.1.1
 nodejs-angular                DefaultDevfileRegistry  Angular is a development platform, built...  2.0.2
 nodejs-nextjs                 DefaultDevfileRegistry  Next.js gives you the best developer exp...  1.0.3
 nodejs-nuxtjs                 DefaultDevfileRegistry  Nuxt is the backbone of your Vue.js proj...  1.0.3
 nodejs-react                  DefaultDevfileRegistry  React is a free and open-source front-en...  2.0.2
 nodejs-svelte                 DefaultDevfileRegistry  Svelte is a radical new approach to buil...  1.0.3
 nodejs-vue                    DefaultDevfileRegistry  Vue is a JavaScript framework for buildi...  1.0.2
 php-laravel                   DefaultDevfileRegistry  Laravel is an open-source PHP framework,...  1.0.1
 python                        DefaultDevfileRegistry  Python is an interpreted, object-oriente...  2.1.0, 3.0.0
 python-django                 DefaultDevfileRegistry  Django is a high-level Python web framew...  2.1.0

```
</details>
