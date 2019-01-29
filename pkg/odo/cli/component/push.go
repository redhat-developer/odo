package component

import (
	"fmt"

	"github.com/pkg/errors"
	appCmd "github.com/redhat-developer/odo/pkg/odo/cli/application"
	projectCmd "github.com/redhat-developer/odo/pkg/odo/cli/project"
	"github.com/redhat-developer/odo/pkg/odo/util/completion"

	"net/url"
	"os"
	"runtime"

	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"

	"github.com/fatih/color"
	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/util"

	"path/filepath"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
)

var pushCmdExample string = `  # Push source code to the current component
odo push

# Push data to the current component from the original source.
odo push

# Push source code in ~/mycode to component called my-component
odo push my-component --local ~/mycode
  `

// RecommendedPushCommandName is the recommended push command name
const RecommendedPushCommandName = "push"

// PushOptions encapsulates options that push command uses
type PushOptions struct {
	local         string
	componentName string
	sourceType    string
	sourcePath    string
	*genericclioptions.Context
}

// NewPushOptions returns new instance of PushOptions
func NewPushOptions() *PushOptions {
	return &PushOptions{}
}

// Complete completes push args
func (po *PushOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	po.Context = genericclioptions.NewContext(cmd)

	var argComponent string
	if len(args) == 1 {
		argComponent = args[0]
	}
	po.componentName = po.Context.Component(argComponent)

	po.sourceType, po.sourcePath, err = component.GetComponentSource(po.Context.Client, po.componentName, po.Context.Application)
	if err != nil {
		return errors.Wrapf(err, "unable to get component source")
	}

	if len(po.local) != 0 {
		po.sourcePath = util.GenFileURL(po.local, runtime.GOOS)
	}

	if po.sourceType == "binary" || po.sourceType == "local" {
		u, err := url.Parse(po.sourcePath)
		if err != nil {
			return errors.Wrapf(err, "unable to parse source %s from component %s", po.sourcePath, po.componentName)
		}

		if u.Scheme != "" && u.Scheme != "file" {
			return fmt.Errorf("Component %s has invalid source path %s", po.componentName, u.Scheme)
		}
		po.sourcePath = util.ReadFilePath(u, runtime.GOOS)
	}

	return
}

// Validate validates the push parameters
func (po *PushOptions) Validate() (err error) {
	// if the componentName is blank then there is no active component set
	if len(po.componentName) == 0 {
		return fmt.Errorf("No component is set as active. Use 'odo component set' to set an active component.")
	}

	// check if component name exists
	isExists, err := component.Exists(po.Context.Client, po.componentName, po.Context.Application)
	if err != nil {
		return err
	}
	if !isExists {
		return fmt.Errorf("component %s doesn't exist", po.componentName)
	}

	switch po.sourceType {
	case "binary":
		if len(po.local) != 0 {
			return fmt.Errorf("unable to push local directory:%s to component %s that uses binary %s.", po.local, po.componentName, po.sourcePath)
		}
	}

	if po.sourceType == "binary" || po.sourceType == "local" {
		_, err = os.Stat(po.sourcePath)
		if err != nil {
			return err
		}
	}

	return
}

// Run has the logic to perform the required actions as part of command
func (po *PushOptions) Run() (err error) {
	stdout := color.Output

	log.Namef("Pushing changes to component: %v", po.componentName)

	switch po.sourceType {
	case "local", "binary":
		// use value of '--dir' as source if it was used

		if po.sourceType == "local" {
			glog.V(4).Infof("Copying directory %s to pod", po.sourcePath)
			err = component.PushLocal(po.Context.Client, po.componentName, po.Context.Application, po.sourcePath, os.Stdout, []string{}, []string{}, true)
		} else {
			dir := filepath.Dir(po.sourcePath)
			glog.V(4).Infof("Copying file %s to pod", po.sourcePath)
			err = component.PushLocal(po.Context.Client, po.componentName, po.Context.Application, dir, os.Stdout, []string{po.sourcePath}, []string{}, true)
		}
		if err != nil {
			return errors.Wrapf(err, fmt.Sprintf("Failed to push component: %v", po.componentName))
		}

	case "git":
		// currently we don't support changing build type
		// it doesn't make sense to use --dir with git build
		if len(componentLocal) != 0 {
			log.Errorf("Unable to push local directory:%s to component %s that uses Git repository:%s.", po.local, po.componentName, po.sourcePath)
			os.Exit(1)
		}
		err := component.Build(po.Context.Client, po.componentName, po.Context.Application, true, stdout)
		return errors.Wrapf(err, fmt.Sprintf("failed to push component: %v", po.componentName))
	}

	log.Successf("Changes successfully pushed to component: %v", po.componentName)

	return
}

// NewCmdPush implements the push odo command
func NewCmdPush(name, fullName string) *cobra.Command {
	po := NewPushOptions()

	var pushCmd = &cobra.Command{
		Use:     fmt.Sprintf("%s [component name]", name),
		Short:   "Push source code to a component",
		Long:    `Push source code to a component.`,
		Example: pushCmdExample,
		Args:    cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			odoutil.LogErrorAndExit(po.Complete(name, cmd, args), "")
			odoutil.LogErrorAndExit(po.Validate(), "")
			odoutil.LogErrorAndExit(po.Run(), "")
		},
	}

	pushCmd.Flags().StringVarP(&po.local, "local", "l", "", "Use given local directory as a source for component. (It must be a local component)")

	// Add a defined annotation in order to appear in the help menu
	pushCmd.Annotations = map[string]string{"command": "component"}
	pushCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)
	completion.RegisterCommandHandler(pushCmd, completion.ComponentNameCompletionHandler)

	//Adding `--project` flag
	projectCmd.AddProjectFlag(pushCmd)
	//Adding `--application` flag
	appCmd.AddApplicationFlag(pushCmd)

	return pushCmd
}
