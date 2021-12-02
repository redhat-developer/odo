---
title:  Writing and running tests
sidebar_position: 5
---

### Setting up test environment

Requires *Go 1.16* and *Ginkgo latest version*.

Testing happens with the above version. Developers are advised to stick to this version if they can but there is no compulsion on Go version.

We use unit, integration and e2e (End to end) tests.
Run `make goget-tools` target to set up the integration test environment. Unit tests do not require any precondition.   

#### Test variables:

There are some test environment variable that helps to get more control over the test run and it's results

- TEST_EXEC_NODES: Env variable TEST_EXEC_NODES is used to pass spec execution type (parallel or sequential) for ginkgo tests. To run the specs sequentially use TEST_EXEC_NODES=1, otherwise by default the specs are run in parallel on 4 ginkgo test node. Any TEST_EXEC_NODES value greater than one runs the spec in parallel on the same number of ginkgo test nodes.

- SLOW_SPEC_THRESHOLD: Env variable SLOW_SPEC_THRESHOLD is used for ginkgo tests. After this time (in second), ginkgo marks test as slow. The default value is set to 120s.

- GINKGO_TEST_ARGS: Env variable GINKGO_TEST_ARGS is used to get control over enabling test flags against each test target run. For example, To enable verbosity export or set env GINKGO_TEST_ARGS like `GINKGO_TEST_ARGS=-v`.

- UNIT_TEST_ARGS: Env variable UNIT_TEST_ARGS is used to get control over enabling test flags along with go test. For example, To enable verbosity export or set env UNIT_TEST_ARGS like `UNIT_TEST_ARGS=-v`.


#### Setting up test environment for integration and e2e tests
- **OpenShift:**
	To run the tests on a 4.x cluster, run `make configure-installer-tests-cluster` which performs the login operation required to run the test. By default, the tests are run against the `odo` binary placed in the `$PATH` which is created by the command `make`.
	
	Make sure that `odo` and `oc` binaries are in `$PATH`. Use the cloned odo directory to launch tests on 4.* clusters.

- **Kubernetes:**
	To run the tests on Kubernetes cluster, set the `KUBERNETES` environment variable:
	```shell
	export KUBERNETES=true
	```
	To communicate with `Kubernetes` cluster use `kubectl`.

Similarly, a 4.x cluster needs to be configured before launching the tests against it. The files `kubeadmin-password` and `kubeconfig` which contain cluster login details should be present in the `auth` directory, and it should reside in the same directory as `Makefile`. If it is not present in the auth directory, please create it, then run `make configure-installer-tests-cluster` to configure the 4.* cluster.

For **ppc64le arch**, run `make configure-installer-tests-cluster-ppc64le` to configure the test environment.

For **s390x arch**, run `make configure-installer-tests-cluster-s390x` to configure the test environment.

### Unit tests

