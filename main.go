package main

import (
	"flag"
	"github.com/posener/complete"
	"github.com/redhat-developer/odo/cmd"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func main() {
	// create the complete command
	root := cmd.RootCmd()
	rootCmp := createCompletion(root)
	cmp := complete.New("odo", rootCmp)

	// AddFlags adds the completion flags to the program flags,
	// in case of using non-default flag set, it is possible to pass
	// it as an argument.
	// it is possible to set custom flags name
	// so when one will type 'self -h', he will see '-complete' to install the
	// completion and -uncomplete to uninstall it.
	cmp.CLI.InstallName = "complete"
	cmp.CLI.UninstallName = "uncomplete"
	cmp.AddFlags(nil)

	// parse the flags - both the program's flags and the completion flags
	flag.Parse()

	// run the completion, in case that the completion was invoked
	// and ran as a completion script or handled a flag that passed
	// as argument, the Run method will return true,
	// in that case, our program have nothing to do and should return.
	if cmp.Complete() {
		return
	}

	// Call commands
	cmd.Execute()
}

func createCompletion(root *cobra.Command) complete.Command {
	rootCmp := complete.Command{}
	rootCmp.Flags = make(complete.Flags)
	addFlags := func(flag *pflag.Flag) {
		if flag.Hidden {
			return
		}
		suggester, ok := cmd.Suggesters[cmd.GetFlagSuggesterName(root, flag.Name)]
		if !ok {
			suggester = complete.PredictAnything
		}

		if len(flag.Shorthand) > 0 {
			rootCmp.Flags["-"+flag.Shorthand] = suggester
		}

		rootCmp.Flags["--"+flag.Name] = suggester
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

	suggester, ok := cmd.Suggesters[cmd.GetCommandSuggesterName(root)]
	if !ok {
		suggester = complete.PredictNothing
	}
	rootCmp.Args = suggester

	return rootCmp
}
