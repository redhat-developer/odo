package component

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	ktemplates "k8s.io/kubectl/pkg/util/templates"

	"github.com/redhat-developer/odo/pkg/labels"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/cli/ui"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	odocontext "github.com/redhat-developer/odo/pkg/odo/context"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
)

// ComponentRecommendedCommandName is the recommended component sub-command name
const ComponentRecommendedCommandName = "component"

var deleteExample = ktemplates.Examples(`
# Delete the component present in the current directory from the cluster
%[1]s

# Delete the component named 'frontend' in the currently active namespace from the cluster
%[1]s --name frontend

# Delete the component named 'frontend' in the 'myproject' namespace from the cluster
%[1]s --name frontend --namespace myproject
`)

type ComponentOptions struct {
	// name of the component to delete, optional
	name string

	// namespace on which to find the component to delete, optional, defaults to current namespace
	namespace string

	// forceFlag forces deletion
	forceFlag bool

	// waitFlag waits for deletion of all resources
	waitFlag bool

	// Context
	*genericclioptions.Context

	// Clients
	clientset *clientset.Clientset
}

var _ genericclioptions.Runnable = (*ComponentOptions)(nil)

// NewComponentOptions returns new instance of ComponentOptions
func NewComponentOptions() *ComponentOptions {
	return &ComponentOptions{}
}

func (o *ComponentOptions) SetClientset(clientset *clientset.Clientset) {
	o.clientset = clientset
}

func (o *ComponentOptions) Complete(ctx context.Context, cmdline cmdline.Cmdline, args []string) (err error) {
	// 1. Name is not passed, and odo has access to devfile.yaml; Name is not passed so we assume that odo has access to the devfile.yaml
	if o.name == "" {
		o.Context, err = genericclioptions.New(genericclioptions.NewCreateParameters(cmdline).NeedDevfile(""))
		return err
	}
	// 2. Name is passed, and odo does not have access to devfile.yaml; if Name is passed, then we assume that odo does not have access to the devfile.yaml
	if o.namespace != "" {
		o.clientset.KubernetesClient.SetNamespace(o.namespace)
	} else {
		o.namespace = o.clientset.KubernetesClient.GetCurrentNamespace()
	}
	return nil
}

func (o *ComponentOptions) Validate(ctx context.Context) (err error) {
	return nil
}

func (o *ComponentOptions) Run(ctx context.Context) error {
	if o.name != "" {
		return o.deleteNamedComponent(ctx)
	}
	return o.deleteDevfileComponent(ctx)
}

// deleteNamedComponent deletes a component given its name
func (o *ComponentOptions) deleteNamedComponent(ctx context.Context) error {
	log.Info("Searching resources to delete, please wait...")
	list, err := o.clientset.DeleteClient.ListClusterResourcesToDelete(ctx, o.name, o.namespace)
	if err != nil {
		return err
	}
	if len(list) == 0 {
		log.Infof("No resource found for component %q in namespace %q\n", o.name, o.namespace)
		return nil
	}
	printDevfileComponents(o.name, o.namespace, list)
	if o.forceFlag || ui.Proceed("Are you sure you want to delete these resources?") {
		failed := o.clientset.DeleteClient.DeleteResources(list, o.waitFlag)
		for _, fail := range failed {
			log.Warningf("Failed to delete the %q resource: %s\n", fail.GetKind(), fail.GetName())
		}
		log.Infof("The component %q is successfully deleted from namespace %q", o.name, o.namespace)
		return nil
	}

	log.Error("Aborting deletion of component")
	return nil
}

