package component

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/openshift/odo/pkg/component"
	"github.com/openshift/odo/pkg/config"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/occlient"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/openshift/odo/pkg/odo/util/completion"
	"github.com/openshift/odo/pkg/project"
	"github.com/openshift/odo/pkg/util"

	odoutil "github.com/openshift/odo/pkg/odo/util"

	ktemplates "k8s.io/kubernetes/pkg/kubectl/cmd/templates"
)

var pushCmdExample = ktemplates.Examples(`  # Push source code to the current component
%[1]s

# Push data to the current component from the original source.
%[1]s

# Push source code in ~/mycode to component called my-component
%[1]s my-component --context ~/mycode
  `)

// PushRecommendedCommandName is the recommended push command name
const PushRecommendedCommandName = "push"

// PushOptions encapsulates options that push command uses
type PushOptions struct {
	ignores []string
	show    bool

	sourceType       config.SrcType
	sourcePath       string
	componentContext string
	client           *occlient.Client
	localConfig      *config.LocalConfigInfo

	pushConfig bool
	pushSource bool

	*genericclioptions.Context
}

// NewPushOptions returns new instance of PushOptions
// with "default" values for certain values, for example, show is "false"
func NewPushOptions() *PushOptions {
	return &PushOptions{
		show: false,
	}
}

// Complete completes push args
func (po *PushOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	po.resolveSrcAndConfigFlags()

	conf, err := config.NewLocalConfigInfo(po.componentContext)
	if err != nil {
		return errors.Wrap(err, "unable to retrieve configuration information")
	}

	// Set the necessary values within WatchOptions
	po.localConfig = conf
	po.sourceType = conf.LocalConfig.GetSourceType()

	glog.V(4).Infof("SourceLocation: %s", po.localConfig.GetSourceLocation())

	// Get SourceLocation here...
	po.sourcePath, err = conf.GetOSSourcePath()
	if err != nil {
		return errors.Wrap(err, "unable to retrieve absolute path to source location")
	}

	glog.V(4).Infof("Source Path: %s", po.sourcePath)

	// Apply ignore information
	err = genericclioptions.ApplyIgnore(&po.ignores, po.sourcePath)
	if err != nil {
		return errors.Wrap(err, "unable to apply ignore information")
	}

	// Set the correct context
	po.Context = genericclioptions.NewContextCreatingAppIfNeeded(cmd)

	// check if project exist
	prjName := po.localConfig.GetProject()
	isPrjExists, err := project.Exists(po.Context.Client, prjName)
	if err != nil {
		return errors.Wrapf(err, "failed to check if project with name %s exists", prjName)
	}
	if !isPrjExists {
		log.Successf("Creating project %s", prjName)
		err = project.Create(po.Context.Client, prjName, true)
		if err != nil {
			log.Errorf("Failed creating project %s", prjName)
			return errors.Wrapf(
				err,
				"project %s does not exist. Failed creating it.Please try after creating project using `odo project create <project_name>`",
				prjName,
			)
		}
		log.Successf("Successfully created project %s", prjName)
	}
	po.Context.Client.Namespace = prjName
	return
}

// Validate validates the push parameters
func (po *PushOptions) Validate() (err error) {
	if err = component.ValidateComponentCreateRequest(po.Context.Client, po.localConfig.GetComponentSettings(), false); err != nil {
		return err
	}

	isCmpExists, err := component.Exists(po.Context.Client, po.localConfig.GetName(), po.localConfig.GetApplication())
	if err != nil {
		return err
	}

	if !isCmpExists && po.pushSource && !po.pushConfig {
		return fmt.Errorf("Component %s does not exist and hence cannot push only source. Please use `odo push` without any flags or with both `--source` and `--config` flags", po.localConfig.GetName())
	}

	return nil
}

func (po *PushOptions) createCmpIfNotExistsAndApplyCmpConfig(stdout io.Writer) error {
	if !po.pushConfig {
		// Not the case of component creation or updation(with new config)
		// So nothing to do here and hence return from here
		return nil
	}

	cmpName := po.localConfig.GetName()
	appName := po.localConfig.GetApplication()
	cmpType := po.localConfig.GetType()

	isCmpExists, err := component.Exists(po.Context.Client, cmpName, appName)
	if err != nil {
		return errors.Wrapf(err, "failed to check if component %s exists or not", cmpName)
	}

	if !isCmpExists {
		log.Successf("Creating %s component with name %s", cmpType, cmpName)
		// Classic case of component creation
		if err = component.CreateComponent(po.Context.Client, *po.localConfig, po.componentContext, stdout); err != nil {
			log.Errorf(
				"Failed to create component with name %s. Please use `odo config view` to view settings used to create component. Error: %+v",
				cmpName,
				err,
			)
			os.Exit(1)
		}
		log.Successf("Successfully created component %s", cmpName)
	}

	// Apply config
	err = component.ApplyConfig(po.Context.Client, *po.localConfig, stdout, isCmpExists)
	if err != nil {
		odoutil.LogErrorAndExit(err, "Failed to update config to component deployed")
	}
	log.Successf("Successfully updated component with name: %v", cmpName)

	return nil
}

