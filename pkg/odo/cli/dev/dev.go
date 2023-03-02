package dev

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"

	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubectl/pkg/util/templates"

	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/dev"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/libdevfile"
	"github.com/redhat-developer/odo/pkg/log"
	clierrors "github.com/redhat-developer/odo/pkg/odo/cli/errors"
	"github.com/redhat-developer/odo/pkg/odo/cli/messages"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/commonflags"
	fcontext "github.com/redhat-developer/odo/pkg/odo/commonflags/context"
	odocontext "github.com/redhat-developer/odo/pkg/odo/context"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/redhat-developer/odo/pkg/podman"
	scontext "github.com/redhat-developer/odo/pkg/segment/context"
	"github.com/redhat-developer/odo/pkg/util"
	"github.com/redhat-developer/odo/pkg/version"
)

// RecommendedCommandName is the recommended command name
const (
	RecommendedCommandName = "dev"
)

type DevOptions struct {
	// Clients
	clientset *clientset.Clientset

	// Variables
	ignorePaths []string
	out         io.Writer
	errOut      io.Writer

	// ctx is used to communicate with WatchAndPush to stop watching and start cleaning up
	ctx context.Context

	// cancel function ensures that any function/method listening on ctx.Done channel stops doing its work
	cancel context.CancelFunc

	// Flags
	noWatchFlag          bool
	randomPortsFlag      bool
	debugFlag            bool
	buildCommandFlag     string
	runCommandFlag       string
	ignoreLocalhostFlag  bool
	forwardLocalhostFlag bool
}

var _ genericclioptions.Runnable = (*DevOptions)(nil)
var _ genericclioptions.SignalHandler = (*DevOptions)(nil)

func NewDevOptions() *DevOptions {
	return &DevOptions{
		out:    log.GetStdout(),
		errOut: log.GetStderr(),
	}
}

var devExample = ktemplates.Examples(`
	# Run your application on the cluster in the Dev mode, using the default run command
	%[1]s

	# Run your application on the cluster in the Dev mode, using the specified run command
	%[1]s --run-command <my-command>

	# Run your application on the cluster in the Dev mode, without automatically syncing the code upon any file changes
	%[1]s --no-watch
`)

func (o *DevOptions) SetClientset(clientset *clientset.Clientset) {
	o.clientset = clientset
}

func (o *DevOptions) PreInit() string {
	return messages.DevInitializeExistingComponent
}

func (o *DevOptions) Complete(ctx context.Context, cmdline cmdline.Cmdline, args []string) error {
	// Define this first so that if user hits Ctrl+c very soon after running odo dev, odo doesn't panic
	o.ctx, o.cancel = context.WithCancel(ctx)

	return nil
}

func (o *DevOptions) Validate(ctx context.Context) error {
	devfileObj := *odocontext.GetDevfileObj(ctx)
	if !o.debugFlag && !libdevfile.HasRunCommand(devfileObj.Data) {
		return clierrors.NewNoCommandInDevfileError("run")
	}
	if o.debugFlag && !libdevfile.HasDebugCommand(devfileObj.Data) {
		return clierrors.NewNoCommandInDevfileError("debug")
	}

	platform := fcontext.GetPlatform(ctx, commonflags.PlatformCluster)
	switch platform {
	case commonflags.PlatformCluster:
		if o.ignoreLocalhostFlag {
			return errors.New("--ignore-localhost cannot be used when running in cluster mode")
		}
		if o.forwardLocalhostFlag {
			return errors.New("--forward-localhost cannot be used when running in cluster mode")
		}
		if o.clientset.KubernetesClient == nil {
			return kclient.NewNoConnectionError()
		}
		scontext.SetPlatform(ctx, o.clientset.KubernetesClient)
	case commonflags.PlatformPodman:
		if o.ignoreLocalhostFlag && o.forwardLocalhostFlag {
			return errors.New("--ignore-localhost and --forward-localhost cannot be used together")
		}
		if o.clientset.PodmanClient == nil {
			return podman.NewPodmanNotFoundError(nil)
		}
		scontext.SetPlatform(ctx, o.clientset.PodmanClient)
	}
	return nil
}

