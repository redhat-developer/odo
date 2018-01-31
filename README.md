# ocdev
[![Build Status](https://travis-ci.org/redhat-developer/ocdev.svg?branch=master)](https://travis-ci.org/redhat-developer/ocdev) [![codecov](https://codecov.io/gh/kadel/ocdev/branch/master/graph/badge.svg)](https://codecov.io/gh/kadel/ocdev)

## What is ocdev?
OpenShift Command line for Developers

## Installation
You can [download](https://dl.bintray.com/ocdev/ocdev/latest/) latest binaries (build automatically from current master branch).

### How to use ocdev as an oc plugin?
- make sure that ocdev binary exists in your $PATH
- copy the [plugin.yaml](./plugin.yaml) file to ~/.kube/plugins/ocdev/
- use the plugin as `oc plugin dev`
