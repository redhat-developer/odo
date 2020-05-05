package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/ghodss/yaml"
	"github.com/jenkins-x/go-scm/scm"
	"github.com/pkg/errors"

	"github.com/jenkins-x/go-scm/scm/factory"
	"github.com/jenkins-x/go-scm/scm/factory/examples/helpers"
)

func main() {
	args := os.Args
	if len(args) < 2 {
		fmt.Printf("usage: org/repo prNumber")
		return
	}
	repo := args[1]

	client, err := factory.NewClientFromEnvironment()
	if err != nil {
		helpers.Fail(err)
		return
	}
	ctx := context.Background()

	if len(args) < 3 {
		fmt.Printf("Getting PRs\n")

		prs, _, err := client.PullRequests.List(ctx, repo, scm.PullRequestListOptions{})
		if err != nil {
			helpers.Fail(err)
			return
		}
		for _, pr := range prs {
			fmt.Printf("Found PullRequest:\n")
			data, err := yaml.Marshal(pr)
			if err != nil {
				helpers.Fail(errors.Wrap(err, "failed to marshal PR as YAML"))
				return
			}
			fmt.Printf("%s:\n", string(data))
		}
		return
	}
	prText := args[2]
	number, err := strconv.Atoi(prText)
	if err != nil {
		helpers.Fail(errors.Wrapf(err, "failed to parse PR number: %s", prText))
		return
	}

	fmt.Printf("Getting PR\n")

	pr, _, err := client.PullRequests.Find(ctx, repo, number)
	if err != nil {
		helpers.Fail(err)
		return
	}
	fmt.Printf("Found PullRequest:\n")
	data, err := yaml.Marshal(pr)
	if err != nil {
		helpers.Fail(errors.Wrap(err, "failed to marshal PR as YAML"))
		return
	}
	fmt.Printf("%s:\n", string(data))
}
