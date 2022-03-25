package utils

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"

	dfutil "github.com/devfile/library/pkg/util"

	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

const (
	terminalCommandName = "terminal"

	ps1 = `
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
	bashPS1Output = ps1 + `
PS1='$(__odo_ps1)'$PS1
`

	// Zsh output
	zshPS1Output = ps1 + `
setopt prompt_subst
PROMPT='$(__odo_ps1)'$PROMPT
`
)

var (
	terminalExample = ktemplates.Examples(`  # Bash terminal PS1 support
  source <(%[1]s bash)

  # Zsh terminal PS1 support
  source <(%[1]s zsh)
`)
	terminalLongDesc = ktemplates.LongDesc(`Add odo terminal support to your development environment.

This will append your PS1 environment variable with odo component and application information.`)
	supportedShells = map[string]string{"bash": bashPS1Output, "zsh": zshPS1Output}
)

// TerminalOptions encapsulates the options for the command
type TerminalOptions struct {
	shellType string
}

// NewTerminalOptions creates a new TerminalOptions instance
func NewTerminalOptions() *TerminalOptions {
	return &TerminalOptions{}
}

func (o *TerminalOptions) SetClientset(clientset *clientset.Clientset) {
}

// Complete completes TerminalOptions after they've been created
func (o *TerminalOptions) Complete(cmdline cmdline.Cmdline, args []string) (err error) {
	o.shellType = args[0]
	return
}

// Validate validates the TerminalOptions based on completed values
func (o *TerminalOptions) Validate() (err error) {
	if _, ok := supportedShells[o.shellType]; !ok {
		return fmt.Errorf("unknown shell type %s, supported shells: %v", o.shellType, getSupportedShells())
	}
	return
}

// Run contains the logic for the command
func (o *TerminalOptions) Run(ctx context.Context) (err error) {
	// shell type is already validated so retrieval will work
	_, err = os.Stdout.Write([]byte(supportedShells[o.shellType]))
	return
}

// NewCmdTerminal implements the utils terminal odo command
func NewCmdTerminal(name, fullName string) *cobra.Command {
	o := NewTerminalOptions()
	terminalCmd := &cobra.Command{
		Use:     name,
		Short:   "Add odo terminal support to your development environment",
		Long:    terminalLongDesc,
		Example: fmt.Sprintf(terminalExample, fullName),
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}
	return terminalCmd
}

func getSupportedShells() []string {
	return dfutil.GetSortedKeys(supportedShells)
}
