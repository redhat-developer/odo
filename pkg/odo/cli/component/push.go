package component

import (
	"fmt"
	"path/filepath"

	scontext "github.com/openshift/odo/pkg/segment/context"

	"github.com/openshift/odo/pkg/component"
	"github.com/openshift/odo/pkg/devfile/validate"
	"github.com/openshift/odo/pkg/envinfo"
	ktemplates "k8s.io/kubectl/pkg/util/templates"

	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/openshift/odo/pkg/devfile"
	"github.com/openshift/odo/pkg/occlient"
	projectCmd "github.com/openshift/odo/pkg/odo/cli/project"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/openshift/odo/pkg/odo/util/completion"
	"github.com/openshift/odo/pkg/util"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	odoutil "github.com/openshift/odo/pkg/odo/util"
	"k8s.io/klog"
)

var pushCmdExample = (`  # Push source code to the current component
%[1]s

# Push data to the current component from the original source
%[1]s

# Push source code in ~/mycode to component called my-component
%[1]s my-component --context ~/mycode

# Push source code with custom devfile commands using --build-command and --run-command
%[1]s --build-command="mybuild" --run-command="myrun"

# Output JSON events corresponding to devfile command execution and log text
%[1]s -o json
  `)

// PushRecommendedCommandName is the recommended push command name
const PushRecommendedCommandName = "push"

// PushOptions encapsulates options that push command uses
type PushOptions struct {
	*CommonPushOptions
	ignores    []string
	sourcePath string
	forceBuild bool

	// devfile path
	DevfilePath string
	Devfile     parser.DevfileObj

	// devfile commands
	devfileInitCommand  string
	devfileBuildCommand string
	devfileRunCommand   string
	devfileDebugCommand string
	debugRun            bool
}

// NewPushOptions returns new instance of PushOptions
// with "default" values for certain values, for example, show is "false"
func NewPushOptions() *PushOptions {
	return &PushOptions{
		CommonPushOptions: NewCommonPushOptions(),
	}
}

// CompleteDevfilePath completes the devfile path from context
func (po *PushOptions) CompleteDevfilePath() {
	if len(po.DevfilePath) > 0 {
		po.DevfilePath = filepath.Join(po.componentContext, po.DevfilePath)
	} else {
		po.DevfilePath = filepath.Join(po.componentContext, devfile.DevfileFilenamesProvider(po.componentContext))
	}
}

// GetComponentContext gets the component context
func (po *PushOptions) GetComponentContext() string {
	return po.componentContext
}

// Complete completes push args
func (po *PushOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	po.CompleteDevfilePath()
	devfileExists := util.CheckPathExists(po.DevfilePath)

	if !devfileExists {
		return fmt.Errorf("the current direcotry doesn't contain a devfile")
	}

	po.Devfile, err = devfile.ParseFromFile(po.DevfilePath)
	if err != nil {
		return errors.Wrap(err, "unable to parse devfile")
	}
	err = validate.ValidateDevfileData(po.Devfile.Data)
	if err != nil {
		return err
	}

	// We retrieve the configuration information. If this does not exist, then BLANK is returned (important!).
	envFileInfo, err := envinfo.NewEnvSpecificInfo(po.componentContext)
	if err != nil {
		return errors.Wrap(err, "unable to retrieve configuration information")
	}

	// If the file does not exist, we should populate the environment file with the correct env.yaml information
	// such as name and namespace.
	if !envFileInfo.Exists() {
		klog.V(4).Info("Environment file does not exist, creating the env.yaml file in order to use 'odo push'")

		// Since the environment file does not exist, we will retrieve a correct namespace from
		// either cmd commands or the current default kubernetes namespace
		namespace, err := retrieveCmdNamespace(cmd)
		if err != nil {
			return errors.Wrap(err, "unable to determine target namespace for the component")
		}
		client, err := genericclioptions.Client()
		if err != nil {
			return err
		}
		if err := checkDefaultProject(client, namespace); err != nil {
			return err
		}

		// Retrieve a default name
		// 1. Use args[0] if the user has supplied a name to be used
		// 2. If the user did not provide a name, use gatherName to retrieve a name from the devfile.Metadata
		// 3. Use the folder name that we are pushing from as a default name if none of the above exist

		var name string
		if len(args) == 1 {
			name = args[0]
		} else {
			name, err = gatherName(po.Devfile, po.DevfilePath)
			if err != nil {
				return errors.Wrap(err, "unable to gather a name to apply to the env.yaml file")
			}
		}

		// Create the environment file. This will actually *create* the env.yaml file in your context directory.
		err = envFileInfo.SetComponentSettings(envinfo.ComponentSettings{Name: name, Project: namespace, AppName: "app"})
		if err != nil {
			return errors.Wrap(err, "failed to create env.yaml for devfile component")
		}

	} else if envFileInfo.GetNamespace() == "" {
		// Since the project name doesn't exist in the environment file, we will retrieve a correct namespace from
		// either cmd commands or the current default kubernetes namespace
		// and write it to the env.yaml
		namespace, err := retrieveCmdNamespace(cmd)
		if err != nil {
			return errors.Wrap(err, "unable to determine target namespace for devfile")
		}
		client, err := genericclioptions.Client()
		if err != nil {
			return err
		}

		if err := checkDefaultProject(client, namespace); err != nil {
			return err
		}

		err = envFileInfo.SetConfiguration("project", namespace)
		if err != nil {
			return errors.Wrap(err, "failed to write the project to the env.yaml for devfile component")
		}
	} else if envFileInfo.GetNamespace() == "default" {
		client, err := genericclioptions.Client()
		if err != nil {
			return err
		}
		if err := checkDefaultProject(client, envFileInfo.GetNamespace()); err != nil {
			return err
		}
	}

	if envFileInfo.GetApplication() == "" {
		err = envFileInfo.SetConfiguration("app", "")
		if err != nil {
			return errors.Wrap(err, "failed to write the app to the env.yaml for devfile component")
		}
	}

	po.EnvSpecificInfo = envFileInfo

	po.Context, err = genericclioptions.NewContext(cmd)
	if err != nil {
		return err
	}

	// set the telemetry data
	cmdCtx := cmd.Context()
	devfileMetadata := po.Devfile.Data.GetMetadata()
	scontext.SetClusterType(cmdCtx, po.Client)
	scontext.SetComponentType(cmdCtx, component.GetComponentTypeFromDevfileMetadata(devfileMetadata))
	scontext.SetLanguage(cmdCtx, component.GetLanguageFromDevfileMetadata(devfileMetadata))
	scontext.SetProjectType(cmdCtx, component.GetProjectTypeFromDevfileMetadata(devfileMetadata))

	return nil
}

