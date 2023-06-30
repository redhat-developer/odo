---
title: odo describe component
---

`odo describe component` command is useful for getting information about a component. 

## Running the command
There are 2 ways to describe a component:
- [Describe with access to Devfile](#describe-with-access-to-devfile)
- [Describe without access to Devfile](#describe-without-access-to-devfile)

### Describe with access to Devfile
```shell
odo describe component
```
<details>
<summary>Example</summary>

```shell
$ odo describe component
Name: my-nodejs
Display Name: Node.js Runtime
Project Type: nodejs
Language: javascript
Version: 1.0.1
Description: Stack with Node.js 14
Tags: NodeJS, Express, ubi8

Running in: Deploy

Supported odo features:
•  Dev: true
•  Deploy: true
•  Debug: true

Commands:
 •  my-install
      Type: exec
      Group: build
      Command Line: "npm install"
      Component: runtime
      Component Type: container
 •  my-run
      Type: exec
      Group: run
      Command Line: "npm start"
      Component: runtime
      Component Type: container
 •  build-image
      Type: apply
      Component: prod-image
      Component Type: image
      Image Name: devfile-nodejs-deploy:latest
 •  deploy-deployment
      Type: apply
      Component: outerloop-deploy
      Component Type: kubernetes
 •  deploy
      Type: composite
      Group: deploy

Container components:
•  runtime

Kubernetes components:
 •  outerloop-deployment
 •  outerloop-service
 •  outerloop-url-ingress
 •  outerloop-url-route

Kubernetes Ingresses:
 •  my-nodejs-app: nodejs.example.com/
 •  my-nodejs-app: nodejs.example.com/foo

Kubernetes Routes:
 •  my-nodejs-app: my-nodejs-app-phmartin-crt-dev.apps.sandbox-m2.ll9k.p1.openshiftapps.com/testpath

```
</details>

This command returns information extracted from the Devfile:
- metadata (name, display name, project type, language, version, description and tags)
- supported odo features, indicating if the Devfile defines necessary information to run `odo dev`, `odo dev --debug` and `odo deploy`
- the list of commands, if any, along with some useful information about each command
- the list of container components,
- the list of Kubernetes components.
- the list of forwarded ports if the component is running in Dev mode.

The command also displays if the component is currently running in the cluster or in Podman on Dev and/or Deploy mode.

### Describe without access to Devfile

```shell
odo describe component --name <component_name> [--namespace <namespace>]
```
<details>
<summary>Example</summary>

```shell
$ odo describe component --name my-nodejs
Name: my-nodejs
Display Name: Unknown
Project Type: nodejs
Language: Unknown
Version: Unknown
Description: Unknown
Tags: 

Running in: Deploy

Supported odo features:
 •  Dev: Unknown
 •  Deploy: Unknown
 •  Debug: Unknown

Kubernetes Ingresses:
 •  my-nodejs-app: nodejs.example.com/
 •  my-nodejs-app: nodejs.example.com/foo

Kubernetes Routes:
 •  my-nodejs-app: my-nodejs-app-phmartin-crt-dev.apps.sandbox-m2.ll9k.p1.openshiftapps.com/testpath

```
</details>

The command extracts information from the labels and annotations attached to the deployed component to display the known metadata of the Devfile used to deploy the component.

The command also displays if the component is currently running in the cluster or in Podman on Dev and/or Deploy mode.

### Targeting a specific platform

By default, `odo describe component` will search components in both the current namespace of the cluster and podman. You can restrict the search to one of the platforms only, using the `--platform` flag, giving a value `cluster` or `podman`.
