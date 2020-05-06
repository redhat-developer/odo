package kubectlwrappers

import (
	"bufio"
	"flag"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	kclientcmd "k8s.io/client-go/tools/clientcmd"
	"k8s.io/kubectl/pkg/cmd/annotate"
	"k8s.io/kubectl/pkg/cmd/apiresources"
	"k8s.io/kubectl/pkg/cmd/apply"
	"k8s.io/kubectl/pkg/cmd/attach"
	"k8s.io/kubectl/pkg/cmd/autoscale"
	"k8s.io/kubectl/pkg/cmd/clusterinfo"
	"k8s.io/kubectl/pkg/cmd/completion"
	"k8s.io/kubectl/pkg/cmd/config"
	kcreate "k8s.io/kubectl/pkg/cmd/create"
	"k8s.io/kubectl/pkg/cmd/delete"
	"k8s.io/kubectl/pkg/cmd/describe"
	"k8s.io/kubectl/pkg/cmd/edit"
	"k8s.io/kubectl/pkg/cmd/exec"
	"k8s.io/kubectl/pkg/cmd/explain"
	kget "k8s.io/kubectl/pkg/cmd/get"
	"k8s.io/kubectl/pkg/cmd/label"
	"k8s.io/kubectl/pkg/cmd/patch"
	"k8s.io/kubectl/pkg/cmd/plugin"
	"k8s.io/kubectl/pkg/cmd/portforward"
	"k8s.io/kubectl/pkg/cmd/proxy"
	"k8s.io/kubectl/pkg/cmd/replace"
	"k8s.io/kubectl/pkg/cmd/run"
	"k8s.io/kubectl/pkg/cmd/scale"
	kcmdutil "k8s.io/kubectl/pkg/cmd/util"
	kwait "k8s.io/kubectl/pkg/cmd/wait"
	"k8s.io/kubectl/pkg/util/templates"
	kcmdauth "k8s.io/kubernetes/pkg/kubectl/cmd/auth"
	"k8s.io/kubernetes/pkg/kubectl/cmd/convert"
	"k8s.io/kubernetes/pkg/kubectl/cmd/cp"

	"github.com/openshift/oc/pkg/cli/create"
	cmdutil "github.com/openshift/oc/pkg/helpers/cmd"
)

func adjustCmdExamples(cmd *cobra.Command, fullName string, name string) {
	for _, subCmd := range cmd.Commands() {
		adjustCmdExamples(subCmd, fullName, cmd.Name())
	}
	cmd.Example = strings.Replace(cmd.Example, "kubectl", fullName, -1)
	tabbing := "  "
	examples := []string{}
	scanner := bufio.NewScanner(strings.NewReader(cmd.Example))
	for scanner.Scan() {
		examples = append(examples, tabbing+strings.TrimSpace(scanner.Text()))
	}
	cmd.Example = strings.Join(examples, "\n")
}

// NewCmdGet is a wrapper for the Kubernetes cli get command
func NewCmdGet(fullName string, f kcmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	return cmdutil.ReplaceCommandName("kubectl", fullName, templates.Normalize(kget.NewCmdGet(fullName, f, streams)))
}

// NewCmdReplace is a wrapper for the Kubernetes cli replace command
func NewCmdReplace(fullName string, f kcmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	return cmdutil.ReplaceCommandName("kubectl", fullName, templates.Normalize(replace.NewCmdReplace(f, streams)))
}

func NewCmdClusterInfo(fullName string, f kcmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	return cmdutil.ReplaceCommandName("kubectl", fullName, templates.Normalize(clusterinfo.NewCmdClusterInfo(f, streams)))
}

// NewCmdPatch is a wrapper for the Kubernetes cli patch command
func NewCmdPatch(fullName string, f kcmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	return cmdutil.ReplaceCommandName("kubectl", fullName, templates.Normalize(patch.NewCmdPatch(f, streams)))
}

// NewCmdDelete is a wrapper for the Kubernetes cli delete command
func NewCmdDelete(fullName string, f kcmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	return cmdutil.ReplaceCommandName("kubectl", fullName, templates.Normalize(delete.NewCmdDelete(f, streams)))
}

