---
layout: default
permalink: /development/
redirect_from: 
  - /docs/development.md/
---

# Development Guide

* TOC
{:toc}

## Building Kedge

Read about building kedge [here](https://github.com/kedgeproject/kedge#building).

## Workflow
### Fork the main repository

1. Go to https://github.com/kedgeproject/kedge
2. Click the "Fork" button (at the top right)

### Clone your fork

The commands below require that you have $GOPATH. We highly recommended you put Kedge' code into your $GOPATH.

```console
git clone https://github.com/$YOUR_GITHUB_USERNAME/kedge.git $GOPATH/src/github.com/kedgeproject/kedge
cd $GOPATH/src/github.com/kedgeproject/kedge
git remote add upstream 'https://github.com/kedgeproject/kedge'
```

### Create a branch and make changes

```console
git checkout -b myfeature
# Make your code changes
```

### Keeping your development fork in sync

```console
git fetch upstream
git rebase upstream/master
```

Note: If you have write access to the main repository at github.com/kedgeproject/kedge, you should modify your git configuration so that you can't accidentally push to upstream:

```console
git remote set-url --push upstream no_push
```

### Committing changes to your fork

```console
git commit
git push -f origin myfeature
```

### Creating a pull request

1. Visit https://github.com/$YOUR_GITHUB_USERNAME/kedge.git
2. Click the "Compare and pull request" button next to your "myfeature" branch.
3. Check out the pull request process for more details

## Dependency management

Kedge uses `glide` to manage dependencies.
If you want to make changes to dependencies please make sure that `glide` is in your `$PATH`.

### Installing glide

There are many ways to build and host golang binaries. Here is an easy way to get utilities like `glide` and `glide-vc` installed:

Ensure that Mercurial and Git are installed on your system. (some of the dependencies use the mercurial source control system).
Use `apt-get install mercurial git` or `yum install mercurial git` on Linux, or `brew.sh` on OS X, or download them directly.

```console
go get -u github.com/Masterminds/glide
```

Check that `glide` is working.

```console
glide --version
```

### Using glide

#### Adding new dependency

1. Update `glide.yaml` file

  Add new packages or subpackages to `glide.yaml` depending if you added whole
  new package as dependency or just new subpackage.

  Kedge vendors OpenShift and all its dependencies. (see comments in `glide.yaml` and `./scripts/vendor-openshift.yaml`)
  It is possible that the dependency you want to add is already in OpenShift as a vendored dependency. If that is true, please make sure 
  that you use the same version OpenShift is using.

2. Get new dependencies

```bash
make vendor-update
```

3. Commit updated glide files and vendor

```bash
git add glide.yaml glide.lock vendor
git commit
```


#### Updating dependencies

1. Set new package version in  `glide.yaml` file.

2. Clear cache

```bash
glide cc
```
This step is necessary if not done glide will pick up old data from it's cache.

3. Get new and updated dependencies

```bash
make vendor-update
```


4. Commit updated glide files and vendor

```bash
git add glide.yaml glide.lock vendor
git commit
```

#### Updating OpenShift

1. Update OPENSHIFT_VERSION within [`scripts/vendor-openshift.sh`](https://github.com/kedgeproject/kedge/blob/c64b0fd7a69edc4db5ef9aab0c52c97a0c9cf10e/scripts/vendor-openshift.sh#L15)


2. Clear the cache to prevent glide from using previous data

```bash
glide cc
```

3. Retrieve new and updated dependencies

```bash
make vendor-update
```

4. Commit updated glide files and vendor

```bash
git add glide.yaml glide.lock vendor
git commit
```

### PR review guidelines

- To merge a PR at least two LGTMs are needed to merge it

- If a PR is opened for more than two weeks, find why it is open for so long
if it is blocked on some other issue/pr label it as blocked and then also link
the issue it is blocked on. If it is outstanding for review and there are no
reviews on it ping maintainers.

- For PRs that have more than 500 LOC break it into pieces and merge it one
by one incrementally so that it is easy to review and going back and forth on
it is easier.

**Note**: Above guidelines are not hard rules use those with discretion

### Running tests

#### Run all tests except end-to-end tests

```bash
make test
```

#### Run end-to-end tests

Before running end to end tests locally make sure [minikube](https://github.com/kubernetes/minikube/)
is running.

```bash
make bin
make test-e2e
```

**Note**: When you run end to end tests, those tests are run in parallel. If
you are low on resources you can limit number of tests that run in parallel by
doing following:

```bash
make test-e2e PARALLEL=4
```

This will run only 4 tests in parallel. By default, it is set to the value of
`GOMAXPROCS`.

You may also add a timeout which will increase the overall timeout period for the tests.

```bash
make test-e2e TIMEOUT=15m
```

### types.go conventions

- Add explanation on top of each struct and struct field in `types.go` to explain what it does,
so that when OpenAPI spec is auto-generated it will show up there.

- Structs that are referred in any other struct in the form of an array should have a comment
of the format `// kedgeSpec: io.kedge.*`, where `*` is name of that struct. This becomes the
identity or reference of that struct in OpenAPI specification.

- If you are embedding a struct, there is no need to add an explanatory comment.

- For any struct that is embedded please add a k8s tag comment:
`// k8s: io.k8s.kubernetes.pkg.api.v1.ServicePort`.

- For all the fields that are optional please include a comment:
`// +optional`.

- Any struct that is defined in same file and is used in another struct, while embedding
please add a ref tag comment:
`// ref: io.kedge.ContainerSpec`.

- To find out what is the key or k8s reference tag for a particular struct in Kubernetes,
please refer to the swagger specification of Kubernetes for any particular release. For e.g
In Kubernetes 1.7, the reference tag for deployment is
`io.k8s.kubernetes.pkg.apis.apps.v1beta1.DeploymentSpec`.

### Validation

In order to facilitate consistent error messages, we ask that validation logic
adheres to the following guidelines whenever possible (though exceptional cases will exist).

* Be as precise as possible.
* Telling users what they CAN do is more useful than telling them what they
CANNOT do.
* When asserting a requirement in the positive, use "must".  Examples: "must be
greater than 0", "must match regex '[a-z]+'".  Words like "should" imply that
the assertion is optional, and must be avoided.
* When asserting a formatting requirement in the negative, use "must not".
Example: "must not contain '..'".  Words like "should not" imply that the
assertion is optional, and must be avoided.
* When asserting a behavioral requirement in the negative, use "may not".
Examples: "may not be specified when otherField is empty", "only `name` may be
specified".
* When referencing a literal string value, indicate the literal in
single-quotes. Example: "must not contain '..'".
* When referencing another field name, indicate the name in back-quotes.
Example: "must be greater than `request`".
* When specifying inequalities, use words rather than symbols.  Examples: "must
be less than 256", "must be greater than or equal to 0".  Do not use words
like "larger than", "bigger than", "more than", "higher than", etc.
* When specifying numeric ranges, use inclusive ranges when possible.

Taken from: [github.com/kubernetes/community/contributors/devel/api-conventions.md](https://github.com/kubernetes/community/blob/2bfe095e4dcd02b4ccd3e21c1f30591ca57518a6/contributors/devel/api-conventions.md#validation)


### Naming conventions

* Go field names must be CamelCase. JSON field names must be camelCase. Other
than capitalization of the initial letter, the two should almost always match.
No underscores nor dashes in either.
* Field and resource names should be declarative, not imperative (DoSomething,
SomethingDoer, DoneBy, DoneAt).
* Use `Node` where referring to
the node resource in the context of the cluster. Use `Host` where referring to
properties of the individual physical/virtual system, such as `hostname`,
`hostPath`, `hostNetwork`, etc.
* `FooController` is a deprecated kind naming convention. Name the kind after
the thing being controlled instead (e.g., `Job` rather than `JobController`).
* The name of a field that specifies the time at which `something` occurs should
be called `somethingTime`. Do not use `stamp` (e.g., `creationTimestamp`).
* We use the `fooSeconds` convention for durations, as discussed in the [units
subsection](#units).
  * `fooPeriodSeconds` is preferred for periodic intervals and other waiting
periods (e.g., over `fooIntervalSeconds`).
  * `fooTimeoutSeconds` is preferred for inactivity/unresponsiveness deadlines.
  * `fooDeadlineSeconds` is preferred for activity completion deadlines.
* Do not use abbreviations in the API, except where they are extremely commonly
used, such as "id", "args", or "stdin".
* Acronyms should similarly only be used when extremely commonly known. All
letters in the acronym should have the same case, using the appropriate case for
the situation. For example, at the beginning of a field name, the acronym should
be all lowercase, such as "httpGet". Where used as a constant, all letters
should be uppercase, such as "TCP" or "UDP".
* The name of a field referring to another resource of kind `Foo` by name should
be called `fooName`. The name of a field referring to another resource of kind
`Foo` by ObjectReference (or subset thereof) should be called `fooRef`.
* More generally, include the units and/or type in the field name if they could
be ambiguous and they are not specified by the value or value type.
* The name of a field expressing a boolean property called 'fooable' should be
called `Fooable`, not `IsFooable`.

Taken from: [github.com/kubernetes/community/contributors/devel/api-conventions.md](https://github.com/kubernetes/community/blob/2bfe095e4dcd02b4ccd3e21c1f30591ca57518a6/contributors/devel/api-conventions.md#naming-conventions)

### Optional vs. Required

Fields must be either optional or required.

Optional fields have the following properties:

- They have the `+optional` comment tag in Go.
- They are a pointer type in the Go definition (e.g. `bool *awesomeFlag`) or
have a built-in `nil` value (e.g. maps and slices).

In most cases, optional fields should also have the `omitempty` struct tag (the 
`omitempty` option specifies that the field should be omitted from the json
encoding if the field has an empty value).


Required fields have the opposite properties, namely:

- They do not have an `+optional` comment tag.
- They do not have an `omitempty` struct tag.
- They are not a pointer type in the Go definition (e.g. `bool otherFlag`).

Using the `+optional` or the `omitempty` tag causes OpenAPI documentation to 
reflect that the field is optional.

Using a pointer allows distinguishing unset from the zero value for that type.
There are examples of this in the codebase. However:

- it can be difficult for implementors to anticipate all cases where an empty
value might need to be distinguished from a zero value
- having a pointer consistently imply optional is clearer

Therefore, we ask that pointers always be used with optional fields that do not
have a built-in `nil` value.

Inspired from: [github.com/kubernetes/community/contributors/devel/api-conventions.md](https://github.com/kubernetes/community/blob/2bfe095e4dcd02b4ccd3e21c1f30591ca57518a6/contributors/devel/api-conventions.md#optional-vs-required)

### General guidelines for developers

- When you add a new function/method

  - Add unit-tests

- When you add a new feature

  - Add an example in docs/example with its explanation README, for e.g. [health](https://github.com/kedgeproject/kedge/tree/master/docs/examples/health).
  - Add an e2e test on above example, for e.g. see test code for [health](https://github.com/kedgeproject/kedge/blob/cfee15ffde02c611d08420699a43869706be2d53/tests/e2e/e2e_test.go#L272).
  - Add this feature information to file-reference, for e.g. see [health section](https://github.com/kedgeproject/kedge/blob/master/docs/file-reference.md#health).

### golang dependency import conventions

Imports MUST be arranged in three sections, separated by an empty line.

```go
stdlib
kedge
thirdparty
```

For example:

```go
"fmt"
"io"
"os/exec"

pkgcmd "github.com/kedgeproject/kedge/pkg/cmd"
"github.com/kedgeproject/kedge/pkg/spec"

"github.com/ghodss/yaml"
"github.com/pkg/errors"
```

Once arranged, let `gofmt` sort the sequence of imports.


## Code instructions

### Things to take care of when adding new type:

- Creating type struct

  - Make a struct type named `ControllerSpecMod` and add it to [types.go](https://github.com/kedgeproject/kedge/blob/cd5297cfdbd2f1daa510824d49c9d3d649ffc0b8/pkg/spec/types.go).
  - Embed this struct with `ControllerSpec` from upstream and [`ControllerFields`](https://github.com/kedgeproject/kedge/blob/cd5297cfdbd2f1daa510824d49c9d3d649ffc0b8/pkg/spec/types.go#L146).
  - Make sure all the comments are placed well and all optional fields are maked as
  [optional](https://github.com/kedgeproject/kedge/blob/cd5297cfdbd2f1daa510824d49c9d3d649ffc0b8/pkg/spec/types.go#L154).
  - For more info about how to add fields read [here](https://github.com/kedgeproject/kedge/blob/3640c31ea44c2aa06e59e127b291bf4e1d49a6b4/docs/development.md#typesgo-conventions).

- Make sure it satisfies the interface [`ControllerInterface`](https://github.com/kedgeproject/kedge/blob/cd5297cfdbd2f1daa510824d49c9d3d649ffc0b8/pkg/spec/controller.go#L29)

  - Define all these methods in interface on the new `ControllerSpecMod` in it's
  own separate file `controller-name.go`, like we have [`deployment.go`](https://github.com/kedgeproject/kedge/blob/d4e27324b444c68a27b552a6eca83baceec4e0df/pkg/spec/deployment.go),
  [`deploymentconfig.go`](https://github.com/kedgeproject/kedge/blob/f4c3808a0285199a5e94f3b89bdfe03eedfe91a3/pkg/spec/deploymentconfig.go), etc.
  - After that, goto function [`GetController`](https://github.com/kedgeproject/kedge/blob/cd5297cfdbd2f1daa510824d49c9d3d649ffc0b8/pkg/spec/controller.go#L46)
  and add a case for this new controller type.

- Interface's Validate method

  - This is to validate what user has given is the right information.
  - In this function we fail, since we can't do much about what user means.
  - Here don't forget to make a call to [`ControllerFields.validateControllerFields()`](https://github.com/kedgeproject/kedge/blob/9b732e3d526b03b197c9ba623627099f221f023c/pkg/spec/resources.go#L512)
  and then also add any other validations that are Controller specific.
  - [Sample Validate method](https://github.com/kedgeproject/kedge/blob/cd5297cfdbd2f1daa510824d49c9d3d649ffc0b8/pkg/spec/deployment.go#L42)
  of `deployment` controller.

- Interface's Fix method

  - This is where we can do most of fixing of user provided data, we can do
  auto-population of information that user has not given us.
  - Make sure to call the [`ControllerFields.fixControllerFields()`](https://github.com/surajssd/kedge/blob/9b732e3d526b03b197c9ba623627099f221f023c/pkg/spec/resources.go#L183).
  - Now after that add any Controller specific Fix methods.
  - [Sample Fix method](https://github.com/kedgeproject/kedge/blob/2f1cf4ee6a90e5f0911ac8d8fcee0d751fc97fa8/pkg/spec/job.go#L42)
  of Job controller.

- Interface's Transform method

  - Create basic Kubernetes/OpenShift objects by calling [`CreateK8sObjects`](https://github.com/surajssd/kedge/blob/9b732e3d526b03b197c9ba623627099f221f023c/pkg/spec/resources.go#L407)
  on the controller.
  - Now create the actual `Controller`.
  - [Sample Transform method](https://github.com/surajssd/kedge/blob/f4c3808a0285199a5e94f3b89bdfe03eedfe91a3/pkg/spec/deploymentconfig.go#L79)
  of DeploymentConfig controller.


##  Issue labeling

Most of the issues should have at least `size` and `priority` and `kind` label.

- Default size is [size/M](#sizeM).
- Default priority is [priority/medium](#priority/medium).

### size/*
size/* labels are for estimating size.
It should estimation how complicated it is going to be to solve given problem.

#### [size/S](https://github.com/kedgeproject/kedge/labels/size%2FS)
Simple change, just few lines (no more than 1 day of work).

#### [size/M](https://github.com/kedgeproject/kedge/labels/size%2FM) 
Considerable change but fair straightforward problem statement.
This is the default size. If you are not sure about the sizes of the task, you can start with marking it as `size/M` and adjust size later on.

#### [size/L](https://github.com/kedgeproject/kedge/labels/size%2FL)
A bit more complicated to solve, maybe requiring small refactoring of existing code.

#### [size/XL](https://github.com/kedgeproject/kedge/labels/size%2FL)
Complex change, new big feature, requiring big refactoring, perhaps is an epic and should be broken into smaller tasks.


### priority/*

#### [priority/low](https://github.com/kedgeproject/kedge/labels/priority%2Flow)
The lowest priority, the issue is not affecting any functionality or it is nice to have feature.

#### [priority/medium](https://github.com/kedgeproject/kedge/labels/priority%2Fmedium)
Medium should be a default priority. If you are not sure about priority of the issue mark it as `priority/medium`. 

#### [priority/high](https://github.com/kedgeproject/kedge/labels/priority%2Fhigh)
High priority means that this is important feature that will help someone start using Kedge, or it is a bug that is affecting existing users.  

#### [priority/urgent](https://github.com/kedgeproject/kedge/labels/priority%2Furgent)
If issue is marked as urgent it means that it should be solved before everything else (Drop everything and work on this). Bug marked as `priority/urget` completely breaks important Kedge functionality.
Usually it doesn't make sense to new features as `priority/urgent`.


### kind/*

#### [kind/blocker](https://github.com/kedgeproject/kedge/labels/kind%2Fblocker)
This issue is blocking some other issue. There should be a commend that is explaining what issue is blocked and why.

Related labels: [status/blocked](#statusblocked)

#### [kind/bug](https://github.com/kedgeproject/kedge/labels/kind%2Fbug)
This is a bug. Something that should be working is broken.

#### [kind/CI-CD](https://github.com/kedgeproject/kedge/labels/kind%2CI-CD)
Issue is touching CI/CD area.

#### [kind/discussion](https://github.com/kedgeproject/kedge/labels/kind%2Fdiscussion)
This issue is discussion. The discussion can be about new features or changes or anything that is related to the project.

Related labels: [status/decided](#statusdecided), [status/undecided](#statusundecided)

#### [kind/documentation](https://github.com/kedgeproject/kedge/labels/kind%2Fdocumentation)
Issues related to documentation.

#### [kind/enhancement](https://github.com/kedgeproject/kedge/labels/kind%2Fenhancement)
This issue is improving already existing functionality.

### [kind/epic](https://github.com/kedgeproject/kedge/labels/kind%2Fepic)
Issues marked as `kind/epic` are usually description of a larger goal or feature. 
Before any actual coding work can be started on this it should be broken down to smaller smaller features ([kind/feature](#kindfeature)) or tasks ([kind/task](#kindtask))

#### [kind/feature](https://github.com/kedgeproject/kedge/labels/kind%2Ffeature)
This is a description or definition of a new feature. 

#### [kind/question](https://github.com/kedgeproject/kedge/labels/kind%2Ffeature)
Someone is asking a question. Once the question is answered the issue should be closed.

#### [kind/refactor](https://github.com/kedgeproject/kedge/labels/kind%2Frefactor)
This is work on definition of some kind of refactoring effort.

#### [kind/task](https://github.com/kedgeproject/kedge/labels/kind%2Ftask)
This is clear definition of a task that can be assigned and work on this can start.

#### [kind/tests](https://github.com/kedgeproject/kedge/labels/kind%2Ftests)
Issue related to test and testing.

#### [kind/user-experience](https://github.com/kedgeproject/kedge/labels/kind%2Fuser-experience)
This issue is touching user experience.

### status/*

#### [status/blocked](https://github.com/kedgeproject/kedge/labels/status%2Fblocked)
This issue is blocked by some other issue. There should be a commend explaining why this issue is blocked 
and what is blocking it. Issue that is blocking it should be marked as `kind/blocker`.

Related labels: [kind/blocker](#kindblocker)

#### [status/decided](https://github.com/kedgeproject/kedge/labels/status%2Fdecided)
This usually goes together with `kind/discussion`. Once a discussion ended and decision was made issue should be labeled `status/decided`.

Related labels: [kind/discussion](#kinddiscussion)

#### [status/discussion-ongoing](https://github.com/kedgeproject/kedge/labels/status%2Fdiscussion-ongoing)

#### [status/do-not-review](https://github.com/kedgeproject/kedge/labels/status%2Fdo-not-review)
Used for PRs. It means that PR is Work In Progress and it will change a lot. Currently it is not worth to do any reviews on it.

#### [status/needs-rebase](https://github.com/kedgeproject/kedge/labels/status%2Fneeds-rebase)
Used for PRs. There are conflicts and branch needs rebase and resolving conflicts.

#### [status/needs-review](https://github.com/kedgeproject/kedge/labels/status%2Fneeds-review)
Used for PRs. All work is done, and PR just need review.

#### [status/undecided](https://github.com/kedgeproject/kedge/labels/status%2Fundecided)
This usually goes together with `kind/discussion`. This means that conclusion was not yet made.

Related labels: [kind/discussion](#kinddiscussion)

#### [status/work-in-progress](https://github.com/kedgeproject/kedge/labels/status%2Fwork-in-progress)
Used for PRs. It means that PR is Work In Progress. Some preliminary reviews can be done, but it can't be merged yet, as something is still missing.


### other

#### [duplicate](https://github.com/kedgeproject/kedge/labels/duplicate)
Issue is a duplicate of other already existing issue.

#### [help wanted](https://github.com/kedgeproject/kedge/labels/help%20wanted)

#### [invalid](https://github.com/kedgeproject/kedge/labels/invalid)

#### [wontfix](https://github.com/kedgeproject/kedge/labels/wontfix)
