package describe

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubectl/pkg/util/templates"

	"github.com/redhat-developer/odo/pkg/api"
	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/machineoutput"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
)

// ComponentRecommendedCommandName is the recommended component sub-command name
const ComponentRecommendedCommandName = "component"

var describeExample = ktemplates.Examples(`
# Describe the component in the current directory
%[1]s

# Describe a component deployed in the cluster
%[1]s --name frontend --namespace myproject
`)

type ComponentOptions struct {
	// nameFlag of the component to describe, optional
	nameFlag string

	// namespaceFlag on which to find the component to describe, optional, defaults to current namespaceFlag
	namespaceFlag string

	// Context
	*genericclioptions.Context

	// Clients
	clientset *clientset.Clientset
}

// NewComponentOptions returns new instance of ComponentOptions
func NewComponentOptions() *ComponentOptions {
	return &ComponentOptions{}
}

func (o *ComponentOptions) SetClientset(clientset *clientset.Clientset) {
	o.clientset = clientset
}

func (o *ComponentOptions) Complete(cmdline cmdline.Cmdline, args []string) (err error) {
	// 1. Name is not passed, and odo has access to devfile.yaml; Name is not passed so we assume that odo has access to the devfile.yaml
	if o.nameFlag == "" {
		o.Context, err = genericclioptions.New(genericclioptions.NewCreateParameters(cmdline).NeedDevfile(""))
		if err != nil {
			return err
		}
		// this ensures that the namespace set in env.yaml is used
		o.clientset.KubernetesClient.SetNamespace(o.GetProject())
		return nil
	}
	// 2. Name is passed, and odo does not have access to devfile.yaml; if Name is passed, then we assume that odo does not have access to the devfile.yaml
	if o.namespaceFlag != "" {
		o.clientset.KubernetesClient.SetNamespace(o.namespaceFlag)
	} else {
		o.namespaceFlag = o.clientset.KubernetesClient.GetCurrentNamespace()
	}
	return nil
}

func (o *ComponentOptions) Validate() (err error) {
	return nil
}

func (o *ComponentOptions) Run(ctx context.Context) error {
	result, err := o.run(ctx)
	if err != nil {
		return err
	}
	_ = result
	return nil
}

// Run contains the logic for the odo command
func (o *ComponentOptions) RunForJsonOutput(ctx context.Context) (out interface{}, err error) {
	return o.run(ctx)
}

func (o *ComponentOptions) run(ctx context.Context) (result api.Component, err error) {
	if o.nameFlag != "" {
		return o.describeNamedComponent(o.nameFlag)
	}
	return o.describeDevfileComponent()
}

// describeNamedComponent describes a component given its name
func (o *ComponentOptions) describeNamedComponent(name string) (result api.Component, err error) {
	forwardedPorts, err := getForwardedPorts()
	if err != nil {
		return api.Component{}, err
	}
	runningIn := component.GetRunningModes(o.clientset.KubernetesClient, name, o.clientset.KubernetesClient.GetCurrentNamespace())
	return api.Component{
		ForwardedPorts: forwardedPorts,
		RunningIn:      runningIn,
		ManagedBy:      "odo",
	}, nil
}

// describeDevfileComponent describes the component defined by the devfile in the current directory
func (o *ComponentOptions) describeDevfileComponent() (result api.Component, err error) {
	devfileObj := o.EnvSpecificInfo.GetDevfileObj()
	path, err := filepath.Abs(".")
	if err != nil {
		return api.Component{}, err
	}
	forwardedPorts, err := getForwardedPorts()
	if err != nil {
		return api.Component{}, err
	}
	runningIn := component.GetRunningModes(o.clientset.KubernetesClient, devfileObj.GetMetadataName(), o.clientset.KubernetesClient.GetCurrentNamespace())
	return api.Component{
		DevfilePath:    path,
		DevfileData:    api.GetDevfileData(devfileObj),
		ForwardedPorts: forwardedPorts,
		RunningIn:      runningIn,
		ManagedBy:      "odo",
	}, nil
}

func getForwardedPorts() ([]api.ForwardedPort, error) {
	// TODO(feloy) when #5676 is done
	return nil, nil
}

// NewCmdComponent implements the component odo sub-command
func NewCmdComponent(name, fullName string) *cobra.Command {
	o := NewComponentOptions()

	var componentCmd = &cobra.Command{
		Use:     name,
		Short:   "Describe a component",
		Long:    "Describe a component",
		Args:    cobra.NoArgs,
		Example: fmt.Sprintf(describeExample, fullName),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}
	componentCmd.Flags().StringVar(&o.nameFlag, "name", "", "Name of the component to describe, optional. By default, the component in the local devfile is described")
	componentCmd.Flags().StringVar(&o.namespaceFlag, "namespace", "", "Namespace in which to find the component to describe, optional. By default, the current namespace defined in kubeconfig is used")
	clientset.Add(componentCmd, clientset.KUBERNETES)
	machineoutput.UsedByCommand(componentCmd)

	return componentCmd
}
