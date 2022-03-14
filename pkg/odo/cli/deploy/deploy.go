package deploy

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/redhat-developer/odo/pkg/devfile/location"
	"github.com/redhat-developer/odo/pkg/envinfo"
	"github.com/redhat-developer/odo/pkg/odo/cli/component"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
	"k8s.io/kubectl/pkg/util/templates"
)

// RecommendedCommandName is the recommended command name
const RecommendedCommandName = "deploy"

// DeployOptions encapsulates the options for the odo command
type DeployOptions struct {
	// Context
	*genericclioptions.Context

	// Clients
	clientset *clientset.Clientset

	// working directory
	contextDir string
}

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

// Complete DeployOptions after they've been created
func (o *DeployOptions) Complete(cmdline cmdline.Cmdline, args []string) (err error) {
	o.contextDir, err = os.Getwd()
	if err != nil {
		return err
	}

	isEmptyDir, err := location.DirIsEmpty(o.clientset.FS, o.contextDir)
	if err != nil {
		return err
	}
	if isEmptyDir {
		return errors.New("this command cannot run in an empty directory, you need to run it in a directory containing source code")
	}

	err = o.clientset.InitClient.InitDevfile(cmdline.GetFlags(), o.contextDir, func(interactiveMode bool) {
		if interactiveMode {
			fmt.Println("The current directory already contains source code. " +
				"odo will try to autodetect the language and project type in order to select the best suited Devfile for your project.")
		}
	})
	if err != nil {
		return err
	}

	o.Context, err = genericclioptions.New(genericclioptions.NewCreateParameters(cmdline).NeedDevfile(o.contextDir).CreateAppIfNeeded())
	if err != nil {
		return err
	}

	envFileInfo, err := envinfo.NewEnvSpecificInfo(o.contextDir)
	if err != nil {
		return errors.Wrap(err, "unable to retrieve configuration information")
	}
	if !envFileInfo.Exists() {
		var cmpName string
		cmpName, err = component.GatherName(o.EnvSpecificInfo.GetDevfileObj(), o.GetDevfilePath())
		if err != nil {
			return errors.Wrap(err, "unable to retrieve component name")
		}
		err = envFileInfo.SetComponentSettings(envinfo.ComponentSettings{Name: cmpName, Project: o.GetProject(), AppName: "app"})
		if err != nil {
			return errors.Wrap(err, "failed to write new env.yaml file")
		}

	} else if envFileInfo.GetComponentSettings().Project != o.GetProject() {
		err = envFileInfo.SetConfiguration("project", o.GetProject())
		if err != nil {
			return errors.Wrap(err, "failed to update project in env.yaml file")
		}
	}

	o.clientset.KubernetesClient.SetNamespace(o.GetProject())
	return
}

// Validate validates the DeployOptions based on completed values
func (o *DeployOptions) Validate() error {
	return nil
}

// Run contains the logic for the odo command
func (o *DeployOptions) Run() error {
	devfileObj := o.EnvSpecificInfo.GetDevfileObj()
	path := filepath.Dir(o.EnvSpecificInfo.GetDevfilePath())
	appName := o.GetApplication()
	return o.clientset.DeployClient.Deploy(devfileObj, path, appName)
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
	clientset.Add(deployCmd, clientset.INIT, clientset.DEPLOY)

	// Add a defined annotation in order to appear in the help menu
	deployCmd.Annotations["command"] = "utility"
	deployCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)
	return deployCmd
}
