package describe

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/golang/glog"
	"github.com/openshift/odo/pkg/catalog"
	"github.com/openshift/odo/pkg/devfile"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/machineoutput"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/openshift/odo/pkg/odo/util/experimental"
	pkgUtil "github.com/openshift/odo/pkg/util"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	ktemplates "k8s.io/kubernetes/pkg/kubectl/util/templates"
)

const componentRecommendedCommandName = "component"

var (
	componentExample = ktemplates.Examples(`  # Describe a component
    %[1]s nodejs`)

	componentLongDesc = ktemplates.LongDesc(`Describe a component type.

This describes the component and its' associated starter projects.
`)
)

// DescribeComponentOptions encapsulates the options for the odo catalog describe component command
type DescribeComponentOptions struct {
	// name of the component to describe, from command arguments
	componentName string
	// if devfile components with name that matches arg[0]
	devfileComponents []catalog.DevfileComponentType
	// if componentName is a classic/odov1 component
	component string
	// generic context options common to all commands
	*genericclioptions.Context
}

// NewDescribeComponentOptions creates a new DescribeComponentOptions instance
func NewDescribeComponentOptions() *DescribeComponentOptions {
	return &DescribeComponentOptions{}
}

// Complete completes DescribeComponentOptions after they've been created
func (o *DescribeComponentOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	o.Context = genericclioptions.NewContext(cmd, true)
	o.componentName = args[0]
	catalogList, err := catalog.ListComponents(o.Client)
	if err != nil {
		if experimental.IsExperimentalModeEnabled() {
			glog.V(4).Info("Please log in to an OpenShift cluster to list OpenShift/s2i components")
		} else {
			return err
		}
	}
	for _, image := range catalogList.Items {
		if image.Name == o.componentName {
			o.component = image.Name
		}
	}

	if experimental.IsExperimentalModeEnabled() {
		catalogDevfileList, err := catalog.ListDevfileComponents()
		if err != nil {
			return err
		}
		for _, devfileComponent := range catalogDevfileList.Items {
			if devfileComponent.Name == o.componentName {
				o.devfileComponents = append(o.devfileComponents, devfileComponent)
			}
		}
	}

	return
}

// Validate validates the DescribeComponentOptions based on completed values
func (o *DescribeComponentOptions) Validate() (err error) {
	if len(o.devfileComponents) == 0 && o.component == "" {
		return errors.Errorf("No components with the name \"%s\" found", o.componentName)
	}

	return nil
}

// Run contains the logic for the command associated with DescribeComponentOptions
func (o *DescribeComponentOptions) Run() (err error) {
	w := tabwriter.NewWriter(os.Stdout, 5, 2, 3, ' ', tabwriter.TabIndent)
	if log.IsJSON() {
		if len(o.devfileComponents) > 0 {
			for _, devfileComponent := range o.devfileComponents {
				data, err := pkgUtil.DownloadFileInMemory(devfileComponent.Registry + devfileComponent.Link)
				if err != nil {
					return errors.Errorf("Failed to download devfile.yaml for devfile component: %v", err)
				}
				devObj, err := devfile.ParseInMemory(data)
				if err != nil {
					return err
				}

				machineoutput.OutputSuccess(devObj)
			}
		}
	} else {
		if len(o.devfileComponents) > 1 {
			fmt.Fprintln(w, "WARNING: There are multiple components named \""+o.componentName+"\" in different multiple devfile registries.\n")
		}
		if len(o.devfileComponents) > 0 {
			fmt.Fprintln(w, "Devfile Component(s):")

			for _, devfileComponent := range o.devfileComponents {
				fmt.Fprintln(w, "\n* Registry: "+devfileComponent.Registry)
				data, err := pkgUtil.DownloadFileInMemory(devfileComponent.Registry + devfileComponent.Link)
				if err != nil {
					return errors.Errorf("Failed to download devfile.yaml for devfile component: %v", err)
				}
				devObj, err := devfile.ParseInMemory(data)
				if err != nil {
					return err
				}

				yamlData, err := yaml.Marshal(devObj)
				if err != nil {
					return errors.Errorf("Failed to marshal devfile object into yaml: %v", err)
				}
				fmt.Printf("---\n%s", string(yamlData))
			}
		} else {
			fmt.Fprintln(w, "There are no Odo devfile components with the name \""+o.componentName+"\"")
		}
		if o.component != "" {
			fmt.Fprintln(w, "\nS2I Based Components:")
			fmt.Fprintln(w, "-"+o.component)
		}
		fmt.Fprintln(w)
	}

	return nil
}

// NewCmdCatalogDescribeComponent implements the odo catalog describe component command
func NewCmdCatalogDescribeComponent(name, fullName string) *cobra.Command {
	o := NewDescribeComponentOptions()
	command := &cobra.Command{
		Use:         name,
		Short:       "Describe a component",
		Long:        componentLongDesc,
		Example:     fmt.Sprintf(componentExample, fullName),
		Args:        cobra.ExactArgs(1),
		Annotations: map[string]string{"machineoutput": "json"},
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	return command
}
