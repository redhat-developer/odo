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
	args := os.Args
	if len(args) < 4 {
		fmt.Printf("arguments: owner repository path\n")
		return
	}

	owner := args[1]
	repo := args[2]
	path := args[3]
	ref := "master"
	if len(args) > 4 {
		ref = args[4]
	}

	fullRepo := scm.Join(owner, repo)

	fmt.Printf("getting content for repository %s/%s and path: %s with ref: %s\n", owner, repo, path, ref)
	content, _, err := client.Contents.Find(ctx, fullRepo, path, ref)
	if err != nil {
		helpers.Fail(err)
		return
	}

	fmt.Printf("result path %s\n", content.Path)
	fmt.Printf("%s\n", string(content.Data))
}

func createListOptions() scm.ListOptions {
	return scm.ListOptions{
		Size: 1000,
	}
}
