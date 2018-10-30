package completion

import (
	"github.com/posener/complete"
	"github.com/redhat-developer/odo/pkg/application"
	"github.com/redhat-developer/odo/pkg/occlient"
	"github.com/redhat-developer/odo/pkg/project"
	"github.com/redhat-developer/odo/pkg/service"
)

// ServiceCompletionHandler provides service name completion for the current project and application
var ServiceCompletionHandler = func(args complete.Args, client *occlient.Client) (completions []string) {
	completions = make([]string, 0)
	applicationName, err := application.GetCurrent(client)
	if err != nil {
		return completions
	}
	projectName := project.GetCurrent(client)
	services, err := service.List(client, applicationName, projectName)
	if err != nil {
		return completions
	}

	for _, class := range services {
		completions = append(completions, class.Name)
	}

	return completions
}

// ServiceClassCompletionHandler provides catalog service class name completion
var ServiceClassCompletionHandler = func(args complete.Args, client *occlient.Client) (completions []string) {
	completions = make([]string, 0)
	services, err := client.GetClusterServiceClasses()
	if err != nil {
		return completions
	}

	for _, class := range services {
		completions = append(completions, class.Spec.ExternalName)
	}

	return completions
}
