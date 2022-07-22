#!/bin/sh

set -ex

rm -f go.mod go.sum

go mod init github.com/redhat-developer/odo

go mod edit -require oras.land/oras-go@v0.4.0 # for github.com/devfile/registry-support

# for odo
go mod edit -require github.com/devfile/api/v2@v2.0.0-20220309195345-48ebbf1e51cf
go mod edit -require github.com/devfile/library@v1.2.1-0.20220602130922-85a4805bd59c
go mod edit -require github.com/devfile/registry-support/index/generator@v0.0.0-20220222194908-7a90a4214f3e
go mod edit -require github.com/devfile/registry-support/registry-library@v0.0.0-20220504150710-21de53798172
go mod edit -require github.com/openshift/oc@v0.0.0-alpha.0.0.20220402064836-f1f09a392fd1
go mod edit -require github.com/kubernetes-sigs/service-catalog@v0.3.1
go mod edit -require k8s.io/utils@v0.0.0-20220210201930-3a6ce19ff2f9
go mod edit -require github.com/redhat-developer/alizer/go@v0.0.0-20220714080930-59b004fd4586
go mod edit -require github.com/redhat-developer/service-binding-operator@v1.0.1-0.20211222115357-5b7bbba3bfb3
go mod edit -require github.com/onsi/ginkgo/v2@v2.1.4
go mod edit -require github.com/segmentio/backo-go@v1.0.1-0.20200129164019-23eae7c10bd3 # fixing this revision because of the missing license in the latest released version
go mod edit -replace gopkg.in/segmentio/analytics-go.v3=github.com/segmentio/analytics-go/v3@v3.2.1
go mod edit -replace github.com/apcera/gssapi=github.com/openshift/gssapi@v0.0.0-20161010215902-5fb4217df13b # for oc

# Upgrades recommended by Dependabot
go mod edit -require github.com/pborman/uuid@v1.2.1
go mod edit -require github.com/posener/complete@v1.2.3
go mod edit -require github.com/fatih/color@v1.13.0
go mod edit -require github.com/jedib0t/go-pretty/v6@v6.3.5
go mod edit -require github.com/golang/mock@v1.6.0
go mod edit -require k8s.io/klog/v2@v2.70.1

go mod tidy -compat=1.17 # why?

go mod vendor