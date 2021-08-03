---
title: Quickstart
sidebar_position: 1
---
This quickstart shows you how to build and deploy a real world application in a cloud-native environment using odo; it uses a Spring Boot application built using Maven for example.

### Objectives

* Deploy the Spring Boot application to a Kubernetes cluster.
* Be able to access the application outside the cluster.
* Create an instance of the Postgres service.
* Link the Spring Boot application with Postgres service.

## Pre-requisite
* Setup a Kubernetes/OpenShift cluster. See guides on setting up a [Kubernetes](cluster-setup/kubernetes.md) or [Openshift](cluster-setup/openshift.md) cluster if you have not already done so.
* Install odo on your system. See the [Installation](installation.md) guide if you have not already installed odo.
* If it is a remote cluster, be logged in to your cluster. Login interactively to your remote cluster with the command below.
  ```shell
  odo login
  ```


## Creating and Deploying the application
1. Clone the git repository and cd into it.
  ```shell
  git clone https://github.com/spring-projects/spring-petclinic.git
  cd spring-petclinic
  ```

2. Create a separate project `petclinic` for our application.
  ```shell
  odo project create petclinic
  ```

3. Build the application.
  ```shell
  odo create java-springboot petclinic
  ```
  `java-springboot` is the type of the application recognized by odo, `petclinic` is the name of the application.

4. Deploy the application into the cluster. This step might take a few minutes depending on your internet connection.
  ```shell
  odo push --show-logs
  ```

5. Check the application description for more information about it.
  ```shell
  odo describe
  ```
  The output can look similar to:
  ```shell
  $ odo describe
  Component Name: petclinic
  Type: java-springboot
  Environment Variables:
   · PROJECTS_ROOT=/projects
   · PROJECT_SOURCE=/projects
   · DEBUG_PORT=5858
  Storage:
   · m2 of size 3Gi mounted to /home/user/.m2
  ```

6. Create an Ingress URL to access this application outside the cluster.

  _OpenShift_ - If you are using an OpenShift cluster, you can skip this step, URL will be automatically created for you.

  _Kubernetes_ - If you are using a Kubernetes cluster, run 
   ```shell
    odo url create  --port 8080 --host <your-ingress-domain>
    ```
  _Minikube_ - If you are using a Minikube cluster, run 
  ```shell
  odo url create --port 8080 --host=$(minikube ip).nip.io
  ```

7. Deploy the changes.
  ```shell
  odo push --show-logs
  ```

8. Check the url list to obtain the URL.
  ```shell
  odo url list
  ```
  The output can look similar to:
  ```shell
  $ odo url list
  Found the following URLs for component petclinic
  NAME         STATE      URL                             PORT     SECURE     KIND
  8080-tcp     Pushed     http://8080-tcp.example.com     8080     false      ingress
  ```
  Alternatively, you can also run the `odo describe` command to obtain the URL.

9. Access the URL via web browser. You should be able to see the Petclinic application.

## Extending the application: Connecting the application to a PostgreSQL service

### Pre-requisite
* If you are using a Kubernetes cluster, install the OLM addon into the cluster. See the guide installing OLM on [Kubernetes](cluster-setup/kubernetes.md) if you have not already installed it.
  
  If you are using an OpenShift cluster, there is no need.

* Install the Service Binding Operator into the cluster. See the guide on installing the Service Binding Operator in a [Kubernetes](cluster-setup/kubernetes.md) or [OpenShift](cluster-setup/openshift.md) cluster, if you have not already installed it.

* Install the PostgreSQL operator into the `petclinic` namespace of the cluster.
  ```shell
  cat <<EOF | kubectl apply -f -
  apiVersion: v1
  kind: Namespace
  metadata:
  name: my-postgresql
  ---
  apiVersion: operators.coreos.com/v1
  kind: OperatorGroup
  metadata:
  name: operatorgroup
  namespace: my-postgresql
  spec:
  targetNamespaces:
  - petclinic
  ---
  apiVersion: operators.coreos.com/v1alpha1
  kind: Subscription
  metadata:
  name: my-postgresql
  namespace: my-postgresql
  spec:
  channel: stable
  name: postgresql
  source: operatorhubio-catalog
  sourceNamespace: olm
  EOF
  ```

* Check if the operator was installed in the cluster by running the following command:
  ```shell
  odo catalog list services
  ```
  The output can look similar to:
  ```shell
  $ odo catalog list services
  Services available through Operators
  NAME                                CRDs
  postgresoperator.v4.7.0             Pgcluster, Pgreplica, Pgpolicy, Pgtask
  service-binding-operator.v0.8.0     ServiceBinding
  ```
  
  Make sure that you see the Postgres and the Service Binding operator.


// TODO:
* Add something about odo exec and operator and linking.
* Remove the Extending Application part. Extending only makes sense if the application can be a standalone functioning unit without the extension.
  Petclinic is probably not capable of it and requires a database connection to be standalone functioning unit.
* Agendas must have their own separate sections
* Look into using MariaDB operator since the petclinin uses MySQL bts.