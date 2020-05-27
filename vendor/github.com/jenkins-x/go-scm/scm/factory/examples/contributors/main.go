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
	args := os.Args
	if len(args) < 2 {
		fmt.Println("usage: repo")
		os.Exit(1)
		return
	}
	repo := args[1]
	client, err := factory.NewClientFromEnvironment()
	if err != nil {
		helpers.Fail(err)
		return
	}

	fmt.Printf("listing collaborators on repository %s\n", repo)

	ctx := context.Background()
	users, _, err := client.Repositories.ListCollaborators(ctx, repo, scm.ListOptions{})
	if err != nil {
		helpers.Fail(err)
		return
	}
	fmt.Printf("Found %d collaborators\n", len(users))

	for _, u := range users {
		fmt.Printf("  user: %#v\n", u)
	}
}
