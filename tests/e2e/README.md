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
| `app describe` | Y |
| `app list` | Y |
| `catalog list` | Y |
| `catalog search` | N |
| `component get` | Y |
| `component set` | Y |
| `create` | Y |
| `create --git` | Y |
| `create --local` | Y |
| `create --binary` | Y |
| `delete` | Y |
| `describe` | Y |
| `link` | N |
| `list` | Y |
| `push` --local | Y |
| `push` --binary | Y |
| `storage create` | Y |
| `storage list` | Y |
| `storage delete` | Y |
| `storage mount` | Y |
| `storage unmount` | Y |
| `update`: `--git` to `--local` | Y |
| `update`: `--local` to `--git` | Y |
| `update`: `--binary` to `--git` | Y |
| `update`: `--git` to `--binary` | Y |
| `update`: `--local` to `--binary` | Y |
| `update`: `--binary` to `--local` | Y |
| `update`: `--binary` to `--binary` | Y |
| `update`: `--local` to `--local` | Y |
| `update`: `--git` to `--git` | Y |
| `url create` | Y |
| `url list` | Y |
| `url delete` | Y |
| `watch` | N |
