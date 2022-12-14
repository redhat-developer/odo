package completion

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/redhat-developer/odo/pkg/odo/util"

	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

const (
	RecommendedCommandName = "completion"
)

var (
	completionExample = ktemplates.Examples(`   # BASH

	## Load into your current shell environment
  source <(%[1]s bash)

	## Load persistently

	### Save the completion to a file
	%[1]s bash > ~/.odo/completion.bash.inc

	### Load the completion from within your $HOME/.bash_profile
	source ~/.odo/completion.bash.inc

  # ZSH

	## Load into your current shell environment
  source <(%[1]s zsh)

	## Load persistently
	%[1]s zsh > "${fpath[1]}/_odo"

	# FISH

	## Load into your current shell environment
	source <(%[1]s fish)

	## Load persistently
	%[1]s fish > ~/.config/fish/completions/odo.fish

	# POWERSHELL

	## Load into your current shell environment
	%[1]s powershell | Out-String | Invoke-Expression

	## Load persistently
	%[1]s powershell >> $PROFILE
`)
	completionLongDesc = ktemplates.LongDesc(`Add odo completion support to your development environment.
This will append your PS1 environment variable with odo component and application information.`)
)

// NewCmdCompletion implements the utils completion odo command
func NewCmdCompletion(name, fullName string) *cobra.Command {
	completionCmd := &cobra.Command{
		Use:                   name,
		Short:                 "Add odo completion support to your development environment",
		Long:                  completionLongDesc,
		Example:               fmt.Sprintf(completionExample, fullName),
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.ExactValidArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Below we ignore the error returns in order to pass golint validation
			// We will handle the error in the main function / output when the user inputs `odo completion`.
			switch args[0] {
			case "bash":
				_ = cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				// Due to https://github.com/spf13/cobra/issues/1529 we cannot load zsh
				// via using source, so we need to add compdef to the beginning of the output so we can easily do:
				// source <(odo completion zsh)
				zsh := "#compdef odo\ncompdef _odo odo\n"
				out := os.Stdout
				_, _ = out.Write([]byte(zsh))
				_ = cmd.Root().GenZshCompletion(out)
			case "fish":
				_ = cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				_ = cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
			}
		},
	}

	completionCmd.SetUsageTemplate(util.CmdUsageTemplate)
	util.SetCommandGroup(completionCmd, util.UtilityGroup)
	return completionCmd
}
