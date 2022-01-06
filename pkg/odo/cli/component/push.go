package component

import (
	"fmt"
	"path/filepath"

	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/preference"
	"github.com/redhat-developer/odo/pkg/project"
	scontext "github.com/redhat-developer/odo/pkg/segment/context"

	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/devfile/location"
	"github.com/redhat-developer/odo/pkg/devfile/validate"
	"github.com/redhat-developer/odo/pkg/envinfo"
	ktemplates "k8s.io/kubectl/pkg/util/templates"

	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/pkg/errors"
	"github.com/redhat-developer/odo/pkg/devfile"
	projectCmd "github.com/redhat-developer/odo/pkg/odo/cli/project"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/util/completion"
	"github.com/redhat-developer/odo/pkg/util"
	"github.com/spf13/cobra"

	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
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
	// Push context
	*CommonPushOptions

	// Flags
	ignoreFlag     []string
	forceBuildFlag bool
	debugFlag      bool

	// devfile commands flags
	initCommandFlag  string
	buildCommandFlag string
	runCommandflag   string
	debugCommandFlag string

	sourcePath string

	// devfile path
	DevfilePath string

	// Devfile content
	Devfile parser.DevfileObj
}

// NewPushOptions returns new instance of PushOptions
// with "default" values for certain values, for example, show is "false"

func NewPushOptions(client component.Client, prjClient project.Client, prefClient preference.Client) *PushOptions {
	return &PushOptions{
		CommonPushOptions: NewCommonPushOptions(prjClient, prefClient, client),
	}
}

// CompleteDevfilePath completes the devfile path from context
func (po *PushOptions) CompleteDevfilePath() {
	if len(po.DevfilePath) > 0 {
		po.DevfilePath = filepath.Join(po.componentContext, po.DevfilePath)
	} else {
		po.DevfilePath = filepath.Join(po.componentContext, location.DevfileFilenamesProvider(po.componentContext))
	}
}

// GetComponentContext gets the component context
func (po *PushOptions) GetComponentContext() string {
	return po.componentContext
}

// Complete completes push args
func (po *PushOptions) Complete(cmdline cmdline.Cmdline, args []string) (err error) {
	po.CompleteDevfilePath()
	devfileExists := util.CheckPathExists(po.DevfilePath)

	if !devfileExists {
		return fmt.Errorf("the current direcotry doesn't contain a devfile")
	}

	po.Devfile, err = devfile.ParseAndValidateFromFile(po.DevfilePath)
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

	err = po.setupEnvFile(envFileInfo, cmdline, args)
	if err != nil {
		return err
	}

	po.EnvSpecificInfo = envFileInfo

	po.Context, err = genericclioptions.New(genericclioptions.NewCreateParameters(cmdline))
	if err != nil {
		return err
	}

	// set the telemetry data
	cmdCtx := cmdline.Context()
	devfileMetadata := po.Devfile.Data.GetMetadata()
	scontext.SetClusterType(cmdCtx, po.KClient)
	scontext.SetComponentType(cmdCtx, component.GetComponentTypeFromDevfileMetadata(devfileMetadata))
	scontext.SetLanguage(cmdCtx, component.GetLanguageFromDevfileMetadata(devfileMetadata))
	scontext.SetProjectType(cmdCtx, component.GetProjectTypeFromDevfileMetadata(devfileMetadata))

	return nil
}

func (po *PushOptions) setupEnvFile(envFileInfo *envinfo.EnvSpecificInfo, cmdline cmdline.Cmdline, args []string) error {
	// If the file does not exist, we should populate the environment file with the correct env.yaml information
	// such as name and namespace.
	if !envFileInfo.Exists() {
		klog.V(4).Info("Environment file does not exist, creating the env.yaml file in order to use 'odo push'")

		// Since the environment file does not exist, we will retrieve a correct namespace from
		// either cmd commands or the current default kubernetes namespace
		namespace, err := retrieveCmdNamespace(cmdline)
		if err != nil {
			return errors.Wrap(err, "unable to determine target namespace for the component")
		}

		if err = po.componentClient.CheckDefaultProject(namespace); err != nil {
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
			name, err = GatherName(po.Devfile, po.DevfilePath)
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
		namespace, err := retrieveCmdNamespace(cmdline)
		if err != nil {
			return errors.Wrap(err, "unable to determine target namespace for devfile")
		}
		if err = po.componentClient.CheckDefaultProject(namespace); err != nil {
			return err
		}

		err = envFileInfo.SetConfiguration("project", namespace)
		if err != nil {
			return errors.Wrap(err, "failed to write the project to the env.yaml for devfile component")
		}
	} else if envFileInfo.GetNamespace() == "default" {
		if err := po.componentClient.CheckDefaultProject(envFileInfo.GetNamespace()); err != nil {
			return err
		}
	}

	if envFileInfo.GetApplication() == "" {
		err := envFileInfo.SetConfiguration("app", "")
		if err != nil {
			return errors.Wrap(err, "failed to write the app to the env.yaml for devfile component")
		}
	}
	return nil
}

// Validate validates the push parameters
func (po *PushOptions) Validate() (err error) {
	return nil
}

// Run has the logic to perform the required actions as part of command
func (po *PushOptions) Run() (err error) {
	// Return Devfile push
	return po.DevfilePush()
}

// NewCmdPush implements the push odo command
func NewCmdPush(name, fullName string) *cobra.Command {
	// The error is not handled at this point, it will be handled during Context creation
	kubclient, _ := kclient.New()
	prefClient, err := preference.NewClient()
	if err != nil {
		odoutil.LogErrorAndExit(err, "unable to set preference, something is wrong with odo, kindly raise an issue at https://github.com/redhat-developer/odo/issues/new?template=Bug.md")
	}
	po := NewPushOptions(component.NewClient(kubclient), project.NewClient(kubclient), prefClient)

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

	odoutil.AddContextFlag(pushCmd, &po.componentContext)
	pushCmd.Flags().BoolVar(&po.showFlag, "show-log", false, "If enabled, logs will be shown when built")
	pushCmd.Flags().StringSliceVar(&po.ignoreFlag, "ignore", []string{}, "Files or folders to be ignored via glob expressions.")
	pushCmd.Flags().BoolVar(&po.configFlag, "config", false, "Use config flag to only apply config on to cluster")
	pushCmd.Flags().BoolVar(&po.sourceFlag, "source", false, "Use source flag to only push latest source on to cluster")
	pushCmd.Flags().BoolVarP(&po.forceBuildFlag, "force-build", "f", false, "Use force-build flag to re-sync the entire source code and re-build the component")

	pushCmd.Flags().StringVar(&po.initCommandFlag, "init-command", "", "Devfile Init Command to execute")
	pushCmd.Flags().StringVar(&po.buildCommandFlag, "build-command", "", "Devfile Build Command to execute")
	pushCmd.Flags().StringVar(&po.runCommandflag, "run-command", "", "Devfile Run Command to execute")
	pushCmd.Flags().BoolVar(&po.debugFlag, "debug", false, "Runs the component in debug mode")
	pushCmd.Flags().StringVar(&po.debugCommandFlag, "debug-command", "", "Devfile Debug Command to execute")

	// Adding `--project` flag
	projectCmd.AddProjectFlag(pushCmd)

	pushCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)
	completion.RegisterCommandHandler(pushCmd, completion.ComponentNameCompletionHandler)
	completion.RegisterCommandFlagHandler(pushCmd, "context", completion.FileCompletionHandler)

	return pushCmd
}
