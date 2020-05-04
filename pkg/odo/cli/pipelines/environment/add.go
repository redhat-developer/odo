package environment

import (
	"fmt"

	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/openshift/odo/pkg/pipelines"
	"github.com/openshift/odo/pkg/pipelines/ioutils"
	"github.com/spf13/cobra"

	ktemplates "k8s.io/kubernetes/pkg/kubectl/util/templates"
)

const (
	// AddEnvRecommendedCommandName the recommended command name
	AddEnvRecommendedCommandName = "add"
)

var (
	addEnvExample = ktemplates.Examples(`
	# Add a new environment to GitOps
	%[1]s 
	`)

	addEnvLongDesc  = ktemplates.LongDesc(`Add a new environment to the GitOps repository`)
	addEnvShortDesc = `Add a new environment`
)

// AddEnvParameters encapsulates the parameters for the odo pipelines init command.
type AddEnvParameters struct {
	envName  string
	output   string
	manifest string
	// generic context options common to all commands
	*genericclioptions.Context
}

// NewAddEnvParameters bootstraps a AddEnvParameters instance.
func NewAddEnvParameters() *AddEnvParameters {
	return &AddEnvParameters{}
}

// Complete completes AddEnvParameters after they've been created.
//
// If the prefix provided doesn't have a "-" then one is added, this makes the
// generated environment names nicer to read.
func (eo *AddEnvParameters) Complete(name string, cmd *cobra.Command, args []string) error {
	return nil
}

// Validate validates the parameters of the EnvParameters.
func (eo *AddEnvParameters) Validate() error {
	return nil
}

// Run runs the project bootstrap command.
func (eo *AddEnvParameters) Run() error {
	options := pipelines.EnvParameters{
		EnvName:          eo.envName,
		ManifestFilename: eo.manifest,
	}
	return pipelines.AddEnv(&options, ioutils.NewFilesystem())
}

// NewCmdAddEnv creates the project add environment command.
func NewCmdAddEnv(name, fullName string) *cobra.Command {
	o := NewAddEnvParameters()

	addEnvCmd := &cobra.Command{
		Use:     name,
		Short:   addEnvShortDesc,
		Long:    addEnvLongDesc,
		Example: fmt.Sprintf(addEnvExample, fullName),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	addEnvCmd.Flags().StringVar(&o.envName, "env-name", "", "name of the environment/namespace")
	addEnvCmd.MarkFlagRequired("env-name")
	addEnvCmd.Flags().StringVar(&o.manifest, "manifest", "pipelines.yaml", "path to manifest file")
	return addEnvCmd
}
