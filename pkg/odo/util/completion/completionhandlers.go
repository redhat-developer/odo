package completion

import (
	applabels "github.com/redhat-developer/odo/pkg/application/labels"
	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/redhat-developer/odo/pkg/preference"
	"github.com/redhat-developer/odo/pkg/registry"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"

	"github.com/posener/complete"
	"github.com/spf13/cobra"
)

// FileCompletionHandler provides suggestions for files and directories
var FileCompletionHandler = func(cmd *cobra.Command, args parsedArgs, context *genericclioptions.Context) (completions []string) {
	completions = append(completions, complete.PredictFiles("*").Predict(args.original)...)
	return
}

// ProjectNameCompletionHandler provides project name completion
var ProjectNameCompletionHandler = func(cmd *cobra.Command, args parsedArgs, context *genericclioptions.Context) (completions []string) {
	completions = make([]string, 0)
	projects, err := context.KClient.ListProjectNames()
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

// CreateCompletionHandler provides component type completion in odo create command
var CreateCompletionHandler = func(cmd *cobra.Command, args parsedArgs, context *genericclioptions.Context) (completions []string) {
	completions = make([]string, 0)
	comps := &completions

	prefClient, err := preference.NewClient()
	if err != nil {
		odoutil.LogErrorAndExit(err, "unable to set preference, something is wrong with odo, kindly raise an issue at https://github.com/redhat-developer/odo/issues/new?template=Bug.md")
	}
	components, _ := registry.NewRegistryClient(filesystem.DefaultFs{}, prefClient).ListDevfileStacks("")
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
	components, err := component.List(context.KClient, selector)

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
