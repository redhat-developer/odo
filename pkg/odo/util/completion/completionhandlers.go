package completion

import (
	"github.com/posener/complete"
	"github.com/redhat-developer/odo/pkg/application"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/project"
	"github.com/redhat-developer/odo/pkg/service"
	"github.com/spf13/cobra"
)

// ServiceCompletionHandler provides service name completion for the current project and application
var ServiceCompletionHandler = func(cmd *cobra.Command, args complete.Args, context *genericclioptions.Context) (completions []string) {
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
var ServiceClassCompletionHandler = func(cmd *cobra.Command, args complete.Args, context *genericclioptions.Context) (completions []string) {
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
var AppCompletionHandler = func(cmd *cobra.Command, args complete.Args, context *genericclioptions.Context) (completions []string) {
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
var FileCompletionHandler = func(cmd *cobra.Command, args complete.Args, context *genericclioptions.Context) (completions []string) {
	completions = append(completions, complete.PredictFiles("*").Predict(args)...)
	return
}

// ProjectNameCompletionHandler provides project name completion
var ProjectNameCompletionHandler = func(cmd *cobra.Command, args complete.Args, context *genericclioptions.Context) (completions []string) {
	completions = make([]string, 0)
	projects, err := project.List(context.Client)
	if err != nil {
		return completions
	}

	var commands []string

	// get only the user typed commands/flags and remove the cobra defined commands
	found := false
	for _, arg := range args.Completed {
		if arg == cmd.Name() {
			found = true
			continue
		}
		if found {
			commands = append(commands, arg)
		}
	}

	// extract the flags and commands
	strippedCommands, strippedFlags := getCommandsAndFlags(commands, cmd)

	// make a map of commands for faster searching
	strippedCommandsMap := make(map[string]bool)
	for _, strippedCommand := range strippedCommands {
		strippedCommandsMap[strippedCommand] = true
	}

	isIncompleteFlagSuggestion := args.LastCompleted == "--project"

	for _, project := range projects {
		// if the user is typing the project flag
		if isIncompleteFlagSuggestion {
			completions = append(completions, project.Name)
			continue
		}
		// if the flag suggestion is done
		if val, ok := strippedFlags["project"]; ok && val != "" {
			return completions
		}
		// we found the project name in the list which means
		// that the project name has been already selected by the user so no need to suggest more
		if val, ok := strippedCommandsMap[project.Name]; ok && val {
			return nil
		}
		completions = append(completions, project.Name)
	}
	return completions
}
