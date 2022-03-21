package describe

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"github.com/redhat-developer/odo/pkg/catalog"
	"github.com/redhat-developer/odo/pkg/devfile/validate"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/machineoutput"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/devfile/library/pkg/devfile/parser/data"
	parsercommon "github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	"github.com/redhat-developer/odo/pkg/devfile"

	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

const componentRecommendedCommandName = "component"

var (
	componentExample = ktemplates.Examples(`  # Describe a component
    %[1]s nodejs`)

	componentLongDesc = ktemplates.LongDesc(`Describe a component type.
This describes the component and its associated starter projects.
`)
)

// DescribeComponentOptions encapsulates the options for the odo catalog describe component command
type DescribeComponentOptions struct {
	// Context
	*genericclioptions.Context

	// Clients
	clientset *clientset.Clientset

	// Parameters
	componentName string

	// devfile components with name that matches componentName
	devfileComponents []catalog.DevfileComponentType
}

// NewDescribeComponentOptions creates a new DescribeComponentOptions instance
func NewDescribeComponentOptions() *DescribeComponentOptions {
	return &DescribeComponentOptions{}
}

func (o *DescribeComponentOptions) SetClientset(clientset *clientset.Clientset) {
	o.clientset = clientset
}

// Complete completes DescribeComponentOptions after they've been created
func (o *DescribeComponentOptions) Complete(cmdline cmdline.Cmdline, args []string) (err error) {
	o.componentName = args[0]

	o.Context, err = genericclioptions.New(genericclioptions.NewCreateParameters(cmdline).IsOffline())
	if err != nil {
		return err
	}

	catalogDevfileList, err := o.clientset.CatalogClient.ListDevfileComponents("")
	if catalogDevfileList.DevfileRegistries == nil {
		log.Warning("Please run 'odo registry add <registry name> <registry URL>' to add registry for listing devfile components\n")
	}
	if err != nil {
		return err
	}

	o.devfileComponents = listDevfileComponentsByName(catalogDevfileList, o.componentName)
	return nil
}

// Validate validates the DescribeComponentOptions based on completed values
func (o *DescribeComponentOptions) Validate() (err error) {
	if len(o.devfileComponents) == 0 {
		return fmt.Errorf("No components with the name \"%s\" found", o.componentName)
	}

	return nil
}

// DevfileComponentDescription represents the JSON output of Devfile component description
// used in odo catalog describe component <name> -o json
type DevfileComponentDescription struct {
	RegistryName string           `json:"RegistryName"`
	Devfile      data.DevfileData `json:"Devfile"`
}

// Run contains the logic for the command associated with DescribeComponentOptions
func (o *DescribeComponentOptions) Run() (err error) {
	w := tabwriter.NewWriter(os.Stdout, 5, 2, 3, ' ', tabwriter.TabIndent)
	if log.IsJSON() {
		if len(o.devfileComponents) > 0 {
			out := []DevfileComponentDescription{}

			for _, devfileComponent := range o.devfileComponents {
				devObj, err := GetDevfile(devfileComponent)
				if err != nil {
					return err
				}
				out = append(out, DevfileComponentDescription{RegistryName: devfileComponent.Registry.Name, Devfile: devObj.Data})
			}
			machineoutput.OutputSuccess(out)
		}
	} else {
		if len(o.devfileComponents) > 1 {
			log.Warningf("There are multiple components named \"%s\" in different multiple devfile registries.\n", o.componentName)
		}
		if len(o.devfileComponents) > 0 {
			fmt.Fprintln(w, "Devfile Component(s):")

			for _, devfileComponent := range o.devfileComponents {
				fmt.Fprintln(w, "\n* Registry: "+devfileComponent.Registry.Name)

				devObj, err := GetDevfile(devfileComponent)
				if err != nil {
					return err
				}

				projects, err := devObj.Data.GetStarterProjects(parsercommon.DevfileOptions{})
				if err != nil {
					return err
				}
				// only print project info if there is at least one project in the devfile
				err = o.PrintDevfileStarterProjects(w, projects, devObj)
				if err != nil {
					return err
				}
			}
		} else {
			fmt.Fprintln(w, "There are no Odo devfile components with the name \""+o.componentName+"\"")
		}
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
	clientset.Add(command, clientset.CATALOG)
	return command
}

// listDevfileComponentsByName lists all the devfile components that have the same name as the specified component
func listDevfileComponentsByName(catalogDevfileList catalog.DevfileComponentTypeList, componentName string) []catalog.DevfileComponentType {
	var components []catalog.DevfileComponentType
	for _, devfileComponent := range catalogDevfileList.Items {
		if devfileComponent.Name == componentName {
			components = append(components, devfileComponent)
		}
	}
	return components
}

// GetDevfile downloads the devfile in memory and return the devfile object
func GetDevfile(devfileComponent catalog.DevfileComponentType) (parser.DevfileObj, error) {
	devObj, err := getDevFileNoValidation(devfileComponent)
	if err != nil {
		return devObj, err
	}

	err = validate.ValidateDevfileData(devObj.Data)
	if err != nil {
		return devObj, err
	}
	return devObj, nil
}

func getDevFileNoValidation(devfileComponent catalog.DevfileComponentType) (parser.DevfileObj, error) {
	if strings.Contains(devfileComponent.Registry.URL, "github") {
		devObj, err := devfile.ParseAndValidateFromURL(devfileComponent.Registry.URL + devfileComponent.Link)
		if err != nil {
			return devObj, fmt.Errorf("Failed to download devfile.yaml from Github-based registry for devfile component: %s: %w", devfileComponent.Name, err)
		}
		return devObj, nil
	}
	registryURL, err := url.Parse(devfileComponent.Registry.URL)
	if err != nil {
		return parser.DevfileObj{}, fmt.Errorf("Failed to parse registry URL for devfile component: %s: %w", devfileComponent.Name, err)
	}
	registryURL.Path = path.Join(registryURL.Path, "devfiles", devfileComponent.Name)
	devObj, err := devfile.ParseAndValidateFromURL(registryURL.String())
	if err != nil {
		return devObj, fmt.Errorf("Failed to download devfile.yaml from OCI-based registry for devfile component: %s: %w", devfileComponent.Name, err)
	}
	return devObj, nil
}

// PrintDevfileStarterProjects prints all the starter projects in a devfile
// If no starter projects exists in the devfile, it prints the whole devfile
func (o *DescribeComponentOptions) PrintDevfileStarterProjects(w *tabwriter.Writer, projects []devfilev1.StarterProject, devObj parser.DevfileObj) error {
	if len(projects) > 0 {
		fmt.Fprintln(w, "\nStarter Projects:")
		for _, project := range projects {
			yamlData, err := yaml.Marshal(project)
			if err != nil {
				return fmt.Errorf("Failed to marshal devfile object into yaml: %w", err)
			}
			fmt.Printf("---\n%s", string(yamlData))
		}
	} else {
		fmt.Fprintln(w, "The Odo devfile component \""+o.componentName+"\" has no starter projects.")
		yamlData, err := yaml.Marshal(devObj)
		if err != nil {
			return fmt.Errorf("Failed to marshal devfile object into yaml: %w", err)
		}
		fmt.Printf("---\n%s", string(yamlData))
	}
	return nil
}
