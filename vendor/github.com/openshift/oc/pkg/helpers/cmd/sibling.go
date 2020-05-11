package cmd

import (
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// SiblingOrNiblingCommand returns the closest command with the specified name
// that is also a sub-command of a direct ancestor of the specified command.
// If the command is not found, just return the name of the command.
func SiblingOrNiblingCommand(cmd *cobra.Command, name string) []string {
	parentCmd := cmd.Parent()
	for parentCmd != nil {
		for _, command := range parentCmd.Commands() {
			if command.Name() == name {
				return append([]string{os.Args[0]}, strings.Split(command.CommandPath(), " ")[1:]...)
			}
		}
		parentCmd = parentCmd.Parent()
	}
	return []string{name}
}
