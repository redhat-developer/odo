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
	%[1]s username org/repo
	`)

	bootstrapLongDesc  = ktemplates.LongDesc(`Bootstrap GitOps pipelines`)
	bootstrapShortDesc = `Bootstrap pipelines`
)

// BootstrapOptions encapsulates the options for the odo pipelines bootstrap
// command.
type BootstrapOptions struct {
	quayUsername       string
	gitRepo            string // e.g. tekton/triggers
	prefix             string // used to generate the environments in a shared cluster
	githubToken        string
	quayIOAuthFilename string
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
	if bo.prefix != "" && !strings.HasSuffix(bo.prefix, "-") {
		bo.prefix = bo.prefix + "-"
	}
	return nil
}

// Validate validates the parameters of the BootstrapOptions.
func (bo *BootstrapOptions) Validate() error {
	// TODO: this won't work with GitLab as the repo can have more path elements.
	if len(strings.Split(bo.gitRepo, "/")) != 2 {
		return fmt.Errorf("repo must be org/repo: %s", bo.gitRepo)
	}
	return nil
}

// Run runs the project bootstrap command.
func (bo *BootstrapOptions) Run() error {
	options := pipelines.BootstrapOptions{
		GithubToken:      bo.githubToken,
		GitRepo:          bo.gitRepo,
		Prefix:           bo.prefix,
		QuayAuthFileName: bo.quayIOAuthFilename,
		QuayUserName:     bo.quayUsername,
	}
	return pipelines.Bootstrap(&options)
}

// NewCmdBootstrap creates the project bootstrap command.
func NewCmdBootstrap(name, fullName string) *cobra.Command {
	o := NewBootstrapOptions()

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
	bootstrapCmd.Flags().StringVar(&o.quayUsername, "quay-username", "", "Image registry username")
	bootstrapCmd.MarkFlagRequired("quay-username")
	bootstrapCmd.Flags().StringVar(&o.githubToken, "github-token", "", "provide the Github token")
	bootstrapCmd.MarkFlagRequired("github-token")
	bootstrapCmd.Flags().StringVar(&o.quayIOAuthFilename, "dockerconfigjson", "", "Docker configuration json filename")
	bootstrapCmd.MarkFlagRequired("dockerconfigjson")
	bootstrapCmd.Flags().StringVar(&o.gitRepo, "git-repository", "", "provide the base repository")
	bootstrapCmd.MarkFlagRequired("git-repository")
	return bootstrapCmd
}
