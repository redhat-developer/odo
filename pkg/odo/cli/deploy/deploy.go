package deploy

import (
	"context"
	"errors"
	"fmt"

	"github.com/devfile/library/pkg/devfile/parser"

	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/devfile/location"
	"github.com/redhat-developer/odo/pkg/envinfo"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
	scontext "github.com/redhat-developer/odo/pkg/segment/context"
	"github.com/redhat-developer/odo/pkg/vars"
	"github.com/redhat-developer/odo/pkg/version"

	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"k8s.io/klog"
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

	// Flags
	varFileFlag string
	varsFlag    []string

	// Variables to override Devfile variables
	variables map[string]string

	// working directory
	contextDir string
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
		return errors.New("this command cannot run in an empty directory, run the command in a directory containing source code or initialize using 'odo init'")
	}

	initFlags := o.clientset.InitClient.GetFlags(cmdline.GetFlags())

	err = o.clientset.InitClient.InitDevfile(initFlags, o.contextDir,
		func(interactiveMode bool) {
			scontext.SetInteractive(cmdline.Context(), interactiveMode)
			if interactiveMode {
				fmt.Println("The current directory already contains source code. " +
					"odo will try to autodetect the language and project type in order to select the best suited Devfile for your project.")
			}
		},
		func(newDevfileObj parser.DevfileObj) error {
			return newDevfileObj.WriteYamlDevfile()
		})
	if err != nil {
		return err
	}

	o.variables, err = vars.GetVariables(o.clientset.FS, o.varFileFlag, o.varsFlag, os.LookupEnv)
	if err != nil {
		return err
	}

	o.Context, err = genericclioptions.New(genericclioptions.NewCreateParameters(cmdline).NeedDevfile(o.contextDir).WithVariables(o.variables).CreateAppIfNeeded())
	if err != nil {
		return err
	}

	// ENV.YAML
	//
	// Set the appropriate variables in the env.yaml file
	//
	// TODO: Eventually refactor with code in dev/dev.go

	envfileinfo, err := envinfo.NewEnvSpecificInfo("")
	if err != nil {
		return fmt.Errorf("unable to retrieve configuration information: %w", err)
	}

	// If the env.yaml does not exist, we will save the project name
	if !envfileinfo.Exists() {
		err = envfileinfo.SetComponentSettings(envinfo.ComponentSettings{Project: o.GetProject()})
		if err != nil {
			return fmt.Errorf("failed to write new env.yaml file: %w", err)
		}
	} else if envfileinfo.Exists() && envfileinfo.GetComponentSettings().Project != o.GetProject() {
		// If the env.yaml exists and the project is set incorrectly, we'll override it.
		klog.V(4).Info("Overriding project name in env.yaml as it's set incorrectly, new project name: ", o.GetProject())
		err = envfileinfo.SetConfiguration("project", o.GetProject())
		if err != nil {
			return fmt.Errorf("failed to update project in env.yaml file: %w", err)
		}
	}

	// END ENV.YAML

	// this ensures that odo deploy uses the namespace set in env.yaml
	o.clientset.KubernetesClient.SetNamespace(o.GetProject())
	return
}

// Validate validates the DeployOptions based on completed values
func (o *DeployOptions) Validate() error {
	return nil
}

// Run contains the logic for the odo command
func (o *DeployOptions) Run(ctx context.Context) error {
	devfileObj := o.EnvSpecificInfo.GetDevfileObj()
	devfileName := devfileObj.GetMetadataName()
	path := filepath.Dir(o.EnvSpecificInfo.GetDevfilePath())
	appName := o.GetApplication()
	namespace := o.GetProject()
	scontext.SetComponentType(ctx, component.GetComponentTypeFromDevfileMetadata(devfileObj.Data.GetMetadata()))
	scontext.SetLanguage(ctx, devfileObj.Data.GetMetadata().Language)
	scontext.SetProjectType(ctx, devfileObj.Data.GetMetadata().ProjectType)
	scontext.SetDevfileName(ctx, devfileName)
	// Output what the command is doing / information
	log.Title("Deploying the application using "+devfileName+" Devfile",
		"Namespace: "+namespace,
		"odo version: "+version.VERSION)

	// Run actual deploy command to be used
	err := o.clientset.DeployClient.Deploy(devfileObj, path, appName)

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
	deployCmd.Flags().StringArrayVar(&o.varsFlag, "var", []string{}, "Variable to override Devfile variable and variables in var-file")
	deployCmd.Flags().StringVar(&o.varFileFlag, "var-file", "", "File containing variables to override Devfile variables")
	clientset.Add(deployCmd, clientset.INIT, clientset.DEPLOY, clientset.FILESYSTEM)

	// Add a defined annotation in order to appear in the help menu
	deployCmd.Annotations["command"] = "main"
	deployCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)
	return deployCmd
}
