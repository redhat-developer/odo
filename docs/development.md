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

Note: If you have write access to the main repository at github.com/redhat-developer/odo, you should modify your git configuration so that you can't accidentally push to upstream:

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
