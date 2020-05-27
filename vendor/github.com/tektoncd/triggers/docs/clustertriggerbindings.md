<!--
---
linkTitle: "Cluster Trigger Binding"
weight: 7
---
-->
# ClusterTriggerBindings

`ClusterTriggerBindings` is similar to TriggerBinding which is used to extract
field from event payload. The only difference is it is cluster-scoped and
designed to encourage reusability clusterwide. You can reference a
ClusterTriggerBinding in any EventListener in any namespace.

<!-- FILE: examples/clustertriggerbindings/clustertriggerbinding.yaml -->
```YAML
apiVersion: triggers.tekton.dev/v1alpha1
kind: ClusterTriggerBinding
metadata:
  name: pipeline-clusterbinding
spec:
  params:
    - name: gitrevision
      value: $(body.head_commit.id)
    - name: gitrepositoryurl
      value: $(body.repository.url)
    - name: contenttype
      value: $(header.Content-Type)
```


You can specify multiple ClusterTriggerBindings in a Trigger. You can use a
ClusterTriggerBinding in multiple Triggers.

In case of using a ClusterTriggerBinding, the `Binding` kind should be added.
The default kind is TriggerBinding which represents a namespaced TriggerBinding.

<!-- FILE: examples/eventlisteners/eventlistener-clustertriggerbinding.yaml -->
```YAML
---
apiVersion: triggers.tekton.dev/v1alpha1
kind: EventListener
metadata:
  name: listener-clustertriggerbinding
spec:
  serviceAccountName: tekton-triggers-example-sa
  triggers:
    - name: foo-trig
      bindings:
        - ref: pipeline-clusterbinding
          kind: ClusterTriggerBinding
        - ref: message-clusterbinding
          kind: ClusterTriggerBinding
      template:
        name: pipeline-template
```

