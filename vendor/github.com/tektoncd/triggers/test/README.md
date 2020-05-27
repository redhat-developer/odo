# Tests

To run tests:

```shell
# Land the latest code
ko apply -f ./config/

# Run unit tests
go test ./...

# Run integration tests (against your current kube cluster)
go test -v -count=1 -tags=e2e -timeout=20m ./test
```

## Unit tests

Unit tests live side by side with the code they are testing and can be run with:

```shell
go test ./...
```

By default `go test` will not run [the end to end tests](#end-to-end-tests),
which need `-tags=e2e` to be enabled.

## End to end tests

### Setup

Environment variables used by end to end tests:

- `KO_DOCKER_REPO` - Set this to an image registry your tests can push images to

### Running

End to end tests live in this directory. To run these tests, you must provide
`go` with `-tags=e2e`. By default the tests run against your current kubeconfig
context, but you can change that and other settings with [the flags](#flags).
Run e2e tests with:

```shell
go test -v -count=1 -tags=e2e -timeout=20m ./test
go test -v -count=1 -tags=e2e -timeout=20m ./test --kubeconfig ~/special/kubeconfig --cluster myspecialcluster
```

You can also use
[all flags defined in `knative/pkg/test`](https://github.com/knative/pkg/tree/master/test#flags).

### Flags

- By default the e2e tests run against the current cluster in `~/.kube/config`
  using the environment specified in
  [your environment variables](/DEVELOPMENT.md#environment-setup).
- Since these tests are fairly slow, running them with logging enabled is
  recommended (`-v`).
- Using [`--logverbose`](#output-verbose-log) will show the verbose log output
  from test as well as from k8s libraries.
- Using `-count=1` is
  [the idiomatic way to disable test caching](https://golang.org/doc/go1.10#test).
- The e2e tests take a long time to run, so a value like `-timeout=20m` can be
  useful depending on what you're running.

You can [use test flags](#flags) to control the environment your tests run
against, i.e. override
[your environment variables](/DEVELOPMENT.md#environment-setup):

```bash
go test -v -tags=e2e -count=1 ./test --kubeconfig ~/special/kubeconfig --cluster myspecialcluster
```

Tests importing [`github.com/tektoncd/triggers/test`](#adding-integration-tests)
recognize the
[flags added by `knative/pkg/test`](https://github.com/knative/pkg/tree/master/test#flags).

Tests are run in a new random namespace prefixed with the word `arakkis-`.
Unless you set the `TEST_KEEP_NAMESPACES` environment variable the namespace
will get automatically cleaned up after running each test.

### Running specific test cases

To run all the test cases with their names starting with the same letters, e.g.
EventListener, use
[the `-run` flag with `go test`](https://golang.org/cmd/go/#hdr-Testing_flags):

```bash
go test -v -tags=e2e -count=1 ./test -run ^TestEventListener
```

### Running YAML tests

To run the YAML e2e tests, run the following command:

```bash
./test/e2e-tests-yaml.sh
```

### Adding integration tests

In the [`test`](/test/) dir you will find several libraries in the `test`
package you can use in your tests.

This library exists partially in this directory and partially in
[`knative/pkg/test`](https://github.com/knative/pkg/tree/master/test).

The libs in this dir can:

- [Setup tests](#setup-tests)
- [Poll resources](#poll-resources)
- [Generate random names](#generate-random-names)

All integration tests _must_ be marked with the `e2e`
[build constraint](https://golang.org/pkg/go/build/) so that `go test ./...` can
be used to run only [the unit tests](#unit-tests), i.e.:

```go
// +build e2e
```

#### Cleaning up cluster-scoped resources

Each integration test runs in its own Namespace; each Namespace is torn down
after its integration test completes. However, cluster-scoped resources will not
be deleted when the Namespace is deleted. So, each test must delete all the
cluster-scoped resources that it creates.

#### Setup tests

The `setup` function in [init_tests.go](./init_test.go) will initialize client
objects, create a new unique Namespace for the test, and initialize anything
needed globally by the tests (i.e. logs and metrics).

```go
clients, namespace := setup(t)
```

The `clients` struct contains initialized clients for accessing:

- Kubernetes resources
- [Pipelines resources](https://github.com/tektoncd/pipeline)
- [Triggers resources](https://github.com/tektoncd/triggers)

_See [init_test.go](./init_test.go) and [clients.go](./clients.go) for more
information._

#### Poll resources

After creating, updating, or deleting kubernetes resources, you will need to
wait for the system to realize these changes. You can use polling methods to
check the resources reach the desired state.

The `WaitFor*` functions use the Kubernetes
[`wait` package](https://godoc.org/k8s.io/apimachinery/pkg/util/wait). For
polling they use
[`PollImmediate`](https://godoc.org/k8s.io/apimachinery/pkg/util/wait#PollImmediate)
with a
[`ConditionFunc`](https://godoc.org/k8s.io/apimachinery/pkg/util/wait#ConditionFunc)
callback function, which returns a `bool` to indicate if the polling should stop
and an `error` to indicate if there was an error.

_See [wait.go](./wait.go) for more information._

#### Generate random names

You can use the
[`names`](https://github.com/tektoncd/pipeline/tree/master/pkg/names) package
from the [`Tekton Pipeline`](https://github.com/tektoncd/pipeline) project to
append a random string, so that your tests can use unique names each time they
run.

```go
import "github.com/tektoncd/pipeline/pkg/names"

namespace := names.SimpleNameGenerator.RestrictLengthWithRandomSuffix("arrakis")
```

### Running presubmit integration tests

The presubmit integration tests entrypoint will run:

- [The e2e tests](#end-to-end-tests)

When run using Prow, integration tests will try to get a new cluster using
[boskos](https://github.com/kubernetes/test-infra/tree/master/boskos), which
only
[the `tektoncd/plumbing` OWNERS](https://github.com/tektoncd/plumbing/blob/master/OWNERS)
have access to.
