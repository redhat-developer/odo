package application

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/redhat-developer/odo/pkg/application"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/machineoutput"
	"github.com/redhat-developer/odo/pkg/odo/cli/project"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

const listRecommendedCommandName = "list"

var (
	listExample = ktemplates.Examples(`  # List all applications in the current project
  %[1]s

  # List all applications in the specified project
  %[1]s --project myproject`)
)

// ListOptions encapsulates the options for the odo command
type ListOptions struct {
	// Context
	*genericclioptions.Context

	// Clients
	appClient application.Client
}

// NewListOptions creates a new ListOptions instance
func NewListOptions(appClient application.Client) *ListOptions {
	return &ListOptions{
		appClient: appClient,
	}
}

// Complete completes ListOptions after they've been created
func (o *ListOptions) Complete(name string, cmdline cmdline.Cmdline, args []string) (err error) {
	o.Context, err = genericclioptions.New(genericclioptions.NewCreateParameters(cmdline))
	return err
}

// Validate validates the ListOptions based on completed values
func (o *ListOptions) Validate() (err error) {
	// list doesn't need the app name
	if o.Context.GetProject() == "" {
		return util.ThrowContextError()
	}
	return nil
}

// Run contains the logic for the odo command
func (o *ListOptions) Run() (err error) {
	apps, err := o.appClient.List()
	if err != nil {
		return fmt.Errorf("unable to get list of applications: %v", err)
	}

	if len(apps) == 0 {
		if o.IsJSON() {
			apps := o.appClient.GetMachineReadableFormatForList([]application.App{})
			machineoutput.OutputSuccess(apps)
			return nil
		}

		log.Infof("There are no applications deployed in the project '%v'", o.GetProject())
		return nil
	}

	if o.IsJSON() {
		var appList []application.App
		for _, app := range apps {
			appDef := o.appClient.GetMachineReadableFormat(app, o.GetProject())
			appList = append(appList, appDef)
		}

		appListDef := o.appClient.GetMachineReadableFormatForList(appList)
		machineoutput.OutputSuccess(appListDef)
		return nil
	}

	log.Infof("The project '%v' has the following applications:", o.GetProject())
	tabWriter := tabwriter.NewWriter(os.Stdout, 5, 2, 3, ' ', tabwriter.TabIndent)
	_, err = fmt.Fprintln(tabWriter, "NAME")
	if err != nil {
		return err
	}
	for _, app := range apps {
		_, err := fmt.Fprintln(tabWriter, app)
		if err != nil {
			return err
		}
	}
	return tabWriter.Flush()

}

// NewCmdList implements the odo command.
func NewCmdList(name, fullName string) *cobra.Command {
	kubclient, _ := kclient.New()
	o := NewListOptions(application.NewClient(kubclient))
	command := &cobra.Command{
		Use:         name,
		Short:       "List all applications in the current project",
		Long:        "List all applications in the current project",
		Example:     fmt.Sprintf(listExample, fullName),
		Args:        cobra.NoArgs,
		Annotations: map[string]string{"machineoutput": "json"},
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	project.AddProjectFlag(command)
	return command
}
