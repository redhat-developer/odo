package application

import (
	"fmt"
	"github.com/redhat-developer/odo/pkg/application"
	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/occlient"
	projectCmd "github.com/redhat-developer/odo/pkg/odo/cli/project"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/redhat-developer/odo/pkg/odo/util/completion"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	applicationShortFlag       bool
	applicationForceDeleteFlag bool
	outputFlag                 string
)

// applicationCmd represents the app command
var applicationCmd = &cobra.Command{
	Use:   "app",
	Short: "Perform application operations",
	Long:  `Performs application operations related to your OpenShift project.`,
	Example: fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s",
		applicationCreateCmd.Example,
		applicationGetCmd.Example,
		applicationDeleteCmd.Example,
		applicationDescribeCmd.Example,
		applicationListCmd.Example,
		applicationSetCmd.Example),
	Aliases: []string{"application"},
	// 'odo app' is the same as 'odo app get'
	// 'odo app <application_name>' is the same as 'odo app set <application_name>'
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 && args[0] != "get" && args[0] != "set" {
			applicationSetCmd.Run(cmd, args)
		} else {
			applicationGetCmd.Run(cmd, args)
		}
	},
}

// getMachineReadableFormat returns resource information in machine readable format
func getMachineReadableFormat(client *occlient.Client, appName string, projectName string, active bool) application.App {
	componentList, _ := component.List(client, appName)
	var compList []string
	for _, comp := range componentList {
		compList = append(compList, comp.ComponentName)
	}
	appDef := application.App{
		TypeMeta: metav1.TypeMeta{
			Kind:       "app",
			APIVersion: "odo.openshift.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      appName,
			Namespace: projectName,
		},
		Spec: application.AppSpec{
			Components: compList,
		},
		Status: application.AppStatus{
			Active: active,
		},
	}
	return appDef
}

// NewCmdApplication implements the odo application command
func NewCmdApplication() *cobra.Command {
	applicationDeleteCmd.Flags().BoolVarP(&applicationForceDeleteFlag, "force", "f", false, "Delete application without prompting")

	applicationGetCmd.Flags().BoolVarP(&applicationShortFlag, "short", "q", false, "If true, display only the application name")

	applicationDescribeCmd.Flags().StringVarP(&outputFlag, "output", "o", "", "output in json format")
	applicationListCmd.Flags().StringVarP(&outputFlag, "output", "o", "", "output in json format")

	// add flags from 'get' to application command
	applicationCmd.Flags().AddFlagSet(applicationGetCmd.Flags())

	applicationCmd.AddCommand(applicationListCmd)
	applicationCmd.AddCommand(applicationDeleteCmd)
	applicationCmd.AddCommand(applicationGetCmd)
	applicationCmd.AddCommand(applicationCreateCmd)
	applicationCmd.AddCommand(applicationSetCmd)
	applicationCmd.AddCommand(applicationDescribeCmd)

	//Adding `--project` flag
	projectCmd.AddProjectFlag(applicationListCmd)
	projectCmd.AddProjectFlag(applicationCreateCmd)
	projectCmd.AddProjectFlag(applicationDeleteCmd)
	projectCmd.AddProjectFlag(applicationDescribeCmd)
	projectCmd.AddProjectFlag(applicationSetCmd)
	projectCmd.AddProjectFlag(applicationGetCmd)

	// Add a defined annotation in order to appear in the help menu
	applicationCmd.Annotations = map[string]string{"command": "other"}
	applicationCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)

	completion.RegisterCommandHandler(applicationDescribeCmd, completion.AppCompletionHandler)
	completion.RegisterCommandHandler(applicationDeleteCmd, completion.AppCompletionHandler)
	completion.RegisterCommandHandler(applicationSetCmd, completion.AppCompletionHandler)

	return applicationCmd
}

// AddApplicationFlag adds a `app` flag to the given cobra command
// Also adds a completion handler to the flag
func AddApplicationFlag(cmd *cobra.Command) {
	cmd.Flags().String(genericclioptions.ApplicationFlagName, "", "Application, defaults to active application")
	completion.RegisterCommandFlagHandler(cmd, "app", completion.AppCompletionHandler)
}

// printDeleteAppInfo will print things which will be deleted
func printDeleteAppInfo(client *occlient.Client, appName string, projectName string) error {
	componentList, err := component.List(client, appName)
	if err != nil {
		return errors.Wrap(err, "failed to get Component list")
	}

	for _, currentComponent := range componentList {
		componentDesc, err := component.GetComponentDesc(client, currentComponent.ComponentName, appName, projectName)
		if err != nil {
			return errors.Wrap(err, "unable to get component description")
		}
		log.Info("Component", currentComponent.ComponentName, "will be deleted.")

		if len(componentDesc.URLs) != 0 {
			fmt.Println("  Externally exposed URLs will be removed")
		}

		for _, store := range componentDesc.Storage {
			fmt.Println("  Storage", store.Name, "of size", store.Size, "will be removed")
		}

	}
	return nil
}
