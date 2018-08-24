# Development Guide

## Workflow

### Fork the main repository

1. Go to https://github.com/redhat-developer/odo
2. Click the "Fork" button (at the top right)

### Clone your fork

The commands below require that you have $GOPATH. We highly recommended you put odo code into your $GOPATH.

```sh
git clone https://github.com/$YOUR_GITHUB_USERNAME/odo.git $GOPATH/src/github.com/redhat-developer/odo
cd $GOPATH/src/github.com/redhat-developer/odo
git remote add upstream 'https://github.com/redhat-developer/odo'
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

**Pull request description:** A PR should contain an accurate description of the feature being implemented as well as a link to an active issue (if any).

### Test Driven Development

We follow Test Driven Development(TDD) workflow in our development process. You can read more about it [here](/docs/tdd-workflow.md).

### Writing Unit tests

#### Unit test for functions consuming client-go functions

Unit-tests for the functions making API calls with client-go library 
can be written by using package (fake)[https://godoc.org/k8s.io/client-go/kubernetes/fake].

The way to mock the api calls is by mocking the actual api calls with functions defined in
(client-go/testing)[https://godoc.org/k8s.io/client-go/testing]
and using (pkg/testing)[https://golang.org/pkg/testing/] for writing the test.


##### How to write unit tests having API calls in a nutshell?

1. Identify the API calls being made by the function during the execution

2. Initialise the relevant clientsets and clients 

3. In case of API calls which are returning any object 
  which is later being processed inside the function, 
  Implement fake functions and use them instead.
  Use addreactor method for adding corresponding clientset ref:(here)[https://godoc.org/k8s.io/client-go/testing#Fake.AddReactor].

4. Use https://godoc.org/k8s.io/client-go/testing#Fake.Actions 
  for validating number of fake actions performed and the values with which the fake calls were made.

##### Example of using fake for testing functions making API calls.

Initialising a fakeclientset and fakeclient properly is the first thing. 

For example : 
take the example of unit test for CreateRoute function in pkg/occlient/occlient.go 
```
/ CreateRoute creates a route object for the given service and with the given
// labels
func (c *Client) CreateRoute(name string, serviceName string, labels map[string]string) (*routev1.Route, error) {
	route := &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
		Spec: routev1.RouteSpec{
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: serviceName,
			},
		},
	}
	r, err := c.routeClient.Routes(c.namespace).Create(route)
	if err != nil {
		return nil, errors.Wrap(err, "error creating route")
	}
	return r, nil
}
```

Looking at the function body and identify how many API calls it is making while execution. 
In this case CreateRoute is making only a single API call which is 
```
r, err := c.routeClient.Routes(c.namespace).Create(route)
```

for example,
add the code for initialising fakeclientset & fakeclient for routeClient on the FakeNew function 
```
    fkclientset.RouteClientset = fakeRouteClientset.NewSimpleClientset()
    client.routeClient = fkclientset.RouteClientset.Route()

```
Initialise a fakeclientset by calling fakeRouteClientset.NewSimpleClientset() 
It returns a simple set of object tracker which can process, creates,updates and deletions 
but without any validations. So it is recommended to implement validations separately if needed.

Writing the test function for CreateRoute,

call the initialisation function
```
fkclient, fkclientset := FakeNew()
```

Once fakeclientset and fakeclient is being sucessfully initialised 
it's possible to make the method calls on the struct which got initialised using FakeNew().
Here in this case it needs to validate the value with which the function is returining
and the number of actions performed on routeclientset.


Then after making that function call inside the test function,
the number of actions performed can be validated using being validated Actions() method on clientset.

for example in this case, `len(fkclientset.RouteClientset.Actions())` 
returns number of actions performed on routeclientset

and values are being validated later.
Ref. (here)[https://github.com/redhat-developer/odo/pull/456/files#diff-54c1e3725d2cfb565cbd1cfdb02bd792R63]

For the API calls which are returning objects that are later being processed inside the function body,
adding reactors for the relevent actions is also needed.

for example take a look at `GetDeploymentConfigsFromSelector` 
https://github.com/redhat-developer/odo/blob/bbcf73f7a9c7af28429d7ec7a9dac274abe0c72a/pkg/occlient/occlient_test.go#L2474

it is making an API call as below 
```
dcList, err := c.appsClient.DeploymentConfigs(c.namespace).List(metav1.ListOptions{
                   LabelSelector: selector,
})
```
which fetch the DeploymentConfigList which later being returned by the `GetDeploymentConfigsFromSelector` function.

so for this we can add a reactor like below, 
which will return tt.dcBefore(a dc object), nil(in place of error)
we can keep the fist return value as `true` for all.

```
fakeClientSet.AppsClientset.PrependReactor("list", "deploymentconfigs", func(action ktesting.Action) (bool, runtime.Object, error) {
    if !reflect.DeepEqual(action.(ktesting.ListAction).GetListRestrictions().Labels.String(), tt.selector) {
        return true, nil, fmt.Errorf("labels not matching with expected values, expected:%s, got:%s", tt.selector, action.(ktesting.ListAction).GetListRestrictions())
    }    
    return true, &listOfDC, nil
})   

```

so during the execution when 
` c.appsClient.DeploymentConfigs(c.namespace).List(metav1.ListOptions{  LabelSelector: selector,})` is being called
the above two values will be returned.

More examples can be found in https://github.com/redhat-developer/odo/blob/master/pkg/occlient/occlient_test.go

 - Reactor is an interface to allow the composition of reaction functions.
 - Verb is get, list, delete...
For more info about reactors Ref: https://godoc.org/k8s.io/client-go/testing

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

# Release guide

## Making a release

Making artifacts for new release is automated. 
When new git tag is created, Travis-ci deploy job automatically builds binaries and uploads it to GitHub release page.

1. Create PR with updated version in following files:
    - [cmd/version.go](/cmd/version.go)
    - [scripts/install.sh](/scripts/install.sh)
    - [README.md](/README.md)
    - [odo.rb](https://github.com/kadel/homebrew-odo/blob/master/Formula/odo.rb) in [kadel/homebrew-odo](https://github.com/kadel/homebrew-odo)

    There is a helper script [scripts/bump-version.sh](/scripts/bump-version.sh) that should change version number in all files listed above (expect odo.rb).

    To update the CLI Structure in README.md, run `make generate-cli-structure` and update the section in [README.md](/README.md#cli-structure)

    To update the CLI reference documentation in docs/cli-reference.md, run `make generate-cli-structure > docs/cli-reference.md`.
2. When PR is merged create and push new git tag for version.
    ```
    git tag v0.0.1
    git push upstream v0.0.1
    ```
    Or create new release using GitHub site (this has to be a proper release, not just draft). 
    Do not upload any binaries for release
    When new tag is created Travis-CI starts a special deploy job.
    This job builds binaries automatically (via `make prepare-release`) and then uploads it to GitHub release page (done using odo-bot user).
3. When job fishes you should see binaries on GitHub release page. Release is now marked as a draft. Update descriptions and publish release.
4. Verify that packages have been uploaded to rpm and deb repositories.
5. Confirm the binaries are available in GitHub release page and update the file `build/VERSION` with latest version number.

## odo-bot
This is GitHub user that does all the automation.

### Scripts using odo-bot

| Script      | What it is doing                          | Access via                                    |
|-------------|-------------------------------------------|-----------------------------------------------|
| .travis.yml | Uploading binaries to GitHub release page | Personal access token `deploy-github-release` |
