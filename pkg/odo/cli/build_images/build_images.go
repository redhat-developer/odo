package build_images

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/templates"

	"github.com/redhat-developer/odo/pkg/devfile/image"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
	"github.com/redhat-developer/odo/pkg/odo/util"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
)

// RecommendedCommandName is the recommended command name
const RecommendedCommandName = "build-images"

// BuildImagesOptions encapsulates the options for the odo command
type BuildImagesOptions struct {
	// Context
	*genericclioptions.Context

	// Clients
	clientset *clientset.Clientset

	// Flags
	pushFlag    bool
	contextFlag string
}

var _ genericclioptions.Runnable = (*BuildImagesOptions)(nil)

var buildImagesExample = templates.Examples(`
  # Build images defined in the devfile
  %[1]s

  # Build images and push them to their registries
  %[1]s --push
`)

// NewBuildImagesOptions creates a new BuildImagesOptions instance
func NewBuildImagesOptions() *BuildImagesOptions {
	return &BuildImagesOptions{}
}

func (o *BuildImagesOptions) SetClientset(clientset *clientset.Clientset) {
	o.clientset = clientset
}

// Complete completes LoginOptions after they've been created
func (o *BuildImagesOptions) Complete(cmdline cmdline.Cmdline, args []string) (err error) {
	o.Context, err = genericclioptions.New(genericclioptions.NewCreateParameters(cmdline).NeedDevfile(o.contextFlag).IsOffline())
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
func (o *BuildImagesOptions) Run(ctx context.Context) (err error) {
	devfileObj := o.Context.EnvSpecificInfo.GetDevfileObj()
	path := filepath.Dir(o.Context.EnvSpecificInfo.GetDevfilePath())
	return image.BuildPushImages(o.clientset.FS, devfileObj, path, o.pushFlag)
}

// NewCmdBuildImages implements the odo command
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
	buildImagesCmd.Annotations = map[string]string{"command": "main"}
	buildImagesCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)
	buildImagesCmd.Flags().BoolVar(&o.pushFlag, "push", false, "If true, build and push the images")
	util.AddContextFlag(buildImagesCmd, &o.contextFlag)
	clientset.Add(buildImagesCmd, clientset.FILESYSTEM)

	return buildImagesCmd
}
