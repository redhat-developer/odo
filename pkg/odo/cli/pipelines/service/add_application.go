package service

import (
	"fmt"
	"strings"

	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/openshift/odo/pkg/pipelines"
	"github.com/spf13/cobra"

	ktemplates "k8s.io/kubernetes/pkg/kubectl/util/templates"
)

const (
	// AddServiceRecommendedCommandName the recommended command name
	AddServiceRecommendedCommandName = "add"
)

var (
	// AddExample an example/description of the command
	AddExample = ktemplates.Examples(`
  # Add applications to OpenShift pipelines in a cluster
  %[1]s
  `)

	// AddLongDesc long description of the command
	AddLongDesc = ktemplates.LongDesc(`Add an application to GitOps CI/CD Pipelines`)

	// AddShortDesc short description of the command
	AddShortDesc = `Add an application repo to GitOps`
)

// AddParameters encapsulates the parameters for the odo pipelines service add command.
type AddParameters struct {
	appName              string
	envName              string
	output               string
	prefix               string
	serviceGitRepo       string
	serviceWebhookSecret string
	skipChecks           bool

	*genericclioptions.Context
}

// NewAddParameters  bootstraps a AddParameters instance.
func NewAddParameters() *AddParameters {
	return &AddParameters{}
}

// Complete completes AddParameters after they've been created.
//
// If the prefix provided doesn't have a "-" then one is added, this makes the
// generated environment names nicer to read.
func (io *AddParameters) Complete(name string, cmd *cobra.Command, args []string) error {
	if io.prefix != "" && !strings.HasSuffix(io.prefix, "-") {
		io.prefix = io.prefix + "-"
	}
	return nil
}

// Validate validates the parameters of the AddParameters
func (io *AddParameters) Validate() error {
	if len(strings.Split(io.serviceGitRepo, "/")) != 2 || len(strings.Split(io.serviceGitRepo, "/")) != 2 {
		return fmt.Errorf("service-git-repo must be org/repo: %s", io.serviceGitRepo)
	}

	return nil
}

// Run runs the project bootstrap command.
func (io *AddParameters) Run() error {
	options := pipelines.AddParameters{
		AppName:              io.appName,
		EnvName:              io.envName,
		Output:               io.output,
		Prefix:               io.prefix,
		ServiceGitRepo:       io.serviceGitRepo,
		ServiceWebhookSecret: io.serviceWebhookSecret,
		SkipChecks:           io.skipChecks,
	}

	return pipelines.CreateApplication(&options)
}

// NewCmdAddService creates the project add service command.
func NewCmdAddService(name, fullName string) *cobra.Command {
	o := NewAddParameters()

	addCmd := &cobra.Command{
		Use:     name,
		Short:   AddShortDesc,
		Long:    AddLongDesc,
		Example: fmt.Sprintf(AddExample, fullName),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	addCmd.Flags().StringVar(&o.output, "output", "", "file path to output folder")
	addCmd.Flags().StringVar(&o.prefix, "prefix", "", "a prefix to the environment names")
	addCmd.Flags().StringVar(&o.appName, "app-name", "", "application name")
	addCmd.Flags().StringVar(&o.serviceWebhookSecret, "service-webhook-secret", "", "Webhook secret of the service Git repository")
	addCmd.Flags().StringVar(&o.envName, "env-name", "", "Add the name of the environment(namespace) to which the pipelines should be bootstrapped")
	addCmd.Flags().StringVar(&o.serviceGitRepo, "service-git-repo", "", "service Git repository in this form <username>/<repository>")
	addCmd.Flags().BoolVarP(&o.skipChecks, "skip-checks", "b", true, "skip Tekton installation checks")
	addCmd.MarkFlagRequired("app-name")
	addCmd.MarkFlagRequired("service-webhook-secret")
	addCmd.MarkFlagRequired("env-name")
	addCmd.MarkFlagRequired("service-git-repo")
	addCmd.MarkFlagRequired("output")

	return addCmd
}
