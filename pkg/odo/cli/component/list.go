package component

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/openshift/odo/pkg/devfile"
	"github.com/openshift/odo/pkg/machineoutput"
	"github.com/openshift/odo/pkg/project"
	"github.com/openshift/odo/pkg/util"
	"github.com/spf13/cobra"
	"k8s.io/klog"

	applabels "github.com/openshift/odo/pkg/application/labels"

	"github.com/openshift/odo/pkg/component"
	"github.com/openshift/odo/pkg/log"
	appCmd "github.com/openshift/odo/pkg/odo/cli/application"
	projectCmd "github.com/openshift/odo/pkg/odo/cli/project"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	odoutil "github.com/openshift/odo/pkg/odo/util"
	"github.com/openshift/odo/pkg/odo/util/completion"

	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

// ListRecommendedCommandName is the recommended watch command name
const ListRecommendedCommandName = "list"

var listExample = ktemplates.Examples(`  # List all components in the application
%[1]s
  `)

// ListOptions is a dummy container to attach complete, validate and run pattern
type ListOptions struct {
	pathFlag         string
	allAppsFlag      bool
	componentContext string
	componentType    string
	devfilePath      string
	*genericclioptions.Context
}

// NewListOptions returns new instance of ListOptions
func NewListOptions() *ListOptions {
	return &ListOptions{}
}

// Complete completes log args
func (lo *ListOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {

	lo.devfilePath = devfile.DevfileLocation(lo.componentContext)

	if util.CheckPathExists(lo.devfilePath) {

		lo.Context, err = genericclioptions.NewContext(cmd)
		if err != nil {
			return err
		}
		devObj, err := devfile.ParseFromFile(lo.devfilePath)
		if err != nil {
			return err
		}
		lo.componentType = component.GetComponentTypeFromDevfileMetadata(devObj.Data.GetMetadata())

	} else {
		// here we use information from user's kubeconfig
		// as odo list should work in a non-component directory too
		if util.CheckKubeConfigExist() {
			klog.V(4).Infof("New Context")
			lo.Context, err = genericclioptions.NewContext(cmd)
			if err != nil {
				return err
			}
		}
	}

	return

}

// Validate validates the list parameters
func (lo *ListOptions) Validate() (err error) {

	if len(lo.Application) != 0 && lo.allAppsFlag {
		klog.V(4).Infof("either --app and --all-apps both provided or provided --all-apps in a folder has app, use --all-apps anyway")
	}

	if util.CheckPathExists(lo.devfilePath) {
		if lo.Application == "" && lo.KClient.GetCurrentNamespace() == "" {
			return odoutil.ThrowContextError()
		}
		return nil
	}
	var project, app string

	if !util.CheckKubeConfigExist() {
		project = lo.EnvSpecificInfo.GetNamespace()
		app = lo.EnvSpecificInfo.GetApplication()
	} else {
		project = lo.Context.Project
		app = lo.Application
	}
	if !lo.allAppsFlag && lo.pathFlag == "" && (project == "" || app == "") {
		return odoutil.ThrowContextError()
	}
	return nil

}

// Run has the logic to perform the required actions as part of command
func (lo *ListOptions) Run(cmd *cobra.Command) (err error) {
	var otherComps []component.Component
	// --path workflow

	if len(lo.pathFlag) != 0 {

		devfileComps, err := component.ListDevfileComponentsInPath(lo.KClient, filepath.SplitList(lo.pathFlag))
		if err != nil {
			return err
		}

		combinedComponents := component.NewCombinedComponentList(devfileComps, otherComps)

		if log.IsJSON() {
			machineoutput.OutputSuccess(combinedComponents)
		} else {
			HumanReadableOutputInPath(os.Stdout, combinedComponents)
		}
		return nil
	}

	// non --path workflow below
	// read the code like
	// -> experimental
	//	or	|-> --all-apps
	//		|-> the current app
	// -> non-experimental
	//	or	|-> --all-apps
	//		|-> the current app

	// experimental workflow

	devfileComponents := []component.Component{}
	var selector string
	// TODO: wrap this into a component list for docker support
	if lo.allAppsFlag {
		selector = project.GetSelector()
	} else {
		selector = applabels.GetSelector(lo.Application)
	}

	currentComponentState := component.StateTypeNotPushed

	if lo.KClient != nil {
		devfileComponentsOut, err := component.ListDevfileComponents(lo.Client, selector)
		if err != nil {
			return err
		}

		devfileComponents = devfileComponentsOut.Items
		for _, comp := range devfileComponents {
			if lo.EnvSpecificInfo != nil {
				// if we can find a component from the listing from server then the local state is pushed
				if lo.EnvSpecificInfo.EnvInfo.MatchComponent(comp.Name, comp.Spec.App, comp.Namespace) {
					currentComponentState = component.StateTypePushed
				}
			}
		}
	}

	// 1st condition - only if we are using the same application or all-apps are provided should we show the current component
	// 2nd condition - if the currentComponentState is unpushed that means it didn't show up in the list above
	if lo.EnvSpecificInfo != nil {
		envinfo := lo.EnvSpecificInfo.EnvInfo
		if (envinfo.GetApplication() == lo.Application || lo.allAppsFlag) && currentComponentState == component.StateTypeNotPushed {
			comp := component.NewComponent(envinfo.GetName())
			comp.Status.State = component.StateTypeNotPushed
			comp.Namespace = envinfo.GetNamespace()
			comp.Spec.App = envinfo.GetApplication()
			comp.Spec.Type = lo.componentType
			devfileComponents = append(devfileComponents, comp)
		}
	}

	// list components managed by other sources/tools
	if lo.allAppsFlag {
		selector = project.GetNonOdoSelector()
	} else {
		selector = applabels.GetNonOdoSelector(lo.Application)
	}

	otherComponents, err := component.List(lo.Client, selector)
	if err != nil {
		return fmt.Errorf("failed to fetch components not managed by odo: %w", err)
	}
	otherComps = otherComponents.Items

	combinedComponents := component.NewCombinedComponentList(devfileComponents, otherComps)
	if log.IsJSON() {
		machineoutput.OutputSuccess(combinedComponents)
	} else {
		HumanReadableOutput(os.Stdout, combinedComponents)
	}

	return
}

// NewCmdList implements the list odo command
func NewCmdList(name, fullName string) *cobra.Command {
	o := NewListOptions()

	var componentListCmd = &cobra.Command{
		Use:         name,
		Short:       "List all components in the current application",
		Long:        "List all components in the current application.",
		Example:     fmt.Sprintf(listExample, fullName),
		Args:        cobra.NoArgs,
		Annotations: map[string]string{"machineoutput": "json", "command": "component"},
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}
	genericclioptions.AddContextFlag(componentListCmd, &o.componentContext)
	componentListCmd.Flags().StringVar(&o.pathFlag, "path", "", "path of the directory to scan for odo component directories")
	componentListCmd.Flags().BoolVar(&o.allAppsFlag, "all-apps", false, "list all components from all applications for the current set project")
	componentListCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)

	//Adding `--project` flag
	projectCmd.AddProjectFlag(componentListCmd)
	//Adding `--application` flag
	appCmd.AddApplicationFlag(componentListCmd)

	completion.RegisterCommandFlagHandler(componentListCmd, "path", completion.FileCompletionHandler)

	return componentListCmd
}

func HumanReadableOutputInPath(wr io.Writer, o component.CombinedComponentList) {
	w := tabwriter.NewWriter(wr, 5, 2, 3, ' ', tabwriter.TabIndent)
	defer w.Flush()

	// if we dont have any components then
	if len(o.DevfileComponents) == 0 {
		fmt.Fprintln(w, "No components found")
		return
	}

	if len(o.DevfileComponents) != 0 {
		fmt.Fprintln(w, "Devfile Components: ")
		fmt.Fprintln(w, "APP", "\t", "NAME", "\t", "PROJECT", "\t", "STATE", "\t", "CONTEXT")
		for _, comp := range o.DevfileComponents {
			fmt.Fprintln(w, comp.Spec.App, "\t", comp.Name, "\t", comp.Namespace, "\t", comp.Status.State, "\t", comp.Status.Context)
		}
		fmt.Fprintln(w)
	}
}

func HumanReadableOutput(wr io.Writer, o component.CombinedComponentList) {
	w := tabwriter.NewWriter(wr, 5, 2, 3, ' ', tabwriter.TabIndent)
	defer w.Flush()

	if len(o.DevfileComponents) == 0 && len(o.OtherComponents) == 0 {
		log.Info("There are no components deployed.")
		return
	}

	if len(o.DevfileComponents) != 0 || len(o.OtherComponents) != 0 {
		fmt.Fprintln(w, "APP", "\t", "NAME", "\t", "PROJECT", "\t", "TYPE", "\t", "STATE", "\t", "MANAGED BY ODO")
		for _, comp := range o.DevfileComponents {
			fmt.Fprintln(w, comp.Spec.App, "\t", comp.Name, "\t", comp.Namespace, "\t", comp.Spec.Type, "\t", comp.Status.State, "\t", "Yes")
		}
		for _, comp := range o.OtherComponents {
			fmt.Fprintln(w, comp.Spec.App, "\t", comp.Name, "\t", comp.Namespace, "\t", comp.Spec.Type, "\t", component.StateTypePushed, "\t", "No")
		}
		fmt.Fprintln(w)
	}
}
