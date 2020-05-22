package component

import (
	"fmt"
	"path/filepath"

	devfileParser "github.com/openshift/odo/pkg/devfile/parser"
	projectCmd "github.com/openshift/odo/pkg/odo/cli/project"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/openshift/odo/pkg/odo/util/completion"
	"github.com/openshift/odo/pkg/odo/util/experimental"
	"github.com/spf13/cobra"

	odoutil "github.com/openshift/odo/pkg/odo/util"

	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

var buildCmdExample = ktemplates.Examples(`  # Build image with devfile information
%[1]s
  `)

// BuildRecommendedCommandName is the recommended build command name
const BuildRecommendedCommandName = "build"

// BuildOptions encapsulates options that build command uses
type BuildOptions struct {
	// devfile path
	componentContext string
	DevfilePath      string
	namespace        string
	tag 			 string
}

// NewBuildOptions returns new instance of BuildOptions
// with "default" values for certain values, for example, show is "false"
func NewBuildOptions() *BuildOptions {
	return &BuildOptions{}
}

// Complete completes push args
func (bo *BuildOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	bo.DevfilePath = filepath.Join(bo.componentContext, bo.DevfilePath)

	return
}

// Validate validates the push parameters
func (bo *BuildOptions) Validate() (err error) {
	return
}

// Run has the logic to perform the required actions as part of command
func (bo *BuildOptions) Run() (err error) {
	// TODO: implement support for the STACK_ROOT variable 

	// Parse devfile
	devObj, err := devfileParser.Parse(bo.DevfilePath)
	if err != nil {
		return err
	}
	component := range devObj.Data.GetComponents()
	fmt.Println(component.Dockerfile)
	if component.Dockerfile != nil {
		dockerfilePath := component.Dockerfile.Path
		// TODO: if path is relative, concatinate with volume stack_root var
		// 		 if remote download it to a hidden folder

		// TODO: run docker build using that dockefile from the current context
		// .     - implement a buildComponent fucntion on each adapter (docker and kube)
		// pass bo.Tag to it


		
	}
	return
}

// NewCmdBuild implements the push odo command
func NewCmdBuild(name, fullName string) *cobra.Command {
	bo := NewBuildOptions()

	var buildCmd = &cobra.Command{
		Use:         fmt.Sprintf("%s [component name]", name),
		Short:       "Build image for component",
		Long:        `Build image for component`,
		Example:     fmt.Sprintf(buildCmdExample, fullName),
		Args:        cobra.MaximumNArgs(1),
		Annotations: map[string]string{"command": "component"},
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(bo, cmd, args)
		},
	}
	genericclioptions.AddContextFlag(buildCmd, &bo.componentContext)

	// enable devfile flag if experimental mode is enabled
	if experimental.IsExperimentalModeEnabled() {
		buildCmd.Flags().StringVar(&bo.DevfilePath, "devfile", "./devfile.yaml", "Path to a devfile.yaml")
		buildCmd.Flags().StringVar(&bo.tag, "tag", "", "Tag used to build the image")
	}

	//Adding `--project` flag
	projectCmd.AddProjectFlag(buildCmd)

	buildCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)
	completion.RegisterCommandHandler(buildCmd, completion.ComponentNameCompletionHandler)
	completion.RegisterCommandFlagHandler(buildCmd, "context", completion.FileCompletionHandler)

	return buildCmd
}