// NewCmdCreate is a wrapper for the Kubernetes cli create command
func NewCmdCreate(fullName string, f kcmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	cmd := cmdutil.ReplaceCommandName("kubectl", fullName, templates.Normalize(kcreate.NewCmdCreate(f, streams)))

	// create subcommands
	cmd.AddCommand(create.NewCmdCreateRoute(fullName, f, streams))
	cmd.AddCommand(create.NewCmdCreateDeploymentConfig(create.DeploymentConfigRecommendedName, fullName+" create "+create.DeploymentConfigRecommendedName, f, streams))
	cmd.AddCommand(create.NewCmdCreateClusterQuota(create.ClusterQuotaRecommendedName, fullName+" create "+create.ClusterQuotaRecommendedName, f, streams))

	cmd.AddCommand(create.NewCmdCreateUser(create.UserRecommendedName, fullName+" create "+create.UserRecommendedName, f, streams))
	cmd.AddCommand(create.NewCmdCreateIdentity(create.IdentityRecommendedName, fullName+" create "+create.IdentityRecommendedName, f, streams))
	cmd.AddCommand(create.NewCmdCreateUserIdentityMapping(create.UserIdentityMappingRecommendedName, fullName+" create "+create.UserIdentityMappingRecommendedName, f, streams))
	cmd.AddCommand(create.NewCmdCreateImageStream(create.ImageStreamRecommendedName, fullName+" create "+create.ImageStreamRecommendedName, f, streams))
	cmd.AddCommand(create.NewCmdCreateImageStreamTag(create.ImageStreamTagRecommendedName, fullName+" create "+create.ImageStreamTagRecommendedName, f, streams))

	adjustCmdExamples(cmd, fullName, "create")

	return cmd
}

var (
	completionLong = templates.LongDesc(`
		Output shell completion code for the specified shell (bash or zsh).
		The shell code must be evaluated to provide interactive
		completion of oc commands.  This can be done by sourcing it from
		the .bash_profile.

		Note for zsh users: [1] zsh completions are only supported in versions of zsh >= 5.2`)
)

func NewCmdCompletion(fullName string, streams genericclioptions.IOStreams) *cobra.Command {
	cmd := cmdutil.ReplaceCommandName("kubectl", fullName, templates.Normalize(completion.NewCmdCompletion(streams.Out, "\n")))
	cmd.Long = completionLong
	// mark all statically included flags as hidden to prevent them appearing in completions
	cmd.PreRun = func(c *cobra.Command, _ []string) {
		pflag.CommandLine.VisitAll(func(flag *pflag.Flag) {
			flag.Hidden = true
		})
		hideGlobalFlags(c.Root(), flag.CommandLine)
	}
	return cmd
}

// hideGlobalFlags marks any flag that is in the global flag set as
// hidden to prevent completion from varying by platform due to conditional
// includes. This means that some completions will not be possible unless
// they are registered in cobra instead of being added to flag.CommandLine.
func hideGlobalFlags(c *cobra.Command, fs *flag.FlagSet) {
	fs.VisitAll(func(flag *flag.Flag) {
		if f := c.PersistentFlags().Lookup(flag.Name); f != nil {
			f.Hidden = true
		}
		if f := c.LocalFlags().Lookup(flag.Name); f != nil {
			f.Hidden = true
		}
	})
	for _, child := range c.Commands() {
		hideGlobalFlags(child, fs)
	}
}

// NewCmdExec is a wrapper for the Kubernetes cli exec command
func NewCmdExec(fullName string, f kcmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	return cmdutil.ReplaceCommandName("kubectl", fullName, templates.Normalize(exec.NewCmdExec(f, streams)))
}

// NewCmdPortForward is a wrapper for the Kubernetes cli port-forward command
func NewCmdPortForward(fullName string, f kcmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	return cmdutil.ReplaceCommandName("kubectl", fullName, templates.Normalize(portforward.NewCmdPortForward(f, streams)))
}

// NewCmdDescribe is a wrapper for the Kubernetes cli describe command
func NewCmdDescribe(fullName string, f kcmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	return cmdutil.ReplaceCommandName("kubectl", fullName, templates.Normalize(describe.NewCmdDescribe(fullName, f, streams)))
}

// NewCmdProxy is a wrapper for the Kubernetes cli proxy command
func NewCmdProxy(fullName string, f kcmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	return cmdutil.ReplaceCommandName("kubectl", fullName, templates.Normalize(proxy.NewCmdProxy(f, streams)))
}

// NewCmdScale is a wrapper for the Kubernetes cli scale command
func NewCmdScale(fullName string, f kcmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	cmd := cmdutil.ReplaceCommandName("kubectl", fullName, templates.Normalize(scale.NewCmdScale(f, streams)))
	cmd.ValidArgs = append(cmd.ValidArgs, "deploymentconfig")
	return cmd
}

