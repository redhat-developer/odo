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

	options genericclioptions.ComponentOptions

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
	po.options.Client = genericclioptions.Client(cmd)

	// Retrieve configuration configuration as well as SourcePath
	conf, err := genericclioptions.RetrieveLocalConfigInfo(po.options.ComponentContext)
	if err != nil {
		return errors.Wrap(err, "unable to retrieve config information")
	}

	// Set the necessary values within WatchOptions
	po.options.LocalConfig = conf.LocalConfig
	po.options.SourcePath = conf.SourcePath
	po.options.SourceType = conf.LocalConfig.GetSourceType()

	// Apply ignore information
	err = genericclioptions.ApplyIgnore(&po.ignores, po.options.SourcePath)
	if err != nil {
		return errors.Wrap(err, "unable to apply ignore information")
	}

	// Set the correct context
	po.Context = genericclioptions.NewContextCreatingAppIfNeeded(cmd)
	return
}

// Validate validates the push parameters
func (po *PushOptions) Validate() (err error) {
	return component.ValidateComponentCreateRequest(po.Context.Client, po.options.LocalConfig.GetComponentSettings(), false)
}

func (po *PushOptions) createCmpIfNotExistsAndApplyCmpConfig(stdout io.Writer) error {
	cmpName := po.options.LocalConfig.GetName()
	prjName := po.options.LocalConfig.GetProject()
	appName := po.options.LocalConfig.GetApplication()
	cmpType := po.options.LocalConfig.GetType()

	isPrjExists, err := project.Exists(po.options.Client, prjName)
	if err != nil {
		return errors.Wrapf(err, "failed to check if project with name %s exists", prjName)
	}
	if !isPrjExists {
		log.Successf("Creating project %s", prjName)
		err = project.Create(po.options.Client, prjName, true)
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
	po.options.Client.Namespace = prjName

	isCmpExists, err := component.Exists(po.options.Client, cmpName, appName)
	if err != nil {
		return errors.Wrapf(err, "failed to check if component %s exists or not", cmpName)
	}

	if !isCmpExists {
		log.Successf("Creating %s component with name %s", cmpType, cmpName)
		// Classic case of component creation
		if err = component.CreateComponent(po.options.Client, *po.options.LocalConfig, po.options.ComponentContext, stdout); err != nil {
			log.Errorf(
				"Failed to create component with name %s. Please use `odo config view` to view settings used to create component. Error: %+v",
				cmpName,
				err,
			)
			os.Exit(1)
		}
		log.Successf("Successfully created component %s", cmpName)
	} else {
		log.Successf("Applying component settings to component: %v", cmpName)
		// Apply config
		err = component.ApplyConfig(po.options.Client, *po.options.LocalConfig, po.options.ComponentContext, stdout)
		if err != nil {
			log.Errorf("Failed to update config to component deployed. Error %+v", err)
			os.Exit(1)
		}
		log.Successf("Successfully created component with name: %v", cmpName)
	}
	return nil
}

// Run has the logic to perform the required actions as part of command
func (po *PushOptions) Run() (err error) {
	stdout := color.Output

	cmpName := po.options.LocalConfig.GetName()
	appName := po.options.LocalConfig.GetApplication()

	err = po.createCmpIfNotExistsAndApplyCmpConfig(stdout)
	if err != nil {
		return
	}

	log.Successf("Pushing changes to component: %v of type %s", cmpName, po.options.SourceType)

	switch po.options.SourceType {
	case config.LOCAL, config.BINARY:

		if po.options.SourceType == config.LOCAL {
			glog.V(4).Infof("Copying directory %s to pod", po.options.SourcePath)
			err = component.PushLocal(
				po.options.Client,
				cmpName,
				appName,
				po.options.LocalConfig.GetSourceLocation(),
				os.Stdout,
				[]string{},
				[]string{},
				true,
				util.GetAbsGlobExps(po.options.SourcePath, po.ignores),
				po.show,
			)
		} else {
			dir := filepath.Dir(po.options.SourcePath)
			glog.V(4).Infof("Copying file %s to pod", po.options.SourcePath)
			err = component.PushLocal(
				po.options.Client,
				cmpName,
				appName,
				dir,
				os.Stdout,
				[]string{po.options.SourcePath},
				[]string{},
				true,
				util.GetAbsGlobExps(po.options.SourcePath, po.ignores),
				po.show,
			)
		}
		if err != nil {
			return errors.Wrapf(err, fmt.Sprintf("Failed to push component: %v", cmpName))
		}

	case "git":
		err := component.Build(
			po.options.Client,
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

	pushCmd.Flags().StringVarP(&po.options.ComponentContext, "context", "c", "", "Use given context directory as a source for component settings")
	pushCmd.Flags().BoolVar(&po.show, "show-log", false, "If enabled, logs will be shown when built")
	pushCmd.Flags().StringSliceVar(&po.ignores, "ignore", []string{}, "Files or folders to be ignored via glob expressions.")

	// Add a defined annotation in order to appear in the help menu
	pushCmd.Annotations = map[string]string{"command": "component"}
	pushCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)
	completion.RegisterCommandHandler(pushCmd, completion.ComponentNameCompletionHandler)

	return pushCmd
}
