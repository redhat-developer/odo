package cli

import (
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/kubectl/pkg/cmd/diff"
	"k8s.io/kubectl/pkg/cmd/kustomize"
	"k8s.io/kubectl/pkg/cmd/plugin"
	kcmdutil "k8s.io/kubectl/pkg/cmd/util"
	ktemplates "k8s.io/kubectl/pkg/util/templates"
	kubecmd "k8s.io/kubernetes/pkg/kubectl/cmd"

	"github.com/openshift/oc/pkg/cli/admin"
	"github.com/openshift/oc/pkg/cli/admin/buildchain"
	"github.com/openshift/oc/pkg/cli/admin/groups/sync"
	"github.com/openshift/oc/pkg/cli/cancelbuild"
	"github.com/openshift/oc/pkg/cli/debug"
	"github.com/openshift/oc/pkg/cli/deployer"
	"github.com/openshift/oc/pkg/cli/experimental/dockergc"
	"github.com/openshift/oc/pkg/cli/expose"
	"github.com/openshift/oc/pkg/cli/extract"
	"github.com/openshift/oc/pkg/cli/idle"
	"github.com/openshift/oc/pkg/cli/image"
	"github.com/openshift/oc/pkg/cli/importimage"
	"github.com/openshift/oc/pkg/cli/kubectlwrappers"
	"github.com/openshift/oc/pkg/cli/login"
	"github.com/openshift/oc/pkg/cli/logout"
	"github.com/openshift/oc/pkg/cli/logs"
	"github.com/openshift/oc/pkg/cli/newapp"
	"github.com/openshift/oc/pkg/cli/newbuild"
	"github.com/openshift/oc/pkg/cli/observe"
	"github.com/openshift/oc/pkg/cli/options"
	"github.com/openshift/oc/pkg/cli/policy"
	"github.com/openshift/oc/pkg/cli/process"
	"github.com/openshift/oc/pkg/cli/project"
	"github.com/openshift/oc/pkg/cli/projects"
	"github.com/openshift/oc/pkg/cli/recycle"
	"github.com/openshift/oc/pkg/cli/registry"
	"github.com/openshift/oc/pkg/cli/requestproject"
	"github.com/openshift/oc/pkg/cli/rollback"
	"github.com/openshift/oc/pkg/cli/rollout"
	"github.com/openshift/oc/pkg/cli/rsh"
	"github.com/openshift/oc/pkg/cli/rsync"
	"github.com/openshift/oc/pkg/cli/secrets"
	"github.com/openshift/oc/pkg/cli/serviceaccounts"
	"github.com/openshift/oc/pkg/cli/set"
	"github.com/openshift/oc/pkg/cli/startbuild"
	"github.com/openshift/oc/pkg/cli/status"
	"github.com/openshift/oc/pkg/cli/tag"
	"github.com/openshift/oc/pkg/cli/version"
	"github.com/openshift/oc/pkg/cli/whoami"
	cmdutil "github.com/openshift/oc/pkg/helpers/cmd"
	"github.com/openshift/oc/pkg/helpers/term"
)

const productName = `OpenShift`

var (
	cliLong = heredoc.Doc(`
    ` + productName + ` Client

    This client helps you develop, build, deploy, and run your applications on any
    OpenShift or Kubernetes cluster. It also includes the administrative
    commands for managing a cluster under the 'adm' subcommand.`)

	cliExplain = heredoc.Doc(`
    To familiarize yourself with OpenShift, login to your cluster and try creating a sample application:

        %[1]s login mycluster.mycompany.com
        %[1]s new-project my-example
        %[1]s new-app django-psql-example
        %[1]s logs -f bc/django-psql-example

    To see what has been created, run:

        %[1]s status

    and get a command shell inside one of the created containers with:

        %[1]s rsh dc/postgresql

    To see the list of available toolchains for building applications, run:

        %[1]s new-app -L

    Since OpenShift runs on top of Kubernetes, your favorite kubectl commands are also present in oc,
    allowing you to quickly switch between development and debugging. You can also run kubectl directly
    against any OpenShift cluster using the kubeconfig file created by 'oc login'.

    For more on OpenShift, see the documentation at https://docs.openshift.com.

    To see the full list of commands supported, run '%[1]s --help'.`)
)

