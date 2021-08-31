---
title: JSON Output
sidebar_position: 1
---
# JSON Output

The `odo` commands that output some content generally accept a `-o json` flag to output this content in a JSON format, suitable for other programs to parse this output more easily.

The output structure is similar to Kubernetes resources, with `kind`, `apiVersion`, `metadata` ,`spec` and `status` fields.

List commands return a `List` resource, containing an `items` (or similar) field listing the items of the list, each item being also similar to Kubernetes resources.

Delete commands return a `Status` resource; see the [Status Kubernetes resource](https://kubernetes.io/docs/reference/kubernetes-api/common-definitions/status/).

Other commands return a resource associated with the command (`Application`, `Storage`', `URL`, etc).

The exhaustive list of commands accepting the `-o json` flag is currently:

| commands                       | Kind (version)                          | Kind (version) of list items                                 | Complete content?         | 
|--------------------------------|-----------------------------------------|--------------------------------------------------------------|---------------------------|
| odo application describe       | Application (odo.dev/v1alpha1)          | *n/a*                                                        |no                         |
| odo application list           | List (odo.dev/v1alpha1)                 | Application (odo.dev/v1alpha1)                               | ?                         |
| odo catalog list components    | List (odo.dev/v1alpha1)                 | *missing* | yes |
| odo catalog list services      | List (odo.dev/v1alpha1)                 | ClusterServiceVersion (operators.coreos.com/v1alpha1)        | ?                         |
| odo catalog describe component | *missing*                               | *n/a*                                                        | yes                       |
| odo catalog describe service   | CRDDescription (odo.dev/v1alpha1)       | *n/a*                                                        | yes                       |
| odo component create           | Component (odo.dev/v1alpha1)            | *n/a*                                                        | yes                       |
| odo component describe         | Component (odo.dev/v1alpha1)            | *n/a*                                                        | yes                       |
| odo component list             | List (odo.dev/v1alpha1)                 | Component (odo.dev/v1alpha1)                                 | yes                       |
| odo config view                | DevfileConfiguration (odo.dev/v1alpha1) | *n/a*                                                        | yes                       |
| odo debug info                 | OdoDebugInfo (odo.dev/v1alpha1)         | *n/a*                                                        | yes                       |
| odo env view                   | EnvInfo (odo.dev/v1alpha1)              | *n/a*                                                        | yes                       |
| odo preference view            | PreferenceList (odo.dev/v1alpha1)       | *n/a*                                                        | yes                       |
| odo project create             | Project (odo.dev/v1alpha1)              | *n/a*                                                        | yes                       |
| odo project delete             | Status (v1)                             | *n/a*                                                        | yes                       |
| odo project get                | Project (odo.dev/v1alpha1)              | *n/a*                                                        | yes                       |
| odo project list               | List (odo.dev/v1alpha1)                 | Project (odo.dev/v1alpha1)                                   | yes                       |
| odo registry list              | List (odo.dev/v1alpha1)                 | *missing*                                                    | yes                       |
| odo service list               | *missing*                               | *depending on services types*                                | ?                         |
| odo storage create             | Storage (odo.dev/v1alpha1)              | *n/a*                                                        | yes                       |
| odo storage delete             | Status (v1)                             | *n/a*                                                        | yes                       |
| odo storage list               | List (odo.dev/v1alpha1)                 | Storage (odo.dev/v1alpha1)                                   | yes                       |
| odo url list                   | List (odo.dev/v1alpha1)                 | URL (odo.dev/v1alpha1)                                       | yes                       |
