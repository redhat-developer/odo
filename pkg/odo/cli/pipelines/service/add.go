package service

import (
	"fmt"

	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/odo/cli/pipelines/utility"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/openshift/odo/pkg/pipelines"
	"github.com/openshift/odo/pkg/pipelines/ioutils"
	"github.com/spf13/cobra"

	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

const (
	addRecommendedCommandName = "add"
)

var (
	addExample = ktemplates.Examples(`	Add a Service to an environment in GitOps 
	%[1]s`)

	addLongDesc  = ktemplates.LongDesc(`Add a Service to an environment in GitOps`)
	addShortDesc = `Add a new service`
)

// AddServiceOptions encapsulates the parameters for service add command
type AddServiceOptions struct {
	*pipelines.AddServiceOptions
	// generic context options common to all commands
	*genericclioptions.Context
}

// Complete is called when the command is completed
func (o *AddServiceOptions) Complete(name string, cmd *cobra.Command, args []string) error {
	o.GitRepoURL = utility.AddGitSuffixIfNecessary(o.GitRepoURL)
	return nil
}

// Validate validates the parameters of the EnvParameters.
func (o *AddServiceOptions) Validate() error {
	return nil
}

// Run runs the project bootstrap command.
func (o *AddServiceOptions) Run() error {
	err := pipelines.AddService(o.AddServiceOptions, ioutils.NewFilesystem())

	if err != nil {
		return err
	}
	log.Successf("Created Service %s sucessfully at environment %s.", o.ServiceName, o.EnvName)
	return nil
}

func newCmdAdd(name, fullName string) *cobra.Command {
	o := &AddServiceOptions{AddServiceOptions: &pipelines.AddServiceOptions{}}

	cmd := &cobra.Command{
		Use:     name,
		Short:   addShortDesc,
		Long:    addLongDesc,
		Example: fmt.Sprintf(addExample, fullName),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	cmd.Flags().StringVar(&o.GitRepoURL, "git-repo-url", "", "source Git repository URL")
	cmd.Flags().StringVar(&o.WebhookSecret, "webhook-secret", "", "source Git repository webhook secret (if not provided, it will be auto-generated)")
	cmd.Flags().StringVar(&o.AppName, "app-name", "", "the name of the application where the service will be added")
	cmd.Flags().StringVar(&o.ServiceName, "service-name", "", "the name of the service to be added")
	cmd.Flags().StringVar(&o.EnvName, "env-name", "", "the name of the environment where the service will be added")
	cmd.Flags().StringVar(&o.ImageRepo, "image-repo", "", "used to push built images")
	cmd.Flags().StringVar(&o.InternalRegistryHostname, "internal-registry-hostname", "image-registry.openshift-image-registry.svc:5000", "internal image registry hostname")
	cmd.Flags().StringVar(&o.PipelinesFilePath, "pipelines-file", "pipelines.yaml", "path to pipelines file")
	cmd.Flags().StringVarP(&o.SealedSecretsNamespace, "sealed-secrets-ns", "", "kube-system", "namespace in which the Sealed Secrets operator is installed, automatically generated secrets are encrypted with this operator")

	// required flags
	_ = cmd.MarkFlagRequired("service-name")
	_ = cmd.MarkFlagRequired("app-name")
	_ = cmd.MarkFlagRequired("env-name")
	return cmd
}
