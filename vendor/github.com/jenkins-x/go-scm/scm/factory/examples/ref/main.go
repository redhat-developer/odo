package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jenkins-x/go-scm/scm/factory"
	"github.com/jenkins-x/go-scm/scm/factory/examples/helpers"
)

func main() {
	args := os.Args
	if len(args) < 3 {
		fmt.Println("usage: repo ref")
		os.Exit(1)
		return
	}
	repo := args[1]
	ref := args[2]

	client, err := factory.NewClientFromEnvironment()
	if err != nil {
		helpers.Fail(err)
		return
	}

	fmt.Printf("finding in repo: %s ref: %s\n", repo, ref)

	ctx := context.Background()
	answer, _, err := client.Git.FindRef(ctx, repo, ref)
	if err != nil {
		helpers.Fail(err)
		return
	}
	fmt.Printf("Found: %s\n", answer)
}
