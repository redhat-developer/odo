package manifest

import (
	"fmt"

	"github.com/openshift/odo/pkg/manifest"
	"github.com/openshift/odo/pkg/manifest/ioutils"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/spf13/cobra"

	ktemplates "k8s.io/kubernetes/pkg/kubectl/util/templates"
)

const (
	// BuildRecommendedCommandName the recommended command name
	BuildRecommendedCommandName = "build"
)

var (
	buildExample = ktemplates.Examples(`
	# Build files from manifest
	%[1]s 
	`)

	buildLongDesc  = ktemplates.LongDesc(`Build GitOps manifest files`)
	buildShortDesc = `Build manifest files`
)

// BuildParameters encapsulates the parameters for the odo manifest build command.
type BuildParameters struct {
	manifest      string
	output        string // path to add Gitops resources
	gitopsRepoURL string
	// generic context options common to all commands
	*genericclioptions.Context
}

// NewBuildParameters bootstraps a BuildParameters instance.
func NewBuildParameters() *BuildParameters {
	return &BuildParameters{}
}

// Complete completes BuildParameters after they've been created.
func (io *BuildParameters) Complete(name string, cmd *cobra.Command, args []string) error {
	return nil
}

// Validate validates the parameters of the BuildParameters.
func (io *BuildParameters) Validate() error {
	return nil
}

// Run runs the project bootstrap command.
func (io *BuildParameters) Run() error {
	options := manifest.BuildParameters{
		ManifestFilename: io.manifest,
		OutputPath:       io.output,
		RepositoryURL:    io.gitopsRepoURL,
	}
	return manifest.BuildResources(&options, ioutils.NewFilesystem())
}

// NewCmdBuild creates the manifest build command.
func NewCmdBuild(name, fullName string) *cobra.Command {
	o := NewBuildParameters()
	buildCmd := &cobra.Command{
		Use:     name,
		Short:   buildShortDesc,
		Long:    buildLongDesc,
		Example: fmt.Sprintf(buildExample, fullName),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	buildCmd.Flags().StringVar(&o.output, "output", ".", "folder path to add Gitops resources")
	buildCmd.Flags().StringVar(&o.manifest, "manifest", "manifest.yaml", "path to manifest file")
	buildCmd.Flags().StringVar(&o.gitopsRepoURL, "gitops-repo-url", "", "full URL for the repository where the manifest and configuration are stored")
	return buildCmd
}
