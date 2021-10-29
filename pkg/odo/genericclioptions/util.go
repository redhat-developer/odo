package genericclioptions

import (
	"fmt"

	"github.com/openshift/odo/pkg/component"
	pkgUtil "github.com/openshift/odo/pkg/util"
	"github.com/spf13/cobra"
)

// getFirstChildOfCommand gets the first child command of the root command of command
func getFirstChildOfCommand(command *cobra.Command) *cobra.Command {
	// If command does not have a parent no point checking
	if command.HasParent() {
		// Get the root command and set current command and its parent
		rootCommand := command.Root()
		parentCommand := command.Parent()
		mainCommand := command
		for {
			// if parent is root, then we have our first child in c
			if parentCommand == rootCommand {
				return mainCommand
			}
			// Traverse backwards making current command as the parent and parent as the grandparent
			mainCommand = parentCommand
			parentCommand = mainCommand.Parent()
		}
	}
	return nil
}

// checkProjectCreateOrDeleteOnlyOnInvalidNamespace errors out if user is trying to create or delete something other than project
// errFormatForCommand must contain one %s
func checkProjectCreateOrDeleteOnlyOnInvalidNamespace(command *cobra.Command, errFormatForCommand string) error {
	// do not error out when its odo delete -a, so that we let users delete the local config on missing namespace
	if command.HasParent() && command.Parent().Name() != "project" && (command.Name() == "create" || (command.Name() == "delete" && !command.Flags().Changed("all"))) {
		return fmt.Errorf(errFormatForCommand, command.Root().Name())
	}
	return nil
}

// checkProjectCreateOrDeleteOnlyOnInvalidNamespaceNoFmt errors out if user is trying to create or delete something other than project
// compare to checkProjectCreateOrDeleteOnlyOnInvalidNamespace, no %s is needed
func checkProjectCreateOrDeleteOnlyOnInvalidNamespaceNoFmt(command *cobra.Command, errFormatForCommand string) error {
	// do not error out when its odo delete -a, so that we let users delete the local config on missing namespace
	if command.HasParent() && command.Parent().Name() != "project" && (command.Name() == "create" || command.Name() == "push" || (command.Name() == "delete" && !command.Flags().Changed("all"))) {
		return fmt.Errorf(errFormatForCommand)
	}
	return nil
}

// checkComponentExistsOrFail checks if the specified component exists with the given context and returns error if not.
func (o *internalCxt) checkComponentExistsOrFail(cmp string) error {
	exists, err := component.Exists(o.KClient, cmp, o.application)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("Component %v does not exist in application %s", cmp, o.application)
	}
	return nil
}

// ApplyIgnore will take the current ignores []string and append the mandatory odo-file-index.json and
// .git ignores; or find the .odoignore/.gitignore file in the directory and use that instead.
func ApplyIgnore(ignores *[]string, sourcePath string) (err error) {
	if len(*ignores) == 0 {
		rules, err := pkgUtil.GetIgnoreRulesFromDirectory(sourcePath)
		if err != nil {
			return err
		}
		*ignores = append(*ignores, rules...)
	}

	indexFile := pkgUtil.GetIndexFileRelativeToContext()
	// check if the ignores flag has the index file
	if !pkgUtil.In(*ignores, indexFile) {
		*ignores = append(*ignores, indexFile)
	}

	// check if the ignores flag has the git dir
	if !pkgUtil.In(*ignores, gitDirName) {
		*ignores = append(*ignores, gitDirName)
	}

	return nil
}

// checkIfConfigurationNeeded checks against a set of commands that do *NOT* need configuration.
func checkIfConfigurationNeeded(command *cobra.Command) (bool, error) {

	// Here we will check for parent commands, if the match a certain criteria, we will skip
	// using the configuration.
	//
	// For example, `odo create` should NOT check to see if there is actually a configuration yet.
	if command.HasParent() {

		// Gather necessary preliminary information
		parentCommand := command.Parent()
		rootCommand := command.Root()
		flagValue := FlagValueIfSet(command, ApplicationFlagName)

		// Find the first child of the command, as some groups are allowed even with non existent configuration
		firstChildCommand := getFirstChildOfCommand(command)

		// This should *never* happen, but added just to be safe
		if firstChildCommand == nil {
			return false, fmt.Errorf("Unable to get first child of command")
		}
		// Case 1 : if command is create operation just allow it
		if command.Name() == "create" && (parentCommand.Name() == "component" || parentCommand.Name() == rootCommand.Name()) {
			return true, nil
		}
		// Case 2 : if command is describe or delete and app flag is used just allow it
		if (firstChildCommand.Name() == "describe" || firstChildCommand.Name() == "delete") && len(flagValue) > 0 {
			return true, nil
		}
		// Case 3 : if command is list, just allow it
		if firstChildCommand.Name() == "list" {
			return true, nil
		}
		// Case 4 : Check if firstChildCommand is project. If so, skip validation of context
		if firstChildCommand.Name() == "project" {
			return true, nil
		}
		// Case 5 : Check if specific flags are set for specific first child commands
		if firstChildCommand.Name() == "app" {
			return true, nil
		}
		// Case 6 : Check if firstChildCommand is catalog and request is to list or search
		if firstChildCommand.Name() == "catalog" {
			return true, nil
		}
		// Case 7: Check if firstChildCommand is component and  request is list
		if (firstChildCommand.Name() == "component" || firstChildCommand.Name() == "service") && command.Name() == "list" {
			return true, nil
		}
		// Case 8 : Check if firstChildCommand is component and app flag is used
		if firstChildCommand.Name() == "component" && len(flagValue) > 0 {
			return true, nil
		}
		// Case 9 : Check if firstChildCommand is logout and app flag is used
		if firstChildCommand.Name() == "logout" {
			return true, nil
		}
		// Case 10: Check if firstChildCommand is service and command is create or delete. Allow it if that's the case
		if firstChildCommand.Name() == "service" && (command.Name() == "create" || command.Name() == "delete") {
			return true, nil
		}

	} else {
		return true, nil
	}

	return false, nil
}
