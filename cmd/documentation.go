package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func getFlags(flags *pflag.FlagSet) []string {
	var f []string
	flags.VisitAll(func(flag *pflag.Flag) {
		f = append(f, fmt.Sprintf("--%v", flag.Name))
	})
	return f
}

func flattenFlags(flags []string) string {
	var flagString string
	for _, flag := range flags {
		flagString = flagString + flag + " "
	}
	return flagString
}

func commandPrinter(command *cobra.Command, level int) string {
	var finalCommand string
	// add indentation
	for i := 0; i < level; i++ {
		finalCommand = finalCommand + "    "
	}
	finalCommand = finalCommand +
		command.Name() +
		" " +
		flattenFlags(getFlags(command.NonInheritedFlags())) +
		": " +
		command.Short +
		"\n"
	for _, subcommand := range command.Commands() {
		finalCommand = finalCommand + commandPrinter(subcommand, level+1)
	}
	return finalCommand
}

func GenerateCLIDocs() string {
	return commandPrinter(rootCmd, 0)
}
