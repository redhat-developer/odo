# ocdev e2e tests

### Pre-requisites

1. To run e2e tests, you need to have an OpenShift cluster running. You could either use [minishift](https://github.com/minishift/minishift) or [oc cluster up](https://github.com/openshift/origin/blob/master/docs/cluster_up_down.md) for the same.

1. You need to have `ocdev` and `oc` binaries in $PATH

### Running tests

Run `make test-e2e` to execute the tests. 

### Test coverage

| `ocdev` command  | Coverage (Y/N) |
| ------------- | ------------- |
| `project create` | Y |
| `project get` | Y |
| `project delete` | Y |
| `application create`  | Y  |
| `application get`  | Y  |
| `application set` | Y |
| `application delete` | Y |
| `application list` | Y |
| `catalog` | N |
| `create --git` | N |
| `create --local` | Y |
| `list` | Y |
| `push` | Y |
| `storage add` | N |
| `storage list` | N |
| `storage remove` | N |
| `update` | N |
| `url create` | Y |
| `url list` | Y |
| `url delete` | N |
