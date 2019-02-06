package component

import (
	"fmt"
	"os"

	"github.com/redhat-developer/odo/pkg/odo/util/completion"
	ktemplates "k8s.io/kubernetes/pkg/kubectl/cmd/templates"

	appCmd "github.com/redhat-developer/odo/pkg/odo/cli/application"
	projectCmd "github.com/redhat-developer/odo/pkg/odo/cli/project"

	pkgUtil "github.com/redhat-developer/odo/pkg/util"

	"github.com/redhat-developer/odo/pkg/log"

	odoutil "github.com/redhat-developer/odo/pkg/odo/util"

	"github.com/fatih/color"
	"github.com/redhat-developer/odo/pkg/component"
	"github.com/spf13/cobra"
)

// UpdateRecommendedCommandName is the recommended update command name
const UpdateRecommendedCommandName = "update"

// UpdateOptions encapsulates the update command options
type UpdateOptions struct {
	binary string
	git    string
	local  string
	ref    string
	*ComponentOptions
}

var updateCmdExample = ktemplates.Examples(`  # Change the source code path of a currently active component to local (use the current directory as a source)
	  %[1]s --local
	
	  # Change the source code path of the frontend component to local with source in ./frontend directory
	  %[1]s frontend --local ./frontend
	
	  # Change the source code path of a currently active component to git 
	  %[1]s --git https://github.com/openshift/nodejs-ex.git
	
	  # Change the source code path of the component named node-ex to git
	  %[1]s node-ex --git https://github.com/openshift/nodejs-ex.git
	
	  # Change the source code path of the component named wildfly to a binary named sample.war in ./downloads directory
	  %[1]s wildfly --binary ./downloads/sample.war
		`)

// NewUpdateOptions returns new instance of UpdateOptions
func NewUpdateOptions() *UpdateOptions {
	return &UpdateOptions{"", "", "", "", &ComponentOptions{}}
}

// Complete completes update args
func (uo *UpdateOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	err = uo.ComponentOptions.Complete(name, cmd, args)
	return
}

// Validate validates the update parameters
func (uo *UpdateOptions) Validate() (err error) {
	checkFlag := 0

	if len(uo.binary) != 0 {
		checkFlag++
	}
	if len(uo.git) != 0 {
		checkFlag++
	}
	if len(uo.local) != 0 {
		checkFlag++
	}

	if checkFlag != 1 {
		return fmt.Errorf("The source can be either --binary or --local or --git")
	}

	// if --git is not specified but --ref is still given then error has to be thrown
	if len(uo.git) == 0 && len(uo.ref) != 0 {
		return fmt.Errorf("The --ref flag is only valid for --git flag")
	}

	if len(uo.Context.Application) == 0 {
		return fmt.Errorf("Cannot update as no application is set as active")
	}

	return
}

// Run has the logic to perform the required actions as part of command
func (uo *UpdateOptions) Run() (err error) {
	stdout := color.Output

	if len(uo.git) != 0 {
		if err := component.Update(uo.Context.Client, uo.componentName, uo.Context.Application, "git", uo.git, uo.ref, stdout); err != nil {
			return err
		}
		log.Successf("The component %s was updated successfully", uo.componentName)
	} else if len(uo.local) != 0 {
		// we want to use and save absolute path for component
		dir, err := pkgUtil.GetAbsPath(uo.local)
		if err != nil {
			return err
		}
		fileInfo, err := os.Stat(dir)
		if err != nil {
			return err
		}
		if !fileInfo.IsDir() {
			return fmt.Errorf("Please provide a path to the directory")
		}
		if err = component.Update(uo.Context.Client, uo.componentName, uo.Context.Application, "local", dir, "", stdout); err != nil {
			return err
		}
		log.Successf("The component %s was updated successfully, please use 'odo push' to push your local changes", uo.componentName)
	} else if len(uo.binary) != 0 {
		path, err := pkgUtil.GetAbsPath(uo.binary)
		if err != nil {
			return err
		}
		if err = component.Update(uo.Context.Client, uo.componentName, uo.Context.Application, "binary", path, "", stdout); err != nil {
			return err
		}
		log.Successf("The component %s was updated successfully, please use 'odo push' to push your local changes", uo.componentName)
	}

	return
}

// NewCmdUpdate implements the Update odo command
func NewCmdUpdate(name, fullName string) *cobra.Command {
	uo := NewUpdateOptions()

	var updateCmd = &cobra.Command{
		Use:     name,
		Args:    cobra.MaximumNArgs(1),
		Short:   "Update the source code path of a component",
		Long:    "Update the source code path of a component",
		Example: fmt.Sprintf(updateCmdExample, fullName),
		Run: func(cmd *cobra.Command, args []string) {
			odoutil.LogErrorAndExit(uo.Complete(name, cmd, args), "")
			odoutil.LogErrorAndExit(uo.Validate(), "")
			odoutil.LogErrorAndExit(uo.Run(), "")
		},
	}

	updateCmd.Flags().StringVarP(&uo.binary, "binary", "b", "", "binary artifact")
	updateCmd.Flags().StringVarP(&uo.git, "git", "g", "", "git source")
	updateCmd.Flags().StringVarP(&uo.local, "local", "l", "", "Use local directory as a source for component.")
	updateCmd.Flags().StringVarP(&uo.ref, "ref", "r", "", "Use a specific ref e.g. commit, branch or tag of the git repository")
	// Add a defined annotation in order to appear in the help menu
	updateCmd.Annotations = map[string]string{"command": "component"}
	updateCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)

	//Adding `--application` flag
	appCmd.AddApplicationFlag(updateCmd)

	//Adding `--project` flag
	projectCmd.AddProjectFlag(updateCmd)

	completion.RegisterCommandFlagHandler(updateCmd, "local", completion.FileCompletionHandler)
	completion.RegisterCommandFlagHandler(updateCmd, "binary", completion.FileCompletionHandler)
	completion.RegisterCommandHandler(updateCmd, completion.ComponentNameCompletionHandler)

	return updateCmd
}
