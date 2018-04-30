# Image Streams

An image stream can be used to automatically perform an action, such as updating a deployment when a new image is created.

Like most of the Kedge constructs, the `ImageStreamSpec` and `ObjectMeta` have been merged at the same YAML level.

A valid OpenShift `ImageStream` resource can be specified at the root level of the Kedge spec in a field called `imageStreams` like we see in [is.yml](is.yml):

```yaml
name: webapp
imageStreams:
- tags:
  - from:
      kind: DockerImage
      name: centos/httpd-24-centos7:2.4
    name: "2.4"
```
