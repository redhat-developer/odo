# Development Guide

- [Workflow](https://github.com/redhat-developer/odo/blob/master/docs/development.md#workflow)
- [Unit Tests](https://github.com/redhat-developer/odo/blob/master/docs/development.md#unit-tests)
- [Integration Tests](https://github.com/redhat-developer/odo/blob/master/docs/development.md#integration-tests)
- [Dependency Management](https://github.com/redhat-developer/odo/blob/master/docs/development.md#dependency-management)
- [Release Guide](https://github.com/redhat-developer/odo/blob/master/docs/development.md#release-guide)
- [Odo Bot](https://github.com/redhat-developer/odo/blob/master/docs/development.md#odo-bot)
- [Licenses](https://github.com/redhat-developer/odo/blob/master/docs/development.md#licenses)

## Workflow

### Fork the main repository

1. Go to https://github.com/redhat-developer/odo
2. Click the "Fork" button (at the top right)

### Clone your fork

The following commands assume that you have the $GOPATH environment variable properly set. We highly recommend you place odo code into $GOPATH.

```sh
git clone https://github.com/$YOUR_GITHUB_USERNAME/odo.git $GOPATH/src/github.com/redhat-developer/odo
cd $GOPATH/src/github.com/redhat-developer/odo
git remote add upstream 'https://github.com/redhat-developer/odo'
```

While cloning Odo, the Windows terminal such as PowerShell or CMD may throw an error of `Filename too long`. To avoid such an error, set your Git config as so:

```sh
git config --system core.longpaths true
```

### Create a branch and make changes

```sh
git checkout -b myfeature
# Make your code changes
```

### Keeping your development fork in sync

```sh
git fetch upstream
git rebase upstream/master
```

**Note for maintainers**: If you have write access to the main repository at github.com/redhat-developer/odo, you should modify your git configuration so that you can't accidentally push to upstream:

```sh
git remote set-url --push upstream no_push
```

### Pushing changes to your fork

```sh
git commit
git push -f origin myfeature
```

### Creating a pull request

1. Visit https://github.com/$YOUR_GITHUB_USERNAME/odo.git
2. Click the "Compare and pull request" button next to your "myfeature" branch.
3. Check out the pull request process for more details

### Requirements for a pull request

A pull request should include:

  - Descriptive context that outlines what has been changed and why
  - A link to an active / open issue (if applicable)

For example:

  ```
  # X feature added
  X is a new feature that has been added to Odo to fix X issue
  ...

  X is used like so:
  ...

  Closes issue X.
  ```

Terminology we use:

  - *WIP (Work in Progress):* If your PR is still in-progress, indicate this with a label or add WIP in your PR title

### Reviewing a pull request

What to look out for when reviewing a pull request:

  - Have tests been added?
  - Does this feature / fix work locally for me? 
  - Am I able to understand the code correctly / have comments been added to the code?

### Test Driven Development

We follow Test Driven Development(TDD) workflow in our development process. You can read more about it [here](/docs/tdd-workflow.md).

## Unit Tests

### Introduction

Unit-tests for Odo functions are written using package [fake](https://godoc.org/k8s.io/client-go/kubernetes/fake). This allows us to create a fake client, and then mock the API calls defined under [OpenShift client-go](https://github.com/openshift/client-go) and [k8s client-go](https://godoc.org/k8s.io/client-go).

The tests are written in golang using the [pkg/testing](https://golang.org/pkg/testing/) package.

### Writing unit tests

 1. Identify the APIs used by the function to be tested.

 2. Initialise the fake client along with the relevant clientsets.

 3. In the case of functions fetching or creating new objects through the APIs, add a [reactor](https://godoc.org/k8s.io/client-go/testing#Fake.AddReactor) interface returning fake objects. 

 4. Verify the objects returned

##### Initialising fake client and creating fake objects

Let us understand the initialisation of fake clients and therefore the creation of fake objects with an example.

The function `GetImageStreams` in [pkg/occlient.go](https://github.com/redhat-developer/odo/blob/master/pkg/occlient/occlient.go) fetches imagestream objects through the API:

```go
func (c *Client) GetImageStreams(namespace string) ([]imagev1.ImageStream, error) {
        imageStreamList, err := c.imageClient.ImageStreams(namespace).List(metav1.ListOptions{})
        if err != nil {
                return nil, errors.Wrap(err, "unable to list imagestreams")
        }
        return imageStreamList.Items, nil
}

```

1. For writing the tests, we start by initialising the fake client using the function `FakeNew()` which initialises the image clientset harnessed by 	`GetImageStreams` funtion:

    ```go
    client, fkclientset := FakeNew()
    ```

2. In the `GetImageStreams` funtions, the list of imagestreams is fetched through the API. While using fake client, this list can be emulated using a [`PrependReactor`](https://github.com/kubernetes/client-go/blob/master/testing/fake.go) interface:
 
   ```go
	fkclientset.ImageClientset.PrependReactor("list", "imagestreams", func(action ktesting.Action) (bool, runtime.Object, error) {
        	return true, fakeImageStreams(tt.args.name, tt.args.namespace), nil
        })
   ```

   The `PrependReactor` expects `resource` and `verb` to be passed in as arguments. We can get this information by looking at the [`List` function for fake imagestream](https://github.com/openshift/client-go/blob/master/image/clientset/versioned/typed/image/v1/fake/fake_imagestream.go):


   	```go
    func (c *FakeImageStreams) List(opts v1.ListOptions) (result *image_v1.ImageStreamList, err error) {
        	obj, err := c.Fake.Invokes(testing.NewListAction(imagestreamsResource, imagestreamsKind, c.ns, opts), &image_v1.ImageStreamList{})
		...
    }
        
    func NewListAction(resource schema.GroupVersionResource, kind schema.GroupVersionKind, namespace string, opts interface{}) ListActionImpl {
        	action := ListActionImpl{}
        	action.Verb = "list"
        	action.Resource = resource
        	action.Kind = kind
        	action.Namespace = namespace
        	labelSelector, fieldSelector, _ := ExtractFromListOptions(opts)
        	action.ListRestrictions = ListRestrictions{labelSelector, fieldSelector}

        	return action
    }
    ```


  The `List` function internally calls `NewListAction` defined in [k8s.io/client-go/testing/actions.go](https://github.com/kubernetes/client-go/blob/master/testing/actions.go).  From these functions, we see that the `resource` and `verb`to be passed into the `PrependReactor` interface are `imagestreams` and `list` respectively. 


  You can see the entire test function `TestGetImageStream` in [pkg/occlient/occlient_test.go](https://github.com/redhat-developer/odo/blob/master/pkg/occlient/occlient_test.go)

**NOTE**: You can use environment variable CUSTOM_HOMEDIR to specify a custom home directory. It can be used in environments where a user and home directory are not resolveable.

## Integration tests

Integration tests, otherwise known as end-2-end (e2e) tests are used within Odo.

All tests can be found in the `tests/e2e` directory and can be called using functions within `makefile`.

Requirements:

 - A `minishift` or OpenShift environment with Service Catalog enabled

```sh
$ MINISHIFT_ENABLE_EXPERIMENTAL=y minishift start --extra-clusterup-flags "--enable=*,service-catalog,automation-service-broker,template-service-broker"
```

 - `odo` and `oc` binaries in $PATH

To deploy an e2e test:

```sh
# The entire suite
make test-e2e

# Just the main tests
make test-main-e2e

# Just component tests
make test-cmp-e2e

# Just service catalog tests
make test-service-e2e
```

Running a subset of tests is possible with ginkgo by using focused specs mechanism
https://onsi.github.io/ginkgo/#focused-specs

### Race conditions

It is not uncommon that during the execution of the integration tests, test failures occur.
Although it's possible that this is due to the tests themselves, more often than not this occurs due to the the actual odo code is written.
For example, the following error has been encountered multiple times:

```
Operation cannot be fulfilled on deploymentconfigs.apps.openshift.io "component-app": the object has been modified; please apply your changes to the latest version and try again
```

The reason this happens is because the `read DeploymentCondif` / `update DC in memory` / `call Update` can potentially fail to due 
the DC being updated concurrently by some other component (usually by Kubernetes/Openshift itself)

For such case it's recommended to avoid the read/update-in-memory-/push-update as much as possible.
One remedy is to use the Patch operation (see `Resource Operations` section from [this](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.11/) docs page)
Another remedy would be to retry the operation when the optimistic concurrency error is encountered  

## Dependency Management

odo uses `glide` to manage dependencies.

They are not strictly required for building odo but they are required when managing dependencies under the `vendor/` directory.

If you want to make changes to dependencies please make sure that `glide` is installed and are in your `$PATH`.

### Installing glide

Get `glide`:

```sh
go get -u github.com/Masterminds/glide
```

Check that `glide` is working

```sh
glide --version
```

### Using glide to add a new dependency

#### Adding new dependency

1. Update `glide.yaml` file. Add new packages or subpackages to `glide.yaml` depending if you added whole new package as dependency or just new subpackage.

2. Run `glide update --strip-vendor` to get new dependencies

3. Commit updated `glide.yaml`, `glide.lock` and `vendor` to git.


#### Updating dependencies

1. Set new package version in  `glide.yaml` file.

2. Run `glide update --strip-vendor` to update dependencies

## Release Guide

### Making a release

Making artifacts for new release is automated. 

When new git tag is created, Travis-ci deploy job automatically builds binaries and uploads it to GitHub release page.

1. Create PR with updated version in following files:

    - [cmd/version.go](/cmd/version.go)
    - [scripts/install.sh](/scripts/install.sh)
    - [README.md](/README.md)

    There is a helper script [scripts/bump-version.sh](/scripts/bump-version.sh) that should change version number in all files listed above (expect odo.rb).

    To update the CLI reference documentation in docs/cli-reference.md, run `make generate-cli-reference`, which will update `docs/cli-reference.md`.

2. Merge the above PR

3. Once the PR is merged create and push new git tag for version.
    ```
    git tag v0.0.1
    git push upstream v0.0.1
    ```
    **Or** create the new release using GitHub site (this has to be a proper release, not just draft). 

    Do not upload any binaries for release

    When new tag is created Travis-CI starts a special deploy job.

    This job builds binaries automatically (via `make prepare-release`) and then uploads it to GitHub release page (done using odo-bot user).

4. When a job finishes you should see binaries on the GitHub release page. Release is now marked as a draft. Update descriptions and publish release.

5. Verify that packages have been uploaded to rpm and deb repositories.

6. We must now update the Homebrew package. Download the current release `.tar.gz` file and retrieve the sha256 value.

    ```sh
    RELEASE=X.X.X
    wget https://github.com/redhat-developer/odo/archive/v$RELEASE.tar.gz
    sha256sum v$RELEASE.tar.gz
    ```

    Then open a PR to update: [odo.rb](https://github.com/kadel/homebrew-odo/blob/master/Formula/odo.rb) in [kadel/homebrew-odo](https://github.com/kadel/homebrew-odo)

7. Confirm the binaries are available in GitHub release page.

8. Create a PR and update the file `build/VERSION` with latest version number.

## Odo Bot

[odo-bot](https://github.com/odo-bot) is the GitHub user that provides automation for certain tasks of Odo.

### Scripts using odo-bot

| Script      | What it is doing                          | Access via                                    |
|-------------|-------------------------------------------|-----------------------------------------------|
| .travis.yml | Uploading binaries to GitHub release page | Personal access token `deploy-github-release` |


## Licenses

[wwhrd](https://github.com/frapposelli/wwhrd) is used in Odo for checking license
compatibilities of vendored packages.

Configuration for `wwhrd` is stored in
[`.wwhrd.yml`](https://github.com/redhat-developer/odo/blob/master/.wwhrd.yml).

The `whitelist` section is for licenses that are always allowed.
The `blacklist` section is for licenses that are never allowed and will
always fail a build. Any licenses that are not explicitly mentioned are considered
to be in a `exceptions` and will need to be explicitly allowed by adding the import
path to the exceptions.

More details about the license compatibility check tool can be found
[here](https://github.com/frapposelli/wwhrd)
