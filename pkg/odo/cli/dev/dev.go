package dev

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/devfile/library/pkg/devfile/parser"
	ktemplates "k8s.io/kubectl/pkg/util/templates"

	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/dev"
	ododevfile "github.com/redhat-developer/odo/pkg/devfile"
	"github.com/redhat-developer/odo/pkg/devfile/adapters"
	kcomponent "github.com/redhat-developer/odo/pkg/devfile/adapters/kubernetes/component"
	"github.com/redhat-developer/odo/pkg/devfile/location"
	"github.com/redhat-developer/odo/pkg/libdevfile"
	"github.com/redhat-developer/odo/pkg/log"
	clierrors "github.com/redhat-developer/odo/pkg/odo/cli/errors"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
	scontext "github.com/redhat-developer/odo/pkg/segment/context"
	"github.com/redhat-developer/odo/pkg/util"
	"github.com/redhat-developer/odo/pkg/vars"
	"github.com/redhat-developer/odo/pkg/version"
	"github.com/redhat-developer/odo/pkg/watch"
)

// RecommendedCommandName is the recommended command name
const RecommendedCommandName = "dev"

type DevOptions struct {
	// Context
	*genericclioptions.Context

	// Clients
	clientset *clientset.Clientset

	// Variables
	ignorePaths []string
	out         io.Writer
	errOut      io.Writer
	// it's called "initial" because it has to be set only once when running odo dev for the first time
	// it is used to compare with updated devfile when we watch the contextDir for changes
	initialDevfileObj parser.DevfileObj
	// ctx is used to communicate with WatchAndPush to stop watching and start cleaning up
	ctx context.Context
	// cancel function ensures that any function/method listening on ctx.Done channel stops doing its work
	cancel context.CancelFunc

	// working directory
	contextDir string

	// Flags
	noWatchFlag      bool
	randomPortsFlag  bool
	debugFlag        bool
	varFileFlag      string
	varsFlag         []string
	buildCommandFlag string
	runCommandFlag   string

	// Variables to override Devfile variables
	variables map[string]string
}

var _ genericclioptions.Runnable = (*DevOptions)(nil)
var _ genericclioptions.SignalHandler = (*DevOptions)(nil)

type Handler struct {
	clientset   clientset.Clientset
	randomPorts bool
	errOut      io.Writer
}

var _ dev.Handler = (*Handler)(nil)

func NewDevOptions() *DevOptions {
	return &DevOptions{
		out:    log.GetStdout(),
		errOut: log.GetStderr(),
	}
}

var devExample = ktemplates.Examples(`
	# Deploy component to the development cluster, using the default run command
	%[1]s

	# Deploy component to the development cluster, using the specified run command
	%[1]s --run-command <my-command>

	# Deploy component to the development cluster without automatically syncing the code upon any file changes
	%[1]s --no-watch
`)

func (o *DevOptions) SetClientset(clientset *clientset.Clientset) {
	o.clientset = clientset
}

func (o *DevOptions) Complete(cmdline cmdline.Cmdline, args []string) error {
	var err error

	// Define this first so that if user hits Ctrl+c very soon after running odo dev, odo doesn't panic
	o.ctx, o.cancel = context.WithCancel(context.Background())

	o.contextDir, err = os.Getwd()
	if err != nil {
		return err
	}

	isEmptyDir, err := location.DirIsEmpty(o.clientset.FS, o.contextDir)
	if err != nil {
		return err
	}
	if isEmptyDir {
		return genericclioptions.NewNoDevfileError(o.contextDir)
	}
	initFlags := o.clientset.InitClient.GetFlags(cmdline.GetFlags())
	err = o.clientset.InitClient.InitDevfile(initFlags, o.contextDir,
		func(interactiveMode bool) {
			scontext.SetInteractive(cmdline.Context(), interactiveMode)
			if interactiveMode {
				fmt.Println("The current directory already contains source code. " +
					"odo will try to autodetect the language and project type in order to select the best suited Devfile for your project.")
			}
		},
		func(newDevfileObj parser.DevfileObj) error {
			return newDevfileObj.WriteYamlDevfile()
		})
	if err != nil {
		return err
	}

	o.variables, err = vars.GetVariables(o.clientset.FS, o.varFileFlag, o.varsFlag, os.LookupEnv)
	if err != nil {
		return err
	}

	o.Context, err = genericclioptions.New(genericclioptions.NewCreateParameters(cmdline).NeedDevfile("").WithVariables(o.variables))
	if err != nil {
		return fmt.Errorf("unable to create context: %v", err)
	}

	o.initialDevfileObj = o.Context.EnvSpecificInfo.GetDevfileObj()

	o.clientset.KubernetesClient.SetNamespace(o.GetProject())

	return nil
}

func (o *DevOptions) Validate() error {
	if !o.debugFlag && !libdevfile.HasRunCommand(o.initialDevfileObj.Data) {
		return clierrors.NewNoCommandInDevfileError("run")
	}
	if o.debugFlag && !libdevfile.HasDebugCommand(o.initialDevfileObj.Data) {
		return clierrors.NewNoCommandInDevfileError("debug")
	}
	return nil
}

