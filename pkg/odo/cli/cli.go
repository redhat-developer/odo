package cli

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/openshift/odo/pkg/odo/cli/application"
	"github.com/openshift/odo/pkg/odo/cli/build_images"
	"github.com/openshift/odo/pkg/odo/cli/catalog"
	"github.com/openshift/odo/pkg/odo/cli/component"
	"github.com/openshift/odo/pkg/odo/cli/config"
	"github.com/openshift/odo/pkg/odo/cli/debug"
	"github.com/openshift/odo/pkg/odo/cli/deploy"
	"github.com/openshift/odo/pkg/odo/cli/env"
	"github.com/openshift/odo/pkg/odo/cli/login"
	"github.com/openshift/odo/pkg/odo/cli/logout"
	"github.com/openshift/odo/pkg/odo/cli/plugins"
	"github.com/openshift/odo/pkg/odo/cli/preference"
	"github.com/openshift/odo/pkg/odo/cli/project"
	"github.com/openshift/odo/pkg/odo/cli/registry"
	"github.com/openshift/odo/pkg/odo/cli/service"
	"github.com/openshift/odo/pkg/odo/cli/storage"
	"github.com/openshift/odo/pkg/odo/cli/telemetry"
	"github.com/openshift/odo/pkg/odo/cli/url"
	"github.com/openshift/odo/pkg/odo/cli/utils"
	"github.com/openshift/odo/pkg/odo/cli/version"
	"github.com/openshift/odo/pkg/odo/util"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

// OdoRecommendedName is the recommended odo command name
const OdoRecommendedName = "odo"

var (
	// We do not use ktemplates.Normalize here as it messed up the newlines..
	odoLong = `odo is a CLI tool for running OpenShift applications in a fast and automated manner.
Reducing the complexity of deployment, odo adds iterative development without the worry of deploying your source code.

Find more information at https://github.com/openshift/odo`

	odoExample = ktemplates.Examples(`  # Creating and deploying a Node.js project
  git clone https://github.com/openshift/nodejs-ex && cd nodejs-ex
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

 git clone https://github.com/openshift/nodejs-ex && cd nodejs-ex
 odo create nodejs
 odo push

Your nodejs application has now been deployed. odo has pushed the source code, built the application and deployed it on OpenShift. 
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
	cobra.AddTemplateFunc("CapitalizeFlagDescriptions", util.CapitalizeFlagDescriptions)
	cobra.AddTemplateFunc("ModifyAdditionalFlags", util.ModifyAdditionalFlags)

	rootCmdList := append([]*cobra.Command{}, application.NewCmdApplication(application.RecommendedCommandName, util.GetFullName(fullName, application.RecommendedCommandName)),
		catalog.NewCmdCatalog(catalog.RecommendedCommandName, util.GetFullName(fullName, catalog.RecommendedCommandName)),
		component.NewCmdComponent(component.RecommendedCommandName, util.GetFullName(fullName, component.RecommendedCommandName)),
		component.NewCmdCreate(component.CreateRecommendedCommandName, util.GetFullName(fullName, component.CreateRecommendedCommandName)),
		component.NewCmdDelete(component.DeleteRecommendedCommandName, util.GetFullName(fullName, component.DeleteRecommendedCommandName)),
		component.NewCmdDescribe(component.DescribeRecommendedCommandName, util.GetFullName(fullName, component.DescribeRecommendedCommandName)),
		component.NewCmdLink(component.LinkRecommendedCommandName, util.GetFullName(fullName, component.LinkRecommendedCommandName)),
		component.NewCmdUnlink(component.UnlinkRecommendedCommandName, util.GetFullName(fullName, component.UnlinkRecommendedCommandName)),
		component.NewCmdList(component.ListRecommendedCommandName, util.GetFullName(fullName, component.ListRecommendedCommandName)),
		component.NewCmdLog(component.LogRecommendedCommandName, util.GetFullName(fullName, component.LogRecommendedCommandName)),
		component.NewCmdPush(component.PushRecommendedCommandName, util.GetFullName(fullName, component.PushRecommendedCommandName)),
		component.NewCmdWatch(component.WatchRecommendedCommandName, util.GetFullName(fullName, component.WatchRecommendedCommandName)),
		component.NewCmdStatus(component.StatusRecommendedCommandName, util.GetFullName(fullName, component.StatusRecommendedCommandName)),
		component.NewCmdExec(component.ExecRecommendedCommandName, util.GetFullName(fullName, component.ExecRecommendedCommandName)),
		login.NewCmdLogin(login.RecommendedCommandName, util.GetFullName(fullName, login.RecommendedCommandName)),
		logout.NewCmdLogout(logout.RecommendedCommandName, util.GetFullName(fullName, logout.RecommendedCommandName)),
		project.NewCmdProject(project.RecommendedCommandName, util.GetFullName(fullName, project.RecommendedCommandName)),
		service.NewCmdService(service.RecommendedCommandName, util.GetFullName(fullName, service.RecommendedCommandName)),
		storage.NewCmdStorage(storage.RecommendedCommandName, util.GetFullName(fullName, storage.RecommendedCommandName)),
		url.NewCmdURL(url.RecommendedCommandName, util.GetFullName(fullName, url.RecommendedCommandName)),
		utils.NewCmdUtils(utils.RecommendedCommandName, util.GetFullName(fullName, utils.RecommendedCommandName)),
		version.NewCmdVersion(version.RecommendedCommandName, util.GetFullName(fullName, version.RecommendedCommandName)),
		config.NewCmdConfiguration(config.RecommendedCommandName, util.GetFullName(fullName, config.RecommendedCommandName)),
		preference.NewCmdPreference(preference.RecommendedCommandName, util.GetFullName(fullName, preference.RecommendedCommandName)),
		debug.NewCmdDebug(debug.RecommendedCommandName, util.GetFullName(fullName, debug.RecommendedCommandName)),
		registry.NewCmdRegistry(registry.RecommendedCommandName, util.GetFullName(fullName, registry.RecommendedCommandName)),
		component.NewCmdTest(component.TestRecommendedCommandName, util.GetFullName(fullName, component.TestRecommendedCommandName)),
		env.NewCmdEnv(env.RecommendedCommandName, util.GetFullName(fullName, env.RecommendedCommandName)),
		telemetry.NewCmdTelemetry(telemetry.RecommendedCommandName),
		build_images.NewCmdBuildImages(build_images.RecommendedCommandName, util.GetFullName(fullName, build_images.RecommendedCommandName)),
		deploy.NewCmdDeploy(deploy.RecommendedCommandName, util.GetFullName(fullName, deploy.RecommendedCommandName)),
	)

	// Add all subcommands to base commands
	rootCmd.AddCommand(rootCmdList...)

	util.VisitCommands(rootCmd, reconfigureCmdWithSubcmd)

	return rootCmd
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