func NewDefaultOcCommand(name string, in io.Reader, out, errout io.Writer) *cobra.Command {
	cmd := NewOcCommand(name, in, out, errout)

	if len(os.Args) <= 1 {
		return cmd
	}

	cmdPathPieces := os.Args[1:]
	pluginHandler := kubecmd.NewDefaultPluginHandler(plugin.ValidPluginFilenamePrefixes)

	// only look for suitable extension executables if
	// the specified command does not already exist
	if _, _, err := cmd.Find(cmdPathPieces); err != nil {
		if err := kubecmd.HandlePluginCommand(pluginHandler, cmdPathPieces); err != nil {
			fmt.Fprintf(errout, "%v\n", err)
			os.Exit(1)
		}
	}

	return cmd
}

func NewOcCommand(name string, in io.Reader, out, errout io.Writer) *cobra.Command {
	// Main command
	cmds := &cobra.Command{
		Use:   name,
		Short: "Command line tools for managing applications",
		Long:  cliLong,
		Run: func(c *cobra.Command, args []string) {
			explainOut := term.NewResponsiveWriter(out)
			c.SetOutput(explainOut)
			kcmdutil.RequireNoArguments(c, args)
			fmt.Fprintf(explainOut, "%s\n\n%s\n", cliLong, fmt.Sprintf(cliExplain, name))
		},
		BashCompletionFunction: bashCompletionFunc,
	}

	kubeConfigFlags := genericclioptions.NewConfigFlags(true)
	kubeConfigFlags.AddFlags(cmds.PersistentFlags())
	matchVersionKubeConfigFlags := kcmdutil.NewMatchVersionFlags(kubeConfigFlags)
	matchVersionKubeConfigFlags.AddFlags(cmds.PersistentFlags())
	cmds.PersistentFlags().AddGoFlagSet(flag.CommandLine)
	f := kcmdutil.NewFactory(matchVersionKubeConfigFlags)

	ioStreams := genericclioptions.IOStreams{In: in, Out: out, ErrOut: errout}

	loginCmd := login.NewCmdLogin(name, f, ioStreams)
	secretcmds := secrets.NewCmdSecrets(name, f, ioStreams)

	groups := ktemplates.CommandGroups{
		{
			Message: "Basic Commands:",
			Commands: []*cobra.Command{
				loginCmd,
				requestproject.NewCmdRequestProject(name, f, ioStreams),
				newapp.NewCmdNewApplication(newapp.NewAppRecommendedCommandName, name, f, ioStreams),
				status.NewCmdStatus(status.StatusRecommendedName, name, name+" "+status.StatusRecommendedName, f, ioStreams),
				project.NewCmdProject(name, f, ioStreams),
				projects.NewCmdProjects(name, f, ioStreams),
				kubectlwrappers.NewCmdExplain(name, f, ioStreams),
			},
		},
		{
			Message: "Build and Deploy Commands:",
			Commands: []*cobra.Command{
				rollout.NewCmdRollout(name, f, ioStreams),
				rollback.NewCmdRollback(name, f, ioStreams),
				newbuild.NewCmdNewBuild(newbuild.NewBuildRecommendedCommandName, name, f, ioStreams),
				startbuild.NewCmdStartBuild(name, f, ioStreams),
				cancelbuild.NewCmdCancelBuild(cancelbuild.CancelBuildRecommendedCommandName, name, f, ioStreams),
				importimage.NewCmdImportImage(name, f, ioStreams),
				tag.NewCmdTag(name, f, ioStreams),
			},
		},
		{
			Message: "Application Management Commands:",
			Commands: []*cobra.Command{
				kubectlwrappers.NewCmdCreate(name, f, ioStreams),
				kubectlwrappers.NewCmdApply(name, f, ioStreams),
				kubectlwrappers.NewCmdGet(name, f, ioStreams),
				kubectlwrappers.NewCmdDescribe(name, f, ioStreams),
				kubectlwrappers.NewCmdEdit(name, f, ioStreams),
				set.NewCmdSet(name, f, ioStreams),
				kubectlwrappers.NewCmdLabel(name, f, ioStreams),
				kubectlwrappers.NewCmdAnnotate(name, f, ioStreams),
				expose.NewCmdExpose(name, f, ioStreams),
				kubectlwrappers.NewCmdDelete(name, f, ioStreams),
				kubectlwrappers.NewCmdScale(name, f, ioStreams),
				kubectlwrappers.NewCmdAutoscale(name, f, ioStreams),
				secretcmds,
				serviceaccounts.NewCmdServiceAccounts(serviceaccounts.ServiceAccountsRecommendedName, name+" "+serviceaccounts.ServiceAccountsRecommendedName, f, ioStreams),
			},
		},
		{
			Message: "Troubleshooting and Debugging Commands:",
			Commands: []*cobra.Command{
				logs.NewCmdLogs(name, f, ioStreams),
				rsh.NewCmdRsh(rsh.RshRecommendedName, name, f, ioStreams),
				rsync.NewCmdRsync(rsync.RsyncRecommendedName, name, f, ioStreams),
				kubectlwrappers.NewCmdPortForward(name, f, ioStreams),
				debug.NewCmdDebug(name, f, ioStreams),
				kubectlwrappers.NewCmdExec(name, f, ioStreams),
				kubectlwrappers.NewCmdProxy(name, f, ioStreams),
				kubectlwrappers.NewCmdAttach(name, f, ioStreams),
				kubectlwrappers.NewCmdRun(name, f, ioStreams),
				kubectlwrappers.NewCmdCp(name, f, ioStreams),
				kubectlwrappers.NewCmdWait(name, f, ioStreams),
			},
		},
		{
			Message: "Advanced Commands:",
			Commands: []*cobra.Command{
				admin.NewCommandAdmin("adm", name+" "+"adm", f, ioStreams),
				kubectlwrappers.NewCmdReplace(name, f, ioStreams),
				kubectlwrappers.NewCmdPatch(name, f, ioStreams),
				process.NewCmdProcess(name, f, ioStreams),
				extract.NewCmdExtract(name, f, ioStreams),
				observe.NewCmdObserve(name, f, ioStreams),
				policy.NewCmdPolicy(policy.PolicyRecommendedName, name+" "+policy.PolicyRecommendedName, f, ioStreams),
				kubectlwrappers.NewCmdAuth(name, f, ioStreams),
				kubectlwrappers.NewCmdConvert(name, f, ioStreams),
				image.NewCmdImage(name, f, ioStreams),
				registry.NewCmd(name, f, ioStreams),
				idle.NewCmdIdle(name, f, ioStreams),
				kubectlwrappers.NewCmdApiVersions(name, f, ioStreams),
				kubectlwrappers.NewCmdApiResources(name, f, ioStreams),
				kubectlwrappers.NewCmdClusterInfo(name, f, ioStreams),
				diff.NewCmdDiff(f, ioStreams),
				kustomize.NewCmdKustomize(ioStreams),
			},
		},
		{
			Message: "Settings Commands:",
			Commands: []*cobra.Command{
				logout.NewCmdLogout("logout", name+" logout", name+" login", f, ioStreams),
				kubectlwrappers.NewCmdConfig(name, "config", f, ioStreams),
				whoami.NewCmdWhoAmI(whoami.WhoAmIRecommendedCommandName, name+" "+whoami.WhoAmIRecommendedCommandName, f, ioStreams),
				kubectlwrappers.NewCmdCompletion(name, ioStreams),
			},
		},
	}
	groups.Add(cmds)

	filters := []string{
		"options",
		"deploy",
	}
	changeSharedFlagDefaults(cmds)
	cmdutil.ActsAsRootCommand(cmds, filters, groups...).
		ExposeFlags(loginCmd, "certificate-authority", "insecure-skip-tls-verify", "token")

	cmds.AddCommand(newExperimentalCommand("ex", name+" ex", f, ioStreams))

	cmds.AddCommand(kubectlwrappers.NewCmdPlugin(name, f, ioStreams))
	cmds.AddCommand(version.NewCmdVersion(name, f, ioStreams))
	cmds.AddCommand(options.NewCmdOptions(ioStreams))

	if cmds.Flag("namespace") != nil {
		if cmds.Flag("namespace").Annotations == nil {
			cmds.Flag("namespace").Annotations = map[string][]string{}
		}
		cmds.Flag("namespace").Annotations[cobra.BashCompCustom] = append(
			cmds.Flag("namespace").Annotations[cobra.BashCompCustom],
			"__oc_get_namespaces",
		)
	}

	return cmds
}

