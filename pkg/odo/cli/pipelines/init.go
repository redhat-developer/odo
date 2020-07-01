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
	// InitRecommendedCommandName the recommended command name
	InitRecommendedCommandName = "init"
)

var (
	initExample = ktemplates.Examples(`
	# Initialize OpenShift GitOps pipelines
	%[1]s 
	`)

	initLongDesc  = ktemplates.LongDesc(`Initialize GitOps pipelines`)
	initShortDesc = `Initialize pipelines`
)

// InitParameters encapsulates the parameters for the odo pipelines init command.
type InitParameters struct {
	*pipelines.InitOptions
	// generic context options common to all commands
	*genericclioptions.Context
}

// NewInitParameters bootstraps a InitParameters instance.
func NewInitParameters() *InitParameters {
	return &InitParameters{InitOptions: &pipelines.InitOptions{}}
}

// Complete completes InitParameters after they've been created.
//
// If the prefix provided doesn't have a "-" then one is added, this makes the
// generated environment names nicer to read.
func (io *InitParameters) Complete(name string, cmd *cobra.Command, args []string) error {
	io.Prefix = utility.MaybeCompletePrefix(io.Prefix)
	io.GitOpsRepoURL = utility.AddGitSuffixIfNecessary(io.GitOpsRepoURL)
	return nil
}

// Validate validates the parameters of the InitParameters.
func (io *InitParameters) Validate() error {
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
func (io *InitParameters) Run() error {
	err := pipelines.Init(io.InitOptions, ioutils.NewFilesystem())
	if err != nil {
		return err
	}
	log.Successf("Intialized GitOps sucessfully.")
	return nil
}

// NewCmdInit creates the project init command.
func NewCmdInit(name, fullName string) *cobra.Command {
	o := NewInitParameters()

	initCmd := &cobra.Command{
		Use:     name,
		Short:   initShortDesc,
		Long:    initLongDesc,
		Example: fmt.Sprintf(initExample, fullName),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	initCmd.Flags().StringVar(&o.GitOpsRepoURL, "gitops-repo-url", "", "GitOps repository e.g. https://github.com/organisation/repository")
	initCmd.MarkFlagRequired("gitops-repo-url")
	initCmd.Flags().StringVar(&o.GitOpsWebhookSecret, "gitops-webhook-secret", "", "provide the GitHub webhook secret for GitOps repository")
	initCmd.MarkFlagRequired("gitops-webhook-secret")
	initCmd.Flags().StringVar(&o.OutputPath, "output", ".", "folder path to add GitOps resources")
	initCmd.Flags().StringVarP(&o.Prefix, "prefix", "p", "", "add a prefix to the environment names")
	initCmd.Flags().StringVar(&o.DockerConfigJSONFilename, "dockercfgjson", "", "dockercfg json pathname")
	initCmd.Flags().StringVar(&o.InternalRegistryHostname, "internal-registry-hostname", "image-registry.openshift-image-registry.svc:5000", "internal image registry hostname")
	initCmd.Flags().StringVar(&o.ImageRepo, "image-repo", "", "image repository in this form <registry>/<username>/<repository> or <project>/<app> for internal registry")
	initCmd.Flags().StringVarP(&o.SealedSecretsNamespace, "sealed-secrets-ns", "", "", "namespace in which the Sealed Secrets operator is installed, automatically generated secrets are encrypted with this operator")
	initCmd.MarkFlagRequired("sealed-secrets-ns")

	return initCmd
}