// deleteDevfileComponent deletes all the components defined by the devfile in the current directory
func (o *ComponentOptions) deleteDevfileComponent(ctx context.Context) error {
	devfileObj := o.DevfileObj

	componentName := o.GetComponentName()

	namespace := odocontext.GetNamespace(ctx)
	appName := odocontext.GetApplication(ctx)

	log.Info("Searching resources to delete, please wait...")
	isInnerLoopDeployed, devfileResources, err := o.clientset.DeleteClient.ListResourcesToDeleteFromDevfile(devfileObj, appName, componentName, labels.ComponentAnyMode)
	if err != nil {
		return err
	}
	if len(devfileResources) == 0 {
		log.Infof("No resource found for component %q in namespace %q\n", componentName, namespace)
		return nil
	}
	// Print all the resources that odo will attempt to delete
	printDevfileComponents(componentName, namespace, devfileResources)

	if o.forceFlag || ui.Proceed(fmt.Sprintf("Are you sure you want to delete %q and all its resources?", componentName)) {
		// Get a list of component's resources present on the cluster
		clusterResources, _ := o.clientset.DeleteClient.ListClusterResourcesToDelete(ctx, componentName, namespace)
		// Get a list of component's resources absent from the devfile, but present on the cluster
		remainingResources := listResourcesMissingFromDevfilePresentOnCluster(componentName, devfileResources, clusterResources)

		// if innerloop deployment resource is present, then execute preStop events
		if isInnerLoopDeployed {
			err = o.clientset.DeleteClient.ExecutePreStopEvents(devfileObj, appName, componentName)
			if err != nil {
				log.Errorf("Failed to execute preStop events")
			}
		}

		// delete all the resources
		failed := o.clientset.DeleteClient.DeleteResources(devfileResources, o.waitFlag)
		for _, fail := range failed {
			log.Warningf("Failed to delete the %q resource: %s\n", fail.GetKind(), fail.GetName())
		}
		log.Infof("The component %q is successfully deleted from namespace %q", componentName, namespace)

		if len(remainingResources) != 0 {
			log.Printf("There are still resources left in the cluster that might be belonging to the deleted component.")
			for _, resource := range remainingResources {
				fmt.Printf("\t- %s: %s\n", resource.GetKind(), resource.GetName())
			}
			log.Infof("If you want to delete those, execute `odo delete component --name %s --namespace %s`", componentName, namespace)
		}
		return nil
	}

	log.Error("Aborting deletion of component")

	return nil
}

// listResourcesMissingFromDevfilePresentOnCluster returns a list of resources belonging to a component name that are present on cluster, but missing from devfile
func listResourcesMissingFromDevfilePresentOnCluster(componentName string, devfileResources, clusterResources []unstructured.Unstructured) []unstructured.Unstructured {
	var remainingResources []unstructured.Unstructured
	// get resources present in k8sResources(present on the cluster) but not in devfileResources(not present in the devfile)
	for _, k8sresource := range clusterResources {
		var present bool
		for _, dresource := range devfileResources {
			//  skip if the cluster and devfile resource are same OR if the cluster resource is the component's Endpoints resource
			if reflect.DeepEqual(dresource, k8sresource) || (k8sresource.GetKind() == "Endpoints" && strings.Contains(k8sresource.GetName(), componentName)) {
				present = true
				break
			}
		}
		if !present {
			remainingResources = append(remainingResources, k8sresource)
		}
	}
	return remainingResources
}

// printDevfileResources prints the devfile components for ComponentOptions.deleteDevfileComponent
func printDevfileComponents(componentName, namespace string, k8sResources []unstructured.Unstructured) {
	log.Infof("This will delete %q from the namespace %q.", componentName, namespace)

	if len(k8sResources) != 0 {
		log.Printf("The component contains the following resources that will get deleted:")
		for _, resource := range k8sResources {
			fmt.Printf("\t- %s: %s\n", resource.GetKind(), resource.GetName())
		}
	}
}

// NewCmdComponent implements the component odo sub-command
func NewCmdComponent(name, fullName string) *cobra.Command {
	o := NewComponentOptions()

	var componentCmd = &cobra.Command{
		Use:     name,
		Short:   "Delete component",
		Long:    "Delete component",
		Args:    genericclioptions.NoArgsAndSilenceJSON,
		Example: fmt.Sprintf(deleteExample, fullName),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}
	componentCmd.Flags().StringVar(&o.name, "name", "", "Name of the component to delete, optional. By default, the component described in the local devfile is deleted")
	componentCmd.Flags().StringVar(&o.namespace, "namespace", "", "Namespace in which to find the component to delete, optional. By default, the current namespace defined in kubeconfig is used")
	componentCmd.Flags().BoolVarP(&o.forceFlag, "force", "f", false, "Delete component without prompting")
	componentCmd.Flags().BoolVarP(&o.waitFlag, "wait", "w", false, "Wait for deletion of all dependent resources")
	clientset.Add(componentCmd, clientset.DELETE_COMPONENT, clientset.KUBERNETES)

	return componentCmd
}
