---
title: Quickstart
sidebar_position: 1
---
This quickstart shows you how to build and deploy a real world application in a cloud-native environment using odo; it uses a Spring Boot application built using Maven for example.

### Objectives

* Deploy the spring boot application to a Kubernetes cluster.
* Be able to access the application outside the cluster.
* Be able to access application environment outside the cluster. 
* Add extra storage to the application.
* Link the spring boot application with another microservice.

### Pre-requisites
Before moving further, make sure you are done with the pre-requisite.
* Setup a Kubernetes/OpenShift cluster.
* Install odo on your system.
* Login to your cluster.

### Deploying the application
1. Clone the git repository and cd.
  ```shell
  $ git clone https://github.com/spring-projects/spring-petclinic.git; cd spring-petclinic
  ```
2. Build the component.
  ```shell
  $ odo create java-springboot petclinic
  ```
3. Deploy the component. This step might take a few minutes depending on your internet connection.
  ```shell
  $ odo push --show-logs
  ```
4. Check the component description.
  ```shell
  $ odo describe
  ```
  Output:
  ```yaml
  Component Name: petclinic
  Type: java-springboot
  Environment Variables:
   · PROJECTS_ROOT=/projects
   · PROJECT_SOURCE=/projects
   · DEBUG_PORT=5858
  Storage:
   · m2 of size 3Gi mounted to /home/user/.m2
  ```
5. Create a URL to access this application outside the cluster.

  _OpenShift_ - If you are using an OpenShift cluster, you can skip this step, URL will be automatically created for you.

  _Kubernetes_ - If you are using a Kubernetes cluster, run 
   ```shell
    $ odo url create  --port 8080 --host <your-ingress-domain>
    ```

  _Minikube_ - If you are using a Minikube cluster, run 
  ```shell
  $ odo url create --port 8080 --host=$(minikube ip).nip.io
  ```

6. Deploy the changes.
  ```shell
  $ odo push --show-logs
  ```
7. Check the url list to obtain the URL.
  ```shell
  $ odo url list
  ```
  Output:
  ```shell
  Found the following URLs for component petclinic
  NAME         STATE      URL                             PORT     SECURE     KIND
  8080-tcp     Pushed     http://8080-tcp.example.com     8080     false      ingress
  ```
  Optionally, you can also run the `odo describe` command to obtain the URL.

8. Access the URL via web browser. You should be able to see the Petclinic application.

### Extending the application

1. Let's say we are running out of storage for our Petclinic application, and we need to add some extra storage; we can do so with odo.
  ```shell
  $ odo storage create extendedStorage --path=/opt/app-root/src/storage --size=1Gi
  ```
  Output:
  ```shell
   ✓  Added storage mystorage to petclinic
  ```
2. Deploy the changes.
  ```shell
  $ odo push --show-logs
  ```
3. Check the storages mounted or attached to our component.
  ```shell
  $ odo storage list
  ```
  Output:
  ```shell
  The component 'petclinic' has the following storage attached:
  NAME          SIZE      PATH                           STATE
  m2            3Gi       /home/user/.m2                 Pushed
  mystorage     100Mi     /opt/app-root/src/storage/     Pushed
  ```

// TODO: Add something about odo exec and operator and linking.
