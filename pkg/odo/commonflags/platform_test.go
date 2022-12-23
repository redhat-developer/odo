package commonflags

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func TestUsePlatformFlagOK(t *testing.T) {
	cmd := &cobra.Command{}
	UsePlatformFlag(cmd)
	err := pflag.CommandLine.Set("platform", "cluster")
	if err != nil {
		t.Errorf("Set error should be nil but is %v", err)
	}
	err = CheckPlatformCommand(cmd)
	if err != nil {
		t.Errorf("Check error should be nil but is %v", err)
	}
}

func TestUsePlatformFlagNotUsed(t *testing.T) {
	cmd := &cobra.Command{}
	err := pflag.CommandLine.Set("platform", "cluster")
	if err != nil {
		t.Errorf("Set error should be nil but is %v", err)
	}
	err = CheckPlatformCommand(cmd)
	if err.Error() != "--platform flag is not supported for this command" {
		t.Errorf("Check error is %v", err)
	}
}

func TestUsePlatformFlagWrongValue(t *testing.T) {
	cmd := &cobra.Command{}
	err := pflag.CommandLine.Set("platform", "wrong-value")
	if err != nil {
		t.Errorf("Set error should be nil but is %v", err)
	}
	err = CheckPlatformCommand(cmd)
	if err.Error() != `wrong-value is not a valid target platform for --platform, please select either "cluster" (default) or "podman" (experimental)` {
		t.Errorf("Check error is %v", err)
	}
}
