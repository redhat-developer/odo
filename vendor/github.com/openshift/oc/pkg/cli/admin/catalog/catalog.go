// +build linux

package catalog

import (
	"github.com/spf13/cobra"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	kcmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/templates"
)

func init() {
	addCommand = func(streams genericclioptions.IOStreams, cmd *cobra.Command) {
		cmd.AddCommand(newCmd(streams))
	}
}

func newCmd(streams genericclioptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "catalog",
		Short: "Tools for managing the OpenShift OLM Catalogs",
		Long: templates.LongDesc(`
			This tool is used to extract and mirror the contents of catalogs for Operator
			Lifecycle Manager.

			The subcommands allow you to build catalog images from a source (such as appregistry) 
			and mirror its content across registries.
			`),
		Run: kcmdutil.DefaultSubCommandRun(streams.ErrOut),
	}
	cmd.AddCommand(NewBuildImage(streams))
	cmd.AddCommand(NewMirrorCatalog(streams))
	return cmd
}
