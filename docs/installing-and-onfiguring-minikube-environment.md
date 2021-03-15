---
# Page settings
layout: default
keywords:
comments: false

# Hero section
title: Setup the minikube environment
description: Setup a Kubernetes cluster that odo can be used with

# Micro navigation
micro_nav: true
---
> **Note**
> 
> The guide only works with minikube configured with kubernetes 1.19.x or lower. `odo link` cannot link services successfully in a Kubernetes 1.20.x or newer environment.

It is recommended that users of this guide obtain a suitable system for running minikube with kubernetes. In practice this should be a 4 core system minimum. Before proceeding to the sample application, please follow the instructions for establishing a minikube environment:

  - You have installed docker. See [the Docker documentation](https://docs.docker.com/engine/install/).

  - You have installed minikube. See [the minikube installation instructions](https://minikube.sigs.k8s.io/docs/start/) for your operating system.

# Starting minikube

  - If you try to run minikube as a root user, minikube throws an error that docker should not be run as root and aborts the startup. To proceed, start minikube in a manner that overrides this protection mechanism:
    
    ``` sh
    $ minikube start --force --driver=docker --kubernetes-version=v1.19.8
    ```

  - If you are a non-root user, start minikube as usual:
    
    ``` sh
    $ minikube start --kubernetes-version=v1.19.8
    ```

# Configuring minikube

## Enabling ingress addon

The application requires an ingress addon to allow the routes to be created easily. It enables `odo url` commands.

1.  onfigure minikube for ingress by adding ingress as a minikube add-on:
    
    ``` sh
    $ minikube addons enable ingress
    ```

## Adding a pull secret to ingress accounts

You may face the DockerHub pull rate limit if you do not have a pull secret for your personal free DockerHub account. During ingress initialization two of the job pods used by ingress may fail to initialize due to pull rate limits. If this happens, and ingress fails to enable, you add a secret for the pulls for the following service accounts:

  - ingress-nginx-admission

  - ingress-nginx

to add a pull secret for these service accounts:

1.  Switch to the kube-system context:
    
    ``` sh
    $ kubectl config set-context --current --namespace=kube-system
    ```

2.  Create a pull secret:
    
    ``` sh
    $ kubectl create secret docker-registry regcred --docker-server=<your-registry-server> --docker-username=<your-name> --docker-password=<your-pword> --docker-email=<your-email>
    ```
    
    where:
    
      - \<your-registry-server\> is the DockerHub Registry FQDN. (<https://index.docker.io/v1/>)
    
      - \<your-name\> is your DockerHub account username.
    
      - \<your-pword\> is your DockerHub account password.
    
      - \<your-email\> is your DockerHub account email.

3.  Add this new secret (`regcred` in the example above) to the default service account in minikube:
    
    ``` sh
    $ kubectl patch serviceaccount ingress-nginx-admission -p '{"imagePullSecrets": [{"name": "regcred"}]}'
    $ kubectl patch serviceaccount ingress-nginx -p '{"imagePullSecrets": [{"name": "regcred"}]}'
    ```

## Patching the default service account

The default service account needs to be patched with a pull secret configured for your personal docker account.

  - Switch to the default context:
    
    ``` sh
    $ kubectl config set-context --current --namespace=default
    ```

  - Create the same docker-registry secret configured for your docker, now for the default minikube context:
    
    ``` sh
    $ kubectl create secret docker-registry regcred --docker-server=<your-registry-server> --docker-username=<your-name> --docker-password=<your-pword> --docker-email=<your-email>
    ```
    
    where:
    
      - \<your-registry-server\> is the DockerHub Registry FQDN. (<https://index.docker.io/v1/>)
    
      - \<your-name\> is your Docker username.
    
      - \<your-pword\> is your Docker password.
    
      - \<your-email\> is your Docker email.

  - Add this cred ('regcred' in the example above) to the default service account in minikube:
    
    ``` sh
    $ kubectl patch serviceaccount default -p '{"imagePullSecrets": [{"name": "regcred"}]}'
    ```

## Enabling an Operator Lifecycle Manager (OLM) addon

Enabling OLM on your minikube instance simplifies installation and upgrades of Operators available from [OperatorHub](https://operatorhub.io). It also enables `odo service` commands to work for Operators.

  - Enable OLM with the following command:
    
    ``` sh
    $ minikube addons enable olm
    ```
