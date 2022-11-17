package commonflags

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func TestUseOutputFlagOK(t *testing.T) {
	cmd := &cobra.Command{}
	UseOutputFlag(cmd)
	err := pflag.CommandLine.Set("o", "json")
	if err != nil {
		t.Errorf("Set error should be nil but is %v", err)
	}
	err = CheckMachineReadableOutputCommand(nil, cmd)
	if err != nil {
		t.Errorf("Check error should be nil but is %v", err)
	}
}

func TestUseOutputFlagNotUsed(t *testing.T) {
	cmd := &cobra.Command{}
	err := pflag.CommandLine.Set("o", "json")
	if err != nil {
		t.Errorf("Set error should be nil but is %v", err)
	}
	err = CheckMachineReadableOutputCommand(nil, cmd)
	if err.Error() != "Machine readable output is not yet implemented for this command" {
		t.Errorf("Check error is %v", err)
	}
}

func TestUseOutputFlagWrongValue(t *testing.T) {
	cmd := &cobra.Command{}
	err := pflag.CommandLine.Set("o", "wrong-value")
	if err != nil {
		t.Errorf("Set error should be nil but is %v", err)
	}
	err = CheckMachineReadableOutputCommand(nil, cmd)
	if err.Error() != "Please input a valid output format for -o, available format: json" {
		t.Errorf("Check error is %v", err)
	}
}
