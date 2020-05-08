package rsh

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/kubectl/pkg/cmd/exec"
	kcmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/templates"
	"k8s.io/kubectl/pkg/util/term"
)

const (
	RshRecommendedName   = "rsh"
	DefaultShell         = "/bin/sh"
	defaultPodRshTimeout = 60 * time.Second
)

var (
	rshUsageStr    = "rsh (POD | TYPE/NAME) [-c CONTAINER] [flags] -- COMMAND [args...]"
	rshUsageErrStr = fmt.Sprintf("expected '%s'.\nPOD or TYPE/NAME is a required argument for the rsh command", rshUsageStr)

	rshLong = templates.LongDesc(`
		Open a remote shell session to a container

		This command will attempt to start a shell session in a pod for the specified resource.
		It works with pods, deployment configs, deployments, jobs, daemon sets, replication controllers
		and replica sets.
		Any of the aforementioned resources (apart from pods) will be resolved to a ready pod.
		It will default to the first container if none is specified, and will attempt to use
		'/bin/sh' as the default shell. You may pass any flags supported by this command before
		the resource name, and an optional command after the resource name, which will be executed
		instead of a login shell. A TTY will be automatically allocated if standard input is
		interactive - use -t and -T to override. A TERM variable is sent to the environment where
		the shell (or command) will be executed. By default its value is the same as the TERM
		variable from the local environment; if not set, 'xterm' is used.

		Note, some containers may not include a shell - use '%[1]s exec' if you need to run commands
		directly.`)

	rshExample = templates.Examples(`
	  # Open a shell session on the first container in pod 'foo'
	  %[1]s foo

	  # Open a shell session on the first container in pod 'foo' and namespace 'bar'
	  # (Note that oc client specific arguments must come before the resource name and its arguments)
	  %[1]s -n bar foo

	  # Run the command 'cat /etc/resolv.conf' inside pod 'foo'
	  %[1]s foo cat /etc/resolv.conf

	  # See the configuration of your internal registry
	  %[1]s dc/docker-registry cat config.yml

	  # Open a shell session on the container named 'index' inside a pod of your job
	  %[1]s -c index job/sheduled`)
)

// RshOptions declare the arguments accepted by the Rsh command
type RshOptions struct {
	ForceTTY   bool
	DisableTTY bool
	Executable string
	Timeout    int
	*exec.ExecOptions
}

func NewRshOptions(parent string, streams genericclioptions.IOStreams) *RshOptions {
	return &RshOptions{
		ForceTTY:   false,
		DisableTTY: false,
		Timeout:    10,
		Executable: DefaultShell,
		ExecOptions: &exec.ExecOptions{
			StreamOptions: exec.StreamOptions{
				IOStreams: streams,
				TTY:       true,
				Stdin:     true,
			},

			ParentCommandName: parent,
			Executor:          &exec.DefaultRemoteExecutor{},
		},
	}
}

// NewCmdRsh returns a command that attempts to open a shell session to the server.
func NewCmdRsh(name string, parent string, f kcmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	options := NewRshOptions(parent, streams)

	cmd := &cobra.Command{
		Use:                   rshUsageStr,
		DisableFlagsInUseLine: true,
		Short:                 "Start a shell session in a container.",
		Long:                  fmt.Sprintf(rshLong, parent),
		Example:               fmt.Sprintf(rshExample, parent+" "+name),
		Run: func(cmd *cobra.Command, args []string) {
			kcmdutil.CheckErr(options.Complete(f, cmd, args))
			kcmdutil.CheckErr(options.Validate())
			kcmdutil.CheckErr(options.Run())
		},
	}
	kcmdutil.AddPodRunningTimeoutFlag(cmd, defaultPodRshTimeout)
	cmd.Flags().BoolVarP(&options.ForceTTY, "tty", "t", options.ForceTTY, "Force a pseudo-terminal to be allocated")
	cmd.Flags().BoolVarP(&options.DisableTTY, "no-tty", "T", options.DisableTTY, "Disable pseudo-terminal allocation")
	cmd.Flags().StringVar(&options.Executable, "shell", options.Executable, "Path to the shell command")
	cmd.Flags().IntVar(&options.Timeout, "timeout", options.Timeout, "Request timeout for obtaining a pod from the server; defaults to 10 seconds")
	cmd.Flags().MarkDeprecated("timeout", "use --request-timeout, instead.")
	cmd.Flags().StringVarP(&options.ContainerName, "container", "c", options.ContainerName, "Container name; defaults to first container")
	cmd.Flags().SetInterspersed(false)
	return cmd
}

// Complete applies the command environment to RshOptions
func (o *RshOptions) Complete(f kcmdutil.Factory, cmd *cobra.Command, args []string) error {
	argsLenAtDash := cmd.ArgsLenAtDash()
	if len(args) == 0 || argsLenAtDash == 0 {
		return kcmdutil.UsageErrorf(cmd, "%s", rshUsageErrStr)
	}

	switch {
	case o.ForceTTY && o.DisableTTY:
		return kcmdutil.UsageErrorf(cmd, "you may not specify -t and -T together")
	case o.ForceTTY:
		o.TTY = true
	case o.DisableTTY:
		o.TTY = false
	default:
		o.TTY = term.IsTerminal(o.In)
	}

	if err := o.ExecOptions.Complete(f, cmd, args, argsLenAtDash); err != nil {
		return err
	}

	// overwrite ExecOptions with rsh specifics
	args = args[1:]
	if len(args) > 0 {
		o.Command = args
	} else {
		o.Command = []string{o.Executable}
	}

	fullCmdName := ""
	cmdParent := cmd.Parent()
	if cmdParent != nil {
		fullCmdName = cmdParent.CommandPath()
	}
	o.ExecOptions.EnableSuggestedCmdUsage = len(fullCmdName) > 0 && kcmdutil.IsSiblingCommandExists(cmd, "describe")

	return nil
}

// Validate ensures that RshOptions are valid
func (o *RshOptions) Validate() error {
	return o.ExecOptions.Validate()
}

// Run starts a remote shell session on the server
func (o *RshOptions) Run() error {
	// Insert the TERM into the command to be run
	if len(o.Command) == 1 && o.Command[0] == DefaultShell {
		term := os.Getenv("TERM")
		if len(term) == 0 {
			term = "xterm"
		}
		termsh := fmt.Sprintf("TERM=%q %s", term, DefaultShell)
		o.Command = append(o.Command, "-c", termsh)
	}
	return o.ExecOptions.Run()
}
