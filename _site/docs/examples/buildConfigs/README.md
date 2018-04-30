# BuildConfig

A build configuration describes a single build definition and a set of triggers for when a new build should be created. A build configuration is defined by a BuildConfig.

In Kedge, defining BuildConfigs is implemented by merging `BuildConfigSpec` and `ObjectMeta` at the root YAML level of the spec.

A BuildConfig can be defined in Kedge as follows -

```yaml
buildConfigs:
- triggers:
  - type: "ImageChange"
  source:
    type: "Git"
    git:
      uri: "https://github.com/openshift/ruby-hello-world"
  strategy:
    type: "Source"
    sourceStrategy:
      from:
        kind: "ImageStreamTag"
        name: "ruby-22-centos7:latest"
  output:
    to:
      kind: "ImageStreamTag"
      name: "origin-ruby-sample:latest"
  postCommit:
      script: "bundle exec rake test"
```
