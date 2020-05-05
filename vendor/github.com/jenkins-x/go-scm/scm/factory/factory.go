package factory

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/jenkins-x/go-scm/scm"
	"github.com/jenkins-x/go-scm/scm/driver/bitbucket"
	"github.com/jenkins-x/go-scm/scm/driver/fake"
	"github.com/jenkins-x/go-scm/scm/driver/gitea"
	"github.com/jenkins-x/go-scm/scm/driver/github"
	"github.com/jenkins-x/go-scm/scm/driver/gitlab"
	"github.com/jenkins-x/go-scm/scm/driver/gogs"
	"github.com/jenkins-x/go-scm/scm/driver/stash"
	"github.com/jenkins-x/go-scm/scm/transport"
	"golang.org/x/oauth2"
)

// MissingGitServerURL the error returned if you use a git driver that needs a git server URL
var MissingGitServerURL = fmt.Errorf("No git serverURL was specified")

type clientOptionFunc func(*scm.Client)

// NewClient creates a new client for a given driver, serverURL and OAuth token
func NewClient(driver, serverURL, oauthToken string, opts ...clientOptionFunc) (*scm.Client, error) {
	if driver == "" {
		driver = "github"
	}
	var client *scm.Client
	var err error

	switch driver {
	case "bitbucket", "bitbucketcloud":
		if serverURL != "" {
			client, err = bitbucket.New(ensureBBCEndpoint(serverURL))
		} else {
			client = bitbucket.NewDefault()
		}
	case "fake":
		client, _ = fake.NewDefault()
	case "gitea":
		if serverURL == "" {
			return nil, MissingGitServerURL
		}
		client, err = gitea.New(serverURL)
	case "github":
		if serverURL != "" {
			client, err = github.New(ensureGHEEndpoint(serverURL))
		} else {
			client = github.NewDefault()
		}
	case "gitlab":
		if serverURL != "" {
			client, err = gitlab.New(serverURL)
		} else {
			client = gitlab.NewDefault()
		}
	case "gogs":
		if serverURL == "" {
			return nil, MissingGitServerURL
		}
		client, err = gogs.New(serverURL)
	case "stash", "bitbucketserver":
		if serverURL == "" {
			return nil, MissingGitServerURL
		}
		client, err = stash.New(serverURL)
	default:
		return nil, fmt.Errorf("Unsupported $GIT_KIND value: %s", driver)
	}
	if err != nil {
		return client, err
	}
	if oauthToken != "" {
		if driver == "gitlab" || driver == "bitbucketcloud" {
			client.Client = &http.Client{
				Transport: &transport.PrivateToken{
					Token: oauthToken,
				},
			}
		} else {
			ts := oauth2.StaticTokenSource(
				&oauth2.Token{AccessToken: oauthToken},
			)
			client.Client = oauth2.NewClient(context.Background(), ts)
		}
	}
	for _, o := range opts {
		o(client)
	}
	return client, err
}

// NewClientFromEnvironment creates a new client using environment variables $GIT_KIND, $GIT_SERVER, $GIT_TOKEN
// defaulting to github if no $GIT_KIND or $GIT_SERVER
func NewClientFromEnvironment() (*scm.Client, error) {
	driver := os.Getenv("GIT_KIND")
	serverURL := os.Getenv("GIT_SERVER")
	oauthToken := os.Getenv("GIT_TOKEN")
	if oauthToken == "" {
		return nil, fmt.Errorf("No Git OAuth token specified for $GIT_TOKEN")
	}
	client, err := NewClient(driver, serverURL, oauthToken)
	if driver == "" {
		driver = client.Driver.String()
	}
	fmt.Printf("using driver: %s and serverURL: %s\n", driver, serverURL)
	return client, err
}

// ensureGHEEndpoint lets ensure we have the /api/v3 suffix on the URL
func ensureGHEEndpoint(u string) string {
	if strings.HasPrefix(u, "https://github.com") || strings.HasPrefix(u, "http://github.com") {
		return "https://api.github.com"
	}
	// lets ensure we use the API endpoint to login
	if strings.Index(u, "/api/") < 0 {
		u = scm.UrlJoin(u, "/api/v3")
	}
	return u
}

// ensureBBCEndpoint lets ensure we have the /api/v3 suffix on the URL
func ensureBBCEndpoint(u string) string {
	if strings.HasPrefix(u, "https://bitbucket.org") || strings.HasPrefix(u, "http://bitbucket.org") {
		return "https://api.bitbucket.org"
	}
	return u
}

func Client(httpClient *http.Client) clientOptionFunc {
	return func(c *scm.Client) {
		c.Client = httpClient
	}
}
