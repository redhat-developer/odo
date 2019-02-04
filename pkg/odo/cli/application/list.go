package application

import (
	"encoding/json"
	"fmt"
	"github.com/redhat-developer/odo/pkg/application"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/cli/project"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ktemplates "k8s.io/kubernetes/pkg/kubectl/cmd/templates"
	"os"
	"text/tabwriter"
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
	outputFormat string
	*genericclioptions.Context
}

// NewListOptions creates a new ListOptions instance
func NewListOptions() *ListOptions {
	return &ListOptions{}
}

// Complete completes ListOptions after they've been created
func (o *ListOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	o.Context = genericclioptions.NewContext(cmd)
	return
}

// Validate validates the ListOptions based on completed values
func (o *ListOptions) Validate() (err error) {
	return util.CheckOutputFlag(o.outputFormat)
}

// Run contains the logic for the odo command
func (o *ListOptions) Run() (err error) {
	apps, err := application.ListInProject(o.Client)
	if err != nil {
		return fmt.Errorf("unable to get list of applications: %v", err)
	}

	if len(apps) > 0 {

		if o.outputFormat == "json" {
			var appList []application.App
			for _, app := range apps {
				appDef := getMachineReadableFormat(o.Client, app.Name, o.Project, app.Active)
				appList = append(appList, appDef)
			}

			appListDef := application.AppList{
				TypeMeta: metav1.TypeMeta{
					Kind:       "List",
					APIVersion: "odo.openshift.io/v1alpha1",
				},
				ListMeta: metav1.ListMeta{},
				Items:    appList,
			}
			out, err := json.Marshal(appListDef)
			if err != nil {
				return err
			}
			fmt.Println(string(out))

		} else {
			log.Infof("The project '%v' has the following applications:", o.Project)
			tabWriter := tabwriter.NewWriter(os.Stdout, 5, 2, 3, ' ', tabwriter.TabIndent)
			_, err := fmt.Fprintln(tabWriter, "ACTIVE", "\t", "NAME")
			if err != nil {
				return err
			}
			for _, app := range apps {
				activeMark := " "
				if app.Active {
					activeMark = "*"
				}
				_, err := fmt.Fprintln(tabWriter, activeMark, "\t", app.Name)
				if err != nil {
					return err
				}
			}
			return tabWriter.Flush()
		}
	} else {
		log.Infof("There are no applications deployed in the project '%v'.", o.Project)
	}
	return
}

// NewCmdList implements the odo command.
func NewCmdList(name, fullName string) *cobra.Command {
	o := NewListOptions()
	command := &cobra.Command{
		Use:     name,
		Short:   "List all applications in the current project",
		Long:    "List all applications in the current project",
		Example: fmt.Sprintf(listExample, fullName),
		Args:    cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			util.LogErrorAndExit(o.Complete(name, cmd, args), "")
			util.LogErrorAndExit(o.Validate(), "")
			util.LogErrorAndExit(o.Run(), "")
		},
	}

	command.Flags().StringVarP(&o.outputFormat, "output", "o", "json", "output in json format")
	project.AddProjectFlag(command)
	return command
}
