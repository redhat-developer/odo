`odo` - Developer-focused CLI for Kubernetes and OpenShift
---

[![GitHub release](https://img.shields.io/github/v/release/openshift/odo?style=for-the-badge)](https://github.com/openshift/odo/releases/latest)
![License](https://img.shields.io/github/license/openshift/odo?style=for-the-badge)
[![Godoc](https://img.shields.io/badge/godoc-reference-007d9c?logo=go&logoColor=white&style=for-the-badge)](https://odo.dev/godoc)
[![Netlify Status](https://api.netlify.com/api/v1/badges/e07867b0-56a4-4905-92a9-a152ceab5f0d/deploy-status)](https://app.netlify.com/sites/odo-docusaurus-preview/deploys)


### Overview

`odo`  is a fast, iterative, and straightforward CLI tool for developers who write, build, and deploy applications on Kubernetes and OpenShift.

Existing tools such as `kubectl` and `oc` are more operations-focused and require a deep-understanding of Kubernetes and OpenShift concepts. `odo` abstracts away complex Kubernetes and OpenShift concepts for the developer.

### Key features

`odo` is designed to be simple and concise with the following key features:

* Simple syntax and design centered around concepts familiar to developers, such as projects, applications, and components.
* Completely client based. No additional server other than Kubernetes or OpenShift is required for deployment.
* Official support for Node.js and Java components.
* Detects changes to local code and deploys it to the cluster automatically, giving instant feedback to validate changes in real time.
* Lists all the available components and services from the cluster.

Learn more about the features provided by odo on [odo.dev](https://odo.dev/docs/getting-started/features).

### Core concepts

Learn more about core concepts of odo on [odo.dev](https://odo.dev/docs/getting-started/basics).


### Usage data

When odo is run the first time, you will be asked to opt-in to Red Hat's telemetry collection program.

With your approval, odo will collect pseudonymized usage data and send it to Red Hat servers to help improve our products and services. Read our [privacy statement](https://developers.redhat.com/article/tool-data-collection) to learn more about it. For the specific data being collected and to configure this data collection process, see [Usage data](USAGE_DATA.md).

### Official documentation

Visit [odo.dev](https://odo.dev/) to learn more about odo.

### Installing `odo`

Please check the [installation guide on odo.dev](https://odo.dev/docs/getting-started/installation/).


### Community, discussion, contribution, and support


#### Chat 

All of our developer and user discussions happen in the [#odo channel on the official Kubernetes Slack](https://kubernetes.slack.com/archives/C01D6L2NUAG).

If you haven't already joined the Kubernetes Slack, you can [invite yourself here](https://slack.k8s.io/).

Ask questions, inquire about odo or even discuss a new feature.

#### Issues

If you find an issue with `odo`, please [file it here](https://github.com/openshift/odo/issues).


#### Contributing

* Code: We are currently working on updating our code contribution guide.
* Documentation: To contribute to the documentation, please have a look at our [Documentation Guide](https://odo.dev/docs/contributing/docs/).

#### Meetings

All our calls are open to public. You are welcome to join any of our calls.

You can find the exact dates of all scheduled odo calls together with sprint dates in the [odo calendar](https://calendar.google.com/calendar/embed?src=gi0s0v5ukfqkjpnn26p6va3jfc%40group.calendar.google.com) ([iCal format](https://calendar.google.com/calendar/ical/gi0s0v5ukfqkjpnn26p6va3jfc%40group.calendar.google.com/public/basic.ics)).

