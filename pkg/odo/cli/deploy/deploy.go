package deploy

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/cli/messages"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/commonflags"
	odocontext "github.com/redhat-developer/odo/pkg/odo/context"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
	scontext "github.com/redhat-developer/odo/pkg/segment/context"
	"github.com/redhat-developer/odo/pkg/version"

	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/templates"
)

// RecommendedCommandName is the recommended command name
const RecommendedCommandName = "deploy"

// DeployOptions encapsulates the options for the odo command
type DeployOptions struct {
	// Clients
	clientset *clientset.Clientset
}

var _ genericclioptions.Runnable = (*DeployOptions)(nil)

var deployExample = templates.Examples(`
  # Deploy components defined in the devfile
  %[1]s
`)

// NewDeployOptions creates a new DeployOptions instance
func NewDeployOptions() *DeployOptions {
	return &DeployOptions{}
}

func (o *DeployOptions) SetClientset(clientset *clientset.Clientset) {
	o.clientset = clientset
}

func (o *DeployOptions) PreInit() string {
	return messages.DeployInitializeExistingComponent
}

// Complete DeployOptions after they've been created
func (o *DeployOptions) Complete(ctx context.Context, cmdline cmdline.Cmdline, args []string) (err error) {
	return nil
}

// Validate validates the DeployOptions based on completed values
func (o *DeployOptions) Validate(ctx context.Context) error {
	devfileObj := odocontext.GetDevfileObj(ctx)
	if devfileObj == nil {
		return genericclioptions.NewNoDevfileError(odocontext.GetWorkingDirectory(ctx))
	}
	return nil
}

// Run contains the logic for the odo command
func (o *DeployOptions) Run(ctx context.Context) error {
	var (
		devfileObj  = odocontext.GetDevfileObj(ctx)
		devfilePath = odocontext.GetDevfilePath(ctx)
		path        = filepath.Dir(devfilePath)
		devfileName = odocontext.GetComponentName(ctx)
		appName     = odocontext.GetApplication(ctx)
		namespace   = odocontext.GetNamespace(ctx)
	)

	scontext.SetComponentType(ctx, component.GetComponentTypeFromDevfileMetadata(devfileObj.Data.GetMetadata()))
	scontext.SetLanguage(ctx, devfileObj.Data.GetMetadata().Language)
	scontext.SetProjectType(ctx, devfileObj.Data.GetMetadata().ProjectType)
	scontext.SetDevfileName(ctx, devfileName)
	// Output what the command is doing / information
	log.Title("Deploying the application using "+devfileName+" Devfile",
		"Namespace: "+namespace,
		"odo version: "+version.VERSION)

	// Run actual deploy command to be used
	err := o.clientset.DeployClient.Deploy(ctx, o.clientset.FS, *devfileObj, path, appName, devfileName)

	if err == nil {
		log.Info("\nYour Devfile has been successfully deployed")
	}

	return err
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
	clientset.Add(deployCmd, clientset.INIT, clientset.DEPLOY, clientset.FILESYSTEM)

	// Add a defined annotation in order to appear in the help menu
	deployCmd.Annotations["command"] = "main"
	deployCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)
	commonflags.UseVariablesFlags(deployCmd)
	return deployCmd
}
