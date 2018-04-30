# ConfigMaps

## Defining a configMap

See the following snippet from [web.yaml](./web.yaml)

```yaml
configMaps:
- data:
    WORDPRESS_DB_NAME: wordpress
    WORDPRESS_DB_HOST: "database:3306"
```

Define a root level field called `configMaps`. It is just a key value pair.
If this is defined a `configMap` with the `name` of app is created.

You can also define the name of `configMap` using field called `name`.

e.g.

```yaml
configMaps:
- name: web
  data:
    WORDPRESS_DB_NAME: wordpress
```

## Consuming the configMap

See the following code snippet from [db.yaml](./db.yaml)

```yaml
  - name: MYSQL_DATABASE
    valueFrom:
      configMapKeyRef:
        key: MYSQL_DATABASE
        name: database
```

This is similar to the way `configMap` is referred in Kubernetes.


## Ref

- [Referring a Config Map](https://kubernetes.io/docs/api-reference/v1.6/#envvarsource-v1-core)
- [ConfigMap APIs](https://kubernetes.io/docs/api-reference/v1.6/#configmap-v1-core)
- [Configure Containers Using a ConfigMap](https://kubernetes.io/docs/tasks/configure-pod-container/configmap/)
- [Use ConfigMap Data in Pods](https://kubernetes.io/docs/tasks/configure-pod-container/configure-pod-configmap/)
