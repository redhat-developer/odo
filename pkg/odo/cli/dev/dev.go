package dev

import (
	"context"
	"errors"
	"fmt"
	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/redhat-developer/odo/pkg/api"
	"io"
	"k8s.io/klog"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

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
	ignorePaths    []string
	out            io.Writer
	errOut         io.Writer
	forwardedPorts []api.ForwardedPort

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
	// TODO: Add doc
	portForwardFlag []string
	addressFlag     string
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

	if o.randomPortsFlag && o.portForwardFlag != nil {
		return errors.New("--random-ports and --port-forward cannot be used together")
	}

	if o.portForwardFlag != nil {
		containerEndpointMapping, err := libdevfile.GetDevfileContainerEndpointMapping(devfileObj, true)
		if err != nil {
			return fmt.Errorf("failed to obtain container endpoints to validate --port-forward ports; cause:%w", err)
		}

		forwardedPorts, err := parsePortForwardFlag(o.portForwardFlag)
		if err != nil {
			return err
		}
		o.forwardedPorts = forwardedPorts

		return validatePortForwardFlagData(forwardedPorts, containerEndpointMapping)
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
	if platform == commonflags.PlatformCluster {
		genericclioptions.WarnIfDefaultNamespace(odocontext.GetNamespace(ctx), o.clientset.KubernetesClient)
	}
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
			IgnorePaths:          o.ignorePaths,
			Debug:                o.debugFlag,
			BuildCommand:         o.buildCommandFlag,
			RunCommand:           o.runCommandFlag,
			RandomPorts:          o.randomPortsFlag,
			WatchFiles:           !o.noWatchFlag,
			IgnoreLocalhost:      o.ignoreLocalhostFlag,
			ForwardLocalhost:     o.forwardLocalhostFlag,
			Variables:            variables,
			CustomForwardedPorts: o.forwardedPorts,
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
	devCmd.Flags().BoolVar(&o.forwardLocalhostFlag, "forward-localhost", false,
		"Whether to enable port-forwarding if app is listening on the container loopback interface. Applicable only if platform is podman.")
	// TODO: Fix usage
	//   kubectl port-forward TYPE/NAME [options] [LOCAL_PORT:]REMOTE_PORT [...[LOCAL_PORT_N:]REMOTE_PORT_N]
	devCmd.Flags().StringArrayVar(&o.portForwardFlag, "port-forward", nil,
		"Define custom port forwards")
	devCmd.Flags().StringVar(&o.addressFlag, "address", "",
		"Which local address the port will be listening")

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

// validatePortForwardFlagData runs validation checks on the --port-forward flag data as follows:
// 1. Every container port defined by the flag is present in the devfile
// 2. Every local port defined by the flag is unique
// 3. If multiple containers have the same container port, the validation fails and asks the user to provide container names
func validatePortForwardFlagData(forwardedPorts []api.ForwardedPort, containerEndpointMapping map[string][]v1alpha2.Endpoint) error {
	// TODO: Should we check for all errors before returning?

	// Validate that local Ports present in forwardedPorts are unique
	var localPorts = make(map[int]struct{})
	for _, fPort := range forwardedPorts {
		if _, ok := localPorts[fPort.LocalPort]; ok {
			return fmt.Errorf("local port %d is used more than once, please use unique local ports", fPort.LocalPort)
		}
		localPorts[fPort.LocalPort] = struct{}{}
	}

	// portContainerMapping maps a target port with a list of containers where it is present.
	// This is useful for checking that no custom port-forwarding config contains duplicate containerPort without a container name
	// {"runtime": {{TargetPort: 9090}, {TargetPort: 8000}}, "tools":   {{TargetPort: 9090}}}
	// >>>> {9090: {"runtime", "tools"}, 8000: {"runtime"}}
	portContainerMapping := make(map[int][]string)
	for container, endpoints := range containerEndpointMapping {
		for _, endpoint := range endpoints {
			portContainerMapping[endpoint.TargetPort] = append(portContainerMapping[endpoint.TargetPort], container)
		}
	}

	// 	Check that all endpoints are valid and present in the Devfile
portLoop:
	for _, fPort := range forwardedPorts {
		if fPort.ContainerName != "" {
			if containerEndpoints, ok := containerEndpointMapping[fPort.ContainerName]; ok {
				for _, endpoint := range containerEndpoints {
					if endpoint.TargetPort == fPort.ContainerPort {
						klog.V(1).Infof("%d container port matches %s endpoints of container:%s", fPort.ContainerPort, endpoint.Name, fPort.ContainerName)
						continue portLoop
					}
				}
				return fmt.Errorf("%d container port does not match any endpoints of container:%s", fPort.ContainerPort, fPort.ContainerName)
			} else {
				return fmt.Errorf("container:%s defined by --port-forward not found", fPort.ContainerName)
			}
		} else {
			// Validate that a custom portforwarding config without a container targets a unique containerPort
			if containers, ok := portContainerMapping[fPort.ContainerPort]; ok && len(containers) > 1 {
				return fmt.Errorf("multiple container component (%s) found with same container port %d, port forwarding must be defined with format <localPort>:<containerName>:<containerPort>", strings.Join(portContainerMapping[fPort.ContainerPort], ", "), fPort.ContainerPort)
			}
			for _, containerEndpoints := range containerEndpointMapping {
				for _, endpoint := range containerEndpoints {
					if endpoint.TargetPort == fPort.ContainerPort {
						klog.V(1).Infof("%d port matches %s endpoints of container %s", fPort.ContainerPort, endpoint.Name, fPort.ContainerName)
						continue portLoop
					}
				}
			}
			return fmt.Errorf("%d container port defined by --port-forward not found", fPort.ContainerPort)
		}
	}
	return nil
}

// parsePortForwardFlag parses custom port forwarding configuration; acceptable patterns: <localPort>:<containerPort>, <localPort>:<containerName>:<containerPort>
func parsePortForwardFlag(portForwardFlag []string) (forwardedPorts []api.ForwardedPort, err error) {
	// acceptable examples: 8000:runtime_123:8080, 8000:9000, 8000:runtime:8080, 20001:20000
	// unacceptable examples: :8000, 80000:
	pattern := `^(\d{1,5})(:\w*)?:(\d{1,5})$`
	const largestPortValue = 65535
	for _, portData := range portForwardFlag {
		if matched, _ := regexp.MatchString(pattern, portData); !matched {
			return nil, errors.New("ports are not defined properly, acceptable formats are: <localPort>:<containerPort>, <localPort>:<containerName>:<containerPort>")
		}
		var portF api.ForwardedPort

		portDataArr := strings.Split(portData, ":")
		switch len(portDataArr) {
		case 2:
			portF.LocalPort, _ = strconv.Atoi(portDataArr[0])
			portF.ContainerPort, _ = strconv.Atoi(portDataArr[1])
		case 3:
			portF.LocalPort, _ = strconv.Atoi(portDataArr[0])
			portF.ContainerName = portDataArr[1]
			portF.ContainerPort, _ = strconv.Atoi(portDataArr[2])
		}
		if !(portF.LocalPort > 0 && portF.LocalPort <= largestPortValue) || !(portF.ContainerPort > 0 && portF.ContainerPort <= largestPortValue) {
			return nil, fmt.Errorf("%s is invalid; port number must not be greater than 0 and less than %d", portData, largestPortValue)
		}
		forwardedPorts = append(forwardedPorts, portF)
	}
	return forwardedPorts, nil
}
