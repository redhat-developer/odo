package describe

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubectl/pkg/util/templates"

	"github.com/redhat-developer/odo/pkg/api"
	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/commonflags"
	odocontext "github.com/redhat-developer/odo/pkg/odo/context"
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

	// Clients
	clientset *clientset.Clientset
}

var _ genericclioptions.Runnable = (*ComponentOptions)(nil)
var _ genericclioptions.JsonOutputter = (*ComponentOptions)(nil)

// NewComponentOptions returns new instance of ComponentOptions
func NewComponentOptions() *ComponentOptions {
	return &ComponentOptions{}
}

func (o *ComponentOptions) SetClientset(clientset *clientset.Clientset) {
	o.clientset = clientset
}

func (o *ComponentOptions) Complete(ctx context.Context, cmdline cmdline.Cmdline, args []string) (err error) {
	// 1. Name is not passed, and odo has access to devfile.yaml; Name is not passed so we assume that odo has access to the devfile.yaml
	if o.nameFlag == "" {
		if len(o.namespaceFlag) > 0 {
			return errors.New("--namespace can be used only with --name")
		}
		devfileObj := odocontext.GetDevfileObj(ctx)
		if devfileObj == nil {
			return genericclioptions.NewNoDevfileError(odocontext.GetWorkingDirectory(ctx))
		}
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

func (o *ComponentOptions) Validate(ctx context.Context) (err error) {
	return nil
}

func (o *ComponentOptions) Run(ctx context.Context) error {
	result, devfileObj, err := o.run(ctx)
	if err != nil {
		return err
	}
	return printHumanReadableOutput(result, devfileObj)
}

// Run contains the logic for the odo command
func (o *ComponentOptions) RunForJsonOutput(ctx context.Context) (out interface{}, err error) {
	result, _, err := o.run(ctx)
	return result, err
}

func (o *ComponentOptions) run(ctx context.Context) (result api.Component, devfileObj *parser.DevfileObj, err error) {
	if o.nameFlag != "" {
		return o.describeNamedComponent(ctx, o.nameFlag)
	}
	return o.describeDevfileComponent(ctx)
}

// describeNamedComponent describes a component given its name
func (o *ComponentOptions) describeNamedComponent(ctx context.Context, name string) (result api.Component, devfileObj *parser.DevfileObj, err error) {
	runningIn, err := component.GetRunningModes(ctx, o.clientset.KubernetesClient, name)
	if err != nil {
		return api.Component{}, nil, err
	}
	devfile, err := component.GetDevfileInfoFromCluster(ctx, o.clientset.KubernetesClient, name)
	if err != nil {
		return api.Component{}, nil, err
	}
	ingresses, routes, err := component.ListRoutesAndIngresses(o.clientset.KubernetesClient, name, odocontext.GetApplication(ctx))
	if err != nil {
		return api.Component{}, nil, fmt.Errorf("failed to get ingresses/routes: %w", err)
	}

	return api.Component{
		DevfileData: &api.DevfileData{
			Devfile: devfile.Data,
		},
		RunningIn: runningIn,
		ManagedBy: "odo",
		Ingresses: ingresses,
		Routes:    routes,
	}, &devfile, nil
}

// describeDevfileComponent describes the component defined by the devfile in the current directory
func (o *ComponentOptions) describeDevfileComponent(ctx context.Context) (result api.Component, devfile *parser.DevfileObj, err error) {
	var (
		devfileObj    = odocontext.GetDevfileObj(ctx)
		devfilePath   = odocontext.GetDevfilePath(ctx)
		componentName = odocontext.GetComponentName(ctx)
	)

	forwardedPorts, err := o.clientset.StateClient.GetForwardedPorts()
	if err != nil {
		return api.Component{}, nil, err
	}

	runningIn, err := component.GetRunningModes(ctx, o.clientset.KubernetesClient, componentName)
	if err != nil {
		if !errors.As(err, &component.NoComponentFoundError{}) {
			return api.Component{}, nil, err
		} else {
			// it is ok if the component is not deployed
			runningIn = nil
		}
	}
	ingresses, routes, err := component.ListRoutesAndIngresses(o.clientset.KubernetesClient, componentName, odocontext.GetApplication(ctx))
	if err != nil {
		return api.Component{}, nil, fmt.Errorf("failed to get ingresses/routes: %w", err)
	}

	return api.Component{
		DevfilePath:       devfilePath,
		DevfileData:       api.GetDevfileData(*devfileObj),
		DevForwardedPorts: forwardedPorts,
		RunningIn:         runningIn,
		ManagedBy:         "odo",
		Ingresses:         ingresses,
		Routes:            routes,
	}, devfileObj, nil
}

func printHumanReadableOutput(cmp api.Component, devfileObj *parser.DevfileObj) error {
	if cmp.DevfileData != nil {
		log.Describef("Name: ", cmp.DevfileData.Devfile.GetMetadata().Name)
		log.Describef("Display Name: ", cmp.DevfileData.Devfile.GetMetadata().DisplayName)
		log.Describef("Project Type: ", cmp.DevfileData.Devfile.GetMetadata().ProjectType)
		log.Describef("Language: ", cmp.DevfileData.Devfile.GetMetadata().Language)
		log.Describef("Version: ", cmp.DevfileData.Devfile.GetMetadata().Version)
		log.Describef("Description: ", cmp.DevfileData.Devfile.GetMetadata().Description)
		log.Describef("Tags: ", strings.Join(cmp.DevfileData.Devfile.GetMetadata().Tags, ", "))
		fmt.Println()
	}

	log.Describef("Running in: ", cmp.RunningIn.String())
	fmt.Println()

	if len(cmp.DevForwardedPorts) > 0 {
		log.Info("Forwarded ports:")
		for _, port := range cmp.DevForwardedPorts {
			log.Printf("%s:%d -> %s:%d", port.LocalAddress, port.LocalPort, port.ContainerName, port.ContainerPort)
		}
		fmt.Println()
	}

	log.Info("Supported odo features:")
	if cmp.DevfileData != nil && cmp.DevfileData.SupportedOdoFeatures != nil {
		log.Printf("Dev: %v", cmp.DevfileData.SupportedOdoFeatures.Dev)
		log.Printf("Deploy: %v", cmp.DevfileData.SupportedOdoFeatures.Deploy)
		log.Printf("Debug: %v", cmp.DevfileData.SupportedOdoFeatures.Debug)
	} else {
		log.Printf("Dev: Unknown")
		log.Printf("Deploy: Unknown")
		log.Printf("Debug: Unknown")
	}
	fmt.Println()

	err := listComponentsNames("Container components:", devfileObj, v1alpha2.ContainerComponentType)
	if err != nil {
		return err
	}

	err = listComponentsNames("Kubernetes components:", devfileObj, v1alpha2.KubernetesComponentType)
	if err != nil {
		return err
	}

	if len(cmp.Ingresses) != 0 {
		log.Info("Kubernetes Ingresses:")
		for _, ing := range cmp.Ingresses {
			for _, rule := range ing.Rules {
				for _, path := range rule.Paths {
					log.Printf("%s: %s%s", ing.Name, rule.Host, path)
				}
			}
			if len(ing.Rules) == 0 {
				log.Printf(ing.Name)
			}
		}
		fmt.Println()
	}

	if len(cmp.Routes) != 0 {
		log.Info("OpenShift Routes:")
		for _, route := range cmp.Routes {
			for _, rule := range route.Rules {
				for _, path := range rule.Paths {
					log.Printf("%s: %s%s", route.Name, rule.Host, path)
				}
			}
			if len(route.Rules) == 0 {
				log.Printf(route.Name)
			}
		}
		fmt.Println()
	}

	return nil
}

func listComponentsNames(title string, devfileObj *parser.DevfileObj, typ v1alpha2.ComponentType) error {
	if devfileObj == nil {
		log.Describef(title, " Unknown")
		return nil
	}
	containers, err := devfileObj.Data.GetComponents(common.DevfileOptions{
		ComponentOptions: common.ComponentOptions{ComponentType: typ},
	})
	if err != nil {
		return err
	}
	if len(containers) == 0 {
		return nil
	}
	log.Info(title)
	for _, container := range containers {
		log.Printf("%s", container.Name)
	}
	fmt.Println()
	return nil
}

// NewCmdComponent implements the component odo sub-command
func NewCmdComponent(name, fullName string) *cobra.Command {
	o := NewComponentOptions()

	var componentCmd = &cobra.Command{
		Use:     name,
		Short:   "Describe a component",
		Long:    "Describe a component",
		Args:    genericclioptions.NoArgsAndSilenceJSON,
		Example: fmt.Sprintf(describeExample, fullName),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}
	componentCmd.Flags().StringVar(&o.nameFlag, "name", "", "Name of the component to describe, optional. By default, the component in the local devfile is described")
	componentCmd.Flags().StringVar(&o.namespaceFlag, "namespace", "", "Namespace in which to find the component to describe, optional. By default, the current namespace defined in kubeconfig is used")
	clientset.Add(componentCmd, clientset.KUBERNETES, clientset.STATE)
	commonflags.UseOutputFlag(componentCmd)

	return componentCmd
}
