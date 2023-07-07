package logs

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/redhat-developer/odo/pkg/kclient"
	odolabels "github.com/redhat-developer/odo/pkg/labels"
	"github.com/redhat-developer/odo/pkg/podman"

	"github.com/redhat-developer/odo/pkg/log"

	"github.com/redhat-developer/odo/pkg/devfile/location"
	"github.com/redhat-developer/odo/pkg/odo/commonflags"
	fcontext "github.com/redhat-developer/odo/pkg/odo/commonflags/context"
	"github.com/redhat-developer/odo/pkg/odo/util"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"

	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubectl/pkg/util/templates"

	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	odocontext "github.com/redhat-developer/odo/pkg/odo/context"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
)

const RecommendedCommandName = "logs"

type LogsOptions struct {
	// clients
	clientset *clientset.Clientset

	// variables
	out io.Writer

	// flags
	devMode    bool
	deployMode bool
	follow     bool
}

var _ genericclioptions.Runnable = (*LogsOptions)(nil)
var _ genericclioptions.SignalHandler = (*LogsOptions)(nil)

type logsMode string

const (
	DevMode    logsMode = "dev"
	DeployMode logsMode = "deploy"
)

func NewLogsOptions() *LogsOptions {
	return &LogsOptions{
		out: log.GetStdout(),
	}
}

var logsExample = ktemplates.Examples(`
	# Show logs of all containers
	%[1]s
`)

func (o *LogsOptions) SetClientset(clientset *clientset.Clientset) {
	o.clientset = clientset
}

func (o *LogsOptions) Complete(ctx context.Context, cmdline cmdline.Cmdline, _ []string) error {
	var err error
	workingDir := odocontext.GetWorkingDirectory(ctx)
	isEmptyDir, err := location.DirIsEmpty(o.clientset.FS, workingDir)
	if err != nil {
		return err
	}
	if isEmptyDir {
		return errors.New("this command cannot run in an empty directory, run the command in a directory containing source code or initialize using 'odo init'")
	}

	devfileObj := odocontext.GetEffectiveDevfileObj(ctx)
	if devfileObj == nil {
		return genericclioptions.NewNoDevfileError(odocontext.GetWorkingDirectory(ctx))
	}
	return nil
}

func (o *LogsOptions) Validate(ctx context.Context) error {

	switch fcontext.GetPlatform(ctx, commonflags.PlatformCluster) {
	case commonflags.PlatformCluster:
		if o.clientset.KubernetesClient == nil {
			return kclient.NewNoConnectionError()
		}
	case commonflags.PlatformPodman:
		if o.clientset.PodmanClient == nil {
			return podman.NewPodmanNotFoundError(nil)
		}
	}

	if o.devMode && o.deployMode {
		return errors.New("pass only one of --dev or --deploy flags; pass no flag to see logs for both modes")
	}
	return nil
}

func (o *LogsOptions) Run(ctx context.Context) error {
	var logMode logsMode

	componentName := odocontext.GetComponentName(ctx)

	if o.devMode {
		logMode = DevMode
	} else if o.deployMode {
		logMode = DeployMode
	}

	var mode string
	switch logMode {
	case DevMode:
		mode = odolabels.ComponentDevMode
	case DeployMode:
		mode = odolabels.ComponentDeployMode
	default:
		mode = odolabels.ComponentAnyMode
	}

	ns := ""
	if o.clientset.KubernetesClient != nil {
		ns = odocontext.GetNamespace(ctx)
	}

	return o.clientset.LogsClient.DisplayLogs(
		ctx,
		mode,
		componentName,
		ns,
		o.follow,
		o.out,
	)
}

func (o *LogsOptions) HandleSignal(ctx context.Context, cancelFunc context.CancelFunc) error {
	cancelFunc()
	select {}
}

func NewCmdLogs(name, fullname string, testClientset clientset.Clientset) *cobra.Command {
	o := NewLogsOptions()
	logsCmd := &cobra.Command{
		Use:   name,
		Short: "Show logs of all containers of the component",
		Long: `odo logs shows logs of all containers of the component. 
By default it shows logs of all containers running in both Dev and Deploy mode. It prefixes each log message with the container name.`,
		Example: fmt.Sprintf(logsExample, fullname),
		Args:    cobra.MaximumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return genericclioptions.GenericRun(o, testClientset, cmd, args)
		},
	}
	logsCmd.Flags().BoolVar(&o.devMode, string(DevMode), false, "Show logs for containers running only in Dev mode")
	logsCmd.Flags().BoolVar(&o.deployMode, string(DeployMode), false, "Show logs for containers running only in Deploy mode")
	logsCmd.Flags().BoolVar(&o.follow, "follow", false, "Follow/tail the logs of the pods")

	clientset.Add(logsCmd, clientset.LOGS, clientset.FILESYSTEM)
	util.SetCommandGroup(logsCmd, util.MainGroup)
	logsCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)
	commonflags.UsePlatformFlag(logsCmd)
	return logsCmd
}
