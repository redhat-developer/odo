package cli

import (
	"flag"
	"fmt"
	"github.com/redhat-developer/odo/pkg/odo/cli/dev"
	"os"
	"strings"
	"unicode"

	"github.com/redhat-developer/odo/pkg/odo/cli/build_images"
	"github.com/redhat-developer/odo/pkg/odo/cli/catalog"
	"github.com/redhat-developer/odo/pkg/odo/cli/component"
	_delete "github.com/redhat-developer/odo/pkg/odo/cli/delete"
	"github.com/redhat-developer/odo/pkg/odo/cli/deploy"
	_init "github.com/redhat-developer/odo/pkg/odo/cli/init"
	"github.com/redhat-developer/odo/pkg/odo/cli/login"
	"github.com/redhat-developer/odo/pkg/odo/cli/logout"
	"github.com/redhat-developer/odo/pkg/odo/cli/plugins"
	"github.com/redhat-developer/odo/pkg/odo/cli/preference"
	"github.com/redhat-developer/odo/pkg/odo/cli/project"
	"github.com/redhat-developer/odo/pkg/odo/cli/telemetry"
	"github.com/redhat-developer/odo/pkg/odo/cli/url"
	"github.com/redhat-developer/odo/pkg/odo/cli/utils"
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

	odoExample = ktemplates.Examples(`  # Creating and deploying a Node.js project
  git clone https://github.com/odo-devfiles/nodejs-ex && cd nodejs-ex
  %[1]s create nodejs
  %[1]s push

  # Accessing your Node.js component
  %[1]s url create`)

	rootUsageTemplate = `Usage:{{if .Runnable}}
  {{if .HasAvailableFlags}}{{appendIfNotPresent .UseLine "[flags]"}}{{else}}{{.UseLine}}{{end}}{{end}}{{if .HasAvailableSubCommands}}
  {{ .CommandPath}} [command]{{end}}{{if gt .Aliases 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{ .Example }}{{end}}{{ if .HasAvailableSubCommands}}

Commands:{{range .Commands}}{{if eq .Annotations.command "main"}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}

Utility Commands:{{range .Commands}}{{if or (eq .Annotations.command "utility") (eq .Name "help") }}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{ if .HasAvailableLocalFlags}}

Component Shortcuts:{{range .Commands}}{{if eq .Annotations.command "component"}}
  {{rpad .Name .NamePadding }} {{.Short}} {{end}}{{end}}{{end}}

Flags:
{{CapitalizeFlagDescriptions .LocalFlags | trimRightSpace }}{{end}}{{ if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimRightSpace}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsHelpCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{ if .HasAvailableSubCommands }}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`

	rootDefaultHelp = odoLong + `

Get started by creating a new application:

 git clone https://github.com/odo-devfiles/nodejs-ex && cd nodejs-ex
 odo create nodejs
 odo push

Your Node.JS application has now been deployed to Kubernetes. odo has pushed the source code, built the application and deployed it.

You can now edit your code in real time and watch as odo automatically deploys your application.

 odo watch

To access your application, create a URL:

 odo url create myurl
 odo push

More information such as logs or what components you've deployed can be accessed with these commands:

 odo describe
 odo list
 odo log

To see a full list of commands, run 'odo --help'`
)

const pluginPrefix = "odo"

