package commonflags

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func TestUseRunOnFlagOK(t *testing.T) {
	cmd := &cobra.Command{}
	UseRunOnFlag(cmd)
	err := pflag.CommandLine.Set("run-on", "cluster")
	if err != nil {
		t.Errorf("Set error should be nil but is %v", err)
	}
	err = CheckRunOnCommand(cmd)
	if err != nil {
		t.Errorf("Check error should be nil but is %v", err)
	}
}

func TestUseRunOnFlagNotUsed(t *testing.T) {
	cmd := &cobra.Command{}
	err := pflag.CommandLine.Set("run-on", "cluster")
	if err != nil {
		t.Errorf("Set error should be nil but is %v", err)
	}
	err = CheckRunOnCommand(cmd)
	if err.Error() != "--run-on flag is not supported for this command" {
		t.Errorf("Check error is %v", err)
	}
}

func TestUseRunOnFlagWrongValue(t *testing.T) {
	cmd := &cobra.Command{}
	err := pflag.CommandLine.Set("run-on", "wrong-value")
	if err != nil {
		t.Errorf("Set error should be nil but is %v", err)
	}
	err = CheckRunOnCommand(cmd)
	if err.Error() != "wrong-value is not a valid target platform for --run-on, please select either cluster (default) or podman" {
		t.Errorf("Check error is %v", err)
	}
}
