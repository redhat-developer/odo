---
sidebar_position: 1
title: Introduction
---

### What is odo?

odo is a fast, iterative and straightforward CLI tool for developers who write, build, and deploy applications on Kubernetes.

odo abstracts the complex Kubernetes terminology so that an application developer can focus on writing code in their favourite framework without having to learn Kubernetes.

odo is focussed on inner loop development with some tooling that would help users transition to the outer loop.

Brendan Burns, the cofounder of Kubernetes, said in the [book Kubernetes Patterns](https://www.redhat.com/cms/managed-files/cm-oreilly-kubernetes-patterns-ebook-f19824-201910-en.pdf):

> It (Kubernetes) is the foundation on which applications will be built, and it provides a large library of APIs and tools for building these applications, but it does little to provide the application architect or developer with any hints or guidance for how these various pieces can be combined into a complete, reliable system that satisfies their business needs and goals.

It is the application architects and developers that odo aims to make Kubernetes easy for.

### What is "inner loop"?

In simplest terms, inner loop is the phase of application development before the developer does `git commit`. odo workflow helps developers get feedback about how their application when deployed to a Kubernetes cluster.   

### Who should use odo?

You should use odo if:
* you are developing applications using Node.js, Spring Boot, or similar framework
* your Ops team deploys your applications to Kubernetes
* you don't want to spend time learning about Kubernetes, but want to develop applications using your favourite framework

Basically, if you are an application developer, you should use odo to deploy your application on a Kubernetes cluster.

### How is odo different from `kubectl` and `oc`?

Both [`kubectl`](https://github.com/kubernetes/kubectl) and [`oc`](https://github.com/openshift/oc/) require deep understanding of Kubernetes concepts.

odo is different from these tools in that it is focussed on application developers and architects. Both `kubectl` and `oc` are Ops oriented tools and help in deploying applications to and maintaining a Kubernetes cluster provided you know Kubernetes well.

You should not use odo:
* to maintain a production Kubernetes cluster
* to perform administration tasks against a Kubernetes cluster