package component

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	appsv1 "k8s.io/api/apps/v1"

	"github.com/openshift/odo/pkg/application"
	"github.com/openshift/odo/pkg/devfile"
	"github.com/openshift/odo/pkg/machineoutput"
	"github.com/openshift/odo/pkg/project"
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

	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

// ListRecommendedCommandName is the recommended watch command name
const ListRecommendedCommandName = "list"

const UnpushedCompState = "Unpushed"
const PushedCompState = "Pushed"

var listExample = ktemplates.Examples(`  # List all components in the application
%[1]s
  `)

// ListOptions is a dummy container to attach complete, validate and run pattern
type ListOptions struct {
	pathFlag             string
	allAppsFlag          bool
	componentContext     string
	componentType        string
	hasDCSupport         bool
	hasDevfileComponents bool
	hasS2IComponents     bool
	devfilePath          string
	*genericclioptions.Context
}

// NewListOptions returns new instance of ListOptions
func NewListOptions() *ListOptions {
	return &ListOptions{}
}

// Complete completes log args
func (lo *ListOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {

	lo.devfilePath = filepath.Join(lo.componentContext, DevfilePath)

	if util.CheckPathExists(lo.devfilePath) {

		lo.Context = genericclioptions.NewDevfileContext(cmd)
		lo.Client = genericclioptions.Client(cmd)
		lo.hasDCSupport, err = lo.Client.IsDeploymentConfigSupported()
		if err != nil {
			return err
		}
		devfile, err := devfile.ParseAndValidate(lo.devfilePath)
		if err != nil {
			return err
		}
		lo.componentType = devfile.Data.GetMetadata().Name

	} else {
		// here we use the config.yaml derived context if its present, else we use information from user's kubeconfig
		// as odo list should work in a non-component directory too

		if util.CheckKubeConfigExist() {
			klog.V(4).Infof("New Context")
			lo.Context = genericclioptions.NewContext(cmd)
			lo.hasDCSupport, err = lo.Client.IsDeploymentConfigSupported()
			if err != nil {
				return err
			}

		} else {
			klog.V(4).Infof("New Config Context")
			lo.Context = genericclioptions.NewConfigContext(cmd)
			// for disconnected situation we just assume we have DC support
			lo.hasDCSupport = true

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
		if lo.Context.Application == "" && lo.Context.KClient.Namespace == "" {
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

		if util.CheckPathExists(lo.devfilePath) {
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

	if util.CheckPathExists(lo.devfilePath) {

		var deploymentList *appsv1.DeploymentList
		var err error

		var selector string
		// TODO: wrap this into a component list for docker support
		if lo.allAppsFlag {
			selector = project.GetSelector()

		} else {
			selector = applabels.GetSelector(lo.Application)
		}

		deploymentList, err = lo.KClient.ListDeployments(selector)

		if err != nil {
			return err
		}

		// Json output is not implemented yet for devfile
		if !log.IsJSON() {
			envinfo := lo.EnvSpecificInfo.EnvInfo
			if len(deploymentList.Items) != 0 || envinfo.GetApplication() == lo.Application {

				currentComponentState := UnpushedCompState
				currentComponentName := envinfo.GetName()
				lo.hasDevfileComponents = true
				w := tabwriter.NewWriter(os.Stdout, 5, 2, 3, ' ', tabwriter.TabIndent)
				fmt.Fprintln(w, "Devfile Components: ")
				fmt.Fprintln(w, "APP", "\t", "NAME", "\t", "PROJECT", "\t", "TYPE", "\t", "STATE")
				for _, comp := range deploymentList.Items {
					app := comp.Labels[applabels.ApplicationLabel]
					cmpType := comp.Labels[componentlabels.ComponentTypeLabel]
					if comp.Name == currentComponentName && app == envinfo.GetApplication() && comp.Namespace == envinfo.GetNamespace() {
						currentComponentState = PushedCompState
					}
					fmt.Fprintln(w, app, "\t", comp.Name, "\t", comp.Namespace, "\t", cmpType, "\t", "Pushed")
				}

				// 1st condition - only if we are using the same application or all-apps are provided should we show the current component
				// 2nd condition - if the currentComponentState is unpushed that means it didn't show up in the list above
				if (envinfo.GetApplication() == lo.Application || lo.allAppsFlag) && currentComponentState == UnpushedCompState {
					fmt.Fprintln(w, envinfo.GetApplication(), "\t", currentComponentName, "\t", envinfo.GetNamespace(), "\t", lo.componentType, "\t", currentComponentState)
				}

				w.Flush()
			}

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
				return errors.Wrapf(err, "failed to fetch component list")
			}
		}

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
