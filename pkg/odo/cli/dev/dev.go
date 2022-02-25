package dev

import (
	"fmt"
	dfutil "github.com/devfile/library/pkg/util"
	"github.com/redhat-developer/odo/pkg/devfile/adapters/kubernetes"
	"github.com/redhat-developer/odo/pkg/envinfo"
	"github.com/redhat-developer/odo/pkg/odo/cli/component"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/redhat-developer/odo/pkg/util"
	"github.com/spf13/cobra"
	"io"
	"k8s.io/kubectl/pkg/util/templates"
	"os"
	"path/filepath"
)

// RecommendedCommandName is the recommended command name
const RecommendedCommandName = "dev"

type DevOptions struct {
	// Context
	*genericclioptions.Context

	// Clients
	clientset *clientset.Clientset

	// Variables
	ignorePaths []string
	out         io.Writer
}

func NewDevOptions() *DevOptions {
	return &DevOptions{
		out: os.Stdout,
	}
}

var devExample = templates.Examples(`
	# Deploy component to the development cluster
	%[1]s
`)

func (o *DevOptions) SetClientset(clientset *clientset.Clientset) {
	o.clientset = clientset
}

func (o *DevOptions) Complete(cmdline cmdline.Cmdline, args []string) error {
	var err error

	o.Context, err = genericclioptions.New(genericclioptions.NewCreateParameters(cmdline).NeedDevfile(""))
	if err != nil {
		return fmt.Errorf("unable to create context: %v", err)
	}

	devfileExists := util.CheckPathExists(o.Context.GetDevfilePath())
	if !devfileExists {
		return fmt.Errorf("there is no devfile.yaml in the current directory")
	}
	fmt.Fprintf(o.out, "Using devfile.yaml from the current directory.\n")

	envFileInfo, err := envinfo.NewEnvSpecificInfo("")
	if err != nil {
		return fmt.Errorf("unable to retrieve configuration information: %v", err)
	}
	if !envFileInfo.Exists() {
		var cmpName string
		cmpName, err = component.GatherName(o.EnvSpecificInfo.GetDevfileObj(), o.GetDevfilePath())
		if err != nil {
			return fmt.Errorf("unable to retrieve component name: %w", err)
		}
		err = envFileInfo.SetComponentSettings(envinfo.ComponentSettings{Name: cmpName, Project: o.GetProject(), AppName: "app"})
		if err != nil {
			return fmt.Errorf("failed to write new env.yaml file: %w", err)
		}
	} else if envFileInfo.GetComponentSettings().Project != o.GetProject() {
		err = envFileInfo.SetConfiguration("project", o.GetProject())
		if err != nil {
			return fmt.Errorf("failed to update project in env.yaml file: %w", err)
		}
	}

	ignores := &[]string{}
	err = genericclioptions.ApplyIgnore(ignores, "")
	if err != nil {
		return err
	}
	sourcePath, err := dfutil.GetAbsPath("")
	if err != nil {
		return fmt.Errorf("unable to get source path: %w", err)
	}

	o.ignorePaths = dfutil.GetAbsGlobExps(sourcePath, *ignores)
	return nil
}

func (o *DevOptions) Validate() error {
	var err error
	return err
}

func (o *DevOptions) Run() error {
	var err error
	var platformContext = kubernetes.KubernetesContext{
		Namespace: o.Context.GetProject(),
	}
	var path = filepath.Dir(o.Context.EnvSpecificInfo.GetDevfilePath())

	fmt.Fprintf(o.out, "Starting your application on cluster in developer mode\n")
	err = o.clientset.DevClient.Start(o.Context.EnvSpecificInfo.GetDevfileObj(), platformContext, o.ignorePaths, path, os.Stdout)
	return err
}

// NewCmdDev implements the odo dev command
func NewCmdDev(name, fullName string) *cobra.Command {
	o := NewDevOptions()
	devCmd := &cobra.Command{
		Use:     name,
		Short:   "Deploy component to development cluster",
		Long:    "Deploy the component to a development cluster. odo dev is a long running command that will automatically sync your source to the cluster",
		Example: fmt.Sprintf(devExample, fullName),
		Args:    cobra.MaximumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	// Add a defined annotation in order to appear in the help menu
	devCmd.Annotations = map[string]string{"command": "utility"}
	clientset.Add(devCmd, clientset.DEV)
	devCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)

	return devCmd
}
