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

    To update the CLI Structure in README.md, run `make generate-cli-docs` and update the section in [README.md](/README.md#cli-structure)
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

## odo-bot
This is GitHub user that does all the automation.

### Scripts using odo-bot

| script | what it is doing | access via | 
|-|-|-|
| .travis.yml | uploading binaries to GitHub release page | personal access token `deploy-github-release` |
