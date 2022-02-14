package cmdline

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/odo/util"

	dfutil "github.com/devfile/library/pkg/util"
)

type Cobra struct {
	cmd *cobra.Command
}

func NewCobra(cmd *cobra.Command) *Cobra {
	return &Cobra{
		cmd: cmd,
	}
}

func (o *Cobra) Context() context.Context {
	return o.cmd.Context()
}

func (o *Cobra) GetArgsAfterDashes(args []string) ([]string, error) {
	l := o.cmd.ArgsLenAtDash()
	if l < 0 {
		return nil, fmt.Errorf("no argument passed after dash")
	}
	return args[l:], nil
}

func (o *Cobra) GetParentName() string {
	if !o.cmd.HasParent() {
		return ""
	}
	return o.cmd.Parent().Name()
}

func (o *Cobra) GetRootName() string {
	return o.cmd.Root().Name()
}

func (o *Cobra) GetName() string {
	return o.cmd.Name()
}

func (o *Cobra) IsFlagSet(flagName string) bool {
	return o.cmd.Flags().Changed(flagName)
}

func (o *Cobra) GetWorkingDirectory() (string, error) {
	contextDir := o.FlagValueIfSet(util.ContextFlagName)

	// Grab the absolute path of the configuration
	if contextDir != "" {
		fAbs, err := dfutil.GetAbsPath(contextDir)
		if err != nil {
			return "", err
		}
		contextDir = fAbs
	} else {
		fAbs, err := dfutil.GetAbsPath(".")
		if err != nil {
			return "", err
		}
		contextDir = fAbs
	}
	return contextDir, nil
}

// FlagValueIfSet retrieves the value of the specified flag if it is set for the given command
func (o *Cobra) FlagValue(flagName string) (string, error) {
	return o.cmd.Flags().GetString(flagName)
}

// FlagValueIfSet retrieves the value of the specified flag if it is set for the given command
func (o *Cobra) FlagValueIfSet(flagName string) string {
	flag, _ := o.cmd.Flags().GetString(flagName)
	return flag
}

// checkIfConfigurationNeeded checks against a set of commands that do *NOT* need configuration.
func (o *Cobra) CheckIfConfigurationNeeded() (bool, error) {
	// Here we will check for parent commands, if the match a certain criteria, we will skip
	// using the configuration.
	//
	// For example, `odo create` should NOT check to see if there is actually a configuration yet.
	if o.cmd.HasParent() {

		// Gather necessary preliminary information
		parentCommand := o.cmd.Parent()
		rootCommand := o.cmd.Root()
		flagValue := o.FlagValueIfSet(util.ApplicationFlagName)

		// Find the first child of the command, as some groups are allowed even with non existent configuration
		firstChildCommand := getFirstChildOfCommand(o.cmd)

		// This should *never* happen, but added just to be safe
		if firstChildCommand == nil {
			return false, fmt.Errorf("Unable to get first child of command")
		}
		// Case 1 : if command is create operation just allow it
		if o.cmd.Name() == "create" && (parentCommand.Name() == "component" || parentCommand.Name() == rootCommand.Name()) {
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
		if (firstChildCommand.Name() == "component" || firstChildCommand.Name() == "service") && o.cmd.Name() == "list" {
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
		if firstChildCommand.Name() == "service" && (o.cmd.Name() == "create" || o.cmd.Name() == "delete") {
			return true, nil
		}
		// Case 11 : if command is deploy
		if o.cmd.Name() == "deploy" {
			return true, nil
		}

	} else {
		return true, nil
	}

	return false, nil
}

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

func (o *Cobra) GetKubeClient() (kclient.ClientInterface, error) {
	return kclient.New()
}

func (o *Cobra) GetFlags() map[string]string {
	flags := map[string]string{}
	o.cmd.Flags().Visit(func(f *pflag.Flag) {
		flags[f.Name] = f.Value.String()
	})
	return flags
}
