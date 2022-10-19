package cmdline

import (
	"context"
	"errors"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/redhat-developer/odo/pkg/kclient"

	dfutil "github.com/devfile/library/pkg/util"
)

type Cobra struct {
	cmd *cobra.Command
}

var _ Cmdline = (*Cobra)(nil)

func NewCobra(cmd *cobra.Command) *Cobra {
	return &Cobra{
		cmd: cmd,
	}
}

func (o *Cobra) Context() context.Context {
	return o.cmd.Context()
}

func (o *Cobra) GetArgsAfterDashes(args []string) ([]string, error) {
	l := o.cmd.ArgsLenAtDash()
	if l < 0 {
		return nil, errors.New("no argument passed after dash")
	}
	return args[l:], nil
}

func (o *Cobra) GetParentName() string {
	if !o.cmd.HasParent() {
		return ""
	}
	return o.cmd.Parent().Name()
}

func (o *Cobra) GetRootName() string {
	return o.cmd.Root().Name()
}

func (o *Cobra) GetName() string {
	return o.cmd.Name()
}

func (o *Cobra) IsFlagSet(flagName string) bool {
	return o.cmd.Flags().Changed(flagName)
}

func (o *Cobra) GetWorkingDirectory() (string, error) {
	return dfutil.GetAbsPath(".")
}

// FlagValueIfSet retrieves the value of the specified flag if it is set for the given command
func (o *Cobra) FlagValue(flagName string) (string, error) {
	return o.cmd.Flags().GetString(flagName)
}

// FlagValueIfSet retrieves the value of the specified flag if it is set for the given command
func (o *Cobra) FlagValueIfSet(flagName string) string {
	flag, _ := o.cmd.Flags().GetString(flagName)
	return flag
}

// FlagValues retrieves the value of the specified flag
func (o *Cobra) FlagValues(flagName string) ([]string, error) {
	return o.cmd.Flags().GetStringArray(flagName)
}

func (o *Cobra) GetKubeClient() (kclient.ClientInterface, error) {
	return kclient.New()
}

func (o *Cobra) GetFlags() map[string]string {
	flags := map[string]string{}
	o.cmd.Flags().Visit(func(f *pflag.Flag) {
		flags[f.Name] = f.Value.String()
	})
	return flags
}