func moved(fullName, to string, parent, cmd *cobra.Command) string {
	cmd.Long = fmt.Sprintf("DEPRECATED: This command has been moved to \"%s %s\"", fullName, to)
	cmd.Short = fmt.Sprintf("DEPRECATED: %s", to)
	parent.AddCommand(cmd)
	return cmd.Name()
}

// changeSharedFlagDefaults changes values of shared flags that we disagree with.  This can't be done in godep code because
// that would change behavior in our `kubectl` symlink. Defend each change.
// 1. show-all - the most interesting pods are terminated/failed pods.  We don't want to exclude them from printing
func changeSharedFlagDefaults(rootCmd *cobra.Command) {
	cmds := []*cobra.Command{rootCmd}

	for i := 0; i < len(cmds); i++ {
		currCmd := cmds[i]
		cmds = append(cmds, currCmd.Commands()...)

		if showAllFlag := currCmd.Flags().Lookup("show-all"); showAllFlag != nil {
			showAllFlag.DefValue = "true"
			showAllFlag.Value.Set("true")
			showAllFlag.Changed = false
			showAllFlag.Usage = "When printing, show all resources (false means hide terminated pods.)"
		}

		// we want to disable the --validate flag by default when we're running kube commands from oc.  We want to make sure
		// that we're only getting the upstream --validate flags, so check both the flag and the usage
		if validateFlag := currCmd.Flags().Lookup("validate"); (validateFlag != nil) && (validateFlag.Usage == "If true, use a schema to validate the input before sending it") {
			validateFlag.DefValue = "false"
			validateFlag.Value.Set("false")
			validateFlag.Changed = false
		}
	}
}