// Validate validates the push parameters
func (po *PushOptions) Validate() (err error) {
	return nil
}

// Run has the logic to perform the required actions as part of command
func (po *PushOptions) Run(cmd *cobra.Command) (err error) {
	// Return Devfile push
	return po.DevfilePush()
}

// NewCmdPush implements the push odo command
func NewCmdPush(name, fullName string) *cobra.Command {
	po := NewPushOptions()

	var pushCmd = &cobra.Command{
		Use:         fmt.Sprintf("%s [component name]", name),
		Short:       "Push source code to a component",
		Long:        `Push source code to a component.`,
		Example:     fmt.Sprintf(ktemplates.Examples(pushCmdExample), fullName),
		Args:        cobra.MaximumNArgs(2),
		Annotations: map[string]string{"machineoutput": "json", "command": "component"},
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(po, cmd, args)
		},
	}

	genericclioptions.AddContextFlag(pushCmd, &po.componentContext)
	pushCmd.Flags().BoolVar(&po.show, "show-log", false, "If enabled, logs will be shown when built")
	pushCmd.Flags().StringSliceVar(&po.ignores, "ignore", []string{}, "Files or folders to be ignored via glob expressions.")
	pushCmd.Flags().BoolVar(&po.pushConfig, "config", false, "Use config flag to only apply config on to cluster")
	pushCmd.Flags().BoolVar(&po.pushSource, "source", false, "Use source flag to only push latest source on to cluster")
	pushCmd.Flags().BoolVarP(&po.forceBuild, "force-build", "f", false, "Use force-build flag to re-sync the entire source code and re-build the component")

	pushCmd.Flags().StringVar(&po.devfileInitCommand, "init-command", "", "Devfile Init Command to execute")
	pushCmd.Flags().StringVar(&po.devfileBuildCommand, "build-command", "", "Devfile Build Command to execute")
	pushCmd.Flags().StringVar(&po.devfileRunCommand, "run-command", "", "Devfile Run Command to execute")
	pushCmd.Flags().BoolVar(&po.debugRun, "debug", false, "Runs the component in debug mode")
	pushCmd.Flags().StringVar(&po.devfileDebugCommand, "debug-command", "", "Devfile Debug Command to execute")

	//Adding `--project` flag
	projectCmd.AddProjectFlag(pushCmd)

	pushCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)
	completion.RegisterCommandHandler(pushCmd, completion.ComponentNameCompletionHandler)
	completion.RegisterCommandFlagHandler(pushCmd, "context", completion.FileCompletionHandler)

	return pushCmd
}

// checkDefaultProject errors out if the project resource is supported and the value is "default"
func checkDefaultProject(client *occlient.Client, name string) error {
	// Check whether resource "Project" is supported
	projectSupported, err := client.IsProjectSupported()

	if err != nil {
		return errors.Wrap(err, "resource project validation check failed.")
	}

	if projectSupported && name == "default" {
		return errors.New("odo may not work as expected in the default project, please run the odo component in a non-default project")
	}
	return nil
}
