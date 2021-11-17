package component

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/openshift/odo/pkg/util"

	"github.com/spf13/cobra"
	"k8s.io/klog"

	"github.com/openshift/odo/pkg/devfile"
	"github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/log"
	appCmd "github.com/openshift/odo/pkg/odo/cli/application"
	projectCmd "github.com/openshift/odo/pkg/odo/cli/project"
	"github.com/openshift/odo/pkg/odo/cli/ui"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	odoutil "github.com/openshift/odo/pkg/odo/util"
	"github.com/openshift/odo/pkg/odo/util/completion"
	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

// DeleteRecommendedCommandName is the recommended delete command name
const DeleteRecommendedCommandName = "delete"

var deleteExample = ktemplates.Examples(`
# Delete the component present in the current directory from the cluster
%[1]s

# Delete the component named 'frontend' from the cluster
%[1]s frontend

# Delete the component present in the current directory from the cluster and all of its related local config files("devfile.yaml" and ".odo" directory)
%[1]s --all

# Delete the component present in the './frontend' directory from the cluster
%[1]s --context ./frontend

# Delete the component present in the './frontend' directory from the cluster and all of its related local config files("devfile.yaml" and ".odo" directory)
%[1]s --context ./frontend --all

# Delete the component 'frontend' that is a part of 'myapp' app inside the 'myproject' project from the cluster	
%[1]s frontend --app myapp --project myproject`)

// DeleteOptions is a container to attach complete, validate and run pattern
type DeleteOptions struct {
	componentForceDeleteFlag bool
	componentDeleteAllFlag   bool
	componentDeleteWaitFlag  bool
	componentContext         string
	isCmpExists              bool
	*ComponentOptions

	// devfile path
	show bool
}

// NewDeleteOptions returns new instance of DeleteOptions
func NewDeleteOptions() *DeleteOptions {
	return &DeleteOptions{
		componentForceDeleteFlag: false,
		componentDeleteAllFlag:   false,
		componentDeleteWaitFlag:  false,
		componentContext:         "",
		isCmpExists:              false,
		ComponentOptions:         &ComponentOptions{},
		show:                     false,
	}
}

// Complete completes log args
func (do *DeleteOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	do.Context, err = genericclioptions.New(genericclioptions.NewCreateParameters(cmd).NeedDevfile(do.componentContext))
	return err
}

// Validate validates the list parameters
func (do *DeleteOptions) Validate() (err error) {
	return

}

