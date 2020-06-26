<!--
---
linkTitle: "Event Listeners"
weight: 5
---
-->
# EventListener

EventListener is a Kubernetes custom resource that allows users a declarative
way to process incoming HTTP based events with JSON payloads. EventListeners
expose an addressable "Sink" to which incoming events are directed. Users can
declare [TriggerBindings](./triggerbindings.md) to extract fields from events,
and apply them to [TriggerTemplates](./triggertemplates.md) in order to create
Tekton resources. In addition, EventListeners allow lightweight event processing
using [Event Interceptors](#Interceptors).

- [Syntax](#syntax)
  - [ServiceAccountName](#serviceAccountName)
  - [PodTemplate](#podTemplate)
  - [Triggers](#triggers)
    - [Interceptors](#interceptors)
- [Logging](#logging)
- [Labels](#labels)
- [Examples](#examples)
- [Multi-Tenant Concerns](#multi-tenant-concerns)

## Syntax

To define a configuration file for an `EventListener` resource, you can specify
the following fields:

- Required:
  - [`apiVersion`][kubernetes-overview] - Specifies the API version, for example
    `triggers.tekton.dev/v1alpha1`.
  - [`kind`][kubernetes-overview] - Specifies the `EventListener` resource
    object.
  - [`metadata`][kubernetes-overview] - Specifies data to uniquely identify the
    `EventListener` resource object, for example a `name`.
  - [`spec`][kubernetes-overview] - Specifies the configuration information for
    your EventListener resource object. In order for an EventListener to do
    anything, the spec must include:
    - [`triggers`](#triggers) - Specifies a list of Triggers to run
    - [`serviceAccountName`](#serviceAccountName) - Specifies the ServiceAccount
      that the EventListener uses to create resources
- Optional:
  - [`serviceType`](#serviceType) - Specifies what type of service the sink pod
    is exposed as
  - [`podTemplate`](#podTemplate) - Specifies the PodTemplate
    for your EventListener pod

[kubernetes-overview]:
  https://kubernetes.io/docs/concepts/overview/working-with-objects/kubernetes-objects/#required-fields

### ServiceAccountName

The `serviceAccountName` field is required. The ServiceAccount that the
EventListener sink uses to create the Tekton resources. The ServiceAccount needs
a role with the following rules:

<!-- FILE: examples/role-resources/triggerbinding-roles/role.yaml -->
```YAML
kind: Role
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: tekton-triggers-example-minimal
rules:
# Permissions for every EventListener deployment to function
- apiGroups: ["triggers.tekton.dev"]
  resources: ["eventlisteners", "triggerbindings", "triggertemplates"]
  verbs: ["get"]
- apiGroups: [""]
  # secrets are only needed for Github/Gitlab interceptors, serviceaccounts only for per trigger authorization
  resources: ["configmaps", "secrets", "serviceaccounts"]
  verbs: ["get", "list", "watch"]
# Permissions to create resources in associated TriggerTemplates
- apiGroups: ["tekton.dev"]
  resources: ["pipelineruns", "pipelineresources", "taskruns"]
  verbs: ["create"]
```


If your EventListener is using
[`ClusterTriggerBindings`](./clustertriggerbindings.md), you'll need a
ServiceAccount with a
[ClusterRole instead](../examples/role-resources/clustertriggerbinding-roles/clusterrole.yaml).

### Triggers

The `triggers` field is required. Each EventListener can consist of one or more
`triggers`. A Trigger consists of:

- `name` - (Optional) a valid
  [Kubernetes name](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#syntax-and-character-set)
- [`interceptors`](#interceptors) - (Optional) list of interceptors to use
- `bindings` - A list of `TriggerBindings` reference to use or embedded TriggerBindingsSpecs to use.
- `template` - The name of `TriggerTemplate` to use

```yaml
triggers:
  - name: trigger-1
    interceptors:
      - github:
          eventTypes: ["pull_request"]
    bindings:
      - name: pipeline-binding
        ref:  pipeline-binding
      - name: message-binding
        spec:
            params:
              - name: message
                value: Hello from the Triggers EventListener!
    template:
      name: pipeline-template
```

Also, to support multi-tenant styled scenarios, where an administrator may not want all triggers to have
the same permissions as the `EventListener`, a service account can optionally be set at the trigger level
and used if present in place of the `EventListener` service account when creating resources:

```yaml
triggers:
  - name: trigger-1
    serviceAccount: 
      name: trigger-1-sa
      namespace: event-listener-namespace
    interceptors:
      - github:
          eventTypes: ["pull_request"]
    bindings:
      - name: pipeline-binding
        ref:  pipeline-binding
      - name: message-binding
        ref:  message-binding
    template:
      name: pipeline-template
``` 

The default ClusterRole for the EventListener allows for reading ServiceAccounts from any namespace.

### ServiceType

The `serviceType` field is optional. EventListener sinks are exposed via
[Kubernetes Services](https://kubernetes.io/docs/concepts/services-networking/service/#publishing-services-service-types).
By default, the serviceType is `ClusterIP` which means any pods running in the
same Kubernetes cluster can access services' via their cluster DNS. Other valid
values are `NodePort` and `LoadBalancer`. Check the
[Kubernetes Service types](https://kubernetes.io/docs/concepts/services-networking/service/#publishing-services-service-types)
documentations for details.

For external services to connect to your cluster (e.g. GitHub sending webhooks),
check out the guide on [exposing EventListeners](./exposing-eventlisteners.md).

## PodTemplate

The `podTemplate` field is optional. A PodTemplate is specifications for 
creating EventListener pod. A PodTemplate consists of:
- `tolerations` - list of toleration which allows pods to schedule onto the nodes with matching taints.
This is needed only if you want to schedule EventListener pod to a tainted node.

```yaml
spec:
  podTemplate:
    tolerations:
    - key: key
      value: value
      operator: Equal
      effect: NoSchedule
```

### Logging

EventListener sinks are exposed as Kubernetes services that are backed by a Pod
running the sink logic. The logging configuration can be controlled via the
`config-logging-triggers` ConfigMap present in the namespace that the
EventListener was created in. This ConfigMap is automatically created and
contains the default values defined in
[config-logging.yaml](../config/config-logging.yaml).

To access logs for the EventListener sink, you can query for pods with the
`eventlistener` label set to the name of your EventListener resource:

```shell
kubectl get pods --selector eventlistener=my-eventlistener
```

## Labels

By default, EventListeners will attach the following labels automatically to all
resources it creates:

| Name                     | Description                                            |
| ------------------------ | ------------------------------------------------------ |
| triggers.tekton.dev/eventlistener | Name of the EventListener that generated the resource. |
| triggers.tekton.dev/trigger       | Name of the Trigger that generated the resource.       |
| triggers.tekton.dev/eventid       | UID of the incoming event.                             |

Since the EventListener name and Trigger name are used as label values, they
must adhere to the
[Kubernetes syntax and character set requirements](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#syntax-and-character-set)
for label values.

## Interceptors

Triggers within an `EventListener` can optionally specify interceptors, to
modify the behavior or payload of Triggers.

Event Interceptors can take several different forms today:

- [Webhook Interceptors](#Webhook-Interceptors)
- [GitHub Interceptors](#GitHub-Interceptors)
- [GitLab Interceptors](#GitLab-Interceptors)
- [Bitbucket Interceptors](#Bitbucket-Interceptors)
- [CEL Interceptors](#CEL-Interceptors)

### Webhook Interceptors

Webhook Interceptors allow users to configure an external k8s object which
contains business logic. These are currently specified under the `Webhook`
field, which contains an
[`ObjectReference`](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.12/#objectreference-v1-core)
to a Kubernetes Service. If a Webhook Interceptor is specified, the
`EventListener` sink will forward incoming events to the service referenced by
the Interceptor over HTTP. The service is expected to process the event and
return a response back. The status code of the response determines if the
processing is successful - a 200 response means the Interceptor was successful
and that processing should continue, any other status code will halt Trigger
processing. The returned request (body and headers) is used as the new event
payload by the EventListener and passed on the `TriggerBinding`. An Interceptor
has an optional header field with key-value pairs that will be merged with event
headers before being sent;
[canonical](https://golang.org/pkg/net/textproto/#CanonicalMIMEHeaderKey) names
must be specified.

When multiple Interceptors are specified, requests are piped through each
Interceptor sequentially for processing - e.g. the headers/body of the first
Interceptor's response will be sent as the request to the second Interceptor. It
is the responsibility of Interceptors to preserve header/body data if desired.
The response body and headers of the last Interceptor is used for resource
binding/templating.

#### Event Interceptor Services

To be an Event Interceptor, a Kubernetes object should:

- Be fronted by a regular Kubernetes v1 Service over port 80
- Accept JSON payloads over HTTP
- Accept HTTP POST requests with JSON payloads.
- Return a HTTP 200 OK Status if the EventListener should continue processing
  the event
- Return a JSON body back. This will be used by the EventListener as the event
  payload for any further processing. If the Interceptor does not need to modify
  the body, it can simply return the body that it received.
- Return any Headers that might be required by other chained Interceptors or any
  bindings.

**Note**: It is the responsibility of Interceptors to preserve header/body data
if desired. The response body and headers of the last Interceptor is used for
resource binding/templating.

<!-- FILE: examples/eventlisteners/eventlistener-interceptor.yaml -->
```YAML
---
apiVersion: triggers.tekton.dev/v1alpha1
kind: EventListener
metadata:
  name: listener-interceptor
spec:
  serviceAccountName: tekton-triggers-example-sa
  triggers:
    - name: foo-trig
      interceptors:
        - webhook:
            header:
              - name: Foo-Trig-Header1
                value: string-value
              - name: Foo-Trig-Header2
                value:
                  - array-val1
                  - array-val2
            objectRef:
              kind: Service
              name: gh-validate
              apiVersion: v1
              namespace: default
      bindings:
        - ref: pipeline-binding
      template:
        name: pipeline-template
```


### GitHub Interceptors

GitHub Interceptors contain logic to validate and filter webhooks that come from
GitHub. Supported features include validating webhooks actually came from GitHub
using the logic outlined in GitHub
[documentation](https://developer.github.com/webhooks/securing/), as well as
filtering incoming events.

To use this Interceptor as a validator, create a secret string using the method
of your choice, and configure the GitHub webhook to use that secret value.
Create a Kubernetes secret containing this value, and pass that as a reference
to the `github` Interceptor.

To use this Interceptor as a filter, add the event types you would like to
accept to the `eventTypes` field. Valid values can be found in GitHub
[docs](https://developer.github.com/webhooks/#events).

The body/header of the incoming request will be preserved in this Interceptor's
response.

<!-- FILE: examples/github/github-eventlistener-interceptor.yaml -->
```YAML
---
apiVersion: triggers.tekton.dev/v1alpha1
kind: EventListener
metadata:
  name: github-listener-interceptor
spec:
  serviceAccountName: tekton-triggers-github-sa
  triggers:
    - name: github-listener
      interceptors:
        - github:
            secretRef:
              secretName: github-secret
              secretKey: secretToken
            eventTypes:
              - pull_request
      bindings:
        - ref: github-binding
      template:
        name: github-template
```


### GitLab Interceptors

GitLab Interceptors contain logic to validate and filter requests that come from
GitLab. Supported features include validating that a webhook actually came from
GitLab, using the logic outlined in GitLab
[documentation](https://docs.gitlab.com/ee/user/project/integrations/webhooks.html),
and to filter incoming events based on the event types. Event types can be found
in GitLab
[documentation](https://docs.gitlab.com/ee/user/project/integrations/webhooks.html#events).

To use this Interceptor as a validator, create a secret string using the method
of your choice, and configure the GitLab webhook to use that secret value.
Create a Kubernetes secret containing this value, and pass that as a reference
to the `gitlab` Interceptor.

To use this Interceptor as a filter, add the event types you would like to
accept to the `eventTypes` field.

The body/header of the incoming request will be preserved in this Interceptor's
response.

```yaml
apiVersion: triggers.tekton.dev/v1alpha1
kind: EventListener
metadata:
  name: gitlab-listener-interceptor
spec:
  serviceAccountName: tekton-triggers-example-sa
  triggers:
    - name: foo-trig
      interceptors:
        - gitlab:
            secretRef:
              secretName: foo
              secretKey: bar
            eventTypes:
              - Push Hook
      bindings:
        - name: pipeline-binding
          ref:  pipeline-binding
      template:
        name: pipeline-template
```

### Bitbucket Interceptors

The Bitbucket interceptor provides support for hooks originating in [Bitbucket server](https://confluence.atlassian.com/bitbucketserver), providing server hook signature validation and event-filtering.
[Bitbucket cloud](https://support.atlassian.com/bitbucket-cloud/) is not currently supported by this interceptor, as it has no secret validation, so you could match on the incoming requests using the CEL interceptor.

To use this Interceptor as a validator, create a secret string using the method
of your choice, and configure the Bitbucket webhook to use that secret value.
Create a Kubernetes secret containing this value, and pass that as a reference
to the `bitbucket` Interceptor.

To use this Interceptor as a filter, add the event types you would like to
accept to the `eventTypes` field. Valid values can be found in Bitbucket
[docs](https://confluence.atlassian.com/bitbucketserver/event-payload-938025882.html).

The body/header of the incoming request will be preserved in this Interceptor's
response.

<!-- FILE: examples/bitbucket/bitbucket-eventlistener-interceptor.yaml -->
```YAML
---
apiVersion: triggers.tekton.dev/v1alpha1
kind: EventListener
metadata:
  name: bitbucket-listener
spec:
  serviceAccountName: tekton-triggers-bitbucket-sa
  triggers:
    - name: bitbucket-triggers
      interceptors:
        - bitbucket:
            secretRef:
              secretName: bitbucket-secret
              secretKey: secretToken
            eventTypes:
              - repo:refs_changed
      bindings:
        - name: bitbucket-binding
      template:
        name: bitbucket-template
```

### CEL Interceptors

CEL Interceptors can be used to filter or modify incoming events, using the
[CEL](https://github.com/google/cel-go) expression language.

Please read the
[cel-spec language definition](https://github.com/google/cel-spec/blob/master/doc/langdef.md)
for more details on the expression language syntax.

The `cel-trig-with-matches` trigger below filters events that don't have an
`'X-GitHub-Event'` header matching `'pull_request'`.

It also modifies the incoming request, adding an extra key to the JSON body,
with a truncated string coming from the hook body.

<!-- FILE: examples/eventlisteners/cel-eventlistener-interceptor.yaml -->
```YAML
apiVersion: triggers.tekton.dev/v1alpha1
kind: EventListener
metadata:
  name: cel-listener-interceptor
spec:
  serviceAccountName: tekton-triggers-example-sa
  triggers:
    - name: cel-trig-with-matches
      interceptors:
        - cel:
            filter: "header.match('X-GitHub-Event', 'pull_request')"
            overlays:
            - key: extensions.truncated_sha
              expression: "body.pull_request.head.sha.truncate(7)"
      bindings:
      - ref: pipeline-binding
      template:
        name: pipeline-template
    - name: cel-trig-with-canonical
      interceptors:
        - cel:
            filter: "header.canonical('X-GitHub-Event') == 'push'"
      bindings:
      - ref: pipeline-binding
      template:
        name: pipeline-template
```


In addition to the standard expressions provided by CEL, Triggers supports some
useful functions for dealing with event data
[CEL expressions](./cel_expressions.md).

The body/header of the incoming request will be preserved in this Interceptor's
response.

<!-- FILE: examples/eventlisteners/cel-eventlistener-interceptor.yaml -->
```YAML
apiVersion: triggers.tekton.dev/v1alpha1
kind: EventListener
metadata:
  name: cel-listener-interceptor
spec:
  serviceAccountName: tekton-triggers-example-sa
  triggers:
    - name: cel-trig-with-matches
      interceptors:
        - cel:
            filter: "header.match('X-GitHub-Event', 'pull_request')"
            overlays:
            - key: extensions.truncated_sha
              expression: "body.pull_request.head.sha.truncate(7)"
      bindings:
      - ref: pipeline-binding
      template:
        name: pipeline-template
    - name: cel-trig-with-canonical
      interceptors:
        - cel:
            filter: "header.canonical('X-GitHub-Event') == 'push'"
      bindings:
      - ref: pipeline-binding
      template:
        name: pipeline-template
```


The `filter` expression must return a `true` value if this trigger is to be
processed, and the `overlays` applied.

Optionally, no `filter` expression can be provided, and the `overlays` will be
applied to the incoming body.
<!-- FILE: examples/eventlisteners/cel-eventlistener-no-filter.yaml -->
```YAML
apiVersion: triggers.tekton.dev/v1alpha1
kind: EventListener
metadata:
  name: cel-eventlistener-no-filter
spec:
  serviceAccountName: tekton-triggers-example-sa
  triggers:
    - name: cel-trig
      interceptors:
        - cel:
            overlays:
            - key: extensions.truncated_sha
              expression: "body.pull_request.head.sha.truncate(7)"
      bindings:
      - ref: pipeline-binding
      template:
        name: pipeline-template
```


#### Overlays

The CEL interceptor supports "overlays", these are CEL expressions that are
applied to the body before it's returned to the event-listener.

<!-- FILE: examples/eventlisteners/cel-eventlistener-multiple-overlays.yaml -->
```YAML
apiVersion: triggers.tekton.dev/v1alpha1
kind: EventListener
metadata:
  name: example-with-multiple-overlays
spec:
  serviceAccountName: tekton-triggers-example-sa
  triggers:
    - name: cel-trig
      interceptors:
        - cel:
            overlays:
            - key: extensions.truncated_sha
              expression: "body.pull_request.head.sha.truncate(7)"
            - key: extensions.branch_name
              expression: "body.ref.split('/')[2]"
      bindings:
      - ref: pipeline-binding
      template:
        name: pipeline-template
```


In this example, the bindings will see two additional fields:

Assuming that the input body looked something like this:

```json
{
  "ref": "refs/heads/master",
  "pull_request": {
    "head": {
      "sha": "6113728f27ae82c7b1a177c8d03f9e96e0adf246"
    }
  }
}
```

The output body would look like this:

```json
{
  "ref": "refs/heads/master",
  "pull_request": {
    "head": {
      "sha": "6113728f27ae82c7b1a177c8d03f9e96e0adf246"
    }
  },
  "extensions": {
    "truncated_sha": "6113728",
    "branch_name": "master"
  }
}
```

The `key` element of the overlay can create new elements in a body, or, overlay
existing elements.

For example, this expression:

```YAML
- key: body.pull_request.head.short_sha
  expression: "truncate(body.pull_request.head.sha, 7)"
```

Would see the `short_sha` being inserted into the existing body:

```json
{
  "ref": "refs/heads/master",
  "pull_request": {
    "head": {
      "sha": "6113728f27ae82c7b1a177c8d03f9e96e0adf246",
      "short_sha": "6113728"
    }
  }
}
```

It's even possible to replace existing fields, by providing a key that matches
the path to an existing value.

Anything that is applied as an overlay can be extracted using a binding e.g.

<!-- FILE: examples/triggerbindings/cel-example-trigger-binding.yaml -->
```YAML
apiVersion: triggers.tekton.dev/v1alpha1
kind: TriggerBinding
metadata:
  name: pipeline-binding-with-cel-extensions
spec:
  params:
  - name: gitrevision
    value: $(body.extensions.branch_name)
  - name: branch
    value: $(body.pull_request.head.short_sha)
```

## Examples

For complete examples, see
[the examples folder](https://github.com/tektoncd/triggers/tree/master/examples).

## Multi-Tenant Concerns

The EventListener is effectively an additional form of client into Tekton, versus what 
example usage via `kubectl` or `tkn` which you have seen elsewhere.  In particular, the HTTP based
events bypass the normal Kubernetes authentication path you get via `kubeconfig` files 
and the `kubectl config` family of commands.

As such, there are set of items to consider when deciding how to 

- best expose (each) EventListener in your cluster to the outside world.
- best control how (each) EventListener and the underlying API Objects described below access, create,
and update Tekton related API Objects in your cluster.

Minimally, each EventListener has its [ServiceAccountName](#serviceAccountName) as noted below and all
events coming over the "Sink" result in any Tekton resource interactions being done with the permissions 
assigned to that ServiceAccount.

However, if you need differing levels of permissions over a set of Tekton resources across the various
[Triggers](#triggers) and [Interceptors](#Interceptors), where not all Triggers or Interceptors can 
manipulate certain Tekton Resources in the same way, a simple, single EventListener will not suffice.

Your options at that point are as follows:

### Multiple EventListeners (One EventListener Per Namespace)

You can create multiple EventListener objects, where your set of Triggers and Interceptors are spread out across the 
EventListeners.

If you create each of those EventListeners in their own namespace, it becomes easy to assign 
varying permissions to the ServiceAccount of each one to serve your needs.  And often times namespace
creation is coupled with a default set of ServiceAccounts and Secrets that are also defined.
So conceivably some administration steps are taken care of.  You just update the permissions
of the automatically created ServiceAccounts.

Possible drawbacks:
- Namespaces with associated Secrets and ServiceAccounts in an aggregate sense prove to be the most expensive
items in Kubernetes underlying `etcd` store.  In larger clusters `etcd` storage capacity can become a concern.
- Multiple EventListeners means multiple HTTP ports that must be exposed to the external entities accessing 
the "Sink".  If you happen to have a HTTP Firewall between your Cluster and external entities, that means more
administrative cost, opening ports in the firewall for each Service, unless you can employ Kubernetes `Ingress` to
serve as a routing abstraction layer for your set of EventListeners. 

### Multiple EventListeners (Multiple EventListeners per Namespace)

Multiple EventListeners per namespace will most likely mean more ServiceAccount/Secret/RBAC manipulation for
the administrator, as some of the built in generation of those artifacts as part of namespace creation are not
applicable.

However you will save some on the `etcd` storage costs by reducing the number of namespaces.

Multiple EventListeners and potential Firewall concerns still apply (again unless you employ `Ingress`).

### ServiceAccount per EventListenerTrigger

Being able to set a ServiceAccount on an EventListenerTrigger allows for finer grained permissions as well.

You still have to create the additional ServiceAccounts.

But staying within 1 namespace and minimizing the number of EventListeners with their associated "Sinks" minimizes 
concerns around `etcd` storage and port considerations with Firewalls if `Ingress` is not utilized.

---

Except as otherwise noted, the content of this page is licensed under the
[Creative Commons Attribution 4.0 License](https://creativecommons.org/licenses/by/4.0/),
and code samples are licensed under the
[Apache 2.0 License](https://www.apache.org/licenses/LICENSE-2.0).
