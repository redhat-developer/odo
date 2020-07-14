package utils

import (
	"fmt"
	"strconv"
	"strings"

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
	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

// TODO
// lables and annotations that are added on s2i components

const (
	migrateCommandName = "migrate-to-devfile"
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

var migrateLongDesc = ktemplates.LongDesc(`Migrate S2I components to devfile components`)

var migrateExample = ktemplates.Examples(`odo utils migrate-to-devfile`)

// MigrateOptions encapsulates the options for the command
type MigrateOptions struct {
	context          *genericclioptions.Context
	componentContext string
	componentName    string
}

// NewMigrateOptions creates a new MigrateOptions instance
func NewMigrateOptions() *MigrateOptions {
	return &MigrateOptions{}
}

// Complete completes MigrateOptions after they've been created
func (mo *MigrateOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {

	if !util.CheckPathExists(component.ConfigFilePath) {
		return errors.New("this directory does not contain an odo s2i component, Please run the command from odo component directory to migrate s2i component to devfile")
	}

	mo.context = genericclioptions.NewContext(cmd)
	mo.componentContext = component.LocalDirectoryDefaultLocation
	mo.componentName = mo.context.LocalConfigInfo.GetName()
	return nil

}

// Validate validates the MigrateOptions based on completed values
func (mo *MigrateOptions) Validate() (err error) {
	if mo.context.LocalConfigInfo.GetSourceType() == config.GIT {
		return errors.New("migration of git type s2i components to devfile not supported by odo")
	}

	return nil
}

// Run contains the logic for the command
func (mo *MigrateOptions) Run() (err error) {

	/*  This data is yet to be converted

	// Absolute path
	sourcePath, _ := context.LocalConfigInfo.GetOSSourcePath()
	minMemory := context.LocalConfigInfo.GetMinMemory()
	minCPU := context.LocalConfigInfo.GetMinCPU()
	maxCPU := context.LocalConfigInfo.GetMaxCPU()

	*/

	err = generateDevfileYaml(mo)
	if err != nil {
		return errors.Wrap(err, "Error in generating devfile.yaml")
	}

	err = generateEnvYaml(mo)
	if err != nil {
		return errors.Wrap(err, "Error in generating env.yaml")
	}

	// TODO: Delete the s2i component and deploy the devfile component.

	return nil
}

// NewCmdMigrate implements the odo utils migrate-to-devfile command
func NewCmdMigrate(name, fullName string) *cobra.Command {
	o := NewMigrateOptions()
	migrateCmd := &cobra.Command{
		Use:     name,
		Short:   "migrates s2i based components to devfile based components",
		Long:    migrateLongDesc,
		Example: fmt.Sprintf(migrateExample, fullName),
		Args:    cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}
	return migrateCmd
}

func generateDevfileYaml(m *MigrateOptions) error {

	// builder image to use
	componentType := m.context.LocalConfigInfo.GetType()
	// git, local, binary, none
	sourceType := m.context.LocalConfigInfo.GetSourceType()

	imageStream, imageforDevfile, err := getImageforDevfile(m.context.Client, componentType)
	if err != nil {
		return errors.Wrap(err, "Failed to get image details")
	}

	envVarList := m.context.LocalConfigInfo.GetEnvVars()
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
	s2iDevfile.SetMetadata(m.componentName, "1.0.0")
	// set commponents
	setDevfileComponentsForS2I(s2iDevfile, imageforDevfile, m.context.LocalConfigInfo, s2iEnv)
	// set commands
	setDevfileCommandsForS2I(s2iDevfile)

	devObj := parser.DevfileObj{
		Ctx:  devfileCtx.NewDevfileCtx(m.componentContext), //component context needs to be passed here
		Data: s2iDevfile,
	}

	err = devObj.WriteYamlDevfile()
	if err != nil {
		return err
	}
	log.Italic("devfile.yaml is available in current directory, run `odo push` to deploy devfile component and `odo delete` to delete s2i component.\n")
	return nil
}

func generateEnvYaml(m *MigrateOptions) (err error) {

	// list of urls having name, ports, secure
	urls := m.context.LocalConfigInfo.GetURL()
	debugPort := m.context.LocalConfigInfo.GetDebugPort()

	// TODO(adi): Add application in env.yaml once odo list PR gets merged
	// application := context.LocalConfigInfo.GetApplication()

	// Generate env.yaml
	m.context.EnvSpecificInfo, err = envinfo.NewEnvSpecificInfo(m.componentContext)
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
		Name:      m.componentName,
		Namespace: m.context.Project,
		URL:       &urlList,
	}

	if debugPort != 0 || debugPort == config.DefaultDebugPort {
		componentSettings.DebugPort = &debugPort
	}

	return m.context.EnvSpecificInfo.SetComponentSettings(componentSettings)
}

func getImageforDevfile(client *occlient.Client, componentType string) (*imagev1.ImageStreamImage, string, error) {

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

func setDevfileCommandsForS2I(d data.DevfileData) {

	buildCommand := common.DevfileCommand{
		Exec: &common.Exec{
			Id:          buildCommandID,
			Component:   containerName,
			CommandLine: buildCommandS2i,
			Group: &common.Group{
				Kind:      common.BuildCommandGroupType,
				IsDefault: true,
			},
		},
	}

	runCommand := common.DevfileCommand{
		Exec: &common.Exec{
			Id:          runCommandID,
			Component:   containerName,
			CommandLine: runCommandS2i,
			Group: &common.Group{
				Kind:      common.RunCommandGroupType,
				IsDefault: true,
			},
		},
	}

	d.AddCommand(buildCommand)
	d.AddCommand(runCommand)
}

func setDevfileComponentsForS2I(d data.DevfileData, s2iImage string, localConfig *config.LocalConfigInfo, s2iEnv config.EnvVarList) {

	maxMemory := localConfig.GetMaxMemory()
	volumes := localConfig.GetStorage()
	// list of ports taken from builder image and set into local config
	ports := localConfig.GetPorts()

	var endpoints []common.Endpoint
	var envs []common.Env
	var volumeMounts []common.VolumeMount

	// convert s2i storage to devfile volumes
	for _, vol := range volumes {
		volume := common.Volume{
			Name: vol.Name,
			Size: vol.Size,
		}
		d.AddComponent(common.DevfileComponent{Volume: &volume})

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
		portInt, _ := strconv.ParseInt(port[0], 10, 32)
		endpoint := common.Endpoint{
			Name:       fmt.Sprintf("port-%s", port[0]),
			TargetPort: int32(portInt),
			Configuration: &common.Configuration{
				Public: true,
			},
		}

		endpoints = append(endpoints, endpoint)
	}

	container := common.Container{
		Name:          containerName,
		Image:         s2iImage,
		MountSources:  true,
		SourceMapping: sourceMappingS2i,
		MemoryLimit:   maxMemory,
		Endpoints:     endpoints,
		Env:           envs,
		VolumeMounts:  volumeMounts,
	}

	component := common.DevfileComponent{Container: &container}
	d.AddComponent(component)
}
