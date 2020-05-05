package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	githubql "github.com/shurcooL/githubv4"
	"github.com/sirupsen/logrus"

	"github.com/jenkins-x/go-scm/scm"
	"github.com/jenkins-x/go-scm/scm/factory"
	"github.com/jenkins-x/go-scm/scm/factory/examples/helpers"
)

var (
	// searchTimeFormat is a time.Time format string for ISO8601 which is the
	// format that GitHub requires for times specified as part of a search query.
	searchTimeFormat = "2006-01-02T15:04:05Z"

	// FoundingYear is the year GitHub was founded. This is just used so that
	// we can lower bound dates related to PRs and issues.
	foundingYear, _ = time.Parse(searchTimeFormat, "2007-01-01T00:00:00Z")
)

// PullRequest holds graphql data about a PR, including its commits and their contexts.
type PullRequest struct {
	Number githubql.Int
	Author struct {
		Login githubql.String
	}
	BaseRef struct {
		Name   githubql.String
		Prefix githubql.String
	}
	HeadRefName githubql.String `graphql:"headRefName"`
	HeadRefOID  githubql.String `graphql:"headRefOid"`
	Mergeable   githubql.MergeableState
	Repository  struct {
		Name          githubql.String
		NameWithOwner githubql.String
		Owner         struct {
			Login githubql.String
		}
	}
	Commits struct {
		Nodes []struct {
			Commit Commit
		}
		// Request the 'last' 4 commits hoping that one of them is the logically 'last'
		// commit with OID matching HeadRefOID. If we don't find it we have to use an
		// additional API token. (see the 'headContexts' func for details)
		// We can't raise this too much or we could hit the limit of 50,000 nodes
		// per query: https://developer.github.com/v4/guides/resource-limitations/#node-limit
	} `graphql:"commits(last: 4)"`
	Labels struct {
		Nodes []struct {
			Name githubql.String
		}
	} `graphql:"labels(first: 100)"`
	Milestone *struct {
		Title githubql.String
	}
	Body      githubql.String
	Title     githubql.String
	UpdatedAt githubql.DateTime
}

// Commit holds graphql data about commits and which contexts they have
type Commit struct {
	Status struct {
		Contexts []Context
	}
	OID githubql.String `graphql:"oid"`
}

// Context holds graphql response data for github contexts.
type Context struct {
	Context     githubql.String
	Description githubql.String
	State       githubql.StatusState
}

type PRNode struct {
	PullRequest PullRequest `graphql:"... on PullRequest"`
}

type searchQuery struct {
	RateLimit struct {
		Cost      githubql.Int
		Remaining githubql.Int
	}
	Search struct {
		PageInfo struct {
			HasNextPage githubql.Boolean
			EndCursor   githubql.String
		}
		Nodes []PRNode
	} `graphql:"search(type: ISSUE, first: 100, after: $searchCursor, query: $query)"`
}

func main() {
	client, err := factory.NewClientFromEnvironment()
	if err != nil {
		helpers.Fail(err)
		return
	}
	args := os.Args
	if len(args) < 2 {
		fmt.Printf("usage: queryString")
		return
	}
	query := args[1]

	fmt.Printf("searching issues and pull requests via GraphQL query %s\n", query)

	graphql := client.GraphQL
	if graphql == nil {
		helpers.Fail(fmt.Errorf("No GraphQL support for driver %s", client.Driver.String()))
		return
	}
	results, err := search(client, logrus.WithField("query", query), query, time.Time{}, time.Now())
	if err != nil {
		helpers.Fail(err)
		return
	}
	fmt.Printf("Found %d results\n", len(results))

	for _, r := range results {
		commits := []string{}
		for _, commit := range r.Commits.Nodes {
			commits = append(commits, string(commit.Commit.OID))
		}
		fmt.Printf("PR %s #%d title: %s commits: %s\n", string(r.Repository.NameWithOwner), r.Number, string(r.Title), strings.Join(commits, ", "))
	}
}

func datedQuery(q string, start, end time.Time) string {
	return fmt.Sprintf("%s %s", q, dateToken(start, end))
}

func floor(t time.Time) time.Time {
	if t.Before(foundingYear) {
		return foundingYear
	}
	return t
}

func search(client *scm.Client, log *logrus.Entry, q string, start, end time.Time) ([]PullRequest, error) {
	start = floor(start)
	end = floor(end)
	log = log.WithFields(logrus.Fields{
		"query": q,
		"start": start.String(),
		"end":   end.String(),
	})
	requestStart := time.Now()
	var cursor *githubql.String
	vars := map[string]interface{}{
		"query":        githubql.String(datedQuery(q, start, end)),
		"searchCursor": cursor,
	}

	var totalCost, remaining int
	var ret []PullRequest
	var sq searchQuery
	ctx := context.Background()
	for {
		log.Debug("Sending query")
		if err := client.GraphQL.Query(ctx, &sq, vars); err != nil {
			if cursor != nil {
				err = fmt.Errorf("cursor: %q, err: %v", *cursor, err)
			}
			return ret, err
		}
		totalCost += int(sq.RateLimit.Cost)
		remaining = int(sq.RateLimit.Remaining)
		for _, n := range sq.Search.Nodes {
			ret = append(ret, n.PullRequest)
		}
		if !sq.Search.PageInfo.HasNextPage {
			break
		}
		cursor = &sq.Search.PageInfo.EndCursor
		vars["searchCursor"] = cursor
		log = log.WithField("searchCursor", *cursor)
	}
	log.WithField("duration", time.Since(requestStart).String()).Debugf("GraphQL returned %d PRs and cost %d point(s). %d remaining.", len(ret), totalCost, remaining)
	return ret, nil
}

// dateToken generates a GitHub search query token for the specified date range.
// See: https://help.github.com/articles/understanding-the-search-syntax/#query-for-dates
func dateToken(start, end time.Time) string {
	// GitHub's GraphQL API silently fails if you provide it with an invalid time
	// string.
	// Dates before 1970 (unix epoch) are considered invalid.
	startString, endString := "*", "*"
	if start.Year() >= 1970 {
		startString = start.Format(searchTimeFormat)
	}
	if end.Year() >= 1970 {
		endString = end.Format(searchTimeFormat)
	}
	return fmt.Sprintf("updated:%s..%s", startString, endString)
}
