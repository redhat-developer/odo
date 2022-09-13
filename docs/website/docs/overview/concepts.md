---
title: Concepts
sidebar_position: 3
---

# Concepts of odo

`odo` abstracts Kubernetes concepts into a developer friendly terminology; in this document, we will take a look at the following terminologies:

## Application

An application in `odo` is a classic application developed with a [cloud-native approach](https://www.redhat.com/en/topics/cloud-native-apps) that is used to perform a particular task.

Examples of applications: Online Video Streaming, Hotel Reservation System, Online Shopping.

## Component

In the cloud-native architecture, an application is a collection of small, independent, and loosely coupled components; a `odo` component is one of these components.
odo uses Devfile  as a definition of a component.

Examples of components: API Backend, Web Frontend, Payment Backend.

## Project

A project helps achieve multi-tenancy: several applications can be run in the same cluster by different teams in different projects.
In Kubernetes Project is represented by **namespace**.

## Context

Context is the directory on the system that contains the source code, tests, libraries and `odo` specific config files for a single component.

## Devfile

Devfile is a portable YAML file containing the definition of a component and its parts(like endpoints, storages and services) Visit [devfile.io](https://devfile.io/) for more information on devfiles.