// NewCmdAutoscale is a wrapper for the Kubernetes cli autoscale command
func NewCmdAutoscale(fullName string, f kcmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	cmd := cmdutil.ReplaceCommandName("kubectl", fullName, templates.Normalize(autoscale.NewCmdAutoscale(f, streams)))
	cmd.Short = "Autoscale a deployment config, deployment, replica set, stateful set, or replication controller"
	cmd.ValidArgs = append(cmd.ValidArgs, "deploymentconfig")
	return cmd
}

// NewCmdRun is a wrapper for the Kubernetes cli run command
func NewCmdRun(fullName string, f kcmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	cmd := cmdutil.ReplaceCommandName("kubectl", fullName, templates.Normalize(run.NewCmdRun(f, streams)))
	cmd.Flags().Set("generator", "")
	cmd.Flag("generator").Usage = "The name of the API generator to use.  Default is 'deploymentconfig/v1' if --restart=Always, otherwise the default is 'run-pod/v1'."
	cmd.Flag("generator").DefValue = ""
	cmd.Flag("generator").Changed = false
	return cmd
}

// NewCmdAttach is a wrapper for the Kubernetes cli attach command
func NewCmdAttach(fullName string, f kcmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	return cmdutil.ReplaceCommandName("kubectl", fullName, templates.Normalize(attach.NewCmdAttach(f, streams)))
}

// NewCmdAnnotate is a wrapper for the Kubernetes cli annotate command
func NewCmdAnnotate(fullName string, f kcmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	return cmdutil.ReplaceCommandName("kubectl", fullName, templates.Normalize(annotate.NewCmdAnnotate(fullName, f, streams)))
}

// NewCmdLabel is a wrapper for the Kubernetes cli label command
func NewCmdLabel(fullName string, f kcmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	return cmdutil.ReplaceCommandName("kubectl", fullName, templates.Normalize(label.NewCmdLabel(f, streams)))
}

// NewCmdApply is a wrapper for the Kubernetes cli apply command
func NewCmdApply(fullName string, f kcmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	return cmdutil.ReplaceCommandName("kubectl", fullName, templates.Normalize(apply.NewCmdApply(fullName, f, streams)))
}

// NewCmdExplain is a wrapper for the Kubernetes cli explain command
func NewCmdExplain(fullName string, f kcmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	return cmdutil.ReplaceCommandName("kubectl", fullName, templates.Normalize(explain.NewCmdExplain(fullName, f, streams)))
}

// NewCmdConvert is a wrapper for the Kubernetes cli convert command
func NewCmdConvert(fullName string, f kcmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	return cmdutil.ReplaceCommandName("kubectl", fullName, templates.Normalize(convert.NewCmdConvert(f, streams)))
}

// NewCmdEdit is a wrapper for the Kubernetes cli edit command
func NewCmdEdit(fullName string, f kcmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	return cmdutil.ReplaceCommandName("kubectl", fullName, templates.Normalize(edit.NewCmdEdit(f, streams)))
}

// NewCmdConfig is a wrapper for the Kubernetes cli config command
func NewCmdConfig(fullName, name string, f kcmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	pathOptions := kclientcmd.NewDefaultPathOptions()

	return cmdutil.ReplaceCommandName("kubectl", fullName, templates.Normalize(config.NewCmdConfig(f, pathOptions, streams)))
}

// NewCmdCp is a wrapper for the Kubernetes cli cp command
func NewCmdCp(fullName string, f kcmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	return cmdutil.ReplaceCommandName("kubectl", fullName, templates.Normalize(cp.NewCmdCp(f, streams)))
}

func NewCmdWait(fullName string, f kcmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	return kwait.NewCmdWait(f, streams)
}

func NewCmdAuth(fullName string, f kcmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	return cmdutil.ReplaceCommandName("kubectl", fullName, templates.Normalize(kcmdauth.NewCmdAuth(f, streams)))
}

func NewCmdPlugin(fullName string, f kcmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	// list of accepted plugin executable filename prefixes that we will look for
	// when executing a plugin. Order matters here, we want to first see if a user
	// has prefixed their plugin with "oc-", before defaulting to upstream behavior.
	plugin.ValidPluginFilenamePrefixes = []string{"oc", "kubectl"}
	return plugin.NewCmdPlugin(f, streams)
}

func NewCmdApiResources(fullName string, f kcmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	return cmdutil.ReplaceCommandName("kubectl", fullName, templates.Normalize(apiresources.NewCmdAPIResources(f, streams)))
}

func NewCmdApiVersions(fullName string, f kcmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	return cmdutil.ReplaceCommandName("kubectl", fullName, templates.Normalize(apiresources.NewCmdAPIVersions(f, streams)))
}
