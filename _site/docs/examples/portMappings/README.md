# portMappings

```yaml
portMappings:
- <port>:<targetPort>/<protocol>
<Kubernetes Service Spec>
```

Each service is Kubernetes Service spec and added fields.
More info: https://kubernetes.io/docs/api-reference/v1.6/#servicespec-v1-core


The only mandatory part to specify in a portMapping is "port".
There are 4 possible cases here

- When only `port` is specified - `targetPort` is set to `port` and protocol is set to `TCP`
- When `port:targetPort` is specified - protocol is set to `TCP`
- When `port/protocol` is specified - `targetPort` is set to `port`
- When `port:targetPort/protocol` is specified - no auto population is done since all values are provided

Example:
```yaml
name: httpd
containers:
- image: centos/httpd
services:
- name: httpd
  portMappings:
  - 8080:80/TCP
```