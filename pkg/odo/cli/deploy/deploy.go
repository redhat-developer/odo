package deploy

import (
	"fmt"
	"path/filepath"

	"github.com/openshift/odo/pkg/devfile/adapters"
	"github.com/openshift/odo/pkg/devfile/adapters/kubernetes"
	"github.com/openshift/odo/pkg/devfile/location"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	odoutil "github.com/openshift/odo/pkg/odo/util"
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/templates"
)

// RecommendedCommandName is the recommended command name
const RecommendedCommandName = "deploy"

// LoginOptions encapsulates the options for the odo command
type DeployOptions struct {
	*genericclioptions.Context
	componentContext string
}

var deployExample = templates.Examples(`
  # Deploy components defined in the devfile
  %[1]s
`)

// NewLoginOptions creates a new LoginOptions instance
func NewDeployOptions() *DeployOptions {
	return &DeployOptions{}
}

// Complete completes LoginOptions after they've been created
func (o *DeployOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	o.Context, err = genericclioptions.New(genericclioptions.CreateParameters{
		Cmd:              cmd,
		DevfilePath:      location.DevfileFilenamesProvider(o.componentContext),
		ComponentContext: o.componentContext,
	})
	if err != nil {
		return err
	}
	return
}

// Validate validates the LoginOptions based on completed values
func (o *DeployOptions) Validate() (err error) {
	return
}

// Run contains the logic for the odo command
func (o *DeployOptions) Run(cmd *cobra.Command) error {
	platformContext := kubernetes.KubernetesContext{
		Namespace: o.KClient.GetCurrentNamespace(),
	}

	devfileHandler, err := adapters.NewComponentAdapter(o.EnvSpecificInfo.GetName(), filepath.Dir(o.EnvSpecificInfo.GetDevfilePath()), o.Application, o.EnvSpecificInfo.GetDevfileObj(), platformContext)
	if err != nil {
		return err
	}

	return devfileHandler.Deploy()
}

// NewCmdDeploy implements the odo command
func NewCmdDeploy(name, fullName string) *cobra.Command {
	o := NewDeployOptions()
	deployCmd := &cobra.Command{
		Use:     name,
		Short:   "Deploy components",
		Long:    "Deploy the components defined in the devfile",
		Example: fmt.Sprintf(deployExample, fullName),
		Args:    cobra.MaximumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	// Add a defined annotation in order to appear in the help menu
	deployCmd.Annotations = map[string]string{"command": "utility"}
	deployCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)
	genericclioptions.AddContextFlag(deployCmd, &o.componentContext)
	return deployCmd
}
