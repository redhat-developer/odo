package pipelines

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/odo/cli/pipelines/utility"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/openshift/odo/pkg/pipelines"
	"github.com/openshift/odo/pkg/pipelines/ioutils"
	"github.com/spf13/cobra"

	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

const (
	// BootstrapRecommendedCommandName the recommended command name
	BootstrapRecommendedCommandName = "bootstrap"
)

var (
	bootstrapExample = ktemplates.Examples(`
    # Bootstrap OpenShift pipelines.
    %[1]s 
    `)

	bootstrapLongDesc  = ktemplates.LongDesc(`Bootstrap GitOps CI/CD Manifest`)
	bootstrapShortDesc = `Bootstrap pipelines with a starter configuration`
)

// BootstrapParameters encapsulates the parameters for the odo pipelines init command.
type BootstrapParameters struct {
	*pipelines.BootstrapOptions
	// generic context options common to all commands
	*genericclioptions.Context
}

// NewBootstrapParameters bootstraps a BootstrapParameters instance.
func NewBootstrapParameters() *BootstrapParameters {
	return &BootstrapParameters{BootstrapOptions: &pipelines.BootstrapOptions{}}
}

// Complete completes BootstrapParameters after they've been created.
//
// If the prefix provided doesn't have a "-" then one is added, this makes the
// generated environment names nicer to read.
func (io *BootstrapParameters) Complete(name string, cmd *cobra.Command, args []string) error {
	io.Prefix = utility.MaybeCompletePrefix(io.Prefix)
	io.GitOpsRepoURL = utility.AddGitSuffixIfNecessary(io.GitOpsRepoURL)
	io.ServiceRepoURL = utility.AddGitSuffixIfNecessary(io.ServiceRepoURL)
	io.GitOpsRepoURL = utility.AddGitSuffixIfNecessary(io.GitOpsRepoURL)
	io.ServiceRepoURL = utility.AddGitSuffixIfNecessary(io.ServiceRepoURL)
	return nil
}

// Validate validates the parameters of the BootstrapParameters.
func (io *BootstrapParameters) Validate() error {
	gr, err := url.Parse(io.GitOpsRepoURL)
	if err != nil {
		return fmt.Errorf("failed to parse url %s: %v", io.GitOpsRepoURL, err)
	}

	// TODO: this won't work with GitLab as the repo can have more path elements.
	if len(utility.RemoveEmptyStrings(strings.Split(gr.Path, "/"))) != 2 {
		return fmt.Errorf("repo must be org/repo: %s", strings.Trim(gr.Path, ".git"))
	}

	return nil
}

// Run runs the project bootstrap command.
func (io *BootstrapParameters) Run() error {
	err := pipelines.Bootstrap(io.BootstrapOptions, ioutils.NewFilesystem())
	if err != nil {
		return err
	}
	log.Success("Bootstrapped GitOps sucessfully.")
	return nil
}

// NewCmdBootstrap creates the project init command.
func NewCmdBootstrap(name, fullName string) *cobra.Command {
	o := NewBootstrapParameters()

	bootstrapCmd := &cobra.Command{
		Use:     name,
		Short:   bootstrapShortDesc,
		Long:    bootstrapLongDesc,
		Example: fmt.Sprintf(bootstrapExample, fullName),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	bootstrapCmd.Flags().StringVar(&o.GitOpsRepoURL, "gitops-repo-url", "", "GitOps repository e.g. https://github.com/organisation/repository")
	bootstrapCmd.Flags().StringVar(&o.GitOpsWebhookSecret, "gitops-webhook-secret", "", "provide the GitHub webhook secret for GitOps repository (if not provided, it will be auto-generated)")

	bootstrapCmd.Flags().StringVar(&o.ServiceRepoURL, "service-repo-url", "", "Service source e.g. https://github.com/organisation/service")
	bootstrapCmd.Flags().StringVar(&o.ServiceWebhookSecret, "service-webhook-secret", "", "Provide the GitHub webhook secret for Service repository (if not provided, it will be auto-generated)")

	bootstrapCmd.Flags().StringVar(&o.DockerConfigJSONFilename, "dockercfgjson", "", "provide the dockercfgjson path")
	bootstrapCmd.Flags().StringVar(&o.InternalRegistryHostname, "internal-registry-hostname", "image-registry.openshift-image-registry.svc:5000", "internal image registry hostname")
	bootstrapCmd.Flags().StringVar(&o.OutputPath, "output", ".", "Folder path to add Gitops resources")
	bootstrapCmd.Flags().StringVarP(&o.Prefix, "prefix", "p", "", "Add a prefix to the environment names")
	bootstrapCmd.Flags().StringVarP(&o.ImageRepo, "image-repo", "", "", "Used to push built images")

	bootstrapCmd.Flags().StringVarP(&o.SealedSecretsNamespace, "sealed-secrets-ns", "", "", "namespace in which the Sealed Secrets operator is installed, automatically generated secrets are encrypted with this operator")
	bootstrapCmd.MarkFlagRequired("gitops-repo-url")
	bootstrapCmd.MarkFlagRequired("app-repo-url")
	bootstrapCmd.MarkFlagRequired("dockercfgjson")
	bootstrapCmd.MarkFlagRequired("image-repo")

	bootstrapCmd.MarkFlagRequired("gitops-repo-url")
	bootstrapCmd.MarkFlagRequired("gitops-webhook-secret")
	bootstrapCmd.MarkFlagRequired("app-repo-url")
	bootstrapCmd.MarkFlagRequired("app-webhook-secret")
	bootstrapCmd.MarkFlagRequired("dockercfgjson")
	bootstrapCmd.MarkFlagRequired("image-repo")
	bootstrapCmd.MarkFlagRequired("sealed-secrets-ns")

	return bootstrapCmd
}
