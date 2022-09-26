package main

import (
	"context"
	"flag"
	"os"

	"github.com/posener/complete"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/cli"
	"github.com/redhat-developer/odo/pkg/odo/cli/version"
	"github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/redhat-developer/odo/pkg/odo/util/completion"
	"github.com/redhat-developer/odo/pkg/preference"
	segment "github.com/redhat-developer/odo/pkg/segment/context"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/klog"
)

func main() {
	// create the complete command
	klog.InitFlags(nil)

	root := cli.NewCmdOdo(cli.OdoRecommendedName, cli.OdoRecommendedName)
	rootCmp := createCompletion(root)
	cmp := complete.New("odo", rootCmp)

	// AddFlags adds the completion flags to the program flags, specifying custom names
	cmp.CLI.InstallName = "complete"
	cmp.CLI.UninstallName = "uncomplete"
	cmp.AddFlags(nil)

	// add the completion flags to the root command, though they won't appear in completions
	root.Flags().AddGoFlagSet(flag.CommandLine)
	// override usage so that flag.Parse uses root command's usage instead of default one when invoked with -h
	flag.Usage = func() {
		_ = root.Help()
	}

	// parse the flags but hack around to avoid exiting with error code 2 on help
	flag.CommandLine.Init(os.Args[0], flag.ContinueOnError)
	args := os.Args[1:]
	if err := flag.CommandLine.Parse(args); err != nil {
		if err == flag.ErrHelp {
			os.Exit(0)
		}
	}

	// run the completion, in case that the completion was invoked
	// and ran as a completion script or handled a flag that passed
	// as argument, the Run method will return true,
	// in that case, our program have nothing to do and should return.
	if cmp.Complete() {
		return
	}

	cfg, err := preference.NewClient()
	if err != nil {
		util.LogErrorAndExit(err, "")
	}

	// Call commands
	// checking the value of updatenotification in config
	// before proceeding with fetching the latest version
	if cfg.GetUpdateNotification() {
		updateInfo := make(chan string)
		go version.GetLatestReleaseInfo(updateInfo)

		util.LogErrorAndExit(root.ExecuteContext(segment.NewContext(context.Background())), "")
		message := <-updateInfo
		if message != "" {
			log.Info(message)
		}
	} else {
		util.LogErrorAndExit(root.ExecuteContext(segment.NewContext(context.Background())), "")
	}
}

func createCompletion(root *cobra.Command) complete.Command {
	rootCmp := complete.Command{}
	rootCmp.Flags = make(complete.Flags)
	addFlags := func(flag *pflag.Flag) {
		if flag.Hidden {
			return
		}
		var handler complete.Predictor
		handler, ok := completion.GetCommandFlagHandler(root, flag.Name)
		if !ok {
			handler = complete.PredictAnything
		}

		if len(flag.Shorthand) > 0 {
			rootCmp.Flags["-"+flag.Shorthand] = handler
		}

		rootCmp.Flags["--"+flag.Name] = handler
	}
	root.LocalFlags().VisitAll(addFlags)
	root.InheritedFlags().VisitAll(addFlags)
	if root.HasAvailableSubCommands() {
		rootCmp.Sub = make(complete.Commands)
		for _, c := range root.Commands() {
			if !c.Hidden {
				rootCmp.Sub[c.Name()] = createCompletion(c)
			}
		}
	}

	var handler complete.Predictor
	handler, ok := completion.GetCommandHandler(root)
	if !ok {
		handler = complete.PredictNothing
	}
	rootCmp.Args = handler

	return rootCmp
}
