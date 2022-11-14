package commonflags

import (
	"flag"
	"os"
	"testing"

	"github.com/spf13/pflag"
	"k8s.io/klog"

	"github.com/redhat-developer/odo/pkg/odo/cli/feature"
)

func TestMain(m *testing.M) {
	// --run-on is considered experimental for now. As such, to exist, it requires the ODO_EXPERIMENTAL_MODE env var to be set.
	os.Setenv(feature.OdoExperimentalModeEnvVar, feature.OdoExperimentalModeTrue)
	klog.InitFlags(nil)
	AddOutputFlag()
	AddRunOnFlag()
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	os.Exit(m.Run())
}
