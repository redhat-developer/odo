package component

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/openshift/odo/pkg/component"
	"github.com/openshift/odo/pkg/config"
	"github.com/openshift/odo/pkg/log"
	appCmd "github.com/openshift/odo/pkg/odo/cli/application"
	projectCmd "github.com/openshift/odo/pkg/odo/cli/project"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	odoutil "github.com/openshift/odo/pkg/odo/util"
	"github.com/openshift/odo/pkg/odo/util/completion"

	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

// UpdateRecommendedCommandName is the recommended update command name
const UpdateRecommendedCommandName = "update"

// UpdateOptions encapsulates the update command options
type UpdateOptions struct {
	binary string
	git    string
	local  string
	ref    string

	*CommonPushOptions
}

var updateCmdExample = ktemplates.Examples(`  # Change the source code path of currently active component to local with source in ./frontend directory
	  %[1]s --local ./frontend
	
	  # Change the source code path of currently active component to git 
	  %[1]s --git https://github.com/openshift/nodejs-ex.git
		
	  # Change the source code path of of currently active component to a binary named sample.war in ./downloads directory
	  %[1]s --binary ./downloads/sample.war
		`)

// NewUpdateOptions returns new instance of UpdateOptions
func NewUpdateOptions() *UpdateOptions {
	return &UpdateOptions{
		CommonPushOptions: &CommonPushOptions{
			pushConfig: true, // we push everything
			forceBuild: true,
			pushSource: true,
			show:       false,
		}}
}

// Complete completes update args
func (uo *UpdateOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	uo.Context = genericclioptions.NewContext(cmd)
	uo.LocalConfigInfo, err = config.NewLocalConfigInfo(uo.componentContext)
	if err != nil {
		return errors.Wrapf(err, "failed to update component")
	}

	return
}

// Validate validates the update parameters
func (uo *UpdateOptions) Validate() (err error) {

	uo.doesComponentExist, err = component.Exists(uo.Context.Client, uo.LocalConfigInfo.GetName(), uo.LocalConfigInfo.GetApplication())
	if err != nil {
		return errors.Wrapf(err, "failed to check if component of name %s exists in application %s", uo.LocalConfigInfo.GetName(), uo.LocalConfigInfo.GetApplication())
	}

	checkFlag := 0

	if len(uo.binary) != 0 {
		checkFlag++
		uo.sourceType = config.BINARY
		uo.sourcePath = uo.binary

		if strings.HasPrefix(uo.sourcePath, fmt.Sprintf("..%c", filepath.Separator)) {
			return fmt.Errorf("%s binary needs to be inside of the context directory", uo.sourcePath)
		}
	}
	if len(uo.git) != 0 {
		checkFlag++
		uo.sourceType = config.GIT
		uo.sourcePath = uo.git
	}
	if len(uo.local) != 0 {
		checkFlag++
		uo.sourceType = config.LOCAL
		uo.sourcePath = uo.local
		srcLocInfo, err := os.Stat(uo.sourcePath)
		if err != nil {
			return fmt.Errorf("error while validating source path: %v", err.Error())
		}
		if !srcLocInfo.IsDir() {
			return fmt.Errorf("source path for component created for local source needs to be a directory")
		}
	}

	if len(uo.componentContext) == 0 {
		dir, err := os.Getwd()
		if err != nil {
			return errors.Wrapf(err, "failed to update component %s", uo.LocalConfigInfo.GetName())
		}
		uo.componentContext = dir
	}
	fileInfo, err := os.Stat(uo.componentContext)
	if err != nil {
		return err
	}
	if !fileInfo.IsDir() {
		return fmt.Errorf("Please provide a path to the directory as --context")
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

	compSettings := uo.LocalConfigInfo.GetComponentSettings()
	compSettings.SourceLocation = &uo.sourcePath
	compSettings.SourceType = &uo.sourceType
	if len(uo.ref) != 0 {
		compSettings.Ref = &uo.ref
	}

	err = uo.LocalConfigInfo.SetComponentSettings(compSettings)
	if err != nil {
		return err
	}

	if err = uo.Push(); err != nil {
		return errors.Wrap(err, "error while updating")
	}

	cmpName := uo.LocalConfigInfo.GetName()
	log.Successf("The component %s was updated successfully", cmpName)
	return
}

// NewCmdUpdate implements the Update odo command
func NewCmdUpdate(name, fullName string) *cobra.Command {
	uo := NewUpdateOptions()

	var updateCmd = &cobra.Command{
		Use:     name,
		Args:    cobra.MaximumNArgs(0),
		Short:   "Update the source code path of a component",
		Long:    "Update the source code path of a component",
		Example: fmt.Sprintf(updateCmdExample, fullName),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(uo, cmd, args)
		},
	}
	genericclioptions.AddContextFlag(updateCmd, &uo.componentContext)
	updateCmd.Flags().BoolVar(&uo.show, "show-log", false, "If enabled, logs will be shown when built")
	updateCmd.Flags().StringVarP(&uo.git, "git", "g", "", "git source")
	updateCmd.Flags().StringVarP(&uo.local, "local", "l", "", "Use local directory as a source for component.")
	updateCmd.Flags().StringVarP(&uo.ref, "ref", "r", "", "Use a specific ref e.g. commit, branch or tag of the git repository")

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
