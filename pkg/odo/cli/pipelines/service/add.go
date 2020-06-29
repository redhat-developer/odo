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

// AddOptions encapsulates the parameters for service add command
type AddOptions struct {
	appName                  string
	envName                  string
	gitRepoURL               string
	imageRepo                string
	internalRegistryHostname string
	pipelinesFilePath        string
	serviceName              string
	webhookSecret            string

	// generic context options common to all commands
	*genericclioptions.Context
}

// Complete is called when the command is completed
func (o *AddOptions) Complete(name string, cmd *cobra.Command, args []string) error {
	o.gitRepoURL = utility.AddGitSuffixIfNecessary(o.gitRepoURL)
	return nil
}

// Validate validates the parameters of the EnvParameters.
func (o *AddOptions) Validate() error {
	return nil
}

// Run runs the project bootstrap command.
func (o *AddOptions) Run() error {

	err := pipelines.AddService(&pipelines.AddServiceParameters{
		AppName:                  o.appName,
		EnvName:                  o.envName,
		GitRepoURL:               o.gitRepoURL,
		ImageRepo:                o.imageRepo,
		InternalRegistryHostname: o.internalRegistryHostname,
		PipelinesFilePath:        o.pipelinesFilePath,
		ServiceName:              o.serviceName,
		WebhookSecret:            o.webhookSecret,
	}, ioutils.NewFilesystem())

	if err != nil {
		return err
	}
	log.Successf("Created Service %s sucessfully at environment %s.", o.serviceName, o.envName)
	return nil

}

func newCmdAdd(name, fullName string) *cobra.Command {
	o := &AddOptions{}

	cmd := &cobra.Command{
		Use:     name,
		Short:   addShortDesc,
		Long:    addLongDesc,
		Example: fmt.Sprintf(addExample, fullName),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	cmd.Flags().StringVar(&o.gitRepoURL, "git-repo-url", "", "source Git repository URL")
	cmd.Flags().StringVar(&o.webhookSecret, "webhook-secret", "", "source Git repository webhook secret")
	cmd.Flags().StringVar(&o.appName, "app-name", "", "the name of the application where the service will be added")
	cmd.Flags().StringVar(&o.serviceName, "service-name", "", "the name of the service to be added")
	cmd.Flags().StringVar(&o.envName, "env-name", "", "the name of the environment where the service will be added")
	cmd.Flags().StringVar(&o.imageRepo, "image-repo", "", "used to push built images")
	cmd.Flags().StringVar(&o.internalRegistryHostname, "internal-registry-hostname", "image-registry.openshift-image-registry.svc:5000", "internal image registry hostname")
	cmd.Flags().StringVar(&o.pipelinesFilePath, "pipelines-file", "pipelines.yaml", "path to pipelines file")

	// required flags
	_ = cmd.MarkFlagRequired("service-name")
	_ = cmd.MarkFlagRequired("app-name")
	_ = cmd.MarkFlagRequired("env-name")
	return cmd
}
