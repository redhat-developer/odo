package deploy

import (
	"fmt"
	"path/filepath"

	"github.com/redhat-developer/odo/pkg/devfile/adapters"
	"github.com/redhat-developer/odo/pkg/devfile/adapters/kubernetes"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/templates"
)

// RecommendedCommandName is the recommended command name
const RecommendedCommandName = "deploy"

// DeployOptions encapsulates the options for the odo command
type DeployOptions struct {
	// Context
	*genericclioptions.Context

	// Flags
	contextFlag string
}

var deployExample = templates.Examples(`
  # Deploy components defined in the devfile
  %[1]s
`)

// NewDeployOptions creates a new DeployOptions instance
func NewDeployOptions() *DeployOptions {
	return &DeployOptions{}
}

// Complete DeployOptions after they've been created
func (o *DeployOptions) Complete(name string, cmdline cmdline.Cmdline, args []string) (err error) {
	o.Context, err = genericclioptions.New(genericclioptions.NewCreateParameters(cmdline).NeedDevfile(o.contextFlag))
	if err != nil {
		return err
	}
	return
}

// Validate validates the DeployOptions based on completed values
func (o *DeployOptions) Validate() error {
	return nil
}

// Run contains the logic for the odo command
func (o *DeployOptions) Run() error {
	platformContext := kubernetes.KubernetesContext{
		Namespace: o.KClient.GetCurrentNamespace(),
	}

	devfileHandler, err := adapters.NewComponentAdapter(o.EnvSpecificInfo.GetName(), filepath.Dir(o.EnvSpecificInfo.GetDevfilePath()), o.GetApplication(), o.EnvSpecificInfo.GetDevfileObj(), platformContext)
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
	odoutil.AddContextFlag(deployCmd, &o.contextFlag)
	return deployCmd
}
