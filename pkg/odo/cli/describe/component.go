package describe

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/v2/pkg/devfile/parser"
	"github.com/devfile/library/v2/pkg/devfile/parser/data/v2/common"
	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubectl/pkg/util/templates"
	"k8s.io/utils/pointer"

	"github.com/redhat-developer/odo/pkg/api"
	"github.com/redhat-developer/odo/pkg/component/describe"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/log"
	clierrors "github.com/redhat-developer/odo/pkg/odo/cli/errors"
	"github.com/redhat-developer/odo/pkg/odo/cli/feature"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/commonflags"
	fcontext "github.com/redhat-developer/odo/pkg/odo/commonflags/context"
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

func (o *ComponentOptions) UseDevfile(ctx context.Context, cmdline cmdline.Cmdline, args []string) bool {
	return o.nameFlag == ""
}

func (o *ComponentOptions) Complete(ctx context.Context, cmdline cmdline.Cmdline, args []string) (err error) {
	platform := fcontext.GetPlatform(ctx, commonflags.PlatformCluster)

	// 1. Name is not passed, and odo has access to devfile.yaml; Name is not passed so we assume that odo has access to the devfile.yaml
	if o.nameFlag == "" {
		if platform == commonflags.PlatformCluster && len(o.namespaceFlag) > 0 {
			return errors.New("--namespace can be used only with --name")
		}
		devfileObj := odocontext.GetEffectiveDevfileObj(ctx)
		if devfileObj == nil {
			return genericclioptions.NewNoDevfileError(odocontext.GetWorkingDirectory(ctx))
		}
		return nil
	}

	// 2. Name is passed, and odo does not have access to devfile.yaml; if Name is passed, then we assume that odo does not have access to the devfile.yaml
	if o.clientset.KubernetesClient != nil {
		if o.namespaceFlag != "" {
			o.clientset.KubernetesClient.SetNamespace(o.namespaceFlag)
		} else {
			o.namespaceFlag = o.clientset.KubernetesClient.GetCurrentNamespace()
		}
	}
	return nil
}

func (o *ComponentOptions) Validate(ctx context.Context) (err error) {
	switch fcontext.GetPlatform(ctx, commonflags.PlatformCluster) {
	case commonflags.PlatformCluster:
		if o.clientset.KubernetesClient == nil {
			log.Warning(kclient.NewNoConnectionError())
		}
	case commonflags.PlatformPodman:
		if o.namespaceFlag != "" {
			log.Warning("--namespace flag ignored on Podman")
		}
	}

	return nil
}

// Run contains the logic for the odo command
func (o *ComponentOptions) Run(ctx context.Context) error {
	result, devfileObj, err := o.run(ctx)
	if err != nil {
		if clierrors.AsWarning(err) {
			log.Warning(err.Error())
		} else {
			return err
		}
	}
	return printHumanReadableOutput(ctx, result, devfileObj)
}

// RunForJsonOutput contains the logic for the JSON Output
func (o *ComponentOptions) RunForJsonOutput(ctx context.Context) (out interface{}, err error) {
	result, _, err := o.run(ctx) // TODO(feloy) handle warning
	if clierrors.AsWarning(err) {
		err = nil
	}
	return result, err
}

func (o *ComponentOptions) run(ctx context.Context) (result api.Component, devfileObj *parser.DevfileObj, err error) {
	if o.nameFlag != "" {
		return describe.DescribeNamedComponent(ctx, o.nameFlag, o.clientset.KubernetesClient, o.clientset.PodmanClient)
	}
	return describe.DescribeDevfileComponent(ctx, o.clientset.KubernetesClient, o.clientset.PodmanClient, o.clientset.StateClient)
}

