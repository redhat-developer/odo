package webhook

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/machineoutput"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/spf13/cobra"

	backend "github.com/openshift/odo/pkg/pipelines/webhook"
	ktemplates "k8s.io/kubernetes/pkg/kubectl/util/templates"
)

const listRecommendedCommandName = "list"

var (
	listExample = ktemplates.Examples(`	# List Git repository webhook IDs 
	%[1]s`)
)

type listOptions struct {
	options
}

// Run contains the logic for the odo command
func (o *listOptions) Run() error {

	ids, err := backend.List(o.accessToken, o.pipelines, o.getAppServiceNames(), o.isCICD)
	if err != nil {
		return fmt.Errorf("Unable to a get list of webhook IDs: %v", err)
	}

	if ids != nil {
		if log.IsJSON() {
			machineoutput.OutputSuccess(ids)
		} else {
			w := tabwriter.NewWriter(os.Stdout, 5, 2, 3, ' ', tabwriter.TabIndent)
			fmt.Fprintln(w, "ID")
			fmt.Fprintln(w, "==")
			for _, id := range ids {
				fmt.Fprintln(w, id)
			}
			w.Flush()
		}
	}

	return nil
}

func newCmdList(name, fullName string) *cobra.Command {

	o := &listOptions{}
	command := &cobra.Command{
		Use:     name,
		Short:   "List existing webhook Ids.",
		Long:    "List existing Git repository webhook IDs of the target repository and listener.",
		Example: fmt.Sprintf(createExample, fullName),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	o.setFlags(command)
	return command
}
