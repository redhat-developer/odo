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
	if io.Prefix != "" && !strings.HasSuffix(io.Prefix, "-") {
		io.Prefix = io.Prefix + "-"
	}
	io.GitOpsRepoURL = utility.AddGitSuffixIfNecessary(io.GitOpsRepoURL)
	io.AppRepoURL = utility.AddGitSuffixIfNecessary(io.AppRepoURL)
	return nil
}

// Validate validates the parameters of the BootstrapParameters.
func (io *BootstrapParameters) Validate() error {
	gr, err := url.Parse(io.GitOpsRepoURL)
	if err != nil {
		return fmt.Errorf("failed to parse url %s: %v", io.GitOpsRepoURL, err)
	}

	// TODO: this won't work with GitLab as the repo can have more path elements.
	if len(removeEmptyStrings(strings.Split(gr.Path, "/"))) != 2 {
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

	initCmd := &cobra.Command{
		Use:     name,
		Short:   bootstrapShortDesc,
		Long:    bootstrapLongDesc,
		Example: fmt.Sprintf(bootstrapExample, fullName),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	initCmd.Flags().StringVar(&o.GitOpsRepoURL, "gitops-repo-url", "", "GitOps repository e.g. https://github.com/organisation/repository")
	initCmd.Flags().StringVar(&o.GitOpsWebhookSecret, "gitops-webhook-secret", "", "provide the GitHub webhook secret for GitOps repository")

	initCmd.Flags().StringVar(&o.AppRepoURL, "app-repo-url", "", "Application source e.g. https://github.com/organisation/application")
	initCmd.Flags().StringVar(&o.AppWebhookSecret, "app-webhook-secret", "", "Provide the GitHub webhook secret for Application repository")

	initCmd.Flags().StringVar(&o.DockerConfigJSONFilename, "dockercfgjson", "", "provide the dockercfgjson path")
	initCmd.Flags().StringVar(&o.InternalRegistryHostname, "internal-registry-hostname", "image-registry.openshift-image-registry.svc:5000", "internal image registry hostname")
	initCmd.Flags().StringVar(&o.OutputPath, "output", ".", "folder path to add Gitops resources")
	initCmd.Flags().StringVarP(&o.Prefix, "prefix", "p", "", "add a prefix to the environment names")
	initCmd.Flags().StringVarP(&o.ImageRepo, "image-repo", "", "", "used to push built images")

	initCmd.MarkFlagRequired("gitops-repo-url")
	initCmd.MarkFlagRequired("gitops-webhook-secret")
	initCmd.MarkFlagRequired("app-repo-url")
	initCmd.MarkFlagRequired("app-webhook-secret")
	initCmd.MarkFlagRequired("dockercfgjson")
	initCmd.MarkFlagRequired("image-repo")

	return initCmd
}

func removeEmptyStrings(s []string) []string {
	nonempty := []string{}
	for _, v := range s {
		if v != "" {
			nonempty = append(nonempty, v)
		}
	}
	return nonempty
}
