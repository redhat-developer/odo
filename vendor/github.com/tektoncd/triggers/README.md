# Tekton Triggers

[![Go Report Card](https://goreportcard.com/badge/tektoncd/triggers)](https://goreportcard.com/report/github.com/tektoncd/triggers)

<p align="center">
<img src="tekton-triggers.png" alt="Tekton Triggers logo (Tekton cat playing with a ball)"></img>
</p>

Triggers is a Kubernetes
[Custom Resource Definition](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/)
(CRD) controller that allows you to extract information from events payloads (a
"trigger") to create Kubernetes resources.

The contents of this repo originated from implementing
[this design](https://docs.google.com/document/d/1fngeNn3kGD4P_FTZjAnfERcEajS7zQhSEUaN7BYIlTw/edit#heading=h.iyqzt1brkg3o)
(visible to members of
[the Tekton mailing list](https://github.com/tektoncd/community/blob/master/contact.md#mailing-list)).

* [Background](#background)
* [Want to start using Triggers?](#want-to-start-using-tekton-triggers)
* [Want to contribute?](#want-to-contribute)
* [Project roadmap](roadmap.md)

## Background

[Tekton](https://github.com/tektoncd/pipeline) is a **Kubernetes-native**,
continuous integration and delivery (CI/CD) framework that enables you to create
containerized, composable, and configurable workloads declaratively through
CRDs. Naturally, CI/CD events contain information that should:

- Identify the kind of event (GitHub Push, Gitlab Issue, Docker Hub Webhook,
  etc.)
- Be accessible from and map to particular pipelines (Take SHA from payload to
  use it in pipeline X)
- Deterministically trigger pipelines (Events/pipelines that trigger pipelines
  based on certain payload values)

The Tekton API enables functionality to be separated from configuration (e.g.
[Pipelines](https://github.com/tektoncd/pipeline/blob/master/docs/pipelines.md)
vs
[PipelineRuns](https://github.com/tektoncd/pipeline/blob/master/docs/pipelineruns.md))
such that steps can be reusable, but it does not provide a mechanism to generate
the resources (notably,
[PipelineRuns](https://github.com/tektoncd/pipeline/blob/master/docs/pipelineruns.md)
and
[PipelineResources](https://github.com/tektoncd/pipeline/blob/master/docs/resources.md#pipelineresources))
that encapsulate these configurations dynamically. Triggers extends the Tekton
architecture with the following CRDs:

- [`TriggerTemplate`](docs/triggertemplates.md) - Templates resources to be
  created (e.g. Create PipelineResources and PipelineRun that uses them)
- [`TriggerBinding`](docs/triggerbindings.md) - Validates events and extracts
  payload fields
- [`EventListener`](docs/eventlisteners.md) - Connects `TriggerBindings` and
  `TriggerTemplates` into an
  [addressable](https://github.com/knative/eventing/blob/master/docs/spec/interfaces.md)
  endpoint (the event sink). It uses the extracted event parameters from each
  `TriggerBinding` (and any supplied static parameters) to create the resources
  specified in the corresponding `TriggerTemplate`. It also optionally allows an
  external service to pre-process the event payload via the `interceptor` field.
- [`ClusterTriggerBinding`](docs/clustertriggerbindings.md) - A cluster-scoped
  TriggerBinding

Using `tektoncd/triggers` in conjunction with `tektoncd/pipeline` enables you to
easily create full-fledged CI/CD systems where the execution is defined
**entirely** through Kubernetes resources.

You can learn more by checking out the [docs](docs/README.md)

## Want to start using Tekton Triggers

[Install](./docs/install.md) Triggers, check out the
[installation guide](./docs/install.md), [examples](./examples/README.md) or
follow the [getting started guide](./docs/getting-started/README.md) to become
familiar with the project. The getting started guide walks through setting up an
end-to-end image building solution, which will be triggered from GitHub `push`
events.

### Read the docs

| Version                                                                                  | Docs                                                                                   | Examples                                                                                | Getting Started                                                                                                                 |
| ---------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------- |
| [HEAD](https://github.com/tektoncd/triggers/blob/master/DEVELOPMENT.md#install-pipeline) | [Docs @ HEAD](https://github.com/tektoncd/triggers/blob/master/docs/README.md)         | [Examples @ HEAD](https://github.com/tektoncd/triggers/blob/master/examples)            | [Getting Started @ HEAD](https://github.com/tektoncd/triggers/blob/master/docs/getting-started#getting-started-with-triggers)   |
| [v0.4.0](https://github.com/tektoncd/triggers/releases/tag/v0.4.0)                       | [Docs @ v0.4.0](https://github.com/tektoncd/triggers/tree/v0.4.0/docs#tekton-triggers) | [Examples @ v0.4.0](https://github.com/tektoncd/triggers/tree/v0.4.0/examples#examples) | [Getting Started @ v0.4.0](https://github.com/tektoncd/triggers/tree/v0.4.0/docs/getting-started#getting-started-with-triggers) |
| [v0.3.1](https://github.com/tektoncd/triggers/releases/tag/v0.3.1)                       | [Docs @ v0.3.1](https://github.com/tektoncd/triggers/tree/v0.3.1/docs#tekton-triggers) | [Examples @ v0.3.1](https://github.com/tektoncd/triggers/tree/v0.3.1/examples#examples) | [Getting Started @ v0.3.1](https://github.com/tektoncd/triggers/tree/v0.3.1/docs/getting-started#getting-started-with-triggers) |
| [v0.3.0](https://github.com/tektoncd/triggers/releases/tag/v0.3.0)                       | [Docs @ v0.3.0](https://github.com/tektoncd/triggers/tree/v0.3.0/docs#tekton-triggers) | [Examples @ v0.3.0](https://github.com/tektoncd/triggers/tree/v0.3.0/examples#examples) | [Getting Started @ v0.3.0](https://github.com/tektoncd/triggers/tree/v0.3.0/docs/getting-started#getting-started-with-triggers) |
| [v0.2.1](https://github.com/tektoncd/triggers/releases/tag/v0.2.1)                       | [Docs @ v0.2.1](https://github.com/tektoncd/triggers/tree/v0.2.1/docs#tekton-triggers) | [Examples @ v0.2.1](https://github.com/tektoncd/triggers/tree/v0.2.1/examples#examples) | [Getting Started @ v0.2.1](https://github.com/tektoncd/triggers/tree/v0.2.1/docs/getting-started#getting-started-with-triggers) |
| [v0.1.0](https://github.com/tektoncd/triggers/releases/tag/v0.1.0)                       | [Docs @ v0.1.0](https://github.com/tektoncd/triggers/tree/v0.1.0/docs#tekton-triggers) | [Examples @ v0.1.0](https://github.com/tektoncd/triggers/tree/v0.1.0/examples#examples) | [Getting Started @ v0.1.0](https://github.com/tektoncd/triggers/tree/v0.1.0/docs/getting-started#getting-started-with-triggers) |

## Want to contribute

Hooray!

- See [CONTRIBUTING.md](CONTRIBUTING.md) for an overview of our processes
- See [DEVELOPMENT.md](DEVELOPMENT.md) for how to get started
- Look at our
  [good first issues](https://github.com/tektoncd/triggers/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22)
  and our
  [help wanted issues](https://github.com/tektoncd/triggers/issues?q=is%3Aissue+is%3Aopen+label%3A%22help+wanted%22)
