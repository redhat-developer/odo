---
title: Develop and Deploy
sidebar_position: 5
---

# Develop and Deploy

The two most important commands in `odo` are `odo dev` and `odo deploy`. 

In some situations, you'd want to use [`odo dev`](/docs/command-reference/dev) over [`odo deploy`](/docs/command-reference/deploy) and vice-versa. This document highlights when you should use either command.

## When should I use `odo dev`?

`odo dev` should be used in the initial development process of your application. 

For example, you should use `odo dev` when you are:
* Making changes constantly
* Want to preview any changes
* Testing initial Kubernetes support for your application
* Want to debug and run tests
* Run the application on a local development environment

## When should I use `odo deploy`?

`odo deploy` should be the deploy stage of development when you are ready for a "production ready" environment.

For example, you should use `odo deploy` when you are working with a production environment and:
* Ready for the application to be viewed publically
* Require building and pushing the container
* Needing custom Kubernetes YAML for your production environment