# Tekton Triggers Roadmap

In 2019 we created a simple system for creating instances of Tekton resources, triggered
by json payloads sent to HTTP endpoints (Tekton Triggers).

In 2020 we would like to add missing features and then push for a `beta` release in the
same year.

We are targeting improving the experience for both end users and operators:

* For end users:
  * [Pluggable core interceptors](https://github.com/tektoncd/triggers/issues/271)
  * [Increased expression support in TriggerBindings](https://github.com/tektoncd/triggers/issues/367)
  * [Using TriggerTemplates outside the context of an event](https://github.com/tektoncd/triggers/issues/200)
  * [Routing to multiple interceptors](https://github.com/tektoncd/triggers/issues/205)
  * [Dynamic TriggerTemplate parameters](https://github.com/tektoncd/triggers/issues/87)
  * Support for poll-based triggering (e.g. when a repo changes state)
  * Support for additional expression languages
  * [GitHub App support](https://github.com/tektoncd/triggers/issues/189)
* For operators:
  * [Improved support for many EventListeners](https://github.com/tektoncd/triggers/issues/370)
  * Increased traceability (e.g. why did my interceptor reject the event?)
  * [Performant Triggers](https://github.com/tektoncd/triggers/issues/406)
  * A scale-to-zero `EventListener`
