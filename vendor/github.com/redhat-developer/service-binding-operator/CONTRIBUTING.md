# How to Contribute

#### Table Of Contents
* [Getting Started](#getting-started)
* [How Can I Contribute?](#how-can-i-contribute)
  * [Reporting Bugs](#reporting-bugs)
  * [Suggesting Enhancements](#suggesting-enhancements)
  * [Pull Requests](#pull-requests)

## Getting Started
* Fork the repository on GitHub
* Read the [README](README.md) file for build and test instructions
* Run the examples
* Explore the the project
* Submit issues and feature requests, submit changes, bug fixes, new features

## How Can I Contribute?

### Reporting Bugs

Bugs are tracked as
[GitHub issues](https://github.com/redhat-developer/service-binding-operator/issues).
Before you log a new bug, review the existing bugs to determine if the problem
that you are seeing has already been reported. If the problem has not already
been reported, then you may log a new bug.

Please describe the problem fully and provide information so that the bug
can be reproduced. Document all the steps that were performed, the
environment used, include screenshots, logs, and any other information
that will be useful in reproducing the bug.

### Suggesting Enhancements

Enhancements and requests for new features are also tracked as
[GitHub issues](https://github.com/redhat-developer/service-binding-operator/issues).
As is the case with bugs, review the existing feature requests before logging
a new request.

Please describe the feature request fully and explain the benefits that will
be derived if the feature is implemented.

### Pull Requests for Code and Documentation

All submitted code and document changes are reviewed by the project
maintainers through pull requests.

To submit a bug fix or enmhancement, log an issue in github, create a new
branch in your fork of the repo and include the issue number as a prefix in
the name of the branch. Include new tests to provide coverage for all new
or changed code. Create a pull request when you have completed code changes.
Include an informative title and full details on the code changed/added in
the git commit message and pull request description.

Before submitting the pull request, verify that all existing tests run
cleanly by executing unit and e2e tests with this make target:

```bash
make test
```
Be sure to run yamllint on all yaml files included in pull requests. Ensure
that all text in files in pull requests is compliant with:
[.editorconfig](.editorconfig)
