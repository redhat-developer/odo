package completion

import (
	"github.com/posener/complete"
	"github.com/redhat-developer/odo/pkg/occlient"
	"github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/redhat-developer/odo/pkg/service"
)

// ServiceCompletionHandler provides service name completion for the current project and application
var ServiceCompletionHandler = func(args complete.Args, client *occlient.Client) (completions []string) {
	completions = make([]string, 0)
	util.GetAndSetNamespace(client)
	applicationName := util.GetAppName(client)

	services, err := service.List(client, applicationName)
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
