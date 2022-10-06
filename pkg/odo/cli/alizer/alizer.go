package alizer

import (
	"context"
	"errors"
	"os"

	"github.com/redhat-developer/odo/pkg/alizer"
	"github.com/redhat-developer/odo/pkg/api"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/commonflags"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"

	"github.com/spf13/cobra"
)

const RecommendedCommandName = "analyze"

type AlizerOptions struct {
	clientset *clientset.Clientset
}

var _ genericclioptions.Runnable = (*AlizerOptions)(nil)
var _ genericclioptions.JsonOutputter = (*AlizerOptions)(nil)

// NewAlizerOptions creates a new AlizerOptions instance
func NewAlizerOptions() *AlizerOptions {
	return &AlizerOptions{}
}

func (o *AlizerOptions) SetClientset(clientset *clientset.Clientset) {
	o.clientset = clientset
}

func (o *AlizerOptions) Complete(ctx context.Context, cmdline cmdline.Cmdline, args []string) (err error) {
	return nil
}

func (o *AlizerOptions) Validate(ctx context.Context) error {
	return nil
}

func (o *AlizerOptions) Run(ctx context.Context) (err error) {
	return errors.New("this command can be run with json output only, please use the flag: -o json")
}

// Run contains the logic for the odo command
func (o *AlizerOptions) RunForJsonOutput(ctx context.Context) (out interface{}, err error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	df, reg, err := o.clientset.AlizerClient.DetectFramework(cwd)
	if err != nil {
		return nil, err
	}
	result := alizer.GetDevfileLocationFromDetection(df, reg)
	return []api.DevfileLocation{*result}, nil
}

func NewCmdAlizer(name, fullName string) *cobra.Command {
	o := NewAlizerOptions()
	alizerCmd := &cobra.Command{
		Use:         name,
		Short:       "Detect devfile to use based on files present in current directory",
		Long:        "Detect devfile to use based on files present in current directory",
		Args:        cobra.MaximumNArgs(0),
		Annotations: map[string]string{},
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}
	clientset.Add(alizerCmd, clientset.ALIZER)
	commonflags.UseOutputFlag(alizerCmd)
	alizerCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)
	alizerCmd.Annotations["command"] = "utility"
	return alizerCmd
}