func (o *DevOptions) Run(ctx context.Context) (err error) {
	var (
		devFileObj  = o.Context.EnvSpecificInfo.GetDevfileObj()
		path        = filepath.Dir(o.Context.EnvSpecificInfo.GetDevfilePath())
		namespace   = o.GetProject()
		devfileName = o.GetComponentName()
	)

	// Output what the command is doing / information
	log.Title("Developing using the "+devfileName+" Devfile",
		"Namespace: "+namespace,
		"odo version: "+version.VERSION)

	// check for .gitignore file and add odo-file-index.json to .gitignore
	gitIgnoreFile, err := util.TouchGitIgnoreFile(path)
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

	log.Section("Deploying to the cluster in developer mode")
	componentStatus, err := o.clientset.DevClient.Start(
		devFileObj,
		devfileName,
		namespace,
		o.ignorePaths,
		path,
		o.debugFlag,
		o.buildCommandFlag,
		o.runCommandFlag,
		o.randomPortsFlag,
		o.errOut,
		o.clientset.FS,
	)
	if err != nil {
		return err
	}

	scontext.SetComponentType(ctx, component.GetComponentTypeFromDevfileMetadata(devFileObj.Data.GetMetadata()))
	scontext.SetLanguage(ctx, devFileObj.Data.GetMetadata().Language)
	scontext.SetProjectType(ctx, devFileObj.Data.GetMetadata().ProjectType)
	scontext.SetDevfileName(ctx, devfileName)

	d := Handler{
		clientset:   *o.clientset,
		randomPorts: o.randomPortsFlag,
		errOut:      o.errOut,
	}
	err = o.clientset.DevClient.Watch(
		o.GetDevfilePath(),
		devFileObj,
		devfileName,
		path,
		o.ignorePaths,
		o.out,
		&d,
		o.ctx,
		o.debugFlag,
		o.buildCommandFlag,
		o.runCommandFlag,
		o.variables,
		o.randomPortsFlag,
		!o.noWatchFlag,
		o.errOut,
		componentStatus,
	)
	return err
}

// RegenerateAdapterAndPush regenerates the adapter and pushes the files to remote pod
func (o *Handler) RegenerateAdapterAndPush(pushParams adapters.PushParameters, watchParams watch.WatchParameters, componentStatus *watch.ComponentStatus) error {
	var adapter kcomponent.ComponentAdapter

	adapter, err := o.regenerateComponentAdapterFromWatchParams(watchParams)
	if err != nil {
		return fmt.Errorf("unable to generate component from watch parameters: %w", err)
	}

	err = adapter.Push(pushParams, componentStatus)
	if err != nil {
		return fmt.Errorf("watch command was unable to push component: %w", err)
	}

	return nil
}

func (o *Handler) regenerateComponentAdapterFromWatchParams(parameters watch.WatchParameters) (kcomponent.ComponentAdapter, error) {
	devObj, err := ododevfile.ParseAndValidateFromFileWithVariables(location.DevfileLocation(""), parameters.Variables)
	if err != nil {
		return nil, err
	}

	return kcomponent.NewKubernetesAdapter(
		o.clientset.KubernetesClient,
		o.clientset.PreferenceClient,
		o.clientset.PortForwardClient,
		o.clientset.BindingClient,
		kcomponent.AdapterContext{
			ComponentName: parameters.ComponentName,
			Context:       parameters.Path,
			AppName:       parameters.ApplicationName,
			Devfile:       devObj,
			FS:            o.clientset.FS,
		},
		o.clientset.KubernetesClient.GetCurrentNamespace(),
	), nil
}

func (o *DevOptions) HandleSignal() error {
	log.Section("Cleanup dev mode")
	log.Print("Stopping the dev environment ...")
	o.cancel()
	// At this point, `ctx.Done()` will be raised, and the cleanup will be done
	// wait for the cleanup to finish and let the main thread finish instead of signal handler go routine from runnable
	select {}
}

func (o *DevOptions) Cleanup(commandError error) {
	if commandError != nil {
		devFileObj := o.Context.EnvSpecificInfo.GetDevfileObj()
		componentName := o.GetComponentName()
		_ = o.clientset.WatchClient.CleanupDevResources(devFileObj, componentName, log.GetStdout())
	}
}

// NewCmdDev implements the odo dev command
func NewCmdDev(name, fullName string) *cobra.Command {
	o := NewDevOptions()
	devCmd := &cobra.Command{
		Use:   name,
		Short: "Deploy component to development cluster",
		Long: `odo dev is a long running command that will automatically sync your source to the cluster.
It forwards endpoints with exposure values 'public' or 'internal' to a port on localhost.`,
		Example: fmt.Sprintf(devExample, fullName),
		Args:    cobra.MaximumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}
	devCmd.Flags().BoolVar(&o.noWatchFlag, "no-watch", false, "Do not watch for file changes")
	devCmd.Flags().BoolVar(&o.randomPortsFlag, "random-ports", false, "Assign random ports to redirected ports")
	devCmd.Flags().BoolVar(&o.debugFlag, "debug", false, "Execute the debug command within the component")
	devCmd.Flags().StringArrayVar(&o.varsFlag, "var", []string{}, "Variable to override Devfile variable and variables in var-file")
	devCmd.Flags().StringVar(&o.varFileFlag, "var-file", "", "File containing variables to override Devfile variables")
	devCmd.Flags().StringVar(&o.buildCommandFlag, "build-command", "",
		"Alternative build command. The default one will be used if this flag is not set.")
	devCmd.Flags().StringVar(&o.runCommandFlag, "run-command", "",
		"Alternative run command to execute. The default one will be used if this flag is not set.")
	clientset.Add(devCmd,
		clientset.BINDING,
		clientset.DEV,
		clientset.FILESYSTEM,
		clientset.INIT,
		clientset.KUBERNETES,
		clientset.PORT_FORWARD,
		clientset.PREFERENCE,
		clientset.WATCH,
	)
	// Add a defined annotation in order to appear in the help menu
	devCmd.Annotations["command"] = "main"
	devCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)

	return devCmd
}
