package main

import (
	"bytes"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"github.com/redhat-developer/odo/pkg/odo/cli"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"os"
)

/*

This "script" generates markdown that can be interpreted by the Slate (https://github.com/lord/slate) format.
Use this to script to generate the documentation needed.

*/

// Uses portions of the help / cmd outputter in cobra13 as part of a CLI reference guide and outputs each command
func referenceCommandFormatter(command *cobra.Command) string {
	return fmt.Sprintf(`## %s

%s

> Example using %s

%s


%s

`,
		command.Name(),
		"`"+command.Use+"`",
		command.Name(),
		"```sh\n"+command.Example+"\n```",
		command.Long)
}

// This prints out the CLI reference
func referencePrinter(command *cobra.Command, level int) string {

	// List each command
	var commandListTable [][]string
	for _, subcommand := range command.Commands() {
		name := fmt.Sprintf("[%s](#%s)", subcommand.Name(), subcommand.Name())
		commandListTable = append(commandListTable, []string{name, subcommand.Short})
	}

	// Create a "table" for listing each command
	tableOutput := new(bytes.Buffer)
	table := tablewriter.NewWriter(tableOutput)
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")
	table.SetHeader([]string{"Name", "Description"})
	table.SetColWidth(10000)
	table.AppendBulk(commandListTable)
	table.Render() // Send output to writer

	// Create documentation for each command
	var commandOutput string
	for _, subcommand := range command.Commands() {
		commandOutput = commandOutput + referenceCommandFormatter(subcommand)
	}

	// The main markdown "template" for everything
	return fmt.Sprintf(`# Overview of the Odo (OpenShift Do) CLI Structure

> Example application

%s 

%s

# Syntax

#### List of Commands

%s

#### CLI Structure

%s

%s
`,
		"```sh\n"+command.Example+"\n```",
		command.Long,
		tableOutput.String(),
		"```sh\n"+fmt.Sprint(commandPrinter(command, 0))+"\n```",
		commandOutput)
}

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

// Generates and returns a markdown-formatted CLI reference page for Odo
func main() {
	var clidoc = &cobra.Command{
		Use:   "cli-doc",
		Short: "Generate CLI reference for Odo",

		Example: `  # Generate a markdown-formatted CLI reference page for Odo
  cli-doc reference > docs/cli-reference.md

  # Generate the CLI structure
  cli-doc structure`,
		Args:      cobra.OnlyValidArgs,
		ValidArgs: []string{"help", "reference", "structure"},

		Run: func(command *cobra.Command, args []string) {
			if len(args) == 0 {
				fmt.Print(command.Usage())
			} else {
				switch args[0] {
				case "reference":
					fmt.Print(referencePrinter(cli.NewCmdOdo(cli.OdoRecommendedName, cli.OdoRecommendedName), 0))
				case "structure":
					fmt.Print(commandPrinter(cli.NewCmdOdo(cli.OdoRecommendedName, cli.OdoRecommendedName), 0))
				default:
					fmt.Print(command.Usage())
				}
			}
		},
	}

	err := clidoc.Execute()
	if err != nil {
		fmt.Println(errors.Cause(err))
		os.Exit(1)
	}
}
