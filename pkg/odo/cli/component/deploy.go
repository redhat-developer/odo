package component

import (
	"bytes"
	"fmt"
	"path/filepath"

	devfileParser "github.com/openshift/odo/pkg/devfile/parser"
	"github.com/openshift/odo/pkg/envinfo"
	"github.com/openshift/odo/pkg/log"
	projectCmd "github.com/openshift/odo/pkg/odo/cli/project"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/openshift/odo/pkg/odo/util/completion"
	"github.com/openshift/odo/pkg/util"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	odoutil "github.com/openshift/odo/pkg/odo/util"

	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

var deployCmdExample = ktemplates.Examples(`  # Build and Deploy the current component 
%[1]s

# Specify the tag for the image by calling
%[1]s --tag <registry>/<namespace>/<image>:<tag>
  `)

// DeployRecommendedCommandName is the recommended build command name
const DeployRecommendedCommandName = "deploy"

// DeployOptions encapsulates options that build command uses
type DeployOptions struct {
	componentContext string
	sourcePath       string
	ignores          []string
	EnvSpecificInfo  *envinfo.EnvSpecificInfo

	DevfilePath     string
	devObj          devfileParser.DevfileObj
	DockerfileURL   string
	DockerfileBytes []byte
	namespace       string
	tag             string
	ManifestSource  []byte
	DeploymentPort  int

	*genericclioptions.Context
}

// NewDeployOptions returns new instance of BuildOptions
// with "default" values for certain values, for example, show is "false"
func NewDeployOptions() *DeployOptions {
	return &DeployOptions{}
}

// CompleteDevfilePath completes the devfile path from context
func (do *DeployOptions) CompleteDevfilePath() {
	if len(do.DevfilePath) > 0 {
		do.DevfilePath = filepath.Join(do.componentContext, do.DevfilePath)
	} else {
		do.DevfilePath = filepath.Join(do.componentContext, "devfile.yaml")
	}
}

// Complete completes deploy args
func (do *DeployOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	do.CompleteDevfilePath()
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

	log.Infof("\nValidation")

	// Validate the --tag
	if do.tag == "" {
		return errors.New("odo deploy requires a tag, in the format <registry>/namespace>/<image>")
	}

	s := log.Spinner("Validating arguments")
	err = util.ValidateTag(do.tag)
	if err != nil {
		s.End(false)
		return err
	}
	s.End(true)

	do.devObj, err = devfileParser.ParseAndValidate(do.DevfilePath)
	if err != nil {
		return err
	}

	s = log.Spinner("Validating build information")
	metadata := do.devObj.Data.GetMetadata()
	dockerfileURL := metadata.Dockerfile

	//Download Dockerfile to .odo, build, then delete from .odo dir
	//If Dockerfile is present in the project already, use that for the build
	//If Dockerfile is present in the project and field is in devfile, build the one already in the project and warn the user.
	if dockerfileURL != "" && util.CheckPathExists(filepath.Join(do.componentContext, "Dockerfile")) {
		// TODO: make clearer more visible output
		log.Warning("Dockerfile already exists in project directory and one is specified in Devfile.")
		log.Warningf("Using Dockerfile specified in devfile from '%s'", dockerfileURL)
	}

	if dockerfileURL != "" {
		dockerfileBytes, err := util.LoadFileIntoMemory(dockerfileURL)
		if err != nil {
			s.End(false)
			return errors.New("unable to download Dockerfile from URL specified in devfile")
		}
		// If we successfully downloaded the Dockerfile into memory, store it in the DeployOptions
		do.DockerfileBytes = dockerfileBytes

		// Validate the file that was downloaded is a Dockerfile
		err = util.ValidateDockerfile(dockerfileBytes)
		if err != nil {
			s.End(false)
			return err
		}

	} else if !util.CheckPathExists(filepath.Join(do.componentContext, "Dockerfile")) {
		s.End(false)
		return errors.New("dockerfile required for build. No 'alpha.build-dockerfile' field found in devfile, or Dockerfile found in project directory")
	}

	s.End(true)

	s = log.Spinner("Validating deployment information")
	manifestURL := metadata.Manifest

	if manifestURL == "" {
		s.End(false)
		return errors.New("Unable to deploy as alpha.deployment-manifest is not defined in devfile.yaml")
	}

	manifestBytes, err := util.LoadFileIntoMemory(manifestURL)
	if err != nil {
		s.End(false)
		return errors.Wrap(err, "unable to download manifest from URL specified in devfile")
	}
	do.ManifestSource = manifestBytes

	// check if manifestSource contains PORT template variable
	// if it does, then check we have an port setup in env.yaml
	do.DeploymentPort = 0
	if bytes.Contains(manifestBytes, []byte("PORT")) {
		deploymentPort, err := do.EnvSpecificInfo.GetPortByURLKind(envinfo.ROUTE)
		if err != nil {
			s.End(false)
			return errors.Wrap(err, "unable to find `port` for deployment. `odo url create` must be run prior to `odo deploy`")
		}
		do.DeploymentPort = deploymentPort
	}

	s.End(true)

	return
}

// Run has the logic to perform the required actions as part of command
func (do *DeployOptions) Run() (err error) {
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
	deployCmd.Flags().StringVar(&do.tag, "tag", "", "Tag used to build the image")
	deployCmd.Flags().StringSliceVar(&do.ignores, "ignore", []string{}, "Files or folders to be ignored via glob expressions.")

	//Adding `--project` flag
	projectCmd.AddProjectFlag(deployCmd)

	deployCmd.AddCommand(deployDeleteCmd)
	deployCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)
	completion.RegisterCommandHandler(deployCmd, completion.ComponentNameCompletionHandler)
	completion.RegisterCommandFlagHandler(deployCmd, "context", completion.FileCompletionHandler)

	return deployCmd
}
