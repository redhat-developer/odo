package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/redhat-developer/odo/pkg/odo/cli/logs"
	"github.com/redhat-developer/odo/pkg/odo/cli/run"
	"github.com/redhat-developer/odo/pkg/odo/commonflags"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"

	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/cli/add"
	"github.com/redhat-developer/odo/pkg/odo/cli/alizer"
	"github.com/redhat-developer/odo/pkg/odo/cli/build_images"
	"github.com/redhat-developer/odo/pkg/odo/cli/completion"
	"github.com/redhat-developer/odo/pkg/odo/cli/create"
	_delete "github.com/redhat-developer/odo/pkg/odo/cli/delete"
	"github.com/redhat-developer/odo/pkg/odo/cli/deploy"
	"github.com/redhat-developer/odo/pkg/odo/cli/describe"
	"github.com/redhat-developer/odo/pkg/odo/cli/dev"
	_init "github.com/redhat-developer/odo/pkg/odo/cli/init"
	"github.com/redhat-developer/odo/pkg/odo/cli/list"
	"github.com/redhat-developer/odo/pkg/odo/cli/login"
	"github.com/redhat-developer/odo/pkg/odo/cli/logout"
	"github.com/redhat-developer/odo/pkg/odo/cli/plugins"
	"github.com/redhat-developer/odo/pkg/odo/cli/preference"
	"github.com/redhat-developer/odo/pkg/odo/cli/registry"
	"github.com/redhat-developer/odo/pkg/odo/cli/remove"
	"github.com/redhat-developer/odo/pkg/odo/cli/set"
	"github.com/redhat-developer/odo/pkg/odo/cli/telemetry"
	"github.com/redhat-developer/odo/pkg/odo/cli/version"
	"github.com/redhat-developer/odo/pkg/odo/util"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

// OdoRecommendedName is the recommended odo command name
const OdoRecommendedName = "odo"

var (
	// We do not use ktemplates.Normalize here as it messed up the newlines..
	odoLong = `  __
 /  \__     odo is a CLI tool for fast iterative application development
 \__/  \    deployed immediately to your kubernetes cluster.
 /  \__/    Find more information at https://odo.dev
 \__/`

	odoExample = ktemplates.Examples(`Initializing your component by taking your pick from multiple languages or frameworks:
  odo init

	After creating your initial component, start development with:
  odo dev

	Want to deploy after development? See it live with:
  odo deploy`)

	rootUsageTemplate = `Usage:{{if .Runnable}}
  {{if .HasAvailableFlags}}{{appendIfNotPresent .UseLine "[flags]"}}{{else}}{{.UseLine}}{{end}}{{end}}{{if .HasAvailableSubCommands}}
  {{ .CommandPath}} [command]{{end}}{{if gt .Aliases 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{ .Example }}{{end}}{{ if .HasAvailableSubCommands}}

Main Commands:{{range .Commands}}{{if eq .Annotations.command "main"}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}

Management Commands:{{range .Commands}}{{if eq .Annotations.command "management"}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}

OpenShift Commands:{{range .Commands}}{{if eq .Annotations.command "openshift"}}
  {{rpad .Name .NamePadding }} {{.Short}} {{end}}{{end}}

Utility Commands:{{range .Commands}}{{if eq .Annotations.command "utility" }}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{ if .HasAvailableLocalFlags}}

Flags:
{{CapitalizeFlagDescriptions .LocalFlags | trimRightSpace }}{{end}}{{ if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimRightSpace}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsHelpCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{ if .HasAvailableSubCommands }}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`

	rootDefaultHelp = fmt.Sprintf("%s\n\nUsage:\n%s\n\n%s", odoLong, odoExample, rootHelpMessage)

	rootHelpMessage = "To see a full list of commands, run 'odo --help'"
)

const pluginPrefix = "odo"

// NewCmdOdo creates a new root command for odo
func NewCmdOdo(ctx context.Context, name, fullName string, testClientset clientset.Clientset) *cobra.Command {
	rootCmd := odoRootCmd(ctx, name, fullName, testClientset)

	if len(os.Args) > 1 {
		cmdPathPieces := os.Args[1:]
		// only look for suitable extension executables if
		// the specified command does not already exist
		cmd, _, err := rootCmd.Find(cmdPathPieces)
		if err == nil && cmd != rootCmd {
			return rootCmd
		}
		handleErr := plugins.HandleCommand(plugins.NewExecHandler(pluginPrefix), cmdPathPieces)
		if handleErr != nil {
			return rootCmd
		}
	}
	return rootCmd
}

