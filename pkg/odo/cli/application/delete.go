package application

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/redhat-developer/odo/pkg/application"
	applabels "github.com/redhat-developer/odo/pkg/application/labels"
	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/cli/project"
	"github.com/redhat-developer/odo/pkg/odo/cli/ui"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	odoUtil "github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/redhat-developer/odo/pkg/odo/util/completion"

	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

const deleteRecommendedCommandName = "delete"

var (
	deleteExample = ktemplates.Examples(`  # Delete the application
  %[1]s myapp`)
)

// DeleteOptions encapsulates the options for the odo command
type DeleteOptions struct {
	// Context
	*genericclioptions.Context

	// Clients
	appClient application.Client

	// Parameters
	appName string

	// Flags
	forceFlag bool
}

// NewDeleteOptions creates a new DeleteOptions instance
func NewDeleteOptions(appClient application.Client) *DeleteOptions {
	return &DeleteOptions{
		appClient: appClient,
	}
}

// Complete completes DeleteOptions after they've been created
func (o *DeleteOptions) Complete(name string, cmdline cmdline.Cmdline, args []string) (err error) {
	o.Context, err = genericclioptions.New(genericclioptions.NewCreateParameters(cmdline))
	if err != nil {
		return err
	}
	o.appName = o.GetApplication()
	if len(args) == 1 {
		// If app name passed, consider it for deletion
		o.appName = args[0]
	}
	return
}

// Validate validates the DeleteOptions based on completed values
func (o *DeleteOptions) Validate() (err error) {
	if o.Context.GetProject() == "" || o.appName == "" {
		return odoUtil.ThrowContextError()
	}

	exist, err := o.appClient.Exists(o.appName)
	if !exist {
		return fmt.Errorf("%s app does not exists", o.appName)
	}
	return err
}

// Run contains the logic for the odo command
func (o *DeleteOptions) Run() (err error) {
	if o.IsJSON() {
		return o.appClient.Delete(o.appName)
	}

	// Print App Information which will be deleted
	err = printAppInfo(o.KClient, o.appName, o.GetProject())
	if err != nil {
		return err
	}

	if o.forceFlag || ui.Proceed(fmt.Sprintf("Are you sure you want to delete the application: %v from project: %v", o.appName, o.GetProject())) {
		err = o.appClient.Delete(o.appName)
		if err != nil {
			return err
		}
		log.Infof("Deleted application: %s from project: %v", o.appName, o.GetProject())
	} else {
		log.Infof("Aborting deletion of application: %v", o.appName)
	}
	return nil
}

// printAppInfo will print things which will be deleted
func printAppInfo(client kclient.ClientInterface, appName string, projectName string) error {
	var selector string
	if appName != "" {
		selector = applabels.GetSelector(appName)
	}
	componentList, err := component.List(client, selector)
	if err != nil {
		return errors.Wrap(err, "failed to get Component list")
	}

	if len(componentList.Items) != 0 {
		log.Info("This application has following components that will be deleted")
		for _, currentComponent := range componentList.Items {
			log.Info("component named", currentComponent.Name)

			if len(currentComponent.Spec.URL) != 0 {
				log.Info("This component has following urls that will be deleted with component")
				for _, u := range currentComponent.Spec.URLSpec {
					log.Info("URL named", u.GetName(), "with host", u.Spec.Host, "having protocol", u.Spec.Protocol, "at port", u.Spec.Port)
				}
			}

			if len(currentComponent.Spec.Storage) != 0 {
				log.Info("The component has following storages which will be deleted with the component")
				for _, storage := range currentComponent.Spec.StorageSpec {
					store := storage
					log.Info("Storage named", store.GetName(), "of size", store.Spec.Size)
				}
			}
		}
	}
	return nil
}

// NewCmdDelete implements the odo command.
func NewCmdDelete(name, fullName string) *cobra.Command {
	kubclient, _ := kclient.New()
	o := NewDeleteOptions(application.NewClient(kubclient))
	command := &cobra.Command{
		Use:     name,
		Short:   "Delete the given application",
		Long:    "Delete the given application",
		Example: fmt.Sprintf(deleteExample, fullName),
		Args:    cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	command.Flags().BoolVarP(&o.forceFlag, "force", "f", false, "Delete application without prompting")

	project.AddProjectFlag(command)
	completion.RegisterCommandHandler(command, completion.AppCompletionHandler)
	return command
}
