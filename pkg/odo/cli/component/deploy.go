package component

import (
	"fmt"
	"os"
	"path/filepath"

	devfileParser "github.com/openshift/odo/pkg/devfile/parser"
	"github.com/openshift/odo/pkg/envinfo"
	"github.com/openshift/odo/pkg/log"
	projectCmd "github.com/openshift/odo/pkg/odo/cli/project"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/openshift/odo/pkg/odo/util/completion"
	"github.com/openshift/odo/pkg/odo/util/experimental"
	"github.com/openshift/odo/pkg/util"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	odoutil "github.com/openshift/odo/pkg/odo/util"

	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

// TODO: add CLI Reference doc
// TODO: add delete example
var deployCmdExample = ktemplates.Examples(`  # Deploys an image and deploys the application 
%[1]s
  `)

// DeployRecommendedCommandName is the recommended build command name
const DeployRecommendedCommandName = "deploy"

// DeployOptions encapsulates options that build command uses
type DeployOptions struct {
	*CommonPushOptions

	// devfile path
	DevfilePath     string
	DockerfileURL   string
	DockerfileBytes []byte
	namespace       string
	tag             string
	ManifestSource  []byte
	deployOnly      bool
}

// NewDeployOptions returns new instance of BuildOptions
// with "default" values for certain values, for example, show is "false"
func NewDeployOptions() *DeployOptions {
	return &DeployOptions{
		CommonPushOptions: NewCommonPushOptions(),
	}
}

// Complete completes push args
func (do *DeployOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	do.DevfilePath = filepath.Join(do.componentContext, do.DevfilePath)
	envInfo, err := envinfo.NewEnvSpecificInfo(do.componentContext)
	if err != nil {
		return errors.Wrap(err, "unable to retrieve configuration information")
	}
	do.EnvSpecificInfo = envInfo
	do.Context = genericclioptions.NewDevfileContext(cmd)

	return nil
}

// Validate validates the push parameters
func (do *DeployOptions) Validate() (err error) {

	// Validate the --tag
	if do.tag == "" {
		return errors.New("odo deploy requires a tag, in the format <registry>/namespace>/<image>")
	}

	err = util.ValidateTag(do.tag)
	if err != nil {
		return err
	}

	return
}

// Run has the logic to perform the required actions as part of command
func (do *DeployOptions) Run() (err error) {
	devObj, err := devfileParser.Parse(do.DevfilePath)
	if err != nil {
		return err
	}

	metadata := devObj.Data.GetMetadata()
	dockerfileURL := metadata.Dockerfile
	localDir, err := os.Getwd()
	if err != nil {
		return err
	}

	// TODO Remove this as it was only put in for testing
	if !do.deployOnly {
		//Download Dockerfile to .odo, build, then delete from .odo dir
		//If Dockerfile is present in the project already, use that for the build
		//If Dockerfile is present in the project and field is in devfile, build the one already in the project and warn the user.
		if dockerfileURL != "" && util.CheckPathExists(filepath.Join(localDir, "Dockerfile")) {
			// TODO: make clearer more visible output
			log.Warning("Dockerfile already exists in project directory and one is specified in Devfile.")
			log.Warningf("Using Dockerfile specified in devfile from '%s'", dockerfileURL)
		}

		if !util.CheckPathExists(filepath.Join(localDir, ".odo")) {
			return errors.Wrap(err, ".odo folder not found")
		}

		if dockerfileURL != "" {
			dockerfileBytes, err := util.DownloadFileInMemory(dockerfileURL)
			if err != nil {
				return errors.New("unable to download Dockerfile from URL specified in devfile")
			}
			// If we successfully downloaded the Dockerfile into memory, store it in the DeployOptions
			do.DockerfileBytes = dockerfileBytes

			// Validate the file that was downloaded is a Dockerfile
			err = util.ValidateDockerfile(dockerfileBytes)
			if err != nil {
				return err
			}

		} else if !util.CheckPathExists(filepath.Join(localDir, "Dockerfile")) {
			return errors.New("dockerfile required for build. No 'dockerfile' field found in devfile, or Dockerfile found in project directory")
		}

		err = do.DevfileBuild()
		if err != nil {
			return err
		}
	}

	manifestURL := metadata.Manifest
	do.ManifestSource, err = util.DownloadFileInMemory(manifestURL)
	if err != nil {
		return errors.Wrap(err, "Unable to download manifest "+manifestURL)
	}

	err = do.DevfileDeploy()
	if err != nil {
		return err
	}

	return nil
}

// Need to use RunE on Cobra command to allow for `odo deploy` and `odo deploy delete`
// See reconfigureCmdWithSubCmd function in cli.go
func (do *DeployOptions) deployRunE(cmd *cobra.Command, args []string) error {
	genericclioptions.GenericRun(do, cmd, args)
	return nil
}

// NewCmdDeploy implements the push odo command
func NewCmdDeploy(name, fullName string) *cobra.Command {
	do := NewDeployOptions()

	deployDeleteCmd := NewCmdDeployDelete(DeployDeleteRecommendedCommandName, odoutil.GetFullName(fullName, DeployDeleteRecommendedCommandName))

	var deployCmd = &cobra.Command{
		Use:         fmt.Sprintf("%s [command] [component name]", name),
		Short:       "Deploy image for component",
		Long:        `Deploy image for component`,
		Example:     fmt.Sprintf(deployCmdExample, fullName),
		Args:        cobra.MaximumNArgs(1),
		Annotations: map[string]string{"command": "component"},
		RunE:        do.deployRunE,
	}
	genericclioptions.AddContextFlag(deployCmd, &do.componentContext)

	// enable devfile flag if experimental mode is enabled
	if experimental.IsExperimentalModeEnabled() {
		deployCmd.Flags().StringVar(&do.DevfilePath, "devfile", "./devfile.yaml", "Path to a devfile.yaml")
		deployCmd.Flags().StringVar(&do.tag, "tag", "", "Tag used to build the image")
		deployCmd.Flags().BoolVar(&do.deployOnly, "deployOnly", false, "Do not build the application, only deploy it")
	}

	//Adding `--project` flag
	projectCmd.AddProjectFlag(deployCmd)

	deployCmd.AddCommand(deployDeleteCmd)
	deployCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)
	completion.RegisterCommandHandler(deployCmd, completion.ComponentNameCompletionHandler)
	completion.RegisterCommandFlagHandler(deployCmd, "context", completion.FileCompletionHandler)

	return deployCmd
}