func odoRootCmd(ctx context.Context, name, fullName string, testClientset clientset.Clientset) *cobra.Command {
	// rootCmd represents the base command when called without any subcommands
	rootCmd := &cobra.Command{
		Use:     name,
		Short:   "odo",
		Long:    odoLong,
		RunE:    ShowHelp,
		Example: odoExample,
	}
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.odo.yaml)")

	commonflags.AddOutputFlag()
	commonflags.AddPlatformFlag(ctx)
	commonflags.AddVariablesFlags()

	// Here we add the necessary "logging" flags.. However, we choose to hide some of these from the user
	// as they are not necessarily needed and more for advanced debugging
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	_ = pflag.CommandLine.Set("logtostderr", "true")
	_ = pflag.CommandLine.MarkHidden("alsologtostderr")
	_ = pflag.CommandLine.MarkHidden("log_backtrace_at")
	_ = pflag.CommandLine.MarkHidden("log_dir")
	_ = pflag.CommandLine.MarkHidden("logtostderr")
	_ = pflag.CommandLine.MarkHidden("stderrthreshold")
	_ = pflag.CommandLine.MarkHidden("add_dir_header")
	_ = pflag.CommandLine.MarkHidden("log_file")
	_ = pflag.CommandLine.MarkHidden("log_file_max_size")
	_ = pflag.CommandLine.MarkHidden("skip_headers")
	_ = pflag.CommandLine.MarkHidden("skip_log_headers")

	// Override the verbosity flag description
	verbosity := pflag.Lookup("v")
	verbosity.Usage += ". Level varies from 0 to 9 (default 0)."

	cobra.AddTemplateFunc("CapitalizeFlagDescriptions", capitalizeFlagDescriptions)
	rootCmd.SetUsageTemplate(rootUsageTemplate)

	// Create a custom help function that will exit when we enter an invalid command, for example:
	// odo foobar --help
	// which will exit with an error message: "unknown command 'foobar', type --help for a list of all commands"
	helpCmd := rootCmd.HelpFunc()
	rootCmd.SetHelpFunc(func(command *cobra.Command, args []string) {
		// Simple way of checking to see if the command has a parent (if it doesn't, it does not exist)
		if !command.HasParent() && len(args) > 0 {
			fmt.Fprintf(log.GetStderr(), "unknown command '%s', type --help for a list of all commands\n", args[0])
			os.Exit(1)
		}
		helpCmd(command, args)
	})

	rootCmdList := append([]*cobra.Command{},
		login.NewCmdLogin(login.RecommendedCommandName, util.GetFullName(fullName, login.RecommendedCommandName), testClientset),
		logout.NewCmdLogout(logout.RecommendedCommandName, util.GetFullName(fullName, logout.RecommendedCommandName), testClientset),
		version.NewCmdVersion(version.RecommendedCommandName, util.GetFullName(fullName, version.RecommendedCommandName), testClientset),
		preference.NewCmdPreference(ctx, preference.RecommendedCommandName, util.GetFullName(fullName, preference.RecommendedCommandName), testClientset),
		telemetry.NewCmdTelemetry(telemetry.RecommendedCommandName, testClientset),
		list.NewCmdList(ctx, list.RecommendedCommandName, util.GetFullName(fullName, list.RecommendedCommandName), testClientset),
		build_images.NewCmdBuildImages(build_images.RecommendedCommandName, util.GetFullName(fullName, build_images.RecommendedCommandName), testClientset),
		deploy.NewCmdDeploy(deploy.RecommendedCommandName, util.GetFullName(fullName, deploy.RecommendedCommandName), testClientset),
		_init.NewCmdInit(_init.RecommendedCommandName, util.GetFullName(fullName, _init.RecommendedCommandName), testClientset),
		_delete.NewCmdDelete(ctx, _delete.RecommendedCommandName, util.GetFullName(fullName, _delete.RecommendedCommandName), testClientset),
		add.NewCmdAdd(add.RecommendedCommandName, util.GetFullName(fullName, add.RecommendedCommandName), testClientset),
		remove.NewCmdRemove(remove.RecommendedCommandName, util.GetFullName(fullName, remove.RecommendedCommandName), testClientset),
		dev.NewCmdDev(ctx, dev.RecommendedCommandName, util.GetFullName(fullName, dev.RecommendedCommandName), testClientset),
		alizer.NewCmdAlizer(alizer.RecommendedCommandName, util.GetFullName(fullName, alizer.RecommendedCommandName), testClientset),
		describe.NewCmdDescribe(ctx, describe.RecommendedCommandName, util.GetFullName(fullName, describe.RecommendedCommandName), testClientset),
		registry.NewCmdRegistry(registry.RecommendedCommandName, util.GetFullName(fullName, registry.RecommendedCommandName), testClientset),
		create.NewCmdCreate(create.RecommendedCommandName, util.GetFullName(fullName, create.RecommendedCommandName), testClientset),
		set.NewCmdSet(set.RecommendedCommandName, util.GetFullName(fullName, set.RecommendedCommandName), testClientset),
		logs.NewCmdLogs(logs.RecommendedCommandName, util.GetFullName(fullName, logs.RecommendedCommandName), testClientset),
		completion.NewCmdCompletion(completion.RecommendedCommandName, util.GetFullName(fullName, completion.RecommendedCommandName)),
		run.NewCmdRun(run.RecommendedCommandName, util.GetFullName(fullName, run.RecommendedCommandName), testClientset),
	)

	// Add all subcommands to base commands
	rootCmd.AddCommand(rootCmdList...)

	visitCommands(rootCmd, reconfigureCmdWithSubcmd)

	return rootCmd
}