func printHumanReadableOutput(ctx context.Context, cmp api.Component, devfileObj *parser.DevfileObj) error {
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

	withPlatformFeature := feature.IsEnabled(ctx, feature.GenericPlatformFlag)

	if withPlatformFeature && len(cmp.RunningOn) > 0 {
		log.Info("Running on:")
		for p, r := range cmp.RunningOn {
			log.Printf("%s: %s", p, r)
		}
		fmt.Println()
	}

	if len(cmp.DevControlPlane) != 0 {
		var webui string
		if feature.IsEnabled(ctx, feature.UIServer) {
			webui = "\n      Web UI: http://%[2]s:%[3]d/"
		}
		const ctrlPlaneHost = "localhost"
		log.Info("Dev Control Plane:")
		for _, dcp := range cmp.DevControlPlane {
			log.Printf(`%[1]s
      API: http://%[2]s:%[3]d/%[4]s`+webui,
				log.Sbold(dcp.Platform),
				ctrlPlaneHost, dcp.LocalPort, strings.TrimPrefix(dcp.APIServerPath, "/"))
		}
		fmt.Println()
	}

	if len(cmp.DevForwardedPorts) > 0 {
		log.Info("Forwarded ports:")
		for _, port := range cmp.DevForwardedPorts {
			details := fmt.Sprintf("%s:%d -> %s:%d", port.LocalAddress, port.LocalPort, port.ContainerName, port.ContainerPort)
			if withPlatformFeature {
				p := port.Platform
				if p == "" {
					p = commonflags.PlatformCluster
				}
				details = fmt.Sprintf("[%s] ", p) + details
			}
			if port.PortName != "" {
				details += "\n    Name: " + port.PortName
			}
			if port.Exposure != "" {
				details += "\n    Exposure: " + port.Exposure
			}
			if port.IsDebug {
				details += "\n    Debug: true"
			}
			log.Printf(details)
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

	if cmp.DevfileData != nil && len(cmp.DevfileData.Commands) != 0 {
		log.Info("Commands:")
		for _, cmd := range cmp.DevfileData.Commands {
			item := cmd.Name
			if pointer.BoolDeref(cmd.IsDefault, false) {
				item = log.Sbold(cmd.Name)
			}
			if cmd.Type != "" {
				item += fmt.Sprintf("\n      Type: %s", cmd.Type)
			}
			if cmd.Group != "" {
				item += fmt.Sprintf("\n      Group: %s", cmd.Group)
			}
			if cmd.CommandLine != "" {
				item += fmt.Sprintf("\n      Command Line: %q", cmd.CommandLine)
			}
			if cmd.Component != "" {
				item += fmt.Sprintf("\n      Component: %s", cmd.Component)
			}
			if cmd.ComponentType != "" {
				item += fmt.Sprintf("\n      Component Type: %s", cmd.ComponentType)
			}
			if cmd.ImageName != "" {
				item += fmt.Sprintf("\n      Image Name: %s", cmd.ImageName)
			}
			log.Printf(item)
		}
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

	err = listComponentsNames("OpenShift components:", devfileObj, v1alpha2.OpenshiftComponentType)
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
	components, err := devfileObj.Data.GetComponents(common.DevfileOptions{
		ComponentOptions: common.ComponentOptions{ComponentType: typ},
	})
	if err != nil {
		return err
	}
	if len(components) == 0 {
		return nil
	}
	log.Info(title)
	for _, component := range components {
		printmsg := component.Name
		if component.Container != nil && component.Container.GetMountSources() {
			printmsg += fmt.Sprintf("\n    Source Mapping: %s", component.Container.SourceMapping)
		}
		log.Printf(printmsg)
	}
	fmt.Println()
	return nil
}

// NewCmdComponent implements the component odo sub-command
func NewCmdComponent(ctx context.Context, name, fullName string, testClientset clientset.Clientset) *cobra.Command {
	o := NewComponentOptions()

	var componentCmd = &cobra.Command{
		Use:     name,
		Short:   "Describe a component",
		Long:    "Describe a component",
		Args:    genericclioptions.NoArgsAndSilenceJSON,
		Example: fmt.Sprintf(describeExample, fullName),
		RunE: func(cmd *cobra.Command, args []string) error {
			return genericclioptions.GenericRun(o, testClientset, cmd, args)
		},
	}
	componentCmd.Flags().StringVar(&o.nameFlag, "name", "", "Name of the component to describe, optional. By default, the component in the local devfile is described")
	componentCmd.Flags().StringVar(&o.namespaceFlag, "namespace", "", "Namespace in which to find the component to describe, optional. By default, the current namespace defined in kubeconfig is used")
	clientset.Add(componentCmd, clientset.KUBERNETES_NULLABLE, clientset.STATE, clientset.FILESYSTEM)
	if feature.IsEnabled(ctx, feature.GenericPlatformFlag) {
		clientset.Add(componentCmd, clientset.PODMAN_NULLABLE)
	}
	commonflags.UseOutputFlag(componentCmd)
	commonflags.UsePlatformFlag(componentCmd)

	return componentCmd
}
