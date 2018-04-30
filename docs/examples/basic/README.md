# Simple example

This directory has simplest of the examples.

## Defining Kubernetes Services

See following snippet from [web.yaml](./web.yaml) the way services info can be given in root level `services` field.

```yaml
services:
- name: wordpress
  type: NodePort
  ports:
  - port: 8080
    targetPort: 80
```

It is a list of [service spec](https://kubernetes.io/docs/api-reference/v1.6/#servicespec-v1-core), which means that each app can contain multiple services defined. Also see that ports are defined in the `services` field but not in the container field. You can choose declare ports in the `containers.ports` as well but, it is not required.

