---
# Page settings
layout: homepage
keywords:

# Hero section
title: odo 
description: "odo is a fast, iterative, and straightforward CLI tool for developers who write, build, and deploy applications on OpenShift.<br><br>Existing tools such as oc are more operations-focused and require a deep-understanding of Kubernetes and OpenShift concepts. odo abstracts away complex Kubernetes and OpenShift concepts for the developer."
project: 
asciinema: '<script id="asciicast-uIcSZvdbrFKKeH2sqrLsFdXym" src="https://asciinema.org/a/uIcSZvdbrFKKeH2sqrLsFdXym.js" async></script>'

buttons:
    - content: Installing odo
      url: '/docs/installing-odo'
      external_url: false
    - icon: github
      content: GitHub
      url: 'https://github.com/openshift/odo'
      external_url: true

# Author box
author:
    title: Official odo Devfile Registry
    title_url: 'https://github.com/odo-devfiles/registry'
    external_url: true
    description: "odo has examples for multiple languages and frameworks. However, with Devfile, you can take any language or framework and deploy it."
    languages:
      - title: Java (Maven)
        url: https://github.com/odo-devfiles/registry/blob/master/devfiles/java-maven/devfile.yaml
      - title: Java (Springboot)
        url: https://github.com/odo-devfiles/registry/blob/master/devfiles/java-springboot/devfile.yaml
      - title: Node.JS
        url: https://github.com/odo-devfiles/registry/blob/master/devfiles/nodejs/devfile.yaml

# Micro navigation
micro_nav: true

# Grid navigation
grid_navigation:
    - title: Installing odo
      excerpt: Installing odo on macOS, Linux and Windows
      cta: Read more
      url: 'docs/installing-odo'

    - title: Understanding odo
      excerpt: Understanding the concepts of odo
      cta: Read more
      url: 'docs/understanding-odo'

    - title: Deploying a devfile using odo
      excerpt: Deploying a portable devfile that decribes your development environment
      cta: Read more
      url: 'docs/deploying-a-devfile-using-odo'
      openshift: true
      kubernetes: true

    - title: Devfile file reference
      excerpt: Devfile reference documentation
      cta: Read more
      url: 'file-reference/'
      openshift: true
      kubernetes: true

    - title: Debugging applications in odo
      excerpt: Learn how to debug an application in odo
      cta: Read more
      url: 'docs/debugging-applications-in-odo'
      openshift: true
      kubernetes: true

    - title: Managing environment variables
      excerpt: Manipulate both config and preference files to your liking
      cta: Read more
      url: 'docs/managing-environment-variables-in-odo'
      openshift: true
      kubernetes: true

    - title: Configuring the odo CLI
      excerpt: Configure your terminal for autocompletion
      cta: Read more
      url: 'docs/configuring-the-odo-cli'

    - title: Architecture of odo
      excerpt: A general overview of the odo architecture
      cta: Read more
      url: 'docs/odo-architecture'

    - title: odo CLI reference
      excerpt: An overview of all the CLI commands related to odo
      cta: Read more
      url: 'docs/odo-cli-reference'

    - title: Introduction to Operators
      excerpt: Deploying an Operator from Operator Hub using odo.
      cta: Read more
      url: 'docs/operator-hub'
      openshift: true
      kubernetes: true
---
