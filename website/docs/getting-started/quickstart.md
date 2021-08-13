---
title: Quickstart
sidebar_position: 1
---
This quickstart shows you how to build and deploy a real world application in a cloud-native environment using odo; it uses a Spring Boot application built using Maven for example.

Agenda:
* Deploy the Spring Boot application to a Kubernetes cluster.
* Be able to access the application outside the cluster.
* Create an instance of the Postgres service.
* Link the Spring Boot application with Postgres service.

## Pre-requisite
* Setup a [Kubernetes](cluster-setup/kubernetes.md)/[OpenShift](cluster-setup/openshift.md) cluster.
* [Install odo](installation.md) on your system.
* If it is a Kubernetes cluster, have [ingress-controller installed](cluster-setup/kubernetes.md) on the cluster.
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

7. Create a URL to access this application outside the cluster.

  _OpenShift_ - If you are using an OpenShift cluster, you can skip this step, URL will be automatically created for you.

  _Kubernetes_ - If you are using a Kubernetes cluster, make sure you have an [ingress controller](https://kubernetes.io/docs/concepts/services-networking/ingress-controllers/) installed on your cluster before creating the URL.
   ```shell
    odo url create  --port 8080 --host <your-ingress-domain>
    ```

  _Minikube_ - If you are using a Minikube cluster, make sure you have enabled the ingress addon before creating the URL.
  ```shell
  minikube addons enable ingress
  ```
  ```shell
  odo url create --port 8080 --host=$(minikube ip).nip.io
  ```
5. Deploy the application into the cluster. This first time deployment might take anywhere between 4-60minutes depending on your internet and cluster connection.
  ```shell
  odo push --show-log
  ```

6. Check the application description for more information about it.
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
  URLs:
  · http://8080-tcp.example.com exposed via 8080
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
  Alternatively, you can also use the URL in the output of `odo describe` from the previous step.

9. Access the URL via web browser. You should be able to see the Petclinic application.

## Extending the application: Connecting the application to a PostgreSQL service

### Pre-requisite
* If it is a Kubernetes cluster,[install the Operator Lifecycle Manager(OLM)](cluster-setup/kubernetes.md) on the cluster.
  
  If it is an OpenShift cluster, there is no need.

* Install the Service Binding Operator on the [Kubernetes](cluster-setup/kubernetes.md)/[OpenShift](cluster-setup/openshift.md) cluster.

* Install the PostgreSQL operator into the `petclinic` namespace of the cluster.
  **Note**: If you are using a namespace other than `petclinic`, then make sure to change the `targetNamespace` in `OperatorGroup` resource inside the `crunchy-postgresql-community.yaml` file.
  ```shell
  kubectl apply -f https://odo.dev/resources/crunchy-postgresql-community.yaml
  ```

* Check if the operators are installed in the cluster by running the following command:
  ```shell
  odo catalog list services
  ```
  The output can look similar to:
  ```shell
  $ odo catalog list services
  Services available through Operators
  NAME                                CRDs
  postgresoperator.v4.7.0             Pgcluster, Pgreplica, Pgpolicy, Pgtask
  service-binding-operator.v0.9.1     ServiceBinding, ServiceBinding
  ```
  
  Make sure that you see the Postgres and the Service Binding operator.

### Create a Postgres Database Instance
Normally you could use odo command to do this. You could run something like this.
```shell
odo service create postgresoperator.v4.7.0/Pgcluster mypg -p name=mypg -p database=mydb -p clustername=mypg -p user=user1
```
But, this is currently not possible due to [odo#issue4916](https://github.com/openshift/odo/issues/4916), so we go with the alternate way.
1. Save the definition file of an operator-backed service.
  ```shell
  odo service create postgresoperator.v4.7.0/Pgcluster mypg --dry-run > db.yaml
  ```

2. Edit the `db.yaml` file and add the required annotations to the list in `.metadata.annotations`.
  ```yaml
  metadata:
    annotations:
      service.binding/database: "path={.spec.database}"
      service.binding/username: "path={.spec.user}"
      service.binding/port: "path={.spec.port}"
  ```
  Edit the other parameters if you like, but for this tutorial we will use all the default values.
3. Create the service from the definition file.
  ```shell
  odo service create --from-file db.yaml
  ```

4. Deploy the changes to the cluster.
  ```shell
  odo push --show-log
  ```
5. See the database instance.
  ```shell
  odo service list
  ```
  The output can look similar to:
  ```shell
  $ odo list services
  NAME                MANAGED BY ODO      STATE      AGE
  Pgcluster/hippo     Yes (petclinic)     Pushed     1m24s
  ```
  Notice the 'Pushed' state.

### Link the application with the database service
1. Create a link between the Postgres database instance and the application.
  ```shell
  odo link Pgcluster/hippo
  ```
2. Deploy the changes to the cluster.
  ```shell
  odo push --show-log
  ```
3. We now have our application running and database running, but the application is not communicating with database because it is unaware of the password required to connect to the database. The Service Binding Operator helps in linking the application should be able to retrieve the password, but it is not currently possible. So we go obtain the password manually and tell our application about it.
  ```shell
  kubectl get secret hippo-hippo-secret -o "jsonpath={.data['password']}" | base64 -d
  ```
  Take a note of this password.
4. Now, create a new file `src/main/resources/application-postgresql.properties` and add the following data:
  ```properties
  database=postgresql
  spring.datasource.url=jdbc:postgresql://${PGCLUSTER_HOST}:${PGCLUSTER_PORT}/${PGCLUSTER_DATABASE}
  spring.datasource.username=${PGCLUSTER_USERNAME}
  spring.datasource.password=
  
  spring.datasource.initialization-mode=always
  spring.jpa.generate-ddl=true
  ```
  Set the password obtained in the previous step to `spring.datasource.password`.
5. Add postgresql dependency to `pom.xml`.
```xml
<dependencies>
    ...
    <dependency>
        <groupId>org.postgresql</groupId>
        <artifactId>postgresql</artifactId>
        <scope>runtime</scope>
    </dependency>
</dependencies>
```
6. Tell the Spring Boot application to use the newly created postgresql profile.
```shell
odo config set --env SPRING_PROFILES_ACTIVE=postgresql
```
7. Deploy the changes.
```shell
odo push --show-log
```
