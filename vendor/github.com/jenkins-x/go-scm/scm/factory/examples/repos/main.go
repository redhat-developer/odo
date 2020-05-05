package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jenkins-x/go-scm/scm"
	"github.com/jenkins-x/go-scm/scm/factory"
	"github.com/jenkins-x/go-scm/scm/factory/examples/helpers"
)

func main() {
	client, err := factory.NewClientFromEnvironment()
	if err != nil {
		helpers.Fail(err)
		return
	}

	ctx := context.Background()
	var repos []*scm.Repository

	args := os.Args
	if len(args) > 1 {
		owner := args[1]
		fmt.Printf("finding repositories for owner %s\n", owner)

		repos, _, err = client.Repositories.ListOrganisation(ctx, owner, createListOptions())
	} else {
		fmt.Printf("listing repositories\n")

		repos, _, err = client.Repositories.List(ctx, createListOptions())
	}

	if err != nil {
		helpers.Fail(err)
		return
	}
	fmt.Printf("Found %d repositories\n", len(repos))

	for _, r := range repos {
		fmt.Printf("  repo: %#v\n", r)
	}
}

func createListOptions() scm.ListOptions {
	return scm.ListOptions{
		Size: 1000,
	}
}
