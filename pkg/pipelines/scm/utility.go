package scm

import (
	"fmt"
	"net/url"
	"strings"

	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"

	triggersv1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
)

func getDriverName(rawURL string) (string, error) {
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

func invalidRepoPathError(repoURL string) error {
	return fmt.Errorf("unable to determine repo path from: %s", repoURL)
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

func createBindingParam(name string, value string) pipelinev1.Param {
	return pipelinev1.Param{
		Name: name,
		Value: pipelinev1.ArrayOrString{
			StringVal: value,
			Type:      pipelinev1.ParamTypeString,
		},
	}
}