// Run has the logic to perform the required actions as part of command
func (do *DeleteOptions) Run(cmd *cobra.Command) (err error) {
	klog.V(4).Infof("component delete called")
	klog.V(4).Infof("args: %#v", do)
	if do.componentForceDeleteFlag || ui.Proceed(fmt.Sprintf("Are you sure you want to delete the devfile component: %s?", do.EnvSpecificInfo.GetName())) {
		err = do.DevfileComponentDelete()
		if err != nil {
			log.Errorf("error occurred while deleting component, cause: %v", err)
		}
	} else {
		log.Error("Aborting deletion of component")
	}

	if do.componentDeleteAllFlag {
		log.Info("\nDeleting local config")
		// Prompt and delete env folder
		if do.componentForceDeleteFlag || ui.Proceed("Are you sure you want to delete env folder?") {
			if !do.EnvSpecificInfo.Exists() {
				return fmt.Errorf("env folder doesn't exist for the component")
			}
			if err = util.DeleteIndexFile(filepath.Dir(do.GetDevfilePath())); err != nil {
				return err
			}

			err = do.EnvSpecificInfo.DeleteEnvInfoFile()
			if err != nil {
				return err
			}
			err = do.EnvSpecificInfo.DeleteEnvDirIfEmpty()
			if err != nil {
				return err
			}
			err = util.DeletePath(filepath.Join(do.componentContext, util.DotOdoDirectory))
			if err != nil {
				return err
			}
			log.Successf("Successfully deleted env file")
		} else {
			log.Error("Aborting deletion of env folder")
		}

		if do.componentForceDeleteFlag {
			if !util.CheckPathExists(do.GetDevfilePath()) {
				return fmt.Errorf("devfile.yaml does not exist in the current directory")
			}
			if !do.EnvSpecificInfo.IsUserCreatedDevfile() {

				// first remove the uri based files mentioned in the devfile
				devfileObj, err := devfile.ParseAndValidateFromFile(do.GetDevfilePath())
				if err != nil {
					return err
				}

				err = common.RemoveDevfileURIContents(devfileObj, do.componentContext)
				if err != nil {
					return err
				}

				empty, err := util.IsEmpty(filepath.Join(do.componentContext, devfile.UriFolder))
				if err != nil && !os.IsNotExist(err) {
					return err
				}

				if !os.IsNotExist(err) && empty {
					err = os.RemoveAll(filepath.Join(do.componentContext, devfile.UriFolder))
					if err != nil {
						return err
					}
				}

				err = util.DeletePath(do.GetDevfilePath())
				if err != nil {
					return err
				}

				log.Successf("Successfully deleted devfile.yaml file")

			} else {
				log.Info("Didn't delete the devfile as it was user provided")
			}

		} else if ui.Proceed("Are you sure you want to delete devfile.yaml?") {
			if !util.CheckPathExists(do.GetDevfilePath()) {
				return fmt.Errorf("devfile.yaml does not exist in the current directory")
			}

			// first remove the uri based files mentioned in the devfile
			devfileObj, err := devfile.ParseAndValidateFromFile(do.GetDevfilePath())
			if err != nil {
				return err
			}

			err = common.RemoveDevfileURIContents(devfileObj, do.componentContext)
			if err != nil {
				return err
			}

			empty, err := util.IsEmpty(filepath.Join(do.componentContext, devfile.UriFolder))
			if err != nil && !os.IsNotExist(err) {
				return err
			}

			if !os.IsNotExist(err) && empty {
				err = os.RemoveAll(filepath.Join(do.componentContext, devfile.UriFolder))
				if err != nil {
					return err
				}
			}

			err = util.DeletePath(do.GetDevfilePath())
			if err != nil {
				return err
			}

			log.Successf("Successfully deleted devfile.yaml file")
		} else {
			log.Error("Aborting deletion of devfile.yaml file")
		}
	}

	return nil
}

// NewCmdDelete implements the delete odo command
func NewCmdDelete(name, fullName string) *cobra.Command {

	do := NewDeleteOptions()

	var componentDeleteCmd = &cobra.Command{
		Use:         fmt.Sprintf("%s <component_name>", name),
		Short:       "Delete component",
		Long:        "Delete component.",
		Example:     fmt.Sprintf(deleteExample, fullName),
		Args:        cobra.MaximumNArgs(1),
		Annotations: map[string]string{"command": "component"},
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(do, cmd, args)
		},
	}

	componentDeleteCmd.Flags().BoolVarP(&do.componentForceDeleteFlag, "force", "f", false, "Delete component without prompting")
	componentDeleteCmd.Flags().BoolVarP(&do.componentDeleteAllFlag, "all", "a", false, "Delete component and local config")
	componentDeleteCmd.Flags().BoolVarP(&do.componentDeleteWaitFlag, "wait", "w", false, "Wait for complete deletion of component and its dependent")

	componentDeleteCmd.Flags().BoolVar(&do.show, "show-log", false, "If enabled, logs will be shown when deleted")

	componentDeleteCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)
	completion.RegisterCommandHandler(componentDeleteCmd, completion.ComponentNameCompletionHandler)
	//Adding `--context` flag
	genericclioptions.AddContextFlag(componentDeleteCmd, &do.componentContext)

	//Adding `--project` flag
	projectCmd.AddProjectFlag(componentDeleteCmd)
	//Adding `--application` flag
	appCmd.AddApplicationFlag(componentDeleteCmd)

	return componentDeleteCmd
}
