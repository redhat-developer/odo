package pipelines

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/openshift/odo/pkg/pipelines"
	"github.com/openshift/odo/pkg/pipelines/ioutils"
	"github.com/spf13/cobra"

	ktemplates "k8s.io/kubernetes/pkg/kubectl/util/templates"
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
	dockercfgjson            string // filepath name to dockerconfigjson file
	gitOpsRepoURL            string // This is where the pipelines and configuration are.
	gitOpsWebhookSecret      string // used to create Github's shared webhook secret for gitops repo
	output                   string // path to add Gitops resources
	prefix                   string // used to generate the environments in a shared cluster
	imageRepo                string
	internalRegistryHostname string
	// generic context options common to all commands
	*genericclioptions.Context
}

// NewInitParameters bootstraps a InitParameters instance.
func NewInitParameters() *InitParameters {
	return &InitParameters{}
}

// Complete completes InitParameters after they've been created.
//
// If the prefix provided doesn't have a "-" then one is added, this makes the
// generated environment names nicer to read.
func (io *InitParameters) Complete(name string, cmd *cobra.Command, args []string) error {
	if io.prefix != "" && !strings.HasSuffix(io.prefix, "-") {
		io.prefix = io.prefix + "-"
	}
	return nil
}

// Validate validates the parameters of the InitParameters.
func (io *InitParameters) Validate() error {
	gr, err := url.Parse(io.gitOpsRepoURL)
	if err != nil {
		return fmt.Errorf("failed to parse url %s: %w", io.gitOpsRepoURL, err)
	}

	// TODO: this won't work with GitLab as the repo can have more path elements.
	if len(removeEmptyStrings(strings.Split(gr.Path, "/"))) != 2 {
		return fmt.Errorf("repo must be org/repo: %s", strings.Trim(gr.Path, ".git"))
	}
	return nil
}

// Run runs the project bootstrap command.
func (io *InitParameters) Run() error {
	options := pipelines.InitParameters{
		DockerConfigJSONFilename: io.dockercfgjson,
		GitOpsWebhookSecret:      io.gitOpsWebhookSecret,
		GitOpsRepoURL:            io.gitOpsRepoURL,
		OutputPath:               io.output,
		Prefix:                   io.prefix,
		ImageRepo:                io.imageRepo,
		InternalRegistryHostname: io.internalRegistryHostname,
	}
	return pipelines.Init(&options, ioutils.NewFilesystem())
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

	initCmd.Flags().StringVar(&o.gitOpsRepoURL, "gitops-repo-url", "", "GitOps repository e.g. https://github.com/organisation/repository")
	initCmd.MarkFlagRequired("gitops-repo-url")
	initCmd.Flags().StringVar(&o.gitOpsWebhookSecret, "gitops-webhook-secret", "", "provide the GitHub webhook secret for GitOps repository")
	initCmd.MarkFlagRequired("gitops-webhook-secret")
	initCmd.Flags().StringVar(&o.output, "output", ".", "folder path to add GitOps resources")
	initCmd.MarkFlagRequired("output")
	initCmd.Flags().StringVarP(&o.prefix, "prefix", "p", "", "add a prefix to the environment names")
	initCmd.Flags().StringVar(&o.dockercfgjson, "dockercfgjson", "", "dockercfg json pathname")
	initCmd.Flags().StringVar(&o.internalRegistryHostname, "internal-registry-hostname", "image-registry.openshift-image-registry.svc:5000", "internal image registry hostname")
	initCmd.Flags().StringVar(&o.imageRepo, "image-repo", "", "image repository in this form <registry>/<username>/<repository> or <project>/<app> for internal registry")

	return initCmd
}
