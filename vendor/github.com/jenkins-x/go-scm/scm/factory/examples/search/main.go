package main

import (
	"context"
	"fmt"
	"os"
	"strings"

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
	options := scm.SearchOptions{}
	args := os.Args
	if len(args) < 2 {
		fmt.Printf("usage: queryString [sort] [ascending/descending]")
		return
	}
	options.Query = args[1]
	options.Ascending = true
	if len(args) > 2 {
		options.Sort = args[2]
	}
	if len(args) > 3 {
		options.Ascending = strings.ToLower(args[3]) == "true"
	}

	fmt.Printf("searching issues and pull requests using %#v\n", &options)

	ctx := context.Background()
	results, res, err := client.Issues.Search(ctx, options)
	if err != nil {
		helpers.Fail(err)
		return
	}
	fmt.Printf("Found %d results\n", len(results))

	for k, v := range res.Header {
		fmt.Printf("  header: %s: %v\n", k, v)
	}
	for _, r := range results {
		fmt.Printf("  result: %#v\n", r)
	}
}
