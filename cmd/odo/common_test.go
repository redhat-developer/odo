package main

import (
	"bytes"
	"context"
	"flag"
	"os"
	"testing"

	"github.com/redhat-developer/odo/pkg/config"
	envcontext "github.com/redhat-developer/odo/pkg/config/context"
	"github.com/redhat-developer/odo/pkg/odo/cli"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
	"github.com/spf13/pflag"
	"k8s.io/klog"
)

func resetGlobalFlags() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)
	klog.InitFlags(nil)
}

func runCommand(
	t *testing.T,
	args []string,
	clientset clientset.Clientset,
	f func(err error, stdout, stderr string),
) {

	// We are running the test on a new and empty directory (on real filesystem)
	originWd, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}
	defer func() {
		_ = os.Chdir(originWd)
	}()
	cwd := t.TempDir()
	err = os.Chdir(cwd)
	if err != nil {
		t.Error(err)
	}

	ctx := context.Background()
	envConfig, err := config.GetConfiguration()
	if err != nil {
		t.Fatal(err)
	}
	ctx = envcontext.WithEnvConfig(ctx, *envConfig)

	resetGlobalFlags()

	var stdoutB, stderrB bytes.Buffer

	clientset.Stdout = &stdoutB
	clientset.Stderr = &stderrB
	root := cli.NewCmdOdo(ctx, cli.OdoRecommendedName, cli.OdoRecommendedName, clientset)

	root.SetOut(&stdoutB)
	root.SetErr(&stderrB)

	root.SetArgs(args)

	err = root.ExecuteContext(ctx)

	stdout := stdoutB.String()
	stderr := stderrB.String()

	f(err, stdout, stderr)
}

func checkEqual[T comparable](t *testing.T, a, b T) {
	if a != b {
		t.Errorf("Name should be \"%v\" but is \"%v\"", b, a)
	}
}
