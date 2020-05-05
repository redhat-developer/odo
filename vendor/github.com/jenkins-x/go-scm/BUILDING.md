# Prerequisites

You will need:

 - [Git](https://git-scm.com) and a [GitHub](https://github.com) account
 - [Go](https://golang.org/) `1.12` or later
 

## Install Go

We recommend the latest version of go as this ensures the go modules works.

The installation of Go should take only a few minutes. You have more than one option to get Go up and running on your machine.

If you are having trouble following the installation guides for go, check out [Go Bootcamp](http://www.golangbootcamp.com/book/get_setup) which contains setups for every platform or reach out to the Jenkins X community in the [Jenkins X Slack channels](/community/#slack).

### Install Go on macOS

If you are a macOS user and have [Homebrew](https://brew.sh/) installed on your machine, installing Go is as simple as the following command:

```shell
$ brew install go 
```

### Install Go via GVM

More experienced users can use the [Go Version Manager](https://github.com/moovweb/gvm) (GVM). GVM allows you to switch between different Go versions *on the same machine*. If you're a beginner, you probably don't need this feature. However, GVM makes it easy to upgrade to a new released Go version with just a few commands.

GVM comes in especially handy if you follow the development of Jenkins X over a longer period of time. Future versions of Jenkins X will usually be compiled with the latest version of Go. Sooner or later, you will have to upgrade if you want to keep up.

### Install Go on Windows

Simply install the latest version by downloading the [installer](https://golang.org/dl/).


## Clearing your go module cache

If you have used an older version of go you may have old versions of go modules. So its good to run this command to clear your cache if you are having go build issues:

```shell 
go clean -modcache
``` 


# Compile and Test

    go install ./...
    go test ./...
    
