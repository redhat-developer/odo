# odo e2e tests

### Pre-requisites

1. To run e2e tests, you need to have an OpenShift cluster running. You could either use [minishift](https://github.com/minishift/minishift) or [oc cluster up](https://github.com/openshift/origin/blob/master/docs/cluster_up_down.md) for the same.

1. You need to have `odo` and `oc` binaries in $PATH

### Running tests

Run `make test-e2e` to execute the tests. 

### Test coverage

| `odo` command  | Coverage (Y/N) |
| ------------- | ------------- |
| `project create` | Y |
| `project get` | Y |
| `project delete` | Y |
| `project list` | N |
| `project set` | N |
| `app create`  | Y  |
| `app get`  | Y  |
| `app set` | Y |
| `app delete` | Y |
| `app describe` | N |
| `app list` | Y |
| `catalog list` | Y |
| `catalog search` | N |
| `component get` | Y |
| `component set` | Y |
| `create` | Y |
| `create --git` | N |
| `create --local` | Y |
| `create --binary` | Y |
| `delete` | N |
| `describe` | N |
| `link` | Y |
| `list` | Y |
| `push` --local | Y |
| `push` --git | Y |
| `push` --binary | Y |
| `storage create` | Y |
| `storage list` | Y |
| `storage delete` | N |
| `update`: `--git` to `--local` | N |
| `update`: `--local` to `--git` | N |
| `update`: `--binary` to `--git` | N |
| `update`: `--git` to `--binary` | N |
| `update`: `--local` to `--binary` | N |
| `update`: `--binary` to `--local` | N |
| `url create` | Y |
| `url list` | Y |
| `url delete` | N |
| `watch` | N |