// capitalizeFlagDescriptions adds capitalizations
func capitalizeFlagDescriptions(f *pflag.FlagSet) string {
	f.VisitAll(func(f *pflag.Flag) {
		cap := []rune(f.Usage)
		cap[0] = unicode.ToUpper(cap[0])
		f.Usage = string(cap)
	})
	return f.FlagUsages()
}

// visitCommands visits each command within Cobra.
// Adapted from: https://github.com/cppforlife/knctl/blob/612840d3c9729b1c57b20ca0450acab0d6eceeeb/pkg/knctl/cobrautil/misc.go#L23
func visitCommands(cmd *cobra.Command, f func(*cobra.Command)) {
	f(cmd)
	for _, child := range cmd.Commands() {
		visitCommands(child, f)
	}
}

// reconfigureCmdWithSubcmd reconfigures each root command with a list of all subcommands and lists them
// beside the help output
// Adapted from: https://github.com/cppforlife/knctl/blob/612840d3c9729b1c57b20ca0450acab0d6eceeeb/pkg/knctl/cmd/knctl.go#L224
func reconfigureCmdWithSubcmd(cmd *cobra.Command) {
	if len(cmd.Commands()) == 0 {
		return
	}

	if cmd.Args == nil {
		cmd.Args = cobra.ArbitraryArgs
	}
	if cmd.RunE == nil && cmd.Run == nil {
		cmd.RunE = ShowSubcommands
	}

	var strs []string
	for _, subcmd := range cmd.Commands() {
		if !subcmd.Hidden {
			strs = append(strs, strings.Split(subcmd.Use, " ")[0])
		}
	}

	cmd.Short += " (" + strings.Join(strs, ", ") + ")"
}

// ShowSubcommands shows all available subcommands.
// Adapted from: https://github.com/cppforlife/knctl/blob/612840d3c9729b1c57b20ca0450acab0d6eceeeb/pkg/knctl/cmd/knctl.go#L224
func ShowSubcommands(cmd *cobra.Command, args []string) error {
	var strs []string
	for _, subcmd := range cmd.Commands() {
		if !subcmd.Hidden {
			strs = append(strs, subcmd.Name())
		}
	}

	if log.IsJSON() {
		cmd.SilenceUsage = true
		cmd.SilenceErrors = true
	}
	//revive:disable:error-strings This is a top-level error message displayed as is to the end user
	return fmt.Errorf("Subcommand not found, use one of the available commands: %s", strings.Join(strs, ", "))
	//revive:enable:error-strings
}

// ShowHelp will show the help correctly (and whether or not the command is invalid...)
// Taken from: https://github.com/cppforlife/knctl/blob/612840d3c9729b1c57b20ca0450acab0d6eceeeb/pkg/knctl/cmd/knctl.go#L71
func ShowHelp(cmd *cobra.Command, args []string) error {

	if len(args) == 0 {

		// We will show a custom help when typing JUST `odo`, directing the user to use `odo --help` for a full help.
		// Thus we will set cmd.SilenceUsage and cmd.SilenceErrors both to true so we do not output the usage or error out.
		cmd.SilenceUsage = true
		cmd.SilenceErrors = true

		// Print out the default "help" usage
		fmt.Println(rootDefaultHelp)
		return nil
	}

	//revive:disable:error-strings This is a top-level error message displayed as is to the end user
	if log.IsJSON() {
		cmd.SilenceUsage = true
		cmd.SilenceErrors = true
		return errors.New("Invalid command - see available commands/subcommands by running `odo`")
	}
	return errors.New("Invalid command - see available commands/subcommands above")
	//revive:enable:error-strings
}
