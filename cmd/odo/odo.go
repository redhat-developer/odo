package main

import (
	"flag"
	"os"
	"strings"

	"github.com/golang/glog"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/odo/cli"
	"github.com/openshift/odo/pkg/odo/cli/version"
	"github.com/openshift/odo/pkg/odo/util"
	"github.com/openshift/odo/pkg/odo/util/completion"
	"github.com/openshift/odo/pkg/preference"
	"github.com/posener/complete"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func main() {
	// create the complete command
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

	// Override the logging level by the value (if set) by the ODO_LOG_LEVEL env
	// The "-v" flag set on command line will take precedence over ODO_LOG_LEVEL env
	v := flag.CommandLine.Lookup("v").Value.String()
	// if the json flag is passed and is valid, we don't turn on ODO_LOG_LEVEL
	jsonFlagValue := flag.CommandLine.Lookup("o").Value.String()
	if level, ok := os.LookupEnv("ODO_LOG_LEVEL"); ok && v == "0" && strings.ToLower(jsonFlagValue) != "json" {
		_ = flag.CommandLine.Set("v", level)
	}

	// run the completion, in case that the completion was invoked
	// and ran as a completion script or handled a flag that passed
	// as argument, the Run method will return true,
	// in that case, our program have nothing to do and should return.
	if cmp.Complete() {
		return
	}

	// Call commands
	// checking the value of updatenotification in config
	// before proceeding with fetching the latest version
	cfg, err := preference.New()
	if err != nil {
		util.LogErrorAndExit(err, "")
	}
	if cfg.GetUpdateNotification() {
		updateInfo := make(chan string)
		go version.GetLatestReleaseInfo(updateInfo)

		util.LogErrorAndExit(root.Execute(), "")
		select {
		case message := <-updateInfo:
			log.Italic(message)
		default:
			glog.V(4).Info("Could not get the latest release information in time. Never mind, exiting gracefully :)")
		}
	} else {
		util.LogErrorAndExit(root.Execute(), "")
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
