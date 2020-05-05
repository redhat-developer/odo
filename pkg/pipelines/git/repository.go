package git

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/jenkins-x/go-scm/scm"
	"github.com/jenkins-x/go-scm/scm/factory"
)

// Repository represent a Git repository ofa specific Git repository URL
type Repository struct {
	*scm.Client

	// name is the repository name of the form <user>/<repository>
	name string
}

// NewRepository creates a new Git reposiory object
func NewRepository(rawURL, token string) (*Repository, error) {

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	driverName, err := getDriverName(parsedURL)
	if err != nil {
		return nil, err
	}

	client, err := factory.NewClient(driverName, "", token)
	if err != nil {
		return nil, err
	}

	repoName, err := getRepoName(parsedURL)
	if err != nil {
		return nil, err
	}

	return &Repository{name: repoName, Client: client}, nil
}

// ListWebhooks returns a list of webhook IDs of the given listener in this repository
func (r *Repository) ListWebhooks(listenerURL string) ([]string, error) {

	hooks, _, err := r.Client.Repositories.ListHooks(context.Background(), r.name, scm.ListOptions{})
	if err != nil {
		return nil, err
	}

	ids := []string{}
	for _, hook := range hooks {
		if hook.Target == listenerURL {
			ids = append(ids, hook.ID)
		}
	}

	return ids, nil
}

// DeleteWebhooks deletes all webhooks that associate with the given listener in this repository
func (r *Repository) DeleteWebhooks(ids []string) ([]string, error) {

	deleted := []string{}
	for _, id := range ids {
		_, err := r.Client.Repositories.DeleteHook(context.Background(), r.name, id)
		if err != nil {
			return deleted, fmt.Errorf("failed to delete webhook id %s: %w", id, err)
		}
		deleted = append(deleted, id)
	}

	return deleted, nil
}

// CreateWebhook creates a new webhook in the repository
// It returns ID of the created webhook
func (r *Repository) CreateWebhook(listenerURL, secret string) (string, error) {

	in := &scm.HookInput{
		Target: listenerURL,
		Secret: secret,
		Events: scm.HookEvents{
			PullRequest: true,
			Push:        true,
		},
	}

	created, _, err := r.Client.Repositories.CreateHook(context.Background(), r.name, in)
	return created.ID, err
}

func getDriverName(u *url.URL) (string, error) {

	if s := strings.TrimSuffix(u.Host, ".com"); s != u.Host {
		return strings.ToLower(s), nil
	}

	if s := strings.TrimSuffix(u.Host, ".org"); s != u.Host {
		return strings.ToLower(s), nil
	}

	return "", errors.New("unknown Git server: " + u.Host)
}

func getRepoName(u *url.URL) (string, error) {

	var components []string

	for _, s := range strings.Split(u.Path, "/") {
		if s != "" {
			components = append(components, s)
		}
	}

	if len(components) != 2 {
		return "", errors.New("failed to get Git repo: " + u.Path)
	}

	components[1] = strings.TrimSuffix(components[1], ".git")

	for _, s := range components {
		if strings.Index(s, ".") != -1 {
			return "", errors.New("failed to get Git repo: " + u.Path)
		}
	}

	return components[0] + "/" + components[1], nil
}
