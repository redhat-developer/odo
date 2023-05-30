package commonflags

import (
	"context"
	"flag"
	"github.com/redhat-developer/odo/pkg/odo/cli/feature"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"k8s.io/klog"
)

const (
	APIServerFlagName     = "api-server"
	APIServerPortFlagName = "api-server-port"
)

func AddAPIServerFlag(ctx context.Context) {
	if feature.IsEnabled(ctx, feature.APIServerFlag) {
		flag.CommandLine.Bool(APIServerFlagName, false, "Start the API Server; this is an experimental feature")
		flag.CommandLine.Int(APIServerPortFlagName, 0, "Define custom port for API Server; this flag should be used in combination with --api-server flag.")
	}
}

// GetAPIServerValue returns value of --api-server flag or default value
func GetAPIServerValue(cmd cmdline.Cmdline) bool {
	if !cmd.IsFlagSet(APIServerFlagName) {
		return false
	}
	value, err := cmd.FlagValueBool(APIServerFlagName)
	if err != nil {
		klog.V(1).Infof("failed to get parse --%s", APIServerFlagName)
		return false
	}
	return value
}

// GetAPIServerPortValue returns value of --api-server-port flag or default value
func GetAPIServerPortValue(cmd cmdline.Cmdline) int {
	if !cmd.IsFlagSet(APIServerFlagName) {
		return 0
	}
	value, err := cmd.FlagValueInt(APIServerPortFlagName)
	if err != nil {
		klog.V(1).Infof("failed to parse --%s", APIServerPortFlagName)
		return 0
	}
	return value
}
