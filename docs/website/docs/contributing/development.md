---
title:  Development Guide
sidebar_position: 3
---

### Setting up

Requires *Go 1.16*

Testing and release builds happen with the above version. Developers are advised to stick to this version if they can, but it is not compulsory.


**Warning:**
If you are adding any features that require a higher version of golang, than the one mentioned above, please contact the maintainers in order to check if the releasing systems can handle the newer version. If that is ok, please ensure you update the required golang version, both here and in the files below, in your PR.

List of files to update for golang version:
* [scripts/rpm-prepare.sh](https://github.com/openshift/odo/blob/main/scripts/rpm-prepare.sh)
* [Dockerfile.rhel](https://github.com/openshift/odo/blob/main/Dockerfile.rhel)
* [openshift-ci/build-root/Dockerfile](https://github.com/openshift/odo/blob/main/openshift-ci/build-root/Dockerfile)

First setup your fork of the odo project, following the steps below

1. [Fork](https://help.github.com/en/articles/fork-a-repo) the [odo](https://github.com/openshift/odo) repository.

2. Clone your fork:
  NOTE: odo uses `go modules` to manage dependencies which means you can clone the code anywhere you like but for backwards compatibility
  we would be cloning it under `$GOPATH`
  
  ```shell
  git clone https://github.com/<YOUR_GITHUB_USERNAME>/odo.git $GOPATH/src/github.com/openshift/odo
  cd $GOPATH/src/github.com/openshift/odo
  git remote add upstream 'https://github.com/openshift/odo'
  ```
  When cloning `odo`, the Windows terminal such as PowerShell or CMD may throw a *Filename too long* error. To avoid such an error, set your Git configuration as follows:
  
  ```shell
  git config --system core.longpaths true
  ```
  If you are a maintainer and have write access to the `odo` repository, modify your git configuration so that you do not accidentally push to upstream:
  
  ```shell
  git remote set-url --push upstream no_push
  ```

3. Install tools used by the build and test system:
  ```shell
  make goget-tools
  ```

### Useful make targets

1. bin:: (default) `go build` the executable in cmd/odo
2. install:: build and install `odo` in your GOPATH
3. validate:: run gofmt, go vet and other validity checks
4. goget-tools:: download tools used to build & test
5. test:: run all unit tests - same as `go test pkg/\...`
6. test-integration:: run all integration tests
7. test-coverage:: generate test coverage report

Read the [Makefile](https://github.com/openshift/odo/blob/main/Makefile) itself for more information.

### Submitting a pull request(PR)
To submit a PR, you must first create a branch from your fork, commits your changes to the branch, and push them on to GitHub.
A "signed-off" signature is good practice. You may sign your commit using `git commit -s` or `git commit --amend --no-edit -s` to a previously created commit

Refer to the guidelines below, and create a PR with your changes.
1. Descriptive context that outlines what has been changed and why
2. If your PR is still in-progress, indicate this with a label or add WIP in your PR title.
3. A link to the active or open issue it fixes (if applicable)

Once you submit a PR, the @openshift-ci-bot will automatically request two reviews from a reviewer and an approver.

### Setting custom Init Container image for bootstrapping Supervisord
For quick deployment of components, odo uses the [Supervisord](https://github.com/ochinchina/supervisord) process manager.
Supervisord is deployed via [Init Container](https://docs.openshift.com/container-platform/4.1/nodes/containers/nodes-containers-init.html) image. 

`ODO_BOOTSTRAPPER_IMAGE` is an environmental variable which specifies the Init Container image used for Supervisord deployment.  You can modify the value of the variable to use a custom Init Container image.
The default Init Container image is `quay.io/openshiftdo/init` 

To set a custom Init Container image, run:
```shell
export ODO_BOOTSTRAPPER_IMAGE=quay.io/myrepo/myimage:test
```

To revert back to the default Init Container image, unset the variable:

```shell
unset ODO_BOOTSTRAPPER_IMAGE
```

### Dependency management
`odo` uses `go mod` to manage dependencies, and with vendor directory. This means that you should use `-mod=vendor` flag with all `go` commands. Or use `GOFLAGS` to set it permanently (`export GOFLAGS=-mod=vendor`).
Vendor is important to make sure that odo can always be built even offline.


#### Adding a new dependency

1. Just add new `import` to your code.
NOTE:  If you want to use a specific version of a module you can do `go get <pkg>@<version>`, for example (`go get golang.org/x/text@v0.3.2`)

1. Run `go mod tidy` and `go mod vendor`.
2. Commit the updated `go.mod`, `go.sum` and `vendor` files to git.

### Writing machine-readable output code

Some tips to consider when writing machine-readable output code.
- Match similar Kubernetes / OpenShift API structures
- Put as much information as possible within `Spec`
- Use `json:"foobar"` within structs to rename the variables 


Within odo, we unmarshal all information from a struct to json. Within this struct, we use `TypeMeta` and `ObjectMeta` in order to supply meta-data information coming from Kubernetes / OpenShift. 

Below is working example of how we would implement a "GenericSuccess" struct.

```go
package main

import (
  "encoding/json"
  "fmt"

  metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Create the struct. Here we use TypeMeta and ObjectMeta
// as required to create a "Kubernetes-like" API.
type GenericSuccess struct {
  metav1.TypeMeta   `json:",inline"`
  metav1.ObjectMeta `json:"metadata,omitempty"`
  Message           string `json:"message"`
}

func main() {

  // Create the actual struct that we will use
  // you will see that we supply a "Kind" and
  // APIVersion. Name your "Kind" to what you are implementing
  machineOutput := GenericSuccess{
    TypeMeta: metav1.TypeMeta{
      Kind:       "genericsuccess",
      APIVersion: "odo.dev/v1alpha1",
    }, 
    ObjectMeta: metav1.ObjectMeta{
      Name: "MyProject",
    }, 
    Message: "Hello API!",
  }

  // We then marshal the output and print it out
  printableOutput, _ := json.Marshal(machineOutput)
  fmt.Println(string(printableOutput))
}
```

### odo-bot

[odo-bot](https://github.com/odo-bot) is the GitHub user that provides automation for certain tasks in odo.

It uses the `.travis.yml` script to upload binaries to the GitHub release page using the *deploy-github-release*
personal access token.

### Licenses

odo uses [wwhrd](https://github.com/frapposelli/wwhrd) to  check license compatibility of vendor packages. The configuration for `wwhrd` is stored in [.wwhrd.yml](https://github.com/openshift/odo/blob/main/.wwhrd.yml).

The `whitelist` section is for licenses that are always allowed. The `blacklist` section is for licenses that are never allowed and will always fail a build. Any licenses that are not explicitly mentioned come under the `exceptions` section and need to be explicitly allowed by adding the import path to the exceptions.

More details about the license compatibility check tool can be found [here](https://github.com/frapposelli/wwhrd).
