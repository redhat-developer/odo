# Using Routes

`routes` is a field at the root level of the Kedge spec, which can be used
to define OpenShift routes.

An OpenShift route is a way to expose a service by giving it an
externally-reachable hostname.
More info about routes can be found [here](https://docs.openshift.com/enterprise/3.0/architecture/core_concepts/routes.html)

Much like other resource definitions in Kedge, routes have been implemented
by merging `RouteSpec` and `ObjectMeta`, which would mean that you can
define fields like `name` and the rest of the RouteSpec at the same level.

A snippet from [httpd.yml](httpd.yml):

```yaml
...
routes:
- to:
    kind: Service
    name: httpd
```

A more detailed example might look like this:

```yaml
...
routes:
- name: webroute
  host: httpd-web.192.168.42.69.nip.io
  to:
    kind: Service
    name: httpd
    weight: 100
  wildcardPolicy: None
```

## Ref

- [What are OpenShift routes](https://docs.openshift.com/enterprise/3.0/architecture/core_concepts/routes.html) 
- [OpenShift v1.Route API](https://docs.openshift.org/latest/rest_api/apis-route.openshift.io/v1.Route.html)
