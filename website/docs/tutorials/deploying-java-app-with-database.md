---
title: Deploying a Java OpenLiberty application with PostgreSQL
sidebar_position: 1
---

This scenario illustrates deploying a Java application with odo and linking it to an in-cluster PostgreSQL service in the minikube environment.

In this example there are two roles:

1. Cluster Admin - Prepare the cluster by installing the required operators into the cluster
2. Application Developer - Imports a Java application, creates a Database instance, and connect the application to the Database instance.

## Cluster admin
---

This section assumes that you have installed minikube and configured it. See [Getting Started > Cluster Setup > Kubernetes](../getting-started/cluster-setup/kubernetes.md).

The cluster admin must install two Operators into the cluster:

1. Operator Backed Service
2. Service Binding Operator

An _Operator Backed Service_ is an operator that helps in deploying instances of a given service, for example PostgreSQL, MySQL, Redis.

Furthermore, these operators are "bind-able" as they expose information necessary for an application to connect to their instances.

We will use Dev4Devs PostgreSQL Operator found in the [OperatorHub](https://operatorhub.io) to demonstrate a sample use case.

Service Binding Operator is an operator that helps in easily binding an application to other Operator Backed Services. It accomplishes this through automatically collecting binding information and sharing with an application to bind it with operator managed backing services

### Installing the Operator Backed Service

* Run the following `kubectl` command to make the PostgreSQL Operator available in `my-postgresql-operator-dev4devs-com` namespace of your minikube cluster:
```shell
  $ kubectl create -f https://operatorhub.io/install/postgresql-operator-dev4devs-com.yaml
  ```

**Note**: The `my-postgresql-operator-dev4devs-com` Operator will be installed in the `my-postgresql-operator-dev4devs-com` namespace and will be usable from this namespace only.

### Installing the Service Binding Operator
* Run the following `kubectl` command to make the Service Binding Operator available in all namespaces on your minikube:
    ```shell
    $ kubectl create -f https://operatorhub.io/install/service-binding-operator.yaml
    ```
  Refer to [Getting Started > Cluster Setup > Operators](../getting-started/cluster-setup/operators.md) for more information on installing operators.

## Application Developer
---

This section assumes that you have installed `odo`. See [Getting Started > Installation](../getting-started/installation.md).

Since the PostgreSQL Operator installed in above step is available only in `my-postgresql-operator-dev4devs-com` namespace, ensure that `odo` uses this namespace to perform any tasks:

```shell
$ odo project set my-postgresql-operator-dev4devs-com
```
## Importing the demo Java MicroService JPA application

In this example we will use odo to manage a sample [Java MicroServices JPA application](https://github.com/OpenLiberty/application-stack-samples.git).

1. Clone the sample application to your system:
    ```shell
    $ git clone https://github.com/OpenLiberty/application-stack-samples.git
    ```

2. Go to the sample JPA app directory:
    ```shell
    $ cd ./application-stack-samples/jpa
    ```

3. Initialize the project:
    ```shell
    $ odo create java-openliberty mysboproj
    ```

4. Push the application to the cluster:
    ```shell
    $ odo push
    ```

5. The application is now deployed to the cluster - you can view the status of the cluster and the application test results by streaming the OpenShift logs to the terminal.
    ```shell
    $ odo log
    ```

    Notice the failing tests due to an UnknownDatabaseHostException:
    ```shell
    [INFO] [err] java.net.UnknownHostException: ${DATABASE_CLUSTERIP}
    [INFO] [err]    at java.base/java.net.AbstractPlainSocketImpl.connect(AbstractPlainSocketImpl.java:220)
    [INFO] [err]    at java.base/java.net.SocksSocketImpl.connect(SocksSocketImpl.java:403)
    [INFO] [err]    at java.base/java.net.Socket.connect(Socket.java:609)
    [INFO] [err]    at org.postgresql.core.PGStream.<init>(PGStream.java:68)
    [INFO] [err]    at org.postgresql.core.v3.ConnectionFactoryImpl.openConnectionImpl(ConnectionFactoryImpl.java:144)
    [INFO] [err]    ... 86 more
    [ERROR] Tests run: 2, Failures: 1, Errors: 1, Skipped: 0, Time elapsed: 0.706 s <<< FAILURE! - in org.example.app.it.DatabaseIT
    [ERROR] testGetAllPeople  Time elapsed: 0.33 s  <<< FAILURE!
    org.opentest4j.AssertionFailedError: Expected at least 2 people to be registered, but there were only: [] ==> expected: <true> but was: <false>
    at org.example.app.it.DatabaseIT.testGetAllPeople(DatabaseIT.java:57)
    
    [ERROR] testGetPerson  Time elapsed: 0.047 s  <<< ERROR!
    java.lang.NullPointerException
    at org.example.app.it.DatabaseIT.testGetPerson(DatabaseIT.java:41)
    
    [INFO]
    [INFO] Results:
    [INFO]
    [ERROR] Failures:
    [ERROR]   DatabaseIT.testGetAllPeople:57 Expected at least 2 people to be registered, but there were only: [] ==> expected: <true> but was: <false>
    [ERROR] Errors:
    [ERROR]   DatabaseIT.testGetPerson:41 NullPointer
    [INFO]
    [ERROR] Tests run: 2, Failures: 1, Errors: 1, Skipped: 0
    [INFO]
    [ERROR] Integration tests failed: There are test failures.
    ```

6. You can also create an ingress URL with `odo` to access the application:
    ```shell
    $ odo url create --host $(minikube ip).nip.io
    ```

7. Push the URL to activate it:
    ```shell
    $ odo push
    ```

8. Display the created URL:
    ```shell
    $ odo url list
    ```

    You will see a fully formed URL that can be used in a web browser:
    ```shell
    [root@pappuses1 jpa]# odo url list
    Found the following URLs for component mysboproj
    NAME               STATE      URL                                           PORT     SECURE     KIND
    mysboproj-9080     Pushed     http://mysboproj-9080.192.168.49.2.nip.io     9080     false      ingress
    [root@pappuses1 jpa]#
    ```

10. Use the URL to navigate to the `CreatePerson.xhtml` data entry page and enter requested data:
URL/CreatePerson.xhtml' and enter a user's name and age data using the form.

11. Click on the **Save** button when complete


Note that the entry of any data does not result in the data being displayed when you click on the "View Persons Record List" link.

### Creating a database to be used by the sample application

You can use the default configurations of the PostgreSQL Operator to start a Postgres database from it. But since our app uses few specific configuration values, lets make sure they are properly populated in the database service we start.

1. Store the YAML of the service in a file:
    ```shell
    $ odo service create postgresql-operator.v0.1.1/Database --dry-run > db.yaml
    ```

2. Modify and add following values under `metadata:` section in the `db.yaml` file:
    ```yaml
    name: sampledatabase
    annotations:
      service.binding/db_name: 'path={.spec.databaseName}'
      service.binding/db_password: 'path={.spec.databasePassword}'
      service.binding/db_user: 'path={.spec.databaseUser}'
    ```

    This configuration ensures that when a database service is started using this file, appropriate annotations are added to it. Annotations help the Service Binding Operator in injecting those values into the application. Hence, the above configuration will help Service Binding Operator inject the values for `databaseName`, `databasePassword` and `databaseUser` into the application.

3. Change the following values under `spec:` section of the YAML file:
    ```yaml
    databaseName: "sampledb"
    databasePassword: "samplepwd"
    databaseUser: "sampleuser"
    ```

4. Create the database from the YAML file:
    ```shell
    $ odo service create --from-file db.yaml
    $ odo push
    ```

    This action will create a database instance pod in the `my-postgresql-operator-dev4devs-com` namespace. The application will be configured to use this database.

## Binding the database and the application

Now, the only thing that remains is to connect the DB and the application. We will use odo to create a link to the Dev4Devs PostgreSQL Database Operator in order to access the database connection information.

1. Display the services available to odo: - You will see an entry for the PostgreSQL Database Operator displayed:
    ```shell
    $ odo catalog list services
    Operators available in the cluster
    NAME                                             CRDs
    postgresql-operator.v0.1.1                       Backup, Database
    ```

2. List the service associated with the database created via the PostgreSQL Operator:
    ```shell
    $ odo service list
    NAME                       MANAGED BY ODO     STATE     AGE
    Database/sampledatabase   Yes (mysboproj)    Pushed    6m35s
    ```

3. Create a Service Binding Request between the application and the database using the Service Binding Operator service created in the previous step `odo link` command:
    ```shell
    $ odo link Database/sampledatabase
    ```

4. Push this link to the cluster:
    ```shell
    $ odo push
    ```

    After the link has been created and pushed a secret will have been created containing the database connection data that the application requires.
    
    You can inspect the new intermediate secret via the dashboard console in the 'my-postgresql-operator-dev4devs-com' namespace by navigating to Secrets and clicking on the secret named `mysboproj-database-sampledatabase`: notice it contains 4 pieces of data all related to the connection information for your PostgreSQL database instance.
    
    Pushing the newly created link will also terminate the existing application pod and start a new application pod that mounts this secret.

5. Once the new pod has initialized you can see the secret database connection data as it is injected into the pod environment by executing the following:
    ```shell
    $ odo exec -- bash -c 'export | grep DATABASE'
    declare -x DATABASE_CLUSTERIP="10.106.182.173"
    declare -x DATABASE_DB_NAME="sampledb"
    declare -x DATABASE_DB_PASSWORD="samplepwd"
    declare -x DATABASE_DB_USER="sampleuser"
    ```
    Once the new version is up (there will be a slight delay until application is available), navigate to the CreatePerson.xhtml using the URL created in a previous step. Enter requested data and click the **Save** button.
    
    Notice you are re-directed to the `PersonList.xhtml` page, where your data is displayed having been input to the postgreSQL database and retrieved for display purposes.
    
    You may inspect the database instance itself and query the table to see the data in place by using the postgreSQL command line tool, psql.

6. Navigate to the pod containing your db from the Kubernetes Dashboard

7. Click on the terminal tab.

8. At the terminal prompt access psql for your database
    ```shell
    sh-4.2$ psql sampledb
    psql (12.3)
    Type "help" for help.
    
    sampledb=#
    ```

9. Issue the following SQL statement:
    ```shell
    sampledb=# SELECT * FROM person;
    ```

9. You can see the data that appeared in the results of the test run:
    ```shell
    personid | age |  name   
    ----------+-----+---------
    5 |  52 | person1
    (1 row)
    
    sampledb=#
    ```