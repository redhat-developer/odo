---
title: odo service
sidebar_position: 6
---
# odo service

odo can deploy *services* with the help of *operators*.

The list of available operators and services available for installation can be found with the [`odo catalog` command](/docs/command-reference/catalog).

Services are created in the context of a *component*, so you should have run [`odo create`](/docs/command-reference/create) before you deploy services.

The deployment of a service is done in two steps:
1. Define the service and store its definition in the devfile,
2. Deploy the defined service to the cluster, using `odo push`.

## Creating a new service

You can create a new service with the command:

```
odo service create
```

For example, to create an instance of a Redis service named `my-redis-service`, you can run:

```
$ odo catalog list services
Services available through Operators
NAME                      CRDs
redis-operator.v0.8.0     RedisCluster, Redis

$ odo service create redis-operator.v0.8.0/Redis my-redis-service
Successfully added service to the configuration; do 'odo push' to create service on the cluster
```

This command creates a Kubernetes manifest in the `kubernetes/` directory, containing the definition of the service, and this file is referenced from the `devfile.yaml` file.

```
$  cat kubernetes/odo-service-my-redis-service.yaml 
apiVersion: redis.redis.opstreelabs.in/v1beta1
kind: Redis
metadata:
  name: my-redis-service
spec:
  kubernetesConfig:
    image: quay.io/opstree/redis:v6.2.5
    imagePullPolicy: IfNotPresent
    resources:
      limits:
        cpu: 101m
        memory: 128Mi
      requests:
        cpu: 101m
        memory: 128Mi
    serviceType: ClusterIP
  redisExporter:
    enabled: false
    image: quay.io/opstree/redis-exporter:1.0
  storage:
    volumeClaimTemplate:
      spec:
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
```

```
$ cat devfile.yaml
[...]
components:
- kubernetes:
    uri: kubernetes/odo-service-my-redis-service.yaml
  name: my-redis-service
[...]
```

Note that the name of the created instance is optional. If you do not provide a name, it will be the lowercased name of the service. For example, the following command will create an instance of a Redis service named `redis`:

```
$ odo service create redis-operator.v0.8.0/Redis
```

### Inlining the manifest

By default, a new manifest is created in the `kubernetes/` directory, referenced from the `devfile.yaml` file. It is possible to inline the manifest inside the `devfile.yaml` file using the `--inlined` flag:

```
$ odo service create redis-operator.v0.8.0/Redis my-redis-service --inlined
Successfully added service to the configuration; do 'odo push' to create service on the cluster

$ cat devfile.yaml
[...]
components:
- kubernetes:
    inlined: |
      apiVersion: redis.redis.opstreelabs.in/v1beta1
      kind: Redis
      metadata:
        name: my-redis-service
      spec:
        kubernetesConfig:
          image: quay.io/opstree/redis:v6.2.5
          imagePullPolicy: IfNotPresent
          resources:
            limits:
              cpu: 101m
              memory: 128Mi
            requests:
              cpu: 101m
              memory: 128Mi
          serviceType: ClusterIP
        redisExporter:
          enabled: false
          image: quay.io/opstree/redis-exporter:1.0
        storage:
          volumeClaimTemplate:
            spec:
              accessModes:
              - ReadWriteOnce
              resources:
                requests:
                  storage: 1Gi
  name: my-redis-service
[...]
```

### Configuring the service

Without specific indication, the service will be created with a default configuration. You can use either command-line arguments or a file to specify your own configuration.

#### Using command-line arguments

You can use the `--parameters` (or `-p`) flag to specify your own configuration.

In the following example, we will configure the Redis service with three parameters:

```
$ odo service create redis-operator.v0.8.0/Redis my-redis-service \
    -p kubernetesConfig.image=quay.io/opstree/redis:v6.2.5 \
    -p kubernetesConfig.serviceType=ClusterIP \
    -p redisExporter.image=quay.io/opstree/redis-exporter:1.0
Successfully added service to the configuration; do 'odo push' to create service on the cluster

$ cat kubernetes/odo-service-my-redis-service.yaml 
apiVersion: redis.redis.opstreelabs.in/v1beta1
kind: Redis
metadata:
  name: my-redis-service
spec:
  kubernetesConfig:
    image: quay.io/opstree/redis:v6.2.5
    serviceType: ClusterIP
  redisExporter:
    image: quay.io/opstree/redis-exporter:1.0
```

You can obtain the possible parameters for a specific service from the [`odo catalog describe service` command](/docs/command-reference/catalog/#getting-information-about-a-service).

#### Using a file

You can use a YAML manifest to specify your own specification.

In the following example, we will configure the Redis service with three parameters. For this, first create a manifest:

```
$ cat > my-redis.yaml <<EOF
apiVersion: redis.redis.opstreelabs.in/v1beta1
kind: Redis
metadata:
  name: my-redis-service
spec:
  kubernetesConfig:
    image: quay.io/opstree/redis:v6.2.5
    serviceType: ClusterIP
  redisExporter:
    image: quay.io/opstree/redis-exporter:1.0
EOF
```

Then create the service from the manifest:

```
$ odo service create --from-file my-redis.yaml
Successfully added service to the configuration; do 'odo push' to create service on the cluster
```

## Deleting a service

You can delete a service with the command:

```
odo service delete
```

For example:

```
$ odo service list
NAME                       MANAGED BY ODO     STATE               AGE
Redis/my-redis-service     Yes (api)          Deleted locally     5m39s

$ odo service delete Redis/my-redis-service
? Are you sure you want to delete Redis/my-redis-service Yes
Service "Redis/my-redis-service" has been successfully deleted; do 'odo push' to delete service from the cluster
```

You can use the `--force` (or `-f`) flag to force the deletion of the service without confirmation.

## Listing services

You can get the list of services created for your component with the command:

```
odo service list
```

For example:

```
$ odo service list
NAME                       MANAGED BY ODO     STATE             AGE
Redis/my-redis-service-1   Yes (api)          Not pushed     
Redis/my-redis-service-2   Yes (api)          Pushed            52s
Redis/my-redis-service-3   Yes (api)          Deleted locally   1m22s
```

For each service, `STATE` indicates if the service has been pushed to the cluster using `odo push`, or if the service is still running on the cluster but removed from the devfile locally using `odo service delete`.

## Getting information about a service

You can get the details about a service such as its kind, version, name and list of configured parameters with the command:

```
odo service describe
```


For example:

```
$ odo service describe Redis/my-redis-service
Version: redis.redis.opstreelabs.in/v1beta1
Kind: Redis
Name: my-redis-service
Parameters:
NAME                           VALUE
kubernetesConfig.image         quay.io/opstree/redis:v6.2.5
kubernetesConfig.serviceType   ClusterIP
redisExporter.image            quay.io/opstree/redis-exporter:1.0
```
