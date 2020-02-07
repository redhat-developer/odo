package pipelines

import (
	"fmt"
	"strings"

	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/spf13/cobra"

	"github.com/openshift/odo/pkg/pipelines"
	ktemplates "k8s.io/kubernetes/pkg/kubectl/util/templates"
)

const (
	BootstrapRecommendedCommandName = "bootstrap"
)

var (
	bootstrapExample = ktemplates.Examples(`
	# Bootstrap OpenShift pipelines in a cluster
	%[1]s username org/repo
	`)

	bootstrapLongDesc  = ktemplates.LongDesc(`Bootstrap GitOps pipelines`)
	bootstrapShortDesc = `Bootstrap pipelines`
)

// BootstrapOptions encapsulates the options for the odo pipelines bootstrap
// command.
type BootstrapOptions struct {
	quayUsername string
	baseRepo     string // e.g. tekton/triggers
	prefix       string // used to generate the environments in a shared cluster

	// generic context options common to all commands
	*genericclioptions.Context
}

// NewBootstrapOptions bootstraps a BootstrapOptions instance.
func NewBootstrapOptions() *BootstrapOptions {
	return &BootstrapOptions{}
}

// Complete completes BootstrapOptions after they've been created.
//
// If the prefix provided doesn't have a "-" then one is added, this makes the
// generated environment names nicer to read.
func (bo *BootstrapOptions) Complete(name string, cmd *cobra.Command, args []string) error {
	bo.quayUsername = args[0]
	bo.baseRepo = args[1]

	if bo.prefix != "" && !strings.HasSuffix(bo.prefix, "-") {
		bo.prefix = bo.prefix + "-"
	}
	return nil
}

// Validate validates the parameters of the BootstrapOptions.
func (bo *BootstrapOptions) Validate() error {
	// TODO: this won't work with GitLab as the repo can have more path elements.
	if len(strings.Split(bo.baseRepo, "/")) != 2 {
		return fmt.Errorf("repo must be org/repo: %s", bo.baseRepo)
	}
	return nil
}

// Run runs the project bootstrap command.
func (bo *BootstrapOptions) Run() error {
	return pipelines.Bootstrap(bo.quayUsername, bo.baseRepo, bo.prefix)
}

// NewCmdBootstrap creates the project bootstrap command.
func NewCmdBootstrap(name, fullName string) *cobra.Command {
	o := NewBootstrapOptions()

	bootstrapCmd := &cobra.Command{
		Use:     name,
		Short:   bootstrapShortDesc,
		Long:    bootstrapLongDesc,
		Example: fmt.Sprintf(bootstrapExample, fullName),
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	bootstrapCmd.Flags().StringVarP(&o.prefix, "prefix", "p", "", "add a prefix to the environment names")
	return bootstrapCmd
}
