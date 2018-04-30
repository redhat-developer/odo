# Ingress

An Ingress is a collection of rules that allow inbound connections to
reach the cluster services.

Like the way most of the Kedge constructs are, for an Ingress resource
ObjectMeta and IngressSpec have been merged at the same YAML level.

A valid `Ingress` resource can be specified at the rool level of the spec,
in a field called `ingresses` like we see in the [web.yaml](web.yaml):

```yaml
ingresses:
- name: root-ingress
  rules:
  - host: minikube.external
    http:
      paths:
      - backend:
          serviceName: wordpress
          servicePort: 8080
        path: /
```
