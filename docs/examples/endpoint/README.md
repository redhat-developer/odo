# Endpoint

See the following snippet from [web.yaml](web.yaml).

```yaml
services:
- name: wordpress
  ports:
  - port: 8080
    targetPort: 80
    endpoint: minikube.local
```

Services here is an extension of [`service` spec](https://kubernetes.io/docs/api-reference/v1.6/#servicespec-v1-core),
such that the `ServicePort` struct is extended with an `endpoint` field.  So if
you want a service to be exposed via an ingress set the `endpoint` field in the
format `ingress_host/ingress_path`.

When the `generate` command is run against this file, we can see that an ingress
resource is populated automatically with the following parameters -

```yaml
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  creationTimestamp: null
  labels:
    app: web
  name: wordpress-8080
spec:
  rules:
  - host: minikube.local
    http:
      paths:
      - backend:
          serviceName: wordpress
          servicePort: 8080
        path: /foo
status:
  loadBalancer: {}
```
