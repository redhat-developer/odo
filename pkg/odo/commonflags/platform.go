package commonflags

import (
	"context"
	"errors"
	"flag"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/redhat-developer/odo/pkg/odo/cli/feature"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
)

const (
	// PlatformFlagName is the name of the flag allowing user to specify target platform
	PlatformFlagName = "platform"
	PlatformCluster  = "cluster"
	PlatformPodman   = "podman"
	PlatformDefault  = PlatformCluster
)

// UsePlatformFlag indicates that a command accepts the --platform flag
func UsePlatformFlag(cmd *cobra.Command) {
	if cmd.Annotations == nil {
		cmd.Annotations = map[string]string{}
	}
	cmd.Annotations["platform"] = "true"
}

// AddPlatformFlag adds the --platform output flag to all commands
// We use "flag" in order to make this accessible throughtout ALL of odo, rather than the
// above traditional "persistentflags" usage that does not make it a pointer within the 'pflag'
// package
func AddPlatformFlag(ctx context.Context) {
	if feature.IsEnabled(ctx, feature.GenericPformFlag) {
		flag.CommandLine.String(PlatformFlagName, "", `Specify target platform, supported platforms: "cluster" (default), "podman" (experimental)`)
		_ = pflag.CommandLine.MarkHidden(PlatformFlagName)
	}
}

// CheckPlatformCommand checks if commands enabling platform flag are used correctly
func CheckPlatformCommand(cmd *cobra.Command) error {

	// Get the needed values
	platformFlag := pflag.Lookup(PlatformFlagName)
	hasFlagChanged := platformFlag != nil && platformFlag.Changed
	platform := cmd.Annotations["platform"]

	// Check the valid output
	if hasFlagChanged && platformFlag.Value.String() != PlatformPodman && platformFlag.Value.String() != PlatformCluster {
		return fmt.Errorf(`%s is not a valid target platform for --platform, please select either "cluster" (default) or "podman" (experimental)`, platformFlag.Value.String())
	}

	// Check that if -o json has been passed, that the command actually USES json.. if not, error out.
	if hasFlagChanged && platformFlag.Value.String() != "" && platform == "" {
		return errors.New("--platform flag is not supported for this command")
	}

	return nil
}

// GetPlatformValue returns value of --platform flag or default value
func GetPlatformValue(cmd cmdline.Cmdline) string {
	return cmd.FlagValueIfSet(PlatformFlagName)
}
