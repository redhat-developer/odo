---
title: Logging in odo
sidebar_position: 3
---
odo uses [Klog](https://github.com/kubernetes/klog) for V style logging.

### Logging categories

Each logging output is categorizes into logging levels.

|Logging value  |Output                                                 |
|---------------|-------------------------------------------------------|
|V(0)           |General information                                    |
|V(1)           |Experimental debugging output / testing suite          |
|V(2)           |Devfile debugging output                               |
|V(3)           |Kubernetes and OpenShift client debugging output       |
|V(4)           |Generic debugging information                          |
|V(10)          |Kubernetes and OpenShift HTTP library debugging output |


### Working

Every `odo` command takes an optional flag `-v` that must be used with an integer log level in the range from 0-9. Any INFO severity log statement that is logged at a level lesser than or equal to the one passed with the command alongside `-v` flag will be logged to STDOUT.

Alternatively, environment variable `ODO_LOG_LEVEL` can be used to set the verbosity of the log level in the integer range 0-9. The value set by `ODO_LOG_LEVEL` can be overridden by explicitly passing the `-v` command line flag to the `odo` command, in such an event `-v` flag takes precedence over the environment variable `ODO_LOG_LEVEL`.

All ERROR severity level log messages will always be logged regardless of the values of `v` or `ODO_LOG_LEVEL`.

In order to list multiple log levels, you can provide `--vmodule` to odo. `--vmodule` takes a key=value command-delimited output. For example: `--vmodule "*=1,*=2,*=3"` to list log levels 1, 2 and 3 for all files (`*`), while `--vmodule "occlient=3"` would list all 3 level log messages in the `occlient` folder.

### Usage

Every source file that requires logging will need to import klog:

```go
import "k8s.io/klog"
```

Any default debug level severity messages need to be logged using:

```go
klog.V(4).Infof(msg, args...)
```

For more information on level logging conventions please refer [here](https://kubernetes.io/docs/reference/kubectl/cheatsheet/#kubectl-output-verbosity-and-debugging).

Error messages can be logged as under:

```go
klog.Errorf(msg, args...)
```

Warning messages can be logged as under:

```go
klog.Warningf(msg, args...)
```

In addition to the above logging, the following hidden flags are available for debugging:


```shell
--add_dir_header                   If true, adds the file directory to the header
--alsologtostderr                  Log to standard error as well as files
--log_backtrace_at traceLocation   When logging hits line file:N, emit a stack trace (default :0)
--log_dir string                   If non-empty, write log files in this directory
--log_file string                  If non-empty, use this log file
--log_file_max_size uint           Defines the maximum size a log file can grow to. Unit is megabytes. If the value is 0, the maximum file size is unlimited. (default 1800)
--logtostderr                      Log to standard error instead of files (default true)
--skip_headers                     If true, avoid header prefixes in the log messages
--skip_log_headers                 If true, avoid headers when opening log files
--stderrthreshold severity         Logs at or above this threshold go to stderr (default 2)
```
