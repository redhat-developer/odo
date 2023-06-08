package main

import (
	"flag"
	"os"

	"github.com/spf13/pflag"
	"k8s.io/klog"
)

func resetGlobalFlags() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)
	klog.InitFlags(nil)
}