// Run has the logic to perform the required actions as part of command
func (po *PushOptions) Run() (err error) {
	stdout := color.Output

	cmpName := po.localConfig.GetName()
	appName := po.localConfig.GetApplication()

	err = po.createCmpIfNotExistsAndApplyCmpConfig(stdout)
	if err != nil {
		return
	}

	if !po.pushSource {
		// If source is not requested for update, return
		return nil
	}

	log.Successf("Pushing changes to component: %v of type %s", cmpName, po.sourceType)

	// Get SourceLocation here...
	po.sourcePath, err = po.localConfig.GetOSSourcePath()
	if err != nil {
		return errors.Wrap(err, "unable to retrieve OS source path to source location")
	}

	switch po.sourceType {
	case config.LOCAL:
		glog.V(4).Infof("Copying directory %s to pod", po.sourcePath)
		err = component.PushLocal(
			po.Context.Client,
			cmpName,
			appName,
			po.sourcePath,
			os.Stdout,
			[]string{},
			[]string{},
			true,
			util.GetAbsGlobExps(po.sourcePath, po.ignores),
			po.show,
		)

		if err != nil {
			return errors.Wrapf(err, fmt.Sprintf("Failed to push component: %v", cmpName))
		}

	case config.BINARY:

		// We will pass in the directory, NOT filepath since this is a binary..
		binaryDirectory := filepath.Dir(po.sourcePath)

		glog.V(4).Infof("Copying binary file %s to pod", po.sourcePath)
		err = component.PushLocal(
			po.Context.Client,
			cmpName,
			appName,
			binaryDirectory,
			os.Stdout,
			[]string{po.sourcePath},
			[]string{},
			true,
			util.GetAbsGlobExps(po.sourcePath, po.ignores),
			po.show,
		)

		if err != nil {
			return errors.Wrapf(err, fmt.Sprintf("Failed to push component: %v", cmpName))
		}

	case config.GIT:
		err := component.Build(
			po.Context.Client,
			cmpName,
			appName,
			true,
			stdout,
			po.show,
		)
		return errors.Wrapf(err, fmt.Sprintf("failed to push component: %v", cmpName))
	}

	log.Successf("Changes successfully pushed to component: %v", cmpName)
	return
}

func (po *PushOptions) resolveSrcAndConfigFlags() {
	// If neither config nor source flag is passed, update both config and source to the component
	if !po.pushConfig && !po.pushSource {
		po.pushConfig = true
		po.pushSource = true
	}
}

// NewCmdPush implements the push odo command
func NewCmdPush(name, fullName string) *cobra.Command {
	po := NewPushOptions()

	var pushCmd = &cobra.Command{
		Use:     fmt.Sprintf("%s [component name]", name),
		Short:   "Push source code to a component",
		Long:    `Push source code to a component.`,
		Example: fmt.Sprintf(pushCmdExample, fullName),
		Args:    cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(po, cmd, args)
		},
	}
	genericclioptions.AddContextFlag(pushCmd, &po.componentContext)
	pushCmd.Flags().BoolVar(&po.show, "show-log", false, "If enabled, logs will be shown when built")
	pushCmd.Flags().StringSliceVar(&po.ignores, "ignore", []string{}, "Files or folders to be ignored via glob expressions.")
	pushCmd.Flags().BoolVar(&po.pushConfig, "config", false, "Use config flag to only apply config on to cluster")
	pushCmd.Flags().BoolVar(&po.pushSource, "source", false, "Use source flag to only push latest source on to cluster")

	// Add a defined annotation in order to appear in the help menu
	pushCmd.Annotations = map[string]string{"command": "component"}
	pushCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)
	completion.RegisterCommandHandler(pushCmd, completion.ComponentNameCompletionHandler)
	completion.RegisterCommandFlagHandler(pushCmd, "context", completion.FileCompletionHandler)

	return pushCmd
}