Unit tests for `odo` functions are written using package [fake](https://godoc.org/k8s.io/client-go/kubernetes/fake). This allows us to create a fake client, and then mock the API calls defined under [OpenShift client-go](https://github.com/openshift/client-go) and [k8s client-go](https://godoc.org/k8s.io/client-go).

The tests are written in golang using the [pkg/testing](https://golang.org/pkg/testing/) package.
Run `make test` to validate unit tests.

### Integration tests

Integration tests utilize [Ginkgo](https://github.com/onsi/ginkgo) and its preferred matcher library [Gomega](https://github.com/onsi/gomega) which define sets of test cases (spec). As per ginkgo test file comprises specs and these test file are controlled by test suite. 

Test and test suite files are located in `tests/integration` directory and can be called using `make test-integration`. 

Integration tests validate and focus on specific fields of odo functionality or individual commands. For example, `cmd_app_test.go` or `generic_test.go`.

By default, the [integration tests](https://github.com/redhat-developer/odo/tree/main/tests/integration/devfile) for the devfile feature run against a `kubernetes` cluster.

#### Running integration tests
Integration tests can be run in two ways, parallel and sequential. By default, the test will run in parallel on 4 ginkgo test node.
- **Parallel Run:** To run the component command integration tests in parallel, on a test cluster:
  ```shell
  make test-cmp-e2e
  ```
  To control the parallel run, use the environment variable `TEST_EXEC_NODES`.

- **Sequential Run:** To run the component command integration tests sequentially or on single ginkgo test node:
  ```shell
  TEST_EXEC_NODES=1 make test-cmd-cmp
  ```
  `make test-cmd-login-logout` doesn't honour environment variable `TEST_EXEC_NODES`. By default, login and logout command integration test suites are run on a single ginkgo test node sequentially to avoid race conditions during a parallel run.

To see the number of available integration test files for validation, press `tab` just after writing `make test-cmd-`. However, there is a test file `generic_test.go` which handles certain test specs easily, and we can run it in parallel by calling `make test-generic`. By calling `make test-integration`, the whole suite will run all the specs in parallel on 4 ginkgo test nodes except `service` and `link`.

To run ONE individual test, you can either:
* Supply the name via command-line:
  ```shell
  ginkgo -focus="When executing catalog list without component directory" tests/integration/
  ```
* Modify the `It` statement to `FIt` and run:
  ```shell
  ginkgo tests/integration/
  ```

If you are running `operatorhub` tests, then you need to install certain operators on the cluster, which can be installed by running [setup-operator.sh](https://github.com/redhat-developer/odo/blob/main/scripts/configure-cluster/common/setup-operators.sh).

### E2e tests

E2e (End to end) uses the same library as integration test. E2e tests and test suite files are located in `tests/e2escenarios` directory and can be called using `.PHONY` within `makefile`. Basically end to end (e2e) test contains user specific scenario that is combination of some features/commands in a single test file.

#### Running E2e tests:

End-to-end(E2e) test run behaves in the similar way like integration test does. To see the number of available e2e test file for execution, press `tab` just after writing `make test-e2e-`. For e2e suite level execution of all e2e test spec use `make test-e2e-all`.

### Writing Tests

Refer to the odo clean test [template](https://github.com/redhat-developer/odo/blob/main/tests/template/template_cleantest_test.go).

#### Test guidelines:
[//]: # (TODO: Writing unit tests using the fake Kubernetes client)

Please follow certain protocol before contributing to odo tests. This helps in contributing to [odo tests](https://github.com/redhat-developer/odo/tree/main/tests). For better understanding of writing test please refer [Ginkgo](https://onsi.github.io/ginkgo/#getting-ginkgo) and it's preferred matcher library [Gomega](http://onsi.github.io/gomega/).

- Before writing tests (Integration/e2e) scenario make sure that the test scenario (Integration or e2e) is identified properly.

  **For example:** For storage feature, storage command will be tested properly includes positive, negative and corner cases whereas in e2e scenario only one or two storage command will be tested in e2e scenario like: _create component -> link -> add storage -> certain operation -> delete storage -> unlink -> delete component_.

- Create a new test file for a new feature and make sure that the feature file name should add proper sense. If the feature test file is already present then update the same test file with new scenario.
	
	**For example:** For storage feature, a new storage test file is created. If a new functionality is added to the storage feature then same file will be updated with new scenario. Naming of the test file should follow a common format like `cmd_<feature name>_test`. So the storage feature test file name will be `cmd_storage_test.go`. Same naming convention can be used for e2e test like `e2e_<release name>_test` or `e2e_<full scenario name>_test`.


- Test description should make sense of what it implements in the specs. Use proper test description in `Describe` block+

  **For example:** For storage feature, the appropriate test description would be `odo storage command tests`.
  ```go
  var _ = Describe("odo storage command tests", func() {
      [...]
  })
  ```

- For a better understanding of what a spec does, use proper description in `Context` and `it` block

  **For example:**
  ```go
  Context("when running help for storage command", func() {
    It("should display the help", func() {
      [...]
    })
  })
  ```

- Do not create a new test spec for the steps which can be run with the existing specs.

- Spec level conditions, pre, and post requirements should be run in ginkgo built-in tear down steps `JustBeforeEach` and `JustAfterEach`

- Due to parallel test run support make sure that the spec should run in isolation, otherwise the test result will lead to race condition. To achieve this ginkgo provides some in build functions `BeforeEach`, `AfterEach` etc.

  **For example:**
  ```go
  var _ = Describe("odo generic", func() {
    var project string
    var context string
    var oc helper.OcRunner
      BeforeEach(func() {
        oc = helper.NewOcRunner("oc")
        SetDefaultEventuallyTimeout(10 * time.Minute)
        context = helper.CreateNewContext()
      })
      AfterEach(func() {
        os.RemoveAll(context)
      })
      Context("deploying a component with a specific image name", func() {
          JustBeforeEach(func() {
              os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "config.yaml"))
              project = helper.CreateRandProject()
          })
	
          JustAfterEach(func() {
              helper.DeleteProject(project)
              os.Unsetenv("GLOBALODOCONFIG")
          })
          It("should deploy the component", func() {
              helper.CopyExample(filepath.Join("source", "nodejs"), context)
              helper.Cmd("odo", "create", "nodejs:latest", "testversioncmp", "--project", project, "--context", context).ShouldPass()
              helper.Cmd("odo", "push", "--context", context).ShouldPass()
              helper.Cmd("odo", "delete", "-f", "--context", context).ShouldPass()
          })
      })
  })
  ```

- Don’t create new test file for issues(bug) and try to add some scenario for each bug fix if applicable

- Don’t use unnecessary text validation in `Expect` of certain command output. Only validation of key text specific to that scenario would be enough.

  **For example:** While running multiple push on same component without changing any source file.
	
  ```go
  helper.Cmd("odo", "push", "--show-log", "--context", context+"/nodejs-ex")
  output := helper.Cmd("odo", "push", "--show-log", "--context", context+"/nodejs-ex").ShouldPass().Out()
  Expect(output).To(ContainSubstring("No file changes detected, skipping build"))
  ```

- If oc, odo or generic library you are looking for is not present in helper package then create a new library function as per the scenario requirement. Avoid unnecessary function implementation within test files. Check to see if there is a helper function already implemented.

- If you are looking for delay with a specific feature test, don't use hard time.Sleep() function. Yes, you can use but as a polling interval of maximum duration. Check the [helper package](https://github.com/redhat-developer/odo/tree/main/tests/helper) for more such reference.

  **For example:**
  ```go
  func RetryInterval(maxRetry, intervalSeconds int, program string, args ...string) string {
    for i := 0; i < maxRetry; i++ {
      session := CmdRunner(program, args...)
      session.Wait()
      if session.ExitCode() == 0 {
        time.Sleep(time.Duration(intervalSeconds) * time.Second)
      } else {
        Consistently(session).ShouldNot(gexec.Exit(0), runningCmd(session.Command))
        return string(session.Err.Contents())
      }
    }
    Fail(fmt.Sprintf("Failed after %d retries", maxRetry))
    return ""
  }
  ```

  There is also an in-built [timeout feature](http://onsi.github.io/ginkgo/#asynchronous-tests) available in Ginkgo.

- The test spec should run in parallel (Default) or sequentially as per choice. Check test template for reference.

- Run tests on local environment before pushing PRs.
