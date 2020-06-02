package scm

import (
	"fmt"
	"net/url"
	"strings"

	triggersv1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
)

// GetDriverName gets Driver name from URL
func GetDriverName(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	if s := strings.TrimSuffix(u.Host, ".com"); s != u.Host {
		return strings.ToLower(s), nil
	}
	return "", invalidRepoTypeError(rawURL)
}

func invalidRepoTypeError(repoURL string) error {
	return fmt.Errorf("unable to determine type of Git host from: %s", repoURL)
}

func invalidRepoPathError(gitType, path string) error {
	return fmt.Errorf("invalid repository path for %s: %s", gitType, path)
}

func unsupportedGitTypeError(gitType string) error {
	return fmt.Errorf("unsupported Git repository type: %s", gitType)
}

func invalidRepoURLError(repoURL, reason string) error {
	return fmt.Errorf("invalid repository URL %s: %s", repoURL, reason)
}

func createEventInterceptor(filter string, repoName string) *triggersv1.EventInterceptor {
	return &triggersv1.EventInterceptor{
		CEL: &triggersv1.CELInterceptor{
			Filter: fmt.Sprintf(filter, repoName),
		},
	}
}

func createListenerTemplate(name string) triggersv1.EventListenerTemplate {
	return triggersv1.EventListenerTemplate{
		Name: name,
	}
}

func createListenerBinding(name string) *triggersv1.EventListenerBinding {
	return &triggersv1.EventListenerBinding{
		Name: name,
	}
}

func createBindings(names []string) []*triggersv1.EventListenerBinding {
	bindings := make([]*triggersv1.EventListenerBinding, len(names))
	for i, name := range names {
		bindings[i] = createListenerBinding(name)
	}
	return bindings
}

func createBindingParam(name string, value string) triggersv1.Param {
	return triggersv1.Param{
		Name:  name,
		Value: value,
	}
}

func processRawURL(rawURL string, processPath func(*url.URL) (string, error)) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	path, err := processPath(parsedURL)
	if err != nil {
		return "", err
	}
	return path, nil
}

func splitRepositoryPath(parsedURL *url.URL) ([]string, error) {
	var components []string
	for _, s := range strings.Split(parsedURL.Path, "/") {
		if s != "" {
			components = append(components, s)
		}
	}
	if len(components) < 1 {
		return nil, invalidRepoURLError(parsedURL.String(), "path is empty")
	}
	components[len(components)-1] = strings.TrimSuffix(components[len(components)-1], ".git")
	return components, nil
}
