package build_images

import (
	"fmt"

	"github.com/openshift/odo/pkg/devfile/image"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	odoutil "github.com/openshift/odo/pkg/odo/util"
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/templates"
)

// RecommendedCommandName is the recommended command name
const RecommendedCommandName = "build-images"

// LoginOptions encapsulates the options for the odo command
type BuildImagesOptions struct {
	push bool
	*genericclioptions.Context
	componentContext string
}

var buildImagesExample = templates.Examples(`
  # Build images defined in the devfile
  %[1]s

  # Build images and push them to their registries
  %[1]s --push
`)

// NewLoginOptions creates a new LoginOptions instance
func NewBuildImagesOptions() *BuildImagesOptions {
	return &BuildImagesOptions{}
}

// Complete completes LoginOptions after they've been created
func (o *BuildImagesOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	o.Context, err = genericclioptions.New(genericclioptions.NewCreateParameters(cmd).NeedDevfile(o.componentContext).IsOffline())
	if err != nil {
		return err
	}
	return
}

// Validate validates the LoginOptions based on completed values
func (o *BuildImagesOptions) Validate() (err error) {
	return
}

// Run contains the logic for the odo command
func (o *BuildImagesOptions) Run(cmd *cobra.Command) (err error) {
	return image.BuildPushImages(o.Context, o.push)
}

// NewCmdLogin implements the odo command
func NewCmdBuildImages(name, fullName string) *cobra.Command {
	o := NewBuildImagesOptions()
	buildImagesCmd := &cobra.Command{
		Use:     name,
		Short:   "Build images",
		Long:    "Build images defined in the devfile",
		Example: fmt.Sprintf(buildImagesExample, fullName),
		Args:    cobra.MaximumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	// Add a defined annotation in order to appear in the help menu
	buildImagesCmd.Annotations = map[string]string{"command": "utility"}
	buildImagesCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)
	buildImagesCmd.Flags().BoolVar(&o.push, "push", false, "If true, build and push the images")
	genericclioptions.AddContextFlag(buildImagesCmd, &o.componentContext)
	return buildImagesCmd
}
