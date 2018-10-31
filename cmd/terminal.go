package cmd

import (
	"fmt"
	"github.com/redhat-developer/odo/pkg/odo/util"
	"io"
	"os"

	"github.com/spf13/cobra"
)

// terminalCmd represents the terminal command
var terminalCmd = &cobra.Command{
	Use:   "terminal",
	Short: "Add Odo terminal support to your development environment",
	Long: `Add Odo terminal support to your development environment. 

This will append your PS1 environment variable with Odo component and application information.`,
	Example: `  # Bash terminal PS1 support
  source <(odo utils terminal bash)

  # Zsh terminal PS1 support
  source <(odo utils terminal zsh)
`,
	RunE: func(cmd *cobra.Command, args []string) error {

		err := TerminalGenerate(os.Stdout, cmd, args)
		util.CheckError(err, "")

		return nil
	},
}

// Generates the PS1 output for Odo terminal support (appends to current PS1 environment variable)
func TerminalGenerate(out io.Writer, cmd *cobra.Command, args []string) error {
	// Check the passed in arguments
	if len(args) == 0 {
		return fmt.Errorf("Shell not specified. ex. odo completion [bash|zsh]")
	}
	if len(args) > 1 {
		return fmt.Errorf("Too many arguments. Expected only the shell type. ex. odo completion [bash|zsh]")
	}
	shell := args[0]

	// sh function for retrieving component information
	var PS1 = `
__odo_ps1() {

		# Get application context
    APP=$(odo application get -q --skip-connection-check)

		if [ "$APP" = "" ]; then
		APP="<no application>"
		fi

    # Get current context
    COMPONENT=$(odo component get -q --skip-connection-check)

		if [ "$COMPONENT" = "" ]; then
		COMPONENT="<no component>"
    fi

    if [ -n "$COMPONENT" ] || [ -n "$APP" ]; then
        echo "[${APP}/${COMPONENT}]"
    fi
}
`

	// Bash output
	var bashPS1Output = PS1 + `
PS1='$(__odo_ps1)'$PS1
`

	// Zsh output
	var zshPS1Output = PS1 + `
setopt prompt_subst
PROMPT='$(__odo_ps1)'$PROMPT
`

	if shell == "bash" {
		out.Write([]byte(bashPS1Output))
	} else if shell == "zsh" {
		out.Write([]byte(zshPS1Output))
	} else {
		return fmt.Errorf("not a compatible shell, bash and zsh are only supported")
	}

	return nil
}
