Custom Resource Status
======================

[![Go Report Card](https://goreportcard.com/badge/github.com/openshift/custom-resource-status)](https://goreportcard.com/report/github.com/openshift/custom-resource-status)
[![Go Doc](https://godoc.org/github.com/openshift/custom-resource-status?status.svg)](http://godoc.org/github.com/openshift/custom-resource-status)
[![GitHub Issues](https://img.shields.io/github/issues/openshift/custom-resource-status.svg)](https://github.com/openshift/custom-resource-status/issues)
[![Licensed under Apache License version 2.0](https://img.shields.io/github/license/openshift/custom-resource-status.svg?maxAge=2592000)](https://www.apache.org/licenses/LICENSE-2.0)

The purpose of this project is to provide some level of standardization and
best-practices with respect to managing the status of custom resources. This project
steals, err draws from:

* [Cluster Version Operator (CVO)](https://github.com/openshift/cluster-version-operator)
  that manages essential OpenShift operators.
* [ClusterOperator Custom Resource](https://github.com/openshift/cluster-version-operator/blob/master/docs/dev/clusteroperator.md#what-should-an-operator-report-with-clusteroperator-custom-resource)
  that exists for operators managed by CVO to communicate their status.
* [openshift/library-go ClusterOperator status helpers](https://github.com/openshift/library-go/blob/master/pkg/config/clusteroperator/v1helpers/status.go)
  that makes it easy to manage the status on a ClusterOperator resource.

The goal here is to prescribe, without mandate, how to meaningfully populate the
status of the Custom Resources your operator manages. Types, constants, and
functions are provided for the following:

* [Conditions](conditions/README.md)
* [Object References](objectreferences/README.md)
