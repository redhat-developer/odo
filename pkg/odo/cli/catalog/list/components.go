package list

import (
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/golang/glog"
	"github.com/openshift/odo/pkg/catalog"
	"github.com/openshift/odo/pkg/devfile"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/machineoutput"
	"github.com/openshift/odo/pkg/odo/cli/catalog/util"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/openshift/odo/pkg/odo/util/experimental"
	"github.com/openshift/odo/pkg/odo/util/pushtarget"
	pkgUtil "github.com/openshift/odo/pkg/util"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

const componentsRecommendedCommandName = "components"

// DevfilePath is the path of temporary devfile.yaml, which we need to cleanup after calling list"
var DevfilePath = "./tmp-devfile.yaml"
var componentsExample = `  # Get the supported components
  %[1]s`

// ListComponentsOptions encapsulates the options for the odo catalog list components command
type ListComponentsOptions struct {
	// display both supported and unsupported devfile components
	listAllDevfileComponents bool
	// display both supported and unsupported devfile components
	listAllDevfileComponentsProjectInfo bool
	// list of known images
	catalogList catalog.ComponentTypeList
	// list of known devfiles
	catalogDevfileList catalog.DevfileComponentTypeList
	// generic context options common to all commands
	*genericclioptions.Context
	// display project information about specified component
	specifiedComponent string
}

// NewListComponentsOptions creates a new ListComponentsOptions instance
func NewListComponentsOptions() *ListComponentsOptions {
	return &ListComponentsOptions{}
}

// Complete completes ListComponentsOptions after they've been created
func (o *ListComponentsOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	if !pushtarget.IsPushTargetDocker() {
		o.Context = genericclioptions.NewContext(cmd)
		o.catalogList, err = catalog.ListComponents(o.Client)
		if err != nil {
			if experimental.IsExperimentalModeEnabled() {
				glog.V(4).Info("Please log in to an OpenShift cluster to list OpenShift/s2i components")
			} else {
				return err
			}
		}

		o.catalogList.Items = util.FilterHiddenComponents(o.catalogList.Items)
	}

	if experimental.IsExperimentalModeEnabled() {
		o.catalogDevfileList, err = catalog.ListDevfileComponents()
		if err != nil {
			return err
		}
	}

	return
}

// Validate validates the ListComponentsOptions based on completed values
func (o *ListComponentsOptions) Validate() (err error) {
	if len(o.catalogList.Items) == 0 && len(o.catalogDevfileList.Items) == 0 {
		return fmt.Errorf("no deployable components found")
	}

	return err
}

// Run contains the logic for the command associated with ListComponentsOptions
func (o *ListComponentsOptions) Run() (err error) {
	if log.IsJSON() {
		for i, image := range o.catalogList.Items {
			// here we don't care about the unsupported tags (second return value)
			supported, _ := catalog.SliceSupportedTags(image)
			o.catalogList.Items[i].Spec.SupportedTags = supported
		}
		machineoutput.OutputSuccess(o.catalogList)
	} else {
		w := tabwriter.NewWriter(os.Stdout, 5, 2, 3, ' ', tabwriter.TabIndent)
		var supCatalogList, unsupCatalogList []catalog.ComponentType
		var supDevfileCatalogList, unsupDevfileCatalogList []catalog.DevfileComponentType
		var specifiedComponent catalog.DevfileComponentType
		var supported string

		// if user passes an arg, look through all devfile components
		if o.specifiedComponent != "" {
			o.listAllDevfileComponents = true
		}
		for _, image := range o.catalogList.Items {
			supported, unsupported := catalog.SliceSupportedTags(image)
			if (o.specifiedComponent != "" && image.Name == o.specifiedComponent) || o.specifiedComponent == "" {
				if len(supported) != 0 {
					image.Spec.NonHiddenTags = supported
					supCatalogList = append(supCatalogList, image)
				}
				if len(unsupported) != 0 {
					image.Spec.NonHiddenTags = unsupported
					unsupCatalogList = append(unsupCatalogList, image)
				}
			}
		}

		for _, devfileComponent := range o.catalogDevfileList.Items {
			if (o.specifiedComponent != "" && devfileComponent.Name == o.specifiedComponent) || o.specifiedComponent == "" {
				if devfileComponent.Support {
					supDevfileCatalogList = append(supDevfileCatalogList, devfileComponent)
				} else {
					unsupDevfileCatalogList = append(unsupDevfileCatalogList, devfileComponent)
				}
				specifiedComponent = devfileComponent
			}
		}

		if len(supCatalogList) != 0 || len(unsupCatalogList) != 0 {
			fmt.Fprintln(w, "Odo OpenShift Components:")
			fmt.Fprintln(w, "NAME", "\t", "PROJECT", "\t", "TAGS", "\t", "SUPPORTED")

			if len(supCatalogList) != 0 {
				supported = "YES"
				o.printCatalogList(w, supCatalogList, supported)
			}

			if len(unsupCatalogList) != 0 {
				supported = "NO"
				o.printCatalogList(w, unsupCatalogList, supported)
			}

			fmt.Fprintln(w)
		}

		if len(supDevfileCatalogList) != 0 || (o.listAllDevfileComponents && len(unsupDevfileCatalogList) != 0) {
			fmt.Fprintln(w, "Odo Devfile Components:")
			fmt.Fprintln(w, "NAME", "\t", "DESCRIPTION", "\t", "SUPPORTED")

			if len(supDevfileCatalogList) != 0 {
				supported = "YES"
				o.printDevfileCatalogList(w, supDevfileCatalogList, supported)
			}

			if o.listAllDevfileComponents && len(unsupDevfileCatalogList) != 0 {
				supported = "NO"
				o.printDevfileCatalogList(w, unsupDevfileCatalogList, supported)
			}

			fmt.Fprintln(w)

			// if user passes an arg, download the devfile and get the project information
			if o.specifiedComponent != "" {
				err := pkgUtil.DownloadFile(specifiedComponent.Registry+specifiedComponent.Link, DevfilePath)
				if err != nil {
					return errors.Errorf("Failed to download devfile.yaml for devfile component: %v", err)
				}

				devObj, err := devfile.Parse(DevfilePath)
				if err != nil {
					return err
				}

				projects := devObj.Data.GetProjects()
				// only print project info if there is at least one project in the devfile
				if len(projects) > 0 {
					fmt.Fprintln(w, "Devfile Starter Project(s):")

					for _, project := range projects {
						// if they pass the -p flag, print all project information, else print the default fields
						if o.listAllDevfileComponentsProjectInfo {
							yamlData, err := yaml.Marshal(project)
							if err != nil {
								return errors.Errorf("Failed to marshal devfile object into yaml:", err)
							}
							fmt.Printf("---\n%s\n", string(yamlData))
						} else {
							fmt.Fprintln(w, "NAME", "\t", "TYPE", "\t", "LOCATION")
							fmt.Fprintln(w, project.Name, "\t", project.Source.Type, "\t", project.Source.Location)
						}
					}
					fmt.Fprintln(w)
				} else {
					fmt.Println("This devfile component does not have any available starter projects.")
				}
				err = os.Remove(DevfilePath)
				if err != nil {
					return err
				}
			}
			// download component devfile
		}
		if len(supCatalogList) == 0 && len(unsupCatalogList) == 0 && len(supDevfileCatalogList) == 0 && len(unsupDevfileCatalogList) == 0 {
			return errors.Errorf("The component \"%s\" is not a valid Odo component.", o.specifiedComponent)
		}

		w.Flush()
	}
	return
}

// NewCmdCatalogListComponents implements the odo catalog list components command
func NewCmdCatalogListComponents(name, fullName string) *cobra.Command {
	o := NewListComponentsOptions()

	var componentListCmd = &cobra.Command{
		Use:         name,
		Short:       "List all components",
		Long:        "List all available component types from OpenShift's Image Builder",
		Example:     fmt.Sprintf(componentsExample, fullName),
		Args:        cobra.RangeArgs(0, 1),
		Annotations: map[string]string{"machineoutput": "json"},
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 1 {
				o.specifiedComponent = args[0]
			} else {
				o.specifiedComponent = ""
			}
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	componentListCmd.Flags().BoolVarP(&o.listAllDevfileComponents, "all", "a", false, "List both supported and unsupported devfile components.")
	componentListCmd.Flags().BoolVarP(&o.listAllDevfileComponentsProjectInfo, "projectInfo", "p", false, "List all information about the component's available projects.")

	return componentListCmd
}

func (o *ListComponentsOptions) printCatalogList(w io.Writer, catalogList []catalog.ComponentType, supported string) {
	for _, component := range catalogList {
		componentName := component.Name
		if component.Namespace == o.Project {
			/*
				If current namespace is same as the current component namespace,
				Loop through every other component,
				If there exists a component with same name but in different namespaces,
				mark the one from current namespace with (*)
			*/
			for _, comp := range catalogList {
				if comp.ObjectMeta.Name == component.ObjectMeta.Name && component.Namespace != comp.Namespace {
					componentName = fmt.Sprintf("%s (*)", component.ObjectMeta.Name)
				}
			}
		}
		fmt.Fprintln(w, componentName, "\t", component.ObjectMeta.Namespace, "\t", strings.Join(component.Spec.NonHiddenTags, ","), "\t", supported)
	}
}

func (o *ListComponentsOptions) printDevfileCatalogList(w io.Writer, catalogDevfileList []catalog.DevfileComponentType, supported string) {
	for _, devfileComponent := range catalogDevfileList {
		fmt.Fprintln(w, devfileComponent.Name, "\t", devfileComponent.Description, "\t", supported)
	}
}
