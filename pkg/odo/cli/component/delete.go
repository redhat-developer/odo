package component

import (
	"errors"
	"fmt"
	"github.com/redhat-developer/odo/pkg/devfile/adapters/kubernetes/component"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"

	"github.com/redhat-developer/odo/pkg/devfile"
	"github.com/redhat-developer/odo/pkg/devfile/adapters/common"
	"github.com/redhat-developer/odo/pkg/devfile/consts"
	"github.com/redhat-developer/odo/pkg/log"
	appCmd "github.com/redhat-developer/odo/pkg/odo/cli/application"
	projectCmd "github.com/redhat-developer/odo/pkg/odo/cli/project"
	"github.com/redhat-developer/odo/pkg/odo/cli/ui"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/redhat-developer/odo/pkg/odo/util/completion"
	"github.com/redhat-developer/odo/pkg/util"

	"k8s.io/klog"
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
	// Component Context
	*ComponentOptions

	// Flags
	contextFlag string
	forceFlag   bool
	allFlag     bool
	waitFlag    bool
	showLogFlag bool
	deployFlag  bool
}

// NewDeleteOptions returns new instance of DeleteOptions
func NewDeleteOptions() *DeleteOptions {
	return &DeleteOptions{
		ComponentOptions: &ComponentOptions{},
	}
}

// Complete completes log args
func (do *DeleteOptions) Complete(cmdline cmdline.Cmdline, args []string) (err error) {
	do.Context, err = genericclioptions.New(genericclioptions.NewCreateParameters(cmdline).NeedDevfile(do.contextFlag))
	return err
}

// Validate validates the list parameters
func (do *DeleteOptions) Validate() error {
	return nil
}

// Run has the logic to perform the required actions as part of command
func (do *DeleteOptions) Run() (err error) {
	klog.V(4).Infof("component delete called")
	klog.V(4).Infof("args: %#v", do)
	devfileExists := util.CheckPathExists(do.GetDevfilePath())

	// odo delete --deploy || odo delete --all
	if do.deployFlag || do.allFlag {
		log.Info("Un-deploying the Kubernetes Component")
		if do.forceFlag || ui.Proceed("Are you sure you want to un-deploy?") {
			err = do.DevfileUnDeploy()
			if err != nil {
				// if there is no component in the devfile to undeploy or if the devfile is non-existent, then skip the error log
				if errors.Is(err, &component.NoDefaultDeployCommandFoundError{}) || !devfileExists {
					log.Printf("no kubernetes component to un-deploy")
				} else {
					log.Errorf("error occurred while un-deploying, cause: %v", err)
				}
			}
		} else {
			log.Error("Aborting un-deployment of the component")
		}
	}

	// odo delete || odo delete --all
	if !do.deployFlag || do.allFlag {
		log.Info("Deleting the devfile component")
		if do.forceFlag || ui.Proceed(fmt.Sprintf("Are you sure you want to delete the devfile component: %s?", do.EnvSpecificInfo.GetName())) {
			err = do.DevfileComponentDelete()
			if err != nil {
				if !devfileExists {
					log.Printf("no devfile component to delete")
				} else {
					log.Errorf("error occurred while deleting component, cause: %v", err)
				}
			}
		} else {
			log.Error("Aborting deletion of component")
		}
	}

	// Delete the configuration files
	if do.allFlag {
		log.Info("Deleting local config")
		// Prompt and delete env folder
		if do.forceFlag || ui.Proceed("Are you sure you want to delete env folder?") {
			if !do.EnvSpecificInfo.Exists() {
				log.Printf("env folder doesn't exist for the component")
				return nil
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
			err = util.DeletePath(filepath.Join(do.contextFlag, util.DotOdoDirectory))
			if err != nil {
				return err
			}
			log.Successf("Successfully deleted env file")
		} else {
			log.Error("Aborting deletion of env folder")
		}

		// Prompt and delete the devfile.yaml
		successMessage := "Successfully deleted devfile.yaml file"
		devfileNotExistsMessage := "devfile.yaml does not exist in the current directory"
		if do.forceFlag {
			if !devfileExists {
				log.Printf(devfileNotExistsMessage)
				return nil
			}
			if !do.EnvSpecificInfo.IsUserCreatedDevfile() {

				// first remove the uri based files mentioned in the devfile
				devfileObj, err := devfile.ParseAndValidateFromFile(do.GetDevfilePath())
				if err != nil {
					return err
				}

				err = common.RemoveDevfileURIContents(devfileObj, do.contextFlag)
				if err != nil {
					return err
				}

				empty, err := util.IsEmpty(filepath.Join(do.contextFlag, consts.UriFolder))
				if err != nil && !os.IsNotExist(err) {
					return err
				}

				if !os.IsNotExist(err) && empty {
					err = os.RemoveAll(filepath.Join(do.contextFlag, consts.UriFolder))
					if err != nil {
						return err
					}
				}

				err = util.DeletePath(do.GetDevfilePath())
				if err != nil {
					return err
				}

				log.Successf(successMessage)

			} else {
				log.Info("Didn't delete the devfile as it was user provided")
			}

		} else if ui.Proceed("Are you sure you want to delete devfile.yaml?") {
			if !devfileExists {
				log.Printf(devfileNotExistsMessage)
				return nil
			}

			// first remove the uri based files mentioned in the devfile
			devfileObj, err := devfile.ParseAndValidateFromFile(do.GetDevfilePath())
			if err != nil {
				return err
			}

			err = common.RemoveDevfileURIContents(devfileObj, do.contextFlag)
			if err != nil {
				return err
			}

			empty, err := util.IsEmpty(filepath.Join(do.contextFlag, consts.UriFolder))
			if err != nil && !os.IsNotExist(err) {
				return err
			}

			if !os.IsNotExist(err) && empty {
				err = os.RemoveAll(filepath.Join(do.contextFlag, consts.UriFolder))
				if err != nil {
					return err
				}
			}

			err = util.DeletePath(do.GetDevfilePath())
			if err != nil {
				return err
			}

			log.Successf(successMessage)
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

	componentDeleteCmd.Flags().BoolVarP(&do.forceFlag, "force", "f", false, "Delete component without prompting")
	componentDeleteCmd.Flags().BoolVarP(&do.allFlag, "all", "a", false, "Delete component and local config")
	componentDeleteCmd.Flags().BoolVarP(&do.waitFlag, "wait", "w", false, "Wait for complete deletion of component and its dependent")
	componentDeleteCmd.Flags().BoolVarP(&do.deployFlag, "deploy", "d", false, "Undeploy the Devfile Kubernetes Component deployed via `odo deploy`")

	componentDeleteCmd.Flags().BoolVar(&do.showLogFlag, "show-log", false, "If enabled, logs will be shown when deleted")

	componentDeleteCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)
	completion.RegisterCommandHandler(componentDeleteCmd, completion.ComponentNameCompletionHandler)
	// Adding `--context` flag
	odoutil.AddContextFlag(componentDeleteCmd, &do.contextFlag)

	// Adding `--project` flag
	projectCmd.AddProjectFlag(componentDeleteCmd)
	// Adding `--application` flag
	appCmd.AddApplicationFlag(componentDeleteCmd)

	return componentDeleteCmd
}