// NewCmdOdo creates a new root command for odo
func NewCmdOdo(name, fullName string) *cobra.Command {
	rootCmd := odoRootCmd(name, fullName)

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

func odoRootCmd(name, fullName string) *cobra.Command {
	// rootCmd represents the base command when called without any subcommands
	rootCmd := &cobra.Command{
		Use:     name,
		Short:   "odo",
		Long:    odoLong,
		RunE:    ShowHelp,
		Example: fmt.Sprintf(odoExample, fullName),
	}
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.odo.yaml)")

	// Add the machine readable output flag to all commands
	// We use "flag" in order to make this accessible throughtout ALL of odo, rather than the
	// above traditional "persistentflags" usage that does not make it a pointer within the 'pflag'
	// package
	flag.CommandLine.String("o", "", "Specify output format, supported format: json")

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

	// We will mark the command as hidden and then re-enable if the command
	// supports json output
	_ = pflag.CommandLine.MarkHidden("o")

	// Override the verbosity flag description
	verbosity := pflag.Lookup("v")
	verbosity.Usage += ". Level varies from 0 to 9 (default 0)."

	rootCmd.SetUsageTemplate(rootUsageTemplate)
	cobra.AddTemplateFunc("CapitalizeFlagDescriptions", capitalizeFlagDescriptions)
	cobra.AddTemplateFunc("ModifyAdditionalFlags", modifyAdditionalFlags)

	rootCmdList := append([]*cobra.Command{},
		catalog.NewCmdCatalog(catalog.RecommendedCommandName, util.GetFullName(fullName, catalog.RecommendedCommandName)),
		component.NewCmdComponent(component.RecommendedCommandName, util.GetFullName(fullName, component.RecommendedCommandName)),
		component.NewCmdCreate(component.CreateRecommendedCommandName, util.GetFullName(fullName, component.CreateRecommendedCommandName)),
		component.NewCmdList(component.ListRecommendedCommandName, util.GetFullName(fullName, component.ListRecommendedCommandName)),
		component.NewCmdPush(component.PushRecommendedCommandName, util.GetFullName(fullName, component.PushRecommendedCommandName)),
		login.NewCmdLogin(login.RecommendedCommandName, util.GetFullName(fullName, login.RecommendedCommandName)),
		logout.NewCmdLogout(logout.RecommendedCommandName, util.GetFullName(fullName, logout.RecommendedCommandName)),
		project.NewCmdProject(project.RecommendedCommandName, util.GetFullName(fullName, project.RecommendedCommandName)),
		url.NewCmdURL(url.RecommendedCommandName, util.GetFullName(fullName, url.RecommendedCommandName)),
		utils.NewCmdUtils(utils.RecommendedCommandName, util.GetFullName(fullName, utils.RecommendedCommandName)),
		version.NewCmdVersion(version.RecommendedCommandName, util.GetFullName(fullName, version.RecommendedCommandName)),
		preference.NewCmdPreference(preference.RecommendedCommandName, util.GetFullName(fullName, preference.RecommendedCommandName)),
		telemetry.NewCmdTelemetry(telemetry.RecommendedCommandName),
		build_images.NewCmdBuildImages(build_images.RecommendedCommandName, util.GetFullName(fullName, build_images.RecommendedCommandName)),
		deploy.NewCmdDeploy(deploy.RecommendedCommandName, util.GetFullName(fullName, deploy.RecommendedCommandName)),
		_init.NewCmdInit(_init.RecommendedCommandName, util.GetFullName(fullName, _init.RecommendedCommandName)),
		_delete.NewCmdDelete(_delete.RecommendedCommandName, util.GetFullName(fullName, _delete.RecommendedCommandName)),
		dev.NewCmdDev(dev.RecommendedCommandName, util.GetFullName(fullName, dev.RecommendedCommandName)),
	)

	// Add all subcommands to base commands
	rootCmd.AddCommand(rootCmdList...)

	visitCommands(rootCmd, reconfigureCmdWithSubcmd)

	return rootCmd
}

// modifyAdditionalFlags modifies the flags and updates the descriptions
// as well as changes whether or not machine readable output
// has been passed in..
//
// Return the flag usages for the help output
func modifyAdditionalFlags(cmd *cobra.Command) string {

	// Hide the machine readable output if the command
	// does not have the annotation.
	machineOutput := cmd.Annotations["machineoutput"]
	f := cmd.InheritedFlags()

	f.VisitAll(func(f *pflag.Flag) {
		// Remove json flag if machineoutput has not been passed in
		if f.Name == "o" && machineOutput == "json" {
			f.Hidden = false
		}
	})

	return capitalizeFlagDescriptions(f)
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
	if cmd.RunE == nil {
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
	return fmt.Errorf("Subcommand not found, use one of the available commands: %s", strings.Join(strs, ", "))
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

	return fmt.Errorf("Invalid command - see available commands/subcommands above")
}
