package completion

import (
	applabels "github.com/openshift/odo/pkg/application/labels"

	"github.com/openshift/odo/pkg/application"
	"github.com/openshift/odo/pkg/catalog"
	"github.com/openshift/odo/pkg/component"
	"github.com/openshift/odo/pkg/config"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/openshift/odo/pkg/url"
	"github.com/openshift/odo/pkg/util"
	"github.com/posener/complete"
	"github.com/spf13/cobra"
)

// AppCompletionHandler provides completion for the app commands
var AppCompletionHandler = func(cmd *cobra.Command, args parsedArgs, context *genericclioptions.Context) (completions []string) {
	completions = make([]string, 0)

	applications, err := application.List(context.Client)
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
	completions = make([]string, 0)

	urls, err := url.ListPushed(context.Client, context.Component(), context.Application)
	if err != nil {
		return completions
	}

	for _, url := range urls.Items {
		// we found the url in the list which means
		// that the url name has been already selected by the user so no need to suggest more
		if args.commands[url.Name] {
			return nil
		}
		completions = append(completions, url.Name)
	}
	return
}

// StorageDeleteCompletionHandler provides storage name completion for storage delete
var StorageDeleteCompletionHandler = func(cmd *cobra.Command, args parsedArgs, context *genericclioptions.Context) (completions []string) {
	completions = make([]string, 0)

	localConfig, err := config.New()
	if err != nil {
		return completions
	}

	storageList, err := localConfig.ListStorage()
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
	found := false

	tasks := util.NewConcurrentTasks(2)
	tasks.Add(util.ConcurrentTask{ToRun: func(errChannel chan error) {
		catalogList, _ := catalog.ListComponents(context.Client)
		for _, builder := range catalogList.Items {
			if args.commands[builder.Name] {
				found = true
				return
			}
			if len(builder.Spec.NonHiddenTags) > 0 {
				*comps = append(*comps, builder.Name)
			}
		}
	}})
	tasks.Add(util.ConcurrentTask{ToRun: func(errChannel chan error) {
		components, _ := catalog.ListDevfileComponents("")
		for _, devfile := range components.Items {
			if args.commands[devfile.Name] {
				found = true
				return
			}
			*comps = append(*comps, devfile.Name)
		}
	}})

	_ = tasks.Run()
	if found {
		return nil
	}
	return completions
}

// ComponentNameCompletionHandler provides component name completion
var ComponentNameCompletionHandler = func(cmd *cobra.Command, args parsedArgs, context *genericclioptions.Context) (completions []string) {
	completions = make([]string, 0)
	var selector string
	if context.Application != "" {
		selector = applabels.GetSelector(context.Application)
	}
	components, err := component.List(context.Client, selector, nil)

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
