---
title: Deploying a Java OpenLiberty application with PostgreSQL
sidebar_position: 1
---

This tutorial illustrates deploying a [Java OpenLiberty](https://openliberty.io/) application with odo and linking it to an in-cluster PostgreSQL service in the minikube environment.

There are two roles in this example:

1. Cluster Admin - Prepare the cluster by installing the required operators on the cluster.
2. Application Developer - Imports a Java application, creates a Database instance, and connect the application to the Database instance.

## Cluster admin
---

This section assumes that you have installed [minikube and configured it](../getting-started/cluster-setup/kubernetes.md).

[//]: # (Move this section to Architecture > Service Binding or create a new Operators doc)

We will be using Operators in this guide. An Operator helps in deploying the instances of a given service, for example PostgreSQL, MySQL, Redis.

Furthermore, these Operators are "bind-able". Meaning, if they expose information necessary to connect to them, odo can help connect your component to their instances.

[//]: # (Move until here)

See the [Operator installation guide](../getting-started/cluster-setup/kubernetes.md) to install and configure an Operator in the Kubernetes cluster.

The cluster admin must install two Operators into the cluster:

1. PostgreSQL Operator
2. Service Binding Operator

We will use [Dev4Devs PostgreSQL Operator](https://operatorhub.io/operator/postgresql-operator-dev4devs-com) found on the [OperatorHub](https://operatorhub.io) to demonstrate a sample use case.
   
## Application Developer
---

This section assumes that you have [installed `odo`](../getting-started/installation.md).

Since the PostgreSQL Operator installed in above step is available only in `my-postgresql-operator-dev4devs-com` namespace, ensure that `odo` uses this namespace to perform any tasks:

```shell
odo project set my-postgresql-operator-dev4devs-com
```
### Importing the demo Java MicroService JPA application

In this example we will use odo to manage a sample [Java MicroServices JPA application](https://github.com/OpenLiberty/application-stack-samples.git).

1. Clone the sample application to your system:
    ```shell
    git clone https://github.com/OpenLiberty/application-stack-samples.git
    ```

2. Go to the sample JPA app directory:
    ```shell
    cd ./application-stack-samples/jpa
    ```

3. Initialize the application:
    ```shell
    odo create java-openliberty mysboproj
    ```
   `java-openliberty` is the type of your application and `mysboproject` is the name of your application.

4. Deploy the application to the cluster:
    ```shell
    odo push --show-log
    ```

5. The application is now deployed to the cluster - you can view the status of the cluster, and the application test results by streaming the cluster logs of the component that we pushed to the cluster in the previous step.
    ```shell
    odo log --follow
    ```

    Notice the failing tests due to an `UnknownDatabaseHostException`:
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
   Note: This error will be fixed at a later stage in the tutorial when we connect a database instance to this application.

6. You can also create a URL with `odo` to access the application:
    ```shell
    odo url create --host $(minikube ip).nip.io
    ```

7. Push the URL to activate it:
    ```shell
    odo push --show-log
    ```

8. Display the created URL:
    ```shell
    odo url list
    ```

    You will see a fully formed URL that can be used in a web browser:
    ```shell
    $ odo url list
    Found the following URLs for component mysboproj
    NAME               STATE      URL                                           PORT     SECURE     KIND
    mysboproj-9080     Pushed     http://mysboproj-9080.192.168.49.2.nip.io     9080     false      ingress
    ```

10. Use the URL to navigate to the `CreatePerson.xhtml` data entry page to use the application:
    In case of this tutorial, we will access `http://mysboproj-9080.192.168.49.2.nip.io/CreatePerson.xhtml`. Note that the URL could be different for you. Now, enter a name and age data using the form.

11. Click on the **Save** button when complete

Note that the entry of any data does not result in the data being displayed when you click on the "View Persons Record List" link, until we connect the application to a database.

### Creating a database to be used by the sample application

You can use the default configuration of the PostgreSQL Operator to start a Postgre database from it. But since our app uses few specific configuration values, lets make sure they are properly populated in the database service we start.

1. Store the YAML of the service in a file:
    ```shell
    odo service create postgresql-operator.v0.1.1/Database --dry-run > db.yaml
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
    odo service create --from-file db.yaml
   ```
   ```shell
    odo push --show-log
    ```

    This action will create a database instance pod in the `my-postgresql-operator-dev4devs-com` namespace. The application will be configured to use this database.

### Binding the database and the application

Now, the only thing that remains is to connect the DB and the application. We will use odo to create a link to the Dev4Devs PostgreSQL Database Operator in order to access the database connection information.

1. List the service associated with the database created via the PostgreSQL Operator:
    ```shell
    odo service list
   ```
   Your output should look similar to the following:
   ```shell
   $ odo service list
    NAME                       MANAGED BY ODO     STATE     AGE
    Database/sampledatabase   Yes (mysboproj)    Pushed    6m35s
    ```

2. Create a Service Binding Request between the application, and the database using the Service Binding Operator service created in the previous step:
    ```shell
    odo link Database/sampledatabase
    ```

3. Push this link to the cluster:
    ```shell
    odo push --show-log
    ```

    After the link has been created and pushed, a secret will have been created containing the database connection data that the application requires.
    
    You can inspect the new intermediate secret via the dashboard console in the `my-postgresql-operator-dev4devs-com` namespace by navigating to Secrets and clicking on the secret named `mysboproj-database-sampledatabase`: notice that it contains 4 pieces of data that are related to the connection information for your PostgreSQL database instance.

   Use `minikube dashboard` to launch the dashboard console.

    Note: Pushing the newly created link will terminate the existing application pod and start a new application pod that mounts this secret.

4. Once the new pod has initialized, you can see the secret database connection data as it is injected into the pod environment by executing the following:
    ```shell
    odo exec -- bash -c 'export | grep DATABASE' \
    declare -x DATABASE_CLUSTERIP="10.106.182.173" \
    declare -x DATABASE_DB_NAME="sampledb" \
    declare -x DATABASE_DB_PASSWORD="samplepwd" \
    declare -x DATABASE_DB_USER="sampleuser"
    ```
    Once the new version is up (there will be a slight delay until the application is available), navigate to the `CreatePerson.xhtml` using the URL created in a previous step. Enter the requested data and click the **Save** button.
    
    Notice that you are re-directed to the `PersonList.xhtml` page, where your data is displayed having been input to the postgreSQL database and retrieved for display purposes.
    
    You may inspect the database instance itself and query the table to see the data in place by using the postgreSQL command line tool, `psql`.

5. Navigate to the pod containing your db from the dashboard console. Use `minikube dashboard` to start the console.

6. Click on the terminal tab.

7. At the terminal prompt access `psql` for your database `sampledb`.
    ```shell
   psql sampledb
   ```
   Your output should look similar to the following:
   ```shell
   sh-4.2$ psql sampledb
   psql (12.3)
   Type "help" for help.

   sampledb=#
    ```

8. Issue the following SQL statement from your :
    ```postgresql
    SELECT * FROM person;
    ```

9. You can see the data that appeared in the results of the test run:
    ```shell
    sampledb=# SELECT * FROM person;

    personid | age |  name   
    ----------+-----+---------
    5 |  52 | person1
    (1 row)
    ```
