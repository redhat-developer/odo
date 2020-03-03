package pipelines

import (
	"fmt"
	"strings"

	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/openshift/odo/pkg/pipelines"
	"github.com/spf13/cobra"

	ktemplates "k8s.io/kubernetes/pkg/kubectl/util/templates"
)

const (
	BootstrapRecommendedCommandName = "bootstrap"
)

var (
	bootstrapExample = ktemplates.Examples(`
	# Bootstrap OpenShift pipelines in a cluster
	%[1]s 
	`)

	bootstrapLongDesc  = ktemplates.LongDesc(`Bootstrap GitOps CI/CD Pipelines`)
	bootstrapShortDesc = `Bootstrap pipelines`
)

// BootstrapParameters encapsulates the paratmeters for the odo pipelines bootstrap
// command.
type BootstrapParameters struct {
	deploymentPath           string
	githubToken              string
	gitRepo                  string // e.g. tekton/triggers
	imageRepo                string
	internalRegistryHostname string
	prefix                   string // used to generate the environments in a shared cluster
	dockerConfigJSONFileName string
	skipChecks               bool
	// generic context options common to all commands
	*genericclioptions.Context
}

// NewBootstrapParameters bootstraps a BootstrapParameters instance.
func NewBootstrapParameters() *BootstrapParameters {
	return &BootstrapParameters{}
}

// Complete completes BootstrapParameters after they've been created.
//
// If the prefix provided doesn't have a "-" then one is added, this makes the
// generated environment names nicer to read.
func (bo *BootstrapParameters) Complete(name string, cmd *cobra.Command, args []string) error {
	if bo.prefix != "" && !strings.HasSuffix(bo.prefix, "-") {
		bo.prefix = bo.prefix + "-"
	}
	return nil
}

// Validate validates the parameters of the BootstrapParameters.
func (bo *BootstrapParameters) Validate() error {
	// TODO: this won't work with GitLab as the repo can have more path elements.
	if len(strings.Split(bo.gitRepo, "/")) != 2 {
		return fmt.Errorf("repo must be org/repo: %s", bo.gitRepo)
	}
	return nil
}

// Run runs the project bootstrap command.
func (bo *BootstrapParameters) Run() error {
	options := pipelines.BootstrapParameters{
		DeploymentPath:           bo.deploymentPath,
		GithubToken:              bo.githubToken,
		GitRepo:                  bo.gitRepo,
		ImageRepo:                bo.imageRepo,
		InternalRegistryHostname: bo.internalRegistryHostname,
		Prefix:                   bo.prefix,
		DockerConfigJSONFileName: bo.dockerConfigJSONFileName,
		SkipChecks:               bo.skipChecks,
	}

	return pipelines.Bootstrap(&options)
}

// NewCmdBootstrap creates the project bootstrap command.
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

	bootstrapCmd.Flags().StringVarP(&o.prefix, "prefix", "p", "", "add a prefix to the environment names")
	bootstrapCmd.Flags().StringVar(&o.githubToken, "github-token", "", "provide the Github token")
	bootstrapCmd.MarkFlagRequired("github-token")
	bootstrapCmd.Flags().StringVar(&o.dockerConfigJSONFileName, "dockerconfigjson", "", "Docker configuration json filename")
	bootstrapCmd.Flags().StringVar(&o.gitRepo, "git-repo", "", "git repository in this form <username>/<repository>")
	bootstrapCmd.MarkFlagRequired("git-repo")
	bootstrapCmd.Flags().StringVar(&o.imageRepo, "image-repo", "", "image repository in this form <registry>/<username>/<repository> or <project>/<app> for internal registry")
	bootstrapCmd.MarkFlagRequired("image-repo")
	bootstrapCmd.Flags().StringVar(&o.deploymentPath, "deployment-path", "", "deployment folder path name")
	bootstrapCmd.MarkFlagRequired("deployment-path")
	bootstrapCmd.Flags().BoolVarP(&o.skipChecks, "skip-checks", "b", false, "skip Tekton installation checks")
	bootstrapCmd.Flags().StringVar(&o.internalRegistryHostname, "internal-registry-hostname", "image-registry.openshift-image-registry.svc:5000", "internal image registry hostname")

	return bootstrapCmd
}
