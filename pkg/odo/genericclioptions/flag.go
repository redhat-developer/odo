package genericclioptions

import (
	pkgUtil "github.com/redhat-developer/odo/pkg/util"
	"github.com/spf13/cobra"
)

const (
	// ProjectFlagName is the name of the flag allowing a user to specify which project to operate on
	ProjectFlagName = "project"
	// ApplicationFlagName is the name of the flag allowing a user to specify which application to operate on
	ApplicationFlagName = "app"
	// ComponentFlagName is the name of the flag allowing a user to specify which component to operate on
	ComponentFlagName = "component"
	// OutputFlagName is the name of the flag allowing user to specify output format
	OutputFlagName = "output"
	// ContextFlagName is the name of the flag allowing a user to specify the location of the component settings
	ContextFlagName = "context"
)

// FlagValueIfSet retrieves the value of the specified flag if it is set for the given command
func FlagValueIfSet(cmd *cobra.Command, flagName string) string {
	flag, _ := cmd.Flags().GetString(flagName)
	return flag
}

func GetContextFlagValue(command *cobra.Command) (string, error) {
	contextDir := FlagValueIfSet(command, ContextFlagName)

	// Grab the absolute path of the configuration
	if contextDir != "" {
		fAbs, err := pkgUtil.GetAbsPath(contextDir)
		if err != nil {
			return "", err
		}
		contextDir = fAbs
	} else {
		fAbs, err := pkgUtil.GetAbsPath(".")
		if err != nil {
			return "", err
		}
		contextDir = fAbs
	}
	return contextDir, nil
}

// AddContextFlag adds `context` flag to given cobra command
func AddContextFlag(cmd *cobra.Command, setValueTo *string) {
	helpMessage := "Use given context directory as a source for component settings"
	if setValueTo != nil {
		cmd.Flags().StringVar(setValueTo, ContextFlagName, "", helpMessage)
	} else {
		cmd.Flags().String(ContextFlagName, "", helpMessage)
	}
}

// AddNowFlag adds `now` flag to given cobra command
func AddNowFlag(cmd *cobra.Command, setValueTo *bool) {
	helpMessage := "Push changes to the cluster immediately"
	if setValueTo != nil {
		cmd.Flags().BoolVar(setValueTo, "now", false, helpMessage)
	} else {
		cmd.Flags().Bool("now", false, helpMessage)
	}
}