func newExperimentalCommand(name, fullName string, f kcmdutil.Factory, ioStreams genericclioptions.IOStreams) *cobra.Command {
	experimental := &cobra.Command{
		Use:   name,
		Short: "Experimental commands under active development",
		Long:  "The commands grouped here are under development and may change without notice.",
		Run: func(c *cobra.Command, args []string) {
			c.SetOutput(ioStreams.Out)
			c.Help()
		},
		BashCompletionFunction: admin.BashCompletionFunc,
	}

	experimental.AddCommand(dockergc.NewCmdDockerGCConfig(f, fullName, "dockergc", ioStreams))
	experimental.AddCommand(buildchain.NewCmdBuildChain(name, fullName+" "+buildchain.BuildChainRecommendedCommandName, f, ioStreams))
	experimental.AddCommand(options.NewCmdOptions(ioStreams))

	// these groups also live under `oc adm groups {sync,prune}` and are here only for backwards compatibility
	experimental.AddCommand(sync.NewCmdSync("sync-groups", fullName+" "+"sync-groups", f, ioStreams))
	experimental.AddCommand(sync.NewCmdPrune("prune-groups", fullName+" "+"prune-groups", f, ioStreams))
	return experimental
}

// CommandFor returns the appropriate command for this base name,
// or the OpenShift CLI command.
func CommandFor(basename string) *cobra.Command {
	var cmd *cobra.Command

	in, out, errout := os.Stdin, os.Stdout, os.Stderr

	// Make case-insensitive and strip executable suffix if present
	if runtime.GOOS == "windows" {
		basename = strings.ToLower(basename)
		basename = strings.TrimSuffix(basename, ".exe")
	}

	switch basename {
	case "kubectl":
		cmd = kubecmd.NewDefaultKubectlCommand()
	case "openshift-deploy":
		cmd = deployer.NewCommandDeployer(basename)
	case "openshift-recycle":
		cmd = recycle.NewCommandRecycle(basename, out)
	default:
		shimKubectlForOc()
		cmd = NewDefaultOcCommand("oc", in, out, errout)

		// treat oc as a kubectl plugin
		if strings.HasPrefix(basename, "kubectl-") {
			args := strings.Split(strings.TrimPrefix(basename, "kubectl-"), "-")

			// the plugin mechanism interprets "_" as dashes. Convert any "_" our basename
			// might have in order to find the appropriate command in the `oc` tree.
			for i := range args {
				args[i] = strings.Replace(args[i], "_", "-", -1)
			}

			if targetCmd, _, err := cmd.Find(args); targetCmd != nil && err == nil {
				// since cobra refuses to execute a child command, executing its root
				// any time Execute() is called, we must create a completely new command
				// and "deep copy" the targetCmd information to it.
				newParent := &cobra.Command{
					Use:     targetCmd.Use,
					Short:   targetCmd.Short,
					Long:    targetCmd.Long,
					Example: targetCmd.Example,
					Run:     targetCmd.Run,
				}

				// copy flags
				newParent.Flags().AddFlagSet(cmd.Flags())
				newParent.Flags().AddFlagSet(targetCmd.Flags())
				newParent.PersistentFlags().AddFlagSet(targetCmd.PersistentFlags())

				// copy subcommands
				newParent.AddCommand(targetCmd.Commands()...)
				cmd = newParent
			}
		}
	}

	if cmd.UsageFunc() == nil {
		cmdutil.ActsAsRootCommand(cmd, []string{"options"})
	}
	return cmd
}
