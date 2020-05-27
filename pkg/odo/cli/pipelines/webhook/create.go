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
	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

const createRecommendedCommandName = "create"

var (
	createExample = ktemplates.Examples(`	# Create a new Git repository webhook 
	%[1]s`)
)

type createOptions struct {
	options
}

// Run contains the logic for the odo command
func (o *createOptions) Run() error {

	id, err := backend.Create(o.accessToken, o.manifest, o.getAppServiceNames(), o.isCICD)

	if err != nil {
		return fmt.Errorf("Unable to create webhook: %v", err)
	}

	if id != "" {
		if log.IsJSON() {
			machineoutput.OutputSuccess(id)
		} else {
			w := tabwriter.NewWriter(os.Stdout, 5, 2, 3, ' ', tabwriter.TabIndent)
			fmt.Fprintln(w, "CREATED ID")
			fmt.Fprintln(w, "==========")
			fmt.Fprintln(w, id)
			w.Flush()
		}
	}

	return nil
}

func newCmdCreate(name, fullName string) *cobra.Command {

	o := &createOptions{}
	command := &cobra.Command{
		Use:     name,
		Short:   "Create a new webhook.",
		Long:    "Create a new Git repository webhook that triggers CI/CD pipeline runs.",
		Example: fmt.Sprintf(createExample, fullName),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	o.setFlags(command)
	return command
}
