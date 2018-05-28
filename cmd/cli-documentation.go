package cmd

import (
	"bytes"
	"fmt"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
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
		"```sh\n"+GenerateCLIStructure()+"\n```",
		commandOutput)
}

// Generates and returns a markdown-formatted CLI reference page for Odo
func GenerateCLIReference() string {
	return referencePrinter(rootCmd, 0)
}
