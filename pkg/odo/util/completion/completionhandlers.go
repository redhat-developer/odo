package completion

import (
	"github.com/posener/complete"
	"github.com/redhat-developer/odo/pkg/application"
	"github.com/redhat-developer/odo/pkg/catalog"
	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/project"
	"github.com/redhat-developer/odo/pkg/service"
	"github.com/redhat-developer/odo/pkg/storage"
	"github.com/redhat-developer/odo/pkg/url"
	"github.com/spf13/cobra"
)

// ServiceCompletionHandler provides service name completion for the current project and application
var ServiceCompletionHandler = func(cmd *cobra.Command, args parsedArgs, context *genericclioptions.Context) (completions []string) {
	completions = make([]string, 0)

	services, err := service.List(context.Client, context.Application)
	if err != nil {
		return completions
	}

	for _, class := range services {
		completions = append(completions, class.Name)
	}

	return
}

// ServiceClassCompletionHandler provides catalog service class name completion
var ServiceClassCompletionHandler = func(cmd *cobra.Command, args parsedArgs, context *genericclioptions.Context) (completions []string) {
	completions = make([]string, 0)
	services, err := context.Client.GetClusterServiceClasses()
	if err != nil {
		return completions
	}

	for _, class := range services {
		completions = append(completions, class.Spec.ExternalName)
	}

	return
}

// AppCompletionHandler provides completion for the app commands
var AppCompletionHandler = func(cmd *cobra.Command, args parsedArgs, context *genericclioptions.Context) (completions []string) {
	completions = make([]string, 0)

	applications, err := application.List(context.Client)
	if err != nil {
		return completions
	}

	for _, app := range applications {
		completions = append(completions, app.Name)
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
	projects, err := project.List(context.Client)
	if err != nil {
		return completions
	}

	for _, project := range projects {
		// we found the project name in the list which means
		// that the project name has been already selected by the user so no need to suggest more
		if args.commands[project.Name] {
			return nil
		}
		completions = append(completions, project.Name)
	}
	return completions
}

// URLCompletionHandler provides completion for the url commands
var URLCompletionHandler = func(cmd *cobra.Command, args parsedArgs, context *genericclioptions.Context) (completions []string) {
	completions = make([]string, 0)

	urls, err := url.List(context.Client, context.Component(), context.Application)
	if err != nil {
		return completions
	}

	for _, url := range urls {
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
	storages, err := storage.List(context.Client, context.Component(), context.Application)
	if err != nil {
		return completions
	}

	for _, storage := range storages {
		// we found the storage name in the list which means
		// that the storage name has been already selected by the user so no need to suggest more
		if args.commands[storage.Name] {
			return nil
		}
		completions = append(completions, storage.Name)
	}
	return completions
}

// StorageMountCompletionHandler provides storage name completion for storage mount
var StorageMountCompletionHandler = func(cmd *cobra.Command, args parsedArgs, context *genericclioptions.Context) (completions []string) {
	completions = make([]string, 0)
	storages, err := storage.ListUnmounted(context.Client, context.Application)
	if err != nil {
		return completions
	}

	for _, storage := range storages {
		// we found the storage name in the list which means
		// that the storage name has been already selected by the user so no need to suggest more
		if args.commands[storage.Name] {
			return nil
		}
		completions = append(completions, storage.Name)
	}
	return completions
}

// StorageUnMountCompletionHandler provides storage name completion for storage unmount
var StorageUnMountCompletionHandler = func(cmd *cobra.Command, args parsedArgs, context *genericclioptions.Context) (completions []string) {
	completions = make([]string, 0)
	storages, err := storage.ListMounted(context.Client, context.Component(), context.Application)
	if err != nil {
		return completions
	}

	for _, storage := range storages {
		// we found the storage name in the list which means
		// that the storage name has been already selected by the user so no need to suggest more
		if args.commands[storage.Name] {
			return nil
		}
		completions = append(completions, storage.Name)
	}
	return completions
}

// CreateCompletionHandler provides componet type completion in odo create command
var CreateCompletionHandler = func(cmd *cobra.Command, args parsedArgs, context *genericclioptions.Context) (completions []string) {
	completions = make([]string, 0)
	catalogList, err := catalog.List(context.Client)
	if err != nil {
		return completions
	}

	for _, builder := range catalogList {
		// we found the builder name in the list which means
		// that the builder name has been already selected by the user so no need to suggest more
		if args.commands[builder.Name] {
			return nil
		}
		completions = append(completions, builder.Name)
	}

	return completions
}

// LinkCompletionHandler provides completion for the odo link
// The function returns both components and services
var LinkCompletionHandler = func(cmd *cobra.Command, args parsedArgs, context *genericclioptions.Context) (completions []string) {
	completions = make([]string, 0)

	components, err := component.List(context.Client, context.Application)
	if err != nil {
		return completions
	}

	services, err := service.List(context.Client, context.Application)
	if err != nil {
		return completions
	}

	for _, component := range components {
		// we found the name in the list which means
		// that the name has been already selected by the user so no need to suggest more
		if val, ok := args.commands[component.Name]; ok && val {
			return nil
		}
		// we don't want to show the selected component as a target for linking, so we remove it from the suggestions
		if component.Name != context.Component() {
			completions = append(completions, component.Name)
		}
	}

	for _, service := range services {
		// we found the name in the list which means
		// that the name has been already selected by the user so no need to suggest more
		if val, ok := args.commands[service.Name]; ok && val {
			return nil
		}
		completions = append(completions, service.Name)
	}

	return completions
}

// ComponentNameCompletionHandler provides component name completion
var ComponentNameCompletionHandler = func(cmd *cobra.Command, args parsedArgs, context *genericclioptions.Context) (completions []string) {
	completions = make([]string, 0)
	components, err := component.List(context.Client, context.Application)

	if err != nil {
		return completions
	}

	for _, component := range components {
		// we found the component name in the list which means
		// that the component name has been already selected by the user so no need to suggest more
		if args.commands[component.Name] {
			return nil
		}
		completions = append(completions, component.Name)
	}
	return completions
}