func (o *DevOptions) Run(ctx context.Context) (err error) {
	var (
		devFileObj    = odocontext.GetDevfileObj(ctx)
		devfilePath   = odocontext.GetDevfilePath(ctx)
		path          = filepath.Dir(devfilePath)
		componentName = odocontext.GetComponentName(ctx)
		variables     = fcontext.GetVariables(ctx)
		platform      = fcontext.GetPlatform(ctx, commonflags.PlatformCluster)
	)

	var dest string
	var deployingTo string
	switch platform {
	case commonflags.PlatformPodman:
		dest = "Platform: podman"
		deployingTo = "podman"
	case commonflags.PlatformCluster:
		dest = "Namespace: " + odocontext.GetNamespace(ctx)
		deployingTo = "the cluster"
	default:
		panic(fmt.Errorf("platform %s is not implemented", platform))
	}

	// Output what the command is doing / information
	log.Title("Developing using the \""+componentName+"\" Devfile",
		dest,
		"odo version: "+version.VERSION)

	// check for .gitignore file and add odo-file-index.json to .gitignore.
	// In case the .gitignore was created by odo, it is purposely not reported as candidate for deletion (via a call to files.ReportLocalFileGeneratedByOdo)
	// because a .gitignore file is more likely to be modified by the user afterward (for another usage).
	gitIgnoreFile, _, err := util.TouchGitIgnoreFile(path)
	if err != nil {
		return err
	}

	// add .odo dir to .gitignore
	err = util.AddOdoDirectory(gitIgnoreFile)
	if err != nil {
		return err
	}

	var ignores []string
	err = genericclioptions.ApplyIgnore(&ignores, "")
	if err != nil {
		return err
	}
	// Ignore the devfile, as it will be handled independently
	o.ignorePaths = ignores

	scontext.SetComponentType(ctx, component.GetComponentTypeFromDevfileMetadata(devFileObj.Data.GetMetadata()))
	scontext.SetLanguage(ctx, devFileObj.Data.GetMetadata().Language)
	scontext.SetProjectType(ctx, devFileObj.Data.GetMetadata().ProjectType)
	scontext.SetDevfileName(ctx, componentName)

	log.Sectionf("Running on %s in Dev mode", deployingTo)

	return o.clientset.DevClient.Start(
		o.ctx,
		o.out,
		o.errOut,
		dev.StartOptions{
			IgnorePaths:      o.ignorePaths,
			Debug:            o.debugFlag,
			BuildCommand:     o.buildCommandFlag,
			RunCommand:       o.runCommandFlag,
			RandomPorts:      o.randomPortsFlag,
			WatchFiles:       !o.noWatchFlag,
			IgnoreLocalhost:  o.ignoreLocalhostFlag,
			ForwardLocalhost: o.forwardLocalhostFlag,
			Variables:        variables,
		},
	)
}

func (o *DevOptions) HandleSignal() error {
	o.cancel()
	// At this point, `ctx.Done()` will be raised, and the cleanup will be done
	// wait for the cleanup to finish and let the main thread finish instead of signal handler go routine from runnable
	select {}
}

func (o *DevOptions) Cleanup(ctx context.Context, commandError error) {
	if commandError != nil {
		_ = o.clientset.DevClient.CleanupResources(ctx, log.GetStdout())
	}
	_ = o.clientset.StateClient.SaveExit()
}

// NewCmdDev implements the odo dev command
func NewCmdDev(name, fullName string) *cobra.Command {
	o := NewDevOptions()
	devCmd := &cobra.Command{
		Use:   name,
		Short: "Run your application on the cluster in the Dev mode",
		Long: `odo dev is a long running command that will automatically sync your source to the cluster.
It forwards endpoints with any exposure values ('public', 'internal' or 'none') to a port on localhost.`,
		Example: fmt.Sprintf(devExample, fullName),
		Args:    cobra.MaximumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return genericclioptions.GenericRun(o, cmd, args)
		},
	}
	devCmd.Flags().BoolVar(&o.noWatchFlag, "no-watch", false, "Do not watch for file changes")
	devCmd.Flags().BoolVar(&o.randomPortsFlag, "random-ports", false, "Assign random ports to redirected ports")
	devCmd.Flags().BoolVar(&o.debugFlag, "debug", false, "Execute the debug command within the component")
	devCmd.Flags().StringVar(&o.buildCommandFlag, "build-command", "",
		"Alternative build command. The default one will be used if this flag is not set.")
	devCmd.Flags().StringVar(&o.runCommandFlag, "run-command", "",
		"Alternative run command to execute. The default one will be used if this flag is not set.")
	devCmd.Flags().BoolVar(&o.ignoreLocalhostFlag, "ignore-localhost", false,
		"Whether to ignore errors related to port-forwarding apps listening on the container loopback interface. Applicable only if platform is podman.")
	// TODO Unhide when moving Podman out of the experimental mode : https://github.com/redhat-developer/odo/issues/6592
	_ = devCmd.Flags().MarkHidden("ignore-localhost")
	devCmd.Flags().BoolVar(&o.forwardLocalhostFlag, "forward-localhost", false,
		"Whether to enable port-forwarding if app is listening on the container loopback interface. Applicable only if platform is podman.")
	// TODO Unhide when moving Podman out of the experimental mode : https://github.com/redhat-developer/odo/issues/6592
	_ = devCmd.Flags().MarkHidden("forward-localhost")

	clientset.Add(devCmd,
		clientset.BINDING,
		clientset.DEV,
		clientset.EXEC,
		clientset.FILESYSTEM,
		clientset.INIT,
		clientset.KUBERNETES_NULLABLE,
		clientset.PODMAN_NULLABLE,
		clientset.PORT_FORWARD,
		clientset.PREFERENCE,
		clientset.STATE,
		clientset.SYNC,
		clientset.WATCH,
	)
	// Add a defined annotation in order to appear in the help menu
	odoutil.SetCommandGroup(devCmd, odoutil.MainGroup)
	devCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)
	commonflags.UseVariablesFlags(devCmd)
	commonflags.UsePlatformFlag(devCmd)
	return devCmd
}
