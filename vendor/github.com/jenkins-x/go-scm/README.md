# go-scm

[![Documentation](https://godoc.org/github.com/jenkins-x/go-scm?status.svg)](https://pkg.go.dev/mod/github.com/jenkins-x/go-scm)
[![Go Report Card](https://goreportcard.com/badge/github.com/jenkins-x/go-scm)](https://goreportcard.com/report/github.com/jenkins-x/go-scm)


A small library with minimal depenencies for working with Webhooks, Commits, Issues, Pull Requests, Comments, Reviews, Teams and more on multiple git provider:

* [GitHub](https://github.com/jenkins-x/go-scm/blob/master/scm/driver/github/github.go#L46)
* [GitHub Enterprise](https://github.com/jenkins-x/go-scm/blob/master/scm/driver/github/github.go#L19) (you specify a server URL)
* [BitBucket Server](https://github.com/jenkins-x/go-scm/blob/master/scm/driver/stash/stash.go#L24)
* [BitBucket Cloud](https://github.com/jenkins-x/go-scm/blob/master/scm/driver/bitbucket/bitbucket.go#L20)
* [GitLab](https://github.com/jenkins-x/go-scm/blob/master/scm/driver/gitlab/gitlab.go#L19)
* [Gitea](https://github.com/jenkins-x/go-scm/blob/master/scm/driver/gitea/gitea.go#L22)
* [Gogs](https://github.com/jenkins-x/go-scm/blob/master/scm/driver/gogs/gogs.go#L22)
* [Fake](https://github.com/jenkins-x/go-scm/blob/master/scm/driver/fake/fake.go)

## Building

See the [guide to prerequisites, building and running the code](BUILDING.md)

## Working on the code

Clone this repository and use go test...

``` 
git clone https://github.com/jenkins-x/go-scm.git
cd go-scm
go test ./...
```

## Writing tests

There are lots of tests for each driver; using sample JSON that comes from the git provider together with the expected canonical JSON.

e.g. I added this test for ListTeams on github: https://github.com/jenkins-x/go-scm/blob/master/scm/driver/github/org_test.go#L83-116

you then add some real json from the git provider: https://github.com/jenkins-x/go-scm/blob/master/scm/driver/github/testdata/teams.json and provide the expected json: https://github.com/jenkins-x/go-scm/blob/master/scm/driver/github/testdata/teams.json.golden


## Trying the client on a provider

There are a few little sample programs in [scm/factory/examples](scm/factory/examples) which are individual binaries you can run from the command line or your IDE.

To test against a git provider of your choice try defining these environment variables:

* `GIT_KIND` for the kind of git provider (e.g. `github`, `bitbucketserver`, `gitlab` etc)
* `GIT_SERVER` for the URL of the server to communicate with
* `GIT_TOKEN` for the git OAuth/private token to talk to the git server 

## Git API Reference docs

To help hack on the different drivers here's a list of docs which outline the git providers REST APIs

### GitHub

* REST API reference: https://developer.github.com/v3/
* WebHooks: https://developer.github.com/v3/activity/events/types/

### Bitbucket Server

* REST API reference: https://docs.atlassian.com/bitbucket-server/rest/6.5.1/bitbucket-rest.html
* Webhooks: https://confluence.atlassian.com/bitbucketserver/event-payload-938025882.html

### Bitbucket Cloud

* REST API reference: https://developer.atlassian.com/bitbucket/api/2/reference/

### Gitlab

* REST API reference: https://docs.gitlab.com/ee/api/

## Fake driver for testing

When testing the use of go-scm its really handy to use the [fake](https://github.com/jenkins-x/go-scm/blob/master/scm/driver/fake/fake.go) provider which lets you populate the in memory resources inside the driver or query resources after a test has run.

```go 
client, data := fake.NewDefault()
```    

## Community

We have a [kanban board](https://github.com/jenkins-x/go-scm/projects/1?add_cards_query=is%3Aopen) of stuff to work on if you fancy contributing!

You can also find us [on Slack](http://slack.k8s.io/) at [kubernetes.slack.com](https://kubernetes.slack.com/):

* [\#jenkins-x-dev](https://kubernetes.slack.com/messages/C9LTHT2BB) for developers of Jenkins X and related OSS projects
* [\#jenkins-x-user](https://kubernetes.slack.com/messages/C9MBGQJRH) for users of Jenkins X
