package completion

import (
	applabels "github.com/openshift/odo/pkg/application/labels"
	"github.com/openshift/odo/pkg/envinfo"

	"github.com/openshift/odo/pkg/application"
	"github.com/openshift/odo/pkg/catalog"
	"github.com/openshift/odo/pkg/component"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/posener/complete"
	"github.com/spf13/cobra"
)

// AppCompletionHandler provides completion for the app commands
var AppCompletionHandler = func(cmd *cobra.Command, args parsedArgs, context *genericclioptions.Context) (completions []string) {
	completions = make([]string, 0)

	applications, err := application.List(context.Client.GetKubeClient())
	if err != nil {
		return completions
	}

	for _, app := range applications {
		if args.commands[app] {
			return nil
		}
		completions = append(completions, app)
	}
	return
}

// FileCompletionHandler provides suggestions for files and directories
var FileCompletionHandler = func(cmd *cobra.Command, args parsedArgs, context *genericclioptions.Context) (completions []string) {
	completions = append(completions, complete.PredictFiles("*").Predict(args.original)...)
	return
}

// ProjectNameCompletionHandler provides project name completion
var ProjectNameCompletionHandler = func(cmd *cobra.Command, args parsedArgs, context *genericclioptions.Context) (completions []string) {
	completions = make([]string, 0)
	projects, err := context.Client.ListProjectNames()
	if err != nil {
		return completions
	}

	for _, project := range projects {
		// we found the project name in the list which means
		// that the project name has been already selected by the user so no need to suggest more
		if args.commands[project] {
			return nil
		}
		completions = append(completions, project)
	}
	return completions
}

// URLCompletionHandler provides completion for the url commands
var URLCompletionHandler = func(cmd *cobra.Command, args parsedArgs, context *genericclioptions.Context) (completions []string) {
	return
}

// StorageDeleteCompletionHandler provides storage name completion for storage delete
var StorageDeleteCompletionHandler = func(cmd *cobra.Command, args parsedArgs, context *genericclioptions.Context) (completions []string) {
	completions = make([]string, 0)

	envInfo, err := envinfo.New()
	if err != nil {
		return completions
	}

	storageList, err := envInfo.ListStorage()
	if err != nil {
		return completions
	}

	for _, storage := range storageList {
		// we found the storage name in the list which means
		// that the storage name has been already selected by the user so no need to suggest more
		if args.commands[storage.Name] {
			return nil
		}
		completions = append(completions, storage.Name)
	}
	return completions
}

// CreateCompletionHandler provides component type completion in odo create command
var CreateCompletionHandler = func(cmd *cobra.Command, args parsedArgs, context *genericclioptions.Context) (completions []string) {
	completions = make([]string, 0)
	comps := &completions

	components, _ := catalog.ListDevfileComponents("")
	for _, devfile := range components.Items {
		if args.commands[devfile.Name] {
			return nil
		}
		*comps = append(*comps, devfile.Name)
	}

	return completions
}

// ComponentNameCompletionHandler provides component name completion
var ComponentNameCompletionHandler = func(cmd *cobra.Command, args parsedArgs, context *genericclioptions.Context) (completions []string) {
	completions = make([]string, 0)
	var selector string
	if context.GetApplication() != "" {
		selector = applabels.GetSelector(context.GetApplication())
	}
	components, err := component.List(context.Client, selector)

	if err != nil {
		return completions
	}

	for _, component := range components.Items {
		// we found the component name in the list which means
		// that the component name has been already selected by the user so no need to suggest more
		if args.commands[component.Name] {
			return nil
		}
		completions = append(completions, component.Name)
	}
	return completions
}
