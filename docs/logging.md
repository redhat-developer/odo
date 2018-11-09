---
layout: default
permalink: /logging/
redirect_from: 
  - /docs/logging.md/
---

# Logging in Odo

[Glog](https://godoc.org/github.com/golang/glog) is used for V style logging in Odo.


## Working

Every Odo command takes an optional flag `-v` that must be used with an integer log level in the range from 0-9. Any INFO severity log statement that is logged at a level lesser than or equal to the one passed with the command alongside `-v` flag will be logged to STDOUT.

All ERROR severity level log messages will always be logged regardless of the passed `v` level.


## Usage

Every source file that requires logging will need to import glog:

``` import "github.com/golang/glog" ```

Any default debug level severity messages need to be logged using:

``` glog.V(4).Infof(msg, args...) ```

For more info level logging conventions please refer [here](https://kubernetes.io/docs/reference/kubectl/cheatsheet/#kubectl-output-verbosity-and-debugging).

Error messages can be logged as under:

``` glog.Errorf(msg, args...) ```

Warning messages can be logged as under:

``` glog.Warningf(msg, args...) ```
