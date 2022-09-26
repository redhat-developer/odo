`odo` - Developer-focused CLI for fast & iterative application development on Kubernetes
---

[![GitHub release](https://img.shields.io/github/v/release/redhat-developer/odo?style=for-the-badge)](https://github.com/redhat-developer/odo/releases/latest)
![License](https://img.shields.io/github/license/redhat-developer/odo?style=for-the-badge)
[![Godoc](https://img.shields.io/badge/godoc-reference-007d9c?logo=go&logoColor=white&style=for-the-badge)](https://odo.dev/godoc)
[![Netlify Status](https://api.netlify.com/api/v1/badges/e07867b0-56a4-4905-92a9-a152ceab5f0d/deploy-status)](https://app.netlify.com/sites/odo-docusaurus-preview/deploys)

![logo](/docs/website/static/img/logo_small.png)

----

### Overview

`odo` is a fast, and iterative CLI tool for developers who write, build, and deploy applications on Kubernetes and OpenShift.

Why use `odo`?

* **Fast:** Spend less time maintaining your application deployment infrastructure and more time coding. Immediately have your application running each time you save.
* **Standalone:** `odo` is a standalone tool that communicates directly with the Kubernetes API. There is no requirement for a daemon or server process.
* **No configuration needed:** There is no need to dive into complex Kubernetes yaml configuration files. `odo` abstracts those concepts away and lets you focus on what matters most: code.
* **Containers first:** We provide first class support for both Kubernetes and OpenShift. Choose your favourite container orchestrator and develop your application.
* **Easy to learn:** Simple syntax and design centered around concepts familiar to developers, such as projects, applications, and components.

Learn more about the features provided by odo on [odo.dev](https://odo.dev/docs/overview/features).

----

### Installing `odo`

Please check the [installation guide on odo.dev](https://odo.dev/docs/overview/installation/).

----

### Official documentation

Visit [odo.dev](https://odo.dev/) to learn more about odo.

----

### Community, discussion, contribution, and support

#### Chat 

All of our developer and user discussions happen in the [#odo channel on the official Kubernetes Slack](https://kubernetes.slack.com/archives/C01D6L2NUAG).

If you haven't already joined the Kubernetes Slack, you can [invite yourself here](https://slack.k8s.io/).

Ask questions, inquire about odo or even discuss a new feature.

#### Issues

If you find an issue with `odo`, please [file it here](https://github.com/redhat-developer/odo/issues).

#### Contributing

* Code: We are currently working on updating our code contribution guide.
* Documentation: To contribute to the documentation, please have a look at our [Documentation Guide](https://github.com/redhat-developer/odo/wiki).

We are an open community who welcomes any concerns, changes or ideas for `odo`! Come join the chat and hang out, ask or give feedback and just generally have a good time.

#### Meetings

All our calls are open to public. You are welcome to join any of our calls.

You can find the exact dates of all scheduled odo calls together with sprint dates in the [odo calendar](https://calendar.google.com/calendar/embed?src=gi0s0v5ukfqkjpnn26p6va3jfc%40group.calendar.google.com) ([iCal format](https://calendar.google.com/calendar/ical/gi0s0v5ukfqkjpnn26p6va3jfc%40group.calendar.google.com/public/basic.ics)).

----

### Legal

#### License

Unless otherwise stated (ex. `/vendor` files), all code is licensed under the [Apache 2.0 License](LICENSE). 

#### Usage data

When `odo` is ran for the first time, you will be asked to opt-in to Red Hat's telemetry collection program.

With your approval, `odo` will collect pseudonymized usage data and send it to Red Hat servers to help improve our products and services. Read our [privacy statement](https://developers.redhat.com/article/tool-data-collection) to learn more about it. For the specific data being collected and to configure this data collection process, see [Usage data](USAGE_DATA.md).