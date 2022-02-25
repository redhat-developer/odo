package dev

import (
	"fmt"
	dfutil "github.com/devfile/library/pkg/util"
	"github.com/pkg/errors"
	"github.com/redhat-developer/odo/pkg/devfile/adapters/kubernetes"
	"github.com/redhat-developer/odo/pkg/envinfo"
	"github.com/redhat-developer/odo/pkg/odo/cli/component"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/redhat-developer/odo/pkg/util"
	"github.com/spf13/cobra"
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
}

func NewDevOptions() *DevOptions {
	return &DevOptions{}
}

var devExample = templates.Examples(`
	# Deploy components to the development cluster
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
		return fmt.Errorf("the current directory doesn't contain a devfile")
	}

	envFileInfo, err := envinfo.NewEnvSpecificInfo("")
	if err != nil {
		return fmt.Errorf("unable to retrieve configuration information: %v", err)
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

	ignores := &[]string{}
	err = genericclioptions.ApplyIgnore(ignores, "")
	if err != nil {
		return err
	}
	sourcePath, err := dfutil.GetAbsPath("")
	if err != nil {
		return errors.Wrap(err, "unable to get source path")
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
