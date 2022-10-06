package commonflags

import (
	"flag"
	"os"
	"testing"

	"github.com/spf13/pflag"
	"k8s.io/klog"
)

func TestMain(m *testing.M) {
	klog.InitFlags(nil)
	AddOutputFlag()
	AddRunOnFlag()
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	os.Exit(m.Run())
}
