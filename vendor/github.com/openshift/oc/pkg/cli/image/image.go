package image

import (
	"fmt"

	"github.com/spf13/cobra"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	kcmdutil "k8s.io/kubectl/pkg/cmd/util"
	ktemplates "k8s.io/kubectl/pkg/util/templates"

	"github.com/openshift/oc/pkg/cli/image/append"
	"github.com/openshift/oc/pkg/cli/image/extract"
	"github.com/openshift/oc/pkg/cli/image/info"
	"github.com/openshift/oc/pkg/cli/image/mirror"
	"github.com/openshift/oc/pkg/cli/image/serve"
	cmdutil "github.com/openshift/oc/pkg/helpers/cmd"
)

var (
	imageLong = ktemplates.LongDesc(`
		Manage images on OpenShift

		These commands help you manage images on OpenShift.`)
)

// NewCmdImage exposes commands for modifying images.
func NewCmdImage(fullName string, f kcmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	image := &cobra.Command{
		Use:   "image COMMAND",
		Short: "Useful commands for managing images",
		Long:  imageLong,
		Run:   kcmdutil.DefaultSubCommandRun(streams.ErrOut),
	}

	name := fmt.Sprintf("%s image", fullName)

	groups := ktemplates.CommandGroups{
		{
			Message: "View or copy images:",
			Commands: []*cobra.Command{
				info.NewInfo(name, streams),
				mirror.NewCmdMirrorImage(name, streams),
			},
		},
		{
			Message: "Advanced commands:",
			Commands: []*cobra.Command{
				serve.New(name, streams),
				append.NewCmdAppendImage(name, streams),
				extract.New(name, streams),
			},
		},
	}
	groups.Add(image)
	cmdutil.ActsAsRootCommand(image, []string{"options"}, groups...)
	return image
}
