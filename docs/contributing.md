# Contributing guide



## Making a release

Making artifacts for new release is automated. 
When new git tag is created, Travis-ci deploy job automatically builds binaries and uploads it to GitHub release page.

1. Create PR with updated version in following files:
    - `cmd/version.go`
    - `scripts/install.sh`
2. When PR is merged create and push new git tag for version.
    ```
    git tag v0.0.1
    git push upstream v0.0.1
    ```
    Or create new release using GitHub site (this has to be proper release not just draft). 
    Do not upload any binaries for release
    When new tag is created Travis-CI starts a special deploy job.
    This job builds binaries automatically (via `make prepare-release`) and then uploads it to GitHub release page (done using ocdev-bot user).
3. When job fishes you should see binaries on GitHub release page. Release is now marked as a draft. Update descriptions and publish release.


## ocdev-bot
This is GitHub user that does all the automation.

### Scripts using ocdev-bot

| script | what it is doing | access via | 
|-|-|-|
| .travis.yml | uploading binaries to GitHub release page | personal access token `deploy-github-release` |