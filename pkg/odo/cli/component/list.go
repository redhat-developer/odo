package component

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	appsv1 "k8s.io/api/apps/v1"

	"github.com/openshift/odo/pkg/application"
	"github.com/openshift/odo/pkg/machineoutput"
	"github.com/openshift/odo/pkg/util"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/klog"

	applabels "github.com/openshift/odo/pkg/application/labels"
	componentlabels "github.com/openshift/odo/pkg/component/labels"

	"github.com/openshift/odo/pkg/component"
	"github.com/openshift/odo/pkg/log"
	appCmd "github.com/openshift/odo/pkg/odo/cli/application"
	projectCmd "github.com/openshift/odo/pkg/odo/cli/project"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	odoutil "github.com/openshift/odo/pkg/odo/util"
	"github.com/openshift/odo/pkg/odo/util/completion"
	"github.com/openshift/odo/pkg/odo/util/experimental"

	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

// ListRecommendedCommandName is the recommended watch command name
const ListRecommendedCommandName = "list"

var listExample = ktemplates.Examples(`  # List all components in the application
%[1]s
  `)

// ListOptions is a dummy container to attach complete, validate and run pattern
type ListOptions struct {
	pathFlag             string
	allAppsFlag          bool
	componentContext     string
	hasDCSupport         bool
	hasDevfileComponents bool
	hasS2IComponents     bool
	*genericclioptions.Context
}

// NewListOptions returns new instance of ListOptions
func NewListOptions() *ListOptions {
	return &ListOptions{}
}

// Complete completes log args
func (lo *ListOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {

	if experimental.IsExperimentalModeEnabled() {
		// Add a disclaimer that we are in *experimental mode*
		log.Experimental("Experimental mode is enabled, use at your own risk")

		lo.Context = genericclioptions.NewDevfileContext(cmd)

	} else {
		if util.CheckKubeConfigExist() {
			klog.V(4).Infof("New Context")
			lo.Context = genericclioptions.NewContext(cmd)
		} else {
			klog.V(4).Infof("New Config Context")
			lo.Context = genericclioptions.NewConfigContext(cmd)

		}
	}

	lo.Client = genericclioptions.Client(cmd)

	dcSupport, err := lo.Client.IsDeploymentConfigSupported()
	if err != nil {
		return err
	}

	lo.hasDCSupport = dcSupport

	return

}

// Validate validates the list parameters
func (lo *ListOptions) Validate() (err error) {

	if experimental.IsExperimentalModeEnabled() {
		if lo.Context.Application == "" && lo.Context.Project == "" {
			return odoutil.ThrowContextError()
		}
		return nil
	}
	var project, app string

	if !util.CheckKubeConfigExist() {
		project = lo.LocalConfigInfo.GetProject()
		app = lo.LocalConfigInfo.GetApplication()

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
func (lo *ListOptions) Run() (err error) {

	// --path workflow

	if len(lo.pathFlag) != 0 {

		if experimental.IsExperimentalModeEnabled() {
			log.Experimental("--path flag is not supported for devfile components")
		}
		components, err := component.ListIfPathGiven(lo.Context.Client, filepath.SplitList(lo.pathFlag))
		if err != nil {
			return err
		}
		if log.IsJSON() {
			machineoutput.OutputSuccess(components)
		} else {
			w := tabwriter.NewWriter(os.Stdout, 5, 2, 3, ' ', tabwriter.TabIndent)
			fmt.Fprintln(w, "APP", "\t", "NAME", "\t", "PROJECT", "\t", "TYPE", "\t", "SOURCETYPE", "\t", "STATE", "\t", "CONTEXT")
			for _, comp := range components.Items {
				fmt.Fprintln(w, comp.Spec.App, "\t", comp.Name, "\t", comp.Namespace, "\t", comp.Spec.Type, "\t", comp.Spec.SourceType, "\t", comp.Status.State, "\t", comp.Status.Context)

			}
			w.Flush()
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

	if experimental.IsExperimentalModeEnabled() {

		var dpl *appsv1.DeploymentList
		var err error

		// TODO: wrap this into a component list for docker support
		if lo.allAppsFlag {
			dpl, err = lo.KClient.ListAllDeployments()

		} else {
			dpl, err = lo.KClient.ListDeployments(lo.Application)
		}

		if err != nil {
			return err
		}

		if len(dpl.Items) != 0 {
			lo.hasDevfileComponents = true
			w := tabwriter.NewWriter(os.Stdout, 5, 2, 3, ' ', tabwriter.TabIndent)
			fmt.Fprintln(w, "Devfile Components: ")
			fmt.Fprintln(w, "APP", "\t", "NAME", "\t", "PROJECT", "\t", "TYPE", "\t", "REVISION")
			for _, comp := range dpl.Items {
				app := comp.Labels[applabels.ApplicationLabel]
				cmpType := comp.Labels[componentlabels.ComponentTypeLabel]
				revision := comp.Annotations["deployment.kubernetes.io/revision"]
				fmt.Fprintln(w, app, "\t", comp.Name, "\t", comp.Namespace, "\t", cmpType, "\t", revision)
			}
			w.Flush()

		}

	}

	// non-experimental workflow

	// we now check if DC is supported
	if lo.hasDCSupport {

		var components component.ComponentList

		if lo.allAppsFlag {
			// retrieve list of application
			apps, err := application.List(lo.Client)
			if err != nil {
				return err
			}

			var componentList []component.Component

			if len(apps) == 0 && lo.LocalConfigInfo.ConfigFileExists() {
				comps, err := component.List(lo.Client, lo.LocalConfigInfo.GetApplication(), lo.LocalConfigInfo)
				if err != nil {
					return err
				}
				componentList = append(componentList, comps.Items...)
			}

			// interating over list of application and get list of all components
			for _, app := range apps {
				comps, err := component.List(lo.Client, app, lo.LocalConfigInfo)
				if err != nil {
					return err
				}
				componentList = append(componentList, comps.Items...)
			}
			// Get machine readable component list format
			components = component.GetMachineReadableFormatForList(componentList)
		} else {

			components, err = component.List(lo.Client, lo.Application, lo.LocalConfigInfo)
			if err != nil {
				return errors.Wrapf(err, "failed to fetch components list")
			}
		}

		klog.V(4).Infof("the components are %+v", components)

		if log.IsJSON() {
			machineoutput.OutputSuccess(components)
		} else {
			if len(components.Items) != 0 {
				if lo.hasDevfileComponents {
					fmt.Println()
				}
				lo.hasS2IComponents = true
				w := tabwriter.NewWriter(os.Stdout, 5, 2, 3, ' ', tabwriter.TabIndent)
				fmt.Fprintln(w, "Openshift Components: ")
				fmt.Fprintln(w, "APP", "\t", "NAME", "\t", "PROJECT", "\t", "TYPE", "\t", "SOURCETYPE", "\t", "STATE")
				for _, comp := range components.Items {
					fmt.Fprintln(w, comp.Spec.App, "\t", comp.Name, "\t", comp.Namespace, "\t", comp.Spec.Type, "\t", comp.Spec.SourceType, "\t", comp.Status.State)
				}
				w.Flush()
			}
		}

		// if we dont have any of the components
		if !lo.hasDevfileComponents && !lo.hasS2IComponents {
			log.Error("There are no components deployed.")
			return
		}

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
