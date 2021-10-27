---
sidebar_position: 1
title: Introduction
---

### What is odo?

odo is a fast, iterative and straightforward CLI tool for developers who write, build, and deploy applications on Kubernetes.

odo abstracts the complex Kubernetes terminology so that an application developer can focus on writing code in their favourite framework without having to learn Kubernetes.

odo is focused on [inner loop](./intro#what-is-inner-loop-and-outer-loop) development with some tooling that would help users transition to the [outer loop](./intro#what-is-inner-loop-and-outer-loop).

Brendan Burns, one of the co-founders of Kubernetes, said in the [book Kubernetes Patterns](https://www.redhat.com/cms/managed-files/cm-oreilly-kubernetes-patterns-ebook-f19824-201910-en.pdf):

> It (Kubernetes) is the foundation on which applications will be built, and it provides a large library of APIs and tools for building these applications, but it does little to provide the application architect or developer with any hints or guidance for how these various pieces can be combined into a complete, reliable system that satisfies their business needs and goals.

odo makes Kubernetes easy for application architects and developers.

### What is "inner loop" and "outer loop"?

The inner loop consists of local coding, building, running, and testing the application—all activities that you, as a developer, can control. The outer loop consists of the larger team processes that your code flows through on its way to the cluster: code reviews, integration tests, security and compliance, and so on. The inner loop could happen mostly on your laptop. The outer loop happens on shared servers and runs in containers, and is often automated with continuous integration/continuous delivery (CI/CD) pipelines. Usually, a code commit to source control is the transition point between the inner and outer loops.
*([Source](https://developers.redhat.com/blog/2020/06/16/enterprise-kubernetes-development-with-odo-the-cli-tool-for-developers#improving_the_developer_workflow))*

### Who should use odo?

You should use odo if:
* you are developing applications using Node.js, Spring Boot, or similar framework
* your applications are intended to run in Kubernetes and your Ops team will help deploy them
* you do not want to spend time learning about Kubernetes, and prefer to focus on develop applications using your favourite framework

Basically, if you are an application developer, you should use odo to run your application on a Kubernetes cluster.

### How is odo different from `kubectl` and `oc`?

Both [`kubectl`](https://github.com/kubernetes/kubectl) and [`oc`](https://github.com/openshift/oc/) require deep understanding of Kubernetes concepts.

odo is different from these tools in that it is focused on application developers and architects. Both `kubectl` and `oc` are Ops oriented tools and help in deploying applications to and maintaining a Kubernetes cluster provided you know Kubernetes well.

You should not use odo:
* to maintain a production Kubernetes cluster
* to perform administration tasks against a Kubernetes cluster