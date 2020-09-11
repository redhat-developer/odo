package utils

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/fatih/color"
	imagev1 "github.com/openshift/api/image/v1"

	"github.com/openshift/odo/pkg/config"
	"github.com/openshift/odo/pkg/devfile/parser"
	devfileCtx "github.com/openshift/odo/pkg/devfile/parser/context"
	"github.com/openshift/odo/pkg/devfile/parser/data"
	"github.com/openshift/odo/pkg/devfile/parser/data/common"
	"github.com/openshift/odo/pkg/envinfo"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/occlient"
	"github.com/openshift/odo/pkg/odo/cli/component"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/openshift/odo/pkg/util"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/klog"
	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

const (
	convertCommandName = "convert-to-devfile"
	//build command id to be used in s2i devfile
	buildCommandID = "s2i-assemble"
	// build command to be used in s2i devfile
	buildCommandS2i = "/opt/odo/bin/s2i-setup && /opt/odo/bin/assemble-and-restart"
	// run command id to be used in devfile
	runCommandID = "s2i-run"
	// run command to be used in s2i devfile
	runCommandS2i = "/opt/odo/bin/run"
	// container component name to be used in devfile
	containerName = "s2i-builder"
	// directory to sync s2i source code
	sourceMappingS2i = "/tmp/projects"
	// devfile version
	devfileVersion = "2.0.0"
	// environment variable set for s2i assemble and restart scripts
	// some change in script if scripts is executed for a s2i component converted to devfile
	envS2iConvertedDevfile = "ODO_S2I_CONVERTED_DEVFILE"
)

var convertLongDesc = ktemplates.LongDesc(`Converts odo specific configuration from s2i to devfile. 
It generates devfile.yaml and .odo/env/env.yaml for s2i components`)

//var convertExample = ktemplates.Examples(`odo utils convert-to-devfile`)

var convertExample = ktemplates.Examples(`  # Convert s2i component to devfile component

Note: Run all commands from  s2i component context directory

1. Generate devfile.yaml and env.yaml for s2i component.
%[1]s  

2. Push the devfile component to the cluster.
odo push

3. Verify if devfile component is deployed sucessfully.
odo list

4. Jump to 'rolling back conversion', if devfile component deployment failed.

5. Delete the s2i component.
odo delete --s2i -a

Congratulations, you have successfully converted s2i component to devfile component.

# Rolling back the conversion

1. If devfile component deployment failed, delete the devfile component with 'odo delete -a'. 
   It would delete only devfile component, your s2i component should still be running.
 
   To complete the migration seek help from odo dev community.

`)

// ConvertOptions encapsulates the options for the command
type ConvertOptions struct {
	context          *genericclioptions.Context
	componentContext string
	componentName    string
}

// NewConvertOptions creates a new ConvertOptions instance
func NewConvertOptions() *ConvertOptions {
	return &ConvertOptions{}
}

// Complete completes ConvertOptions after they've been created
func (co *ConvertOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {

	if !util.CheckPathExists(filepath.Join(co.componentContext, component.ConfigFilePath)) {
		return errors.New("this directory does not contain an odo s2i component, Please run the command from odo component directory to convert s2i component to devfile")
	}

	co.context = genericclioptions.NewContext(cmd)
	co.componentName = co.context.LocalConfigInfo.GetName()
	return nil

}

// Validate validates the ConvertOptions based on completed values
func (co *ConvertOptions) Validate() (err error) {
	if co.context.LocalConfigInfo.GetSourceType() == config.GIT {
		return errors.New("migration of git type s2i components to devfile is not supported by odo")
	}

	return nil
}

// Run contains the logic for the command
func (co *ConvertOptions) Run() (err error) {

	/* NOTE: This data is not used in devfile currently so cannot be converted
	   minMemory := context.LocalConfigInfo.GetMinMemory()
	   minCPU := context.LocalConfigInfo.GetMinCPU()
	   maxCPU := context.LocalConfigInfo.GetMaxCPU()
	*/

	err = generateDevfileYaml(co)
	if err != nil {
		return errors.Wrap(err, "Error in generating devfile.yaml")
	}

	err = generateEnvYaml(co)
	if err != nil {
		return errors.Wrap(err, "Error in generating env.yaml")
	}

	printOutput()

	return nil
}

// NewCmdConvert implements the odo utils convert-to-devfile command
func NewCmdConvert(name, fullName string) *cobra.Command {
	o := NewConvertOptions()
	convertCmd := &cobra.Command{
		Use:     name,
		Short:   "converts s2i based components to devfile based components",
		Long:    convertLongDesc,
		Example: fmt.Sprintf(convertExample, fullName),
		Args:    cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	genericclioptions.AddContextFlag(convertCmd, &o.componentContext)

	return convertCmd
}

// generateDevfileYaml generates a devfile.yaml from s2i data.
func generateDevfileYaml(co *ConvertOptions) error {
	klog.V(2).Info("Generating devfile.yaml")

	// builder image to use
	componentType := co.context.LocalConfigInfo.GetType()
	// git, local, binary, none
	sourceType := co.context.LocalConfigInfo.GetSourceType()

	imageStream, imageforDevfile, err := getImageforDevfile(co.context.Client, componentType)
	if err != nil {
		return errors.Wrap(err, "Failed to get image details")
	}

	envVarList := co.context.LocalConfigInfo.GetEnvVars()
	s2iEnv, err := occlient.GetS2IEnvForDevfile(string(sourceType), envVarList, *imageStream)
	if err != nil {
		return err
	}

	s2iDevfile, err := data.NewDevfileData(devfileVersion)
	if err != nil {
		return err
	}

	// set schema version
	s2iDevfile.SetSchemaVersion(devfileVersion)

	// set metadata
	s2iDevfile.SetMetadata(co.componentName, "1.0.0")
	// set commponents
	err = setDevfileComponentsForS2I(s2iDevfile, imageforDevfile, co.context.LocalConfigInfo, s2iEnv)
	if err != nil {
		return err
	}
	// set commands
	setDevfileCommandsForS2I(s2iDevfile)

	ctx := devfileCtx.NewDevfileCtx(filepath.Join(co.componentContext, "devfile.yaml"))
	err = ctx.SetAbsPath()
	if err != nil {
		return err
	}

	devObj := parser.DevfileObj{
		Ctx:  ctx,
		Data: s2iDevfile,
	}

	err = devObj.WriteYamlDevfile()
	klog.V(2).Info("Generated devfile.yaml successfully")

	if err != nil {
		return err
	}

	return nil
}

// generateEnvYaml generates .odo/env.yaml from s2i data.
func generateEnvYaml(co *ConvertOptions) (err error) {
	klog.V(2).Info("Generating env.yaml")

	// list of urls having name, ports, secure
	urls := co.context.LocalConfigInfo.GetURL()
	debugPort := co.context.LocalConfigInfo.GetDebugPort()

	application := co.context.LocalConfigInfo.GetApplication()

	// Generate env.yaml
	co.context.EnvSpecificInfo, err = envinfo.NewEnvSpecificInfo(co.componentContext)
	if err != nil {
		return err
	}

	var urlList []envinfo.EnvInfoURL

	for _, url := range urls {
		urlEnv := envinfo.EnvInfoURL{
			Name:   url.Name,
			Port:   url.Port,
			Secure: url.Secure,
			// s2i components are only run on openshift cluster
			Kind: envinfo.ROUTE,
		}

		urlList = append(urlList, urlEnv)

	}

	componentSettings := envinfo.ComponentSettings{
		Name:      co.componentName,
		Namespace: co.context.Project,
		URL:       &urlList,
		AppName:   application,
	}

	if debugPort != 0 || debugPort == config.DefaultDebugPort {
		componentSettings.DebugPort = &debugPort
	}

	err = co.context.EnvSpecificInfo.SetComponentSettings(componentSettings)
	if err != nil {
		return err
	}
	klog.V(2).Info("Generated env.yaml successfully")

	return
}

// getImageforDevfile gets image details from s2i component type.
func getImageforDevfile(client *occlient.Client, componentType string) (*imagev1.ImageStreamImage, string, error) {
	klog.V(2).Info("Getting container image details")

	imageNS, imageName, imageTag, _, err := occlient.ParseImageName(componentType)
	if err != nil {
		return nil, "", err
	}
	imageStream, err := client.GetImageStream(imageNS, imageName, imageTag)
	if err != nil {
		return nil, "", err
	}

	imageStreamImage, err := client.GetImageStreamImage(imageStream, imageTag)
	if err != nil {
		return nil, "", err
	}

	imageforDevfile := imageStream.Spec.Tags[0].From.Name

	return imageStreamImage, imageforDevfile, nil
}

// setDevfileCommandsForS2I sets command in devfile.yaml from s2i data.
func setDevfileCommandsForS2I(d data.DevfileData) {
	klog.V(2).Info("Set devfile commands from s2i data")

	buildCommand := common.DevfileCommand{
		Id: buildCommandID,
		Exec: &common.Exec{
			Component:   containerName,
			CommandLine: buildCommandS2i,
			Group: &common.Group{
				Kind:      common.BuildCommandGroupType,
				IsDefault: true,
			},
		},
	}

	runCommand := common.DevfileCommand{
		Id: runCommandID,
		Exec: &common.Exec{
			Component:   containerName,
			CommandLine: runCommandS2i,
			Group: &common.Group{
				Kind:      common.RunCommandGroupType,
				IsDefault: true,
			},
		},
	}

	// Ignoring error as we are writing new file
	_ = d.AddCommands(buildCommand, runCommand)

}

// setDevfileComponentsForS2I sets the devfile.yaml components field from s2i data.
func setDevfileComponentsForS2I(d data.DevfileData, s2iImage string, localConfig *config.LocalConfigInfo, s2iEnv config.EnvVarList) error {
	klog.V(2).Info("Set devfile components from s2i data")

	maxMemory := localConfig.GetMaxMemory()
	volumes := localConfig.GetStorage()
	// list of ports taken from builder image and set into local config
	ports := localConfig.GetPorts()

	var endpoints []common.Endpoint
	var envs []common.Env
	var volumeMounts []common.VolumeMount
	var components []common.DevfileComponent

	// convert s2i storage to devfile volumes
	for _, vol := range volumes {
		volume := common.Volume{
			Size: vol.Size,
		}
		components = append(components, common.DevfileComponent{Name: vol.Name, Volume: &volume})

		volumeMount := common.VolumeMount{
			Name: vol.Name,
			Path: vol.Path,
		}

		volumeMounts = append(volumeMounts, volumeMount)
	}

	// Add s2i specific env variable in devfile
	for _, env := range s2iEnv {
		env := common.Env{
			Name:  env.Name,
			Value: env.Value,
		}
		envs = append(envs, env)
	}
	env := common.Env{
		Name:  envS2iConvertedDevfile,
		Value: "true",
	}
	envs = append(envs, env)

	// convert s2i ports to devfile endpoints
	for _, port := range ports {

		port := strings.Split(port, "/")
		// from s2i config.yaml port is of the form 8080/TCP
		// Ignoring TCP and Udp values here
		portInt, err := strconv.ParseInt(port[0], 10, 32)
		if err != nil {
			return errors.Wrapf(err, "Unable to convert port %s from config.yaml to devfile.yaml", port)
		}

		endpoint := common.Endpoint{
			Name:       fmt.Sprintf("port-%s", port[0]),
			TargetPort: int32(portInt),
		}

		endpoints = append(endpoints, endpoint)
	}

	container := common.Container{
		Image:         s2iImage,
		MountSources:  true,
		SourceMapping: sourceMappingS2i,
		MemoryLimit:   maxMemory,
		Endpoints:     endpoints,
		Env:           envs,
		VolumeMounts:  volumeMounts,
	}

	components = append(components, common.DevfileComponent{Name: containerName, Container: &container})

	// Ignoring error here as we are writing a new file
	_ = d.AddComponents(components)

	return nil

}

func printOutput() {

	infoMessage := "devfile.yaml is available in the current directory."

	nextSteps := `
To complete the conversion, run the following steps:

NOTE: At all steps your s2i component is running, It would not be deleted until you do 'odo delete --s2i -a'

1. Deploy devfile component.
$ odo push

2. Verify if the component gets deployed successfully. 
$ odo list

3. If the devfile component was deployed successfully, your application is up, you can safely delete the s2i component. 
$ odo delete --s2i -a

congratulations you have successfully converted s2i component to devfile component :).
`

	rollBackMessage := ` If you see an error or your application not coming up, delete the devfile component with 'odo delete -a' and report this to odo dev community.`

	log.Infof(infoMessage)
	log.Italicf(nextSteps)
	yellow := color.New(color.FgYellow).SprintFunc()
	log.Warning(yellow(rollBackMessage))
}
