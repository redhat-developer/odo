package pipelines

import (
	"fmt"

	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/openshift/odo/pkg/pipelines"
	"github.com/openshift/odo/pkg/pipelines/ioutils"
	"github.com/spf13/cobra"

	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

const (
	// BuildRecommendedCommandName the recommended command name
	BuildRecommendedCommandName = "build"
)

var (
	buildExample = ktemplates.Examples(`
	# Build files from pipelines
	%[1]s 
	`)

	buildLongDesc  = ktemplates.LongDesc(`Build GitOps pipelines files`)
	buildShortDesc = `Build pipelines files`
)

// BuildParameters encapsulates the parameters for the odo pipelines build command.
type BuildParameters struct {
	pipelinesFilePath string
	output            string // path to add Gitops resources
	// generic context options common to all commands
	*genericclioptions.Context
}

// NewBuildParameters bootstraps a BuildParameters instance.
func NewBuildParameters() *BuildParameters {
	return &BuildParameters{}
}

// Complete completes BuildParameters after they've been created.
func (io *BuildParameters) Complete(name string, cmd *cobra.Command, args []string) error {
	return nil
}

// Validate validates the parameters of the BuildParameters.
func (io *BuildParameters) Validate() error {
	return nil
}

// Run runs the project bootstrap command.
func (io *BuildParameters) Run() error {
	options := pipelines.BuildParameters{
		PipelinesFilePath: io.pipelinesFilePath,
		OutputPath:        io.output,
	}
	err := pipelines.BuildResources(&options, ioutils.NewFilesystem())
	if err != nil {
		return err
	}
	log.Success("Built successfully.")
	return nil
}

// NewCmdBuild creates the pipelines build command.
func NewCmdBuild(name, fullName string) *cobra.Command {
	o := NewBuildParameters()
	buildCmd := &cobra.Command{
		Use:     name,
		Short:   buildShortDesc,
		Long:    buildLongDesc,
		Example: fmt.Sprintf(buildExample, fullName),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	buildCmd.Flags().StringVar(&o.output, "output", ".", "folder path to add Gitops resources")
	buildCmd.Flags().StringVar(&o.pipelinesFilePath, "pipelines-file", "pipelines.yaml", "path to pipelines file")
	return buildCmd
}
