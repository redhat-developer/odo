package envinfo

import (
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	"github.com/openshift/odo/pkg/localConfigProvider"
	"github.com/openshift/odo/pkg/testingutil/filesystem"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"

	"github.com/openshift/odo/pkg/util"
)

type JSONEnvInfoRepr struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ComponentSettings `json:"spec" yaml:"spec"`
}

// ComponentSettings holds all component related information
type ComponentSettings struct {
	Name string `yaml:"Name,omitempty" json:"name,omitempty"`

	Project string `yaml:"Project,omitempty" json:"project,omitempty"`

	UserCreatedDevfile bool `yaml:"UserCreatedDevfile,omitempty" json:"UserCreatedDevfile,omitempty"`

	URL *[]localConfigProvider.LocalURL `yaml:"Url,omitempty" json:"url,omitempty"`
	// AppName is the application name. Application is a virtual concept present in odo used
	// for grouping of components. A namespace can contain multiple applications
	AppName string         `yaml:"AppName,omitempty" json:"appName,omitempty"`
	Link    *[]EnvInfoLink `yaml:"Link,omitempty" json:"link,omitempty"`

	// DebugPort controls the port used by the pod to run the debugging agent on
	DebugPort *int `yaml:"DebugPort,omitempty" json:"debugPort,omitempty"`

	// RunMode indicates the mode of run used for a successful push
	RunMode *RUNMode `yaml:"RunMode,omitempty" json:"runMode,omitempty"`
}

type RUNMode string

const (
	Run   RUNMode = "run"
	Debug RUNMode = "debug"
)

const (
	envInfoEnvName  = "ENVINFO"
	envInfoFileName = "env.yaml"

	// DefaultDebugPort is the default port used for debugging on remote pod
	DefaultDebugPort = 5858

	// DefaultRunMode is the default run mode of the component
	DefaultRunMode = Run
)

// EnvInfo holds all the env specific information relevant to a specific Component.
type EnvInfo struct {
	devfileObj        parser.DevfileObj
	isRouteSupported  bool
	updateURL         bool              // this indicates that the URL create operation should be an update operation
	componentSettings ComponentSettings `yaml:"ComponentSettings,omitempty"`
}

// proxyEnvInfo holds all the parameter that envinfo does but exposes all
// of it, used for serialization.
type proxyEnvInfo struct {
	ComponentSettings ComponentSettings `yaml:"ComponentSettings,omitempty"`
}

// EnvSpecificInfo wraps the envinfo and provides helpers to
// serialize it.
type EnvSpecificInfo struct {
	devfilePath       string
	Filename          string `yaml:"FileName,omitempty"`
	fs                filesystem.Filesystem
	EnvInfo           `yaml:",omitempty"`
	envinfoFileExists bool
}

type EnvInfoLink struct {
	// Name of link (same as name of k8s secret)
	Name string `yaml:"Name,omitempty" json:"name,omitempty"`
	// Kind of service with which the component is linked
	ServiceKind string `yaml:"ServiceKind,omitempty" json:"serviceKind,omitempty"`
	// Name of the instance of the ServiceKind that component is linked with
	ServiceName string `yaml:"ServiceName,omitempty" json:"serviceName,omitempty"`
}

func WrapForJSONOutput(compSettings ComponentSettings) JSONEnvInfoRepr {
	return JSONEnvInfoRepr{
		TypeMeta: metav1.TypeMeta{
			Kind:       "EnvInfo",
			APIVersion: "odo.dev/v1alpha1",
		},
		Spec: compSettings,
	}
}

func (esi EnvSpecificInfo) GetDevfilePath() string {
	return esi.devfilePath
}

// getEnvInfoFile first checks for the ENVINFO variable
// then we check for directory and eventually the file (which we return as a string)
func getEnvInfoFile(envDir string) (string, string, error) {
	if env, ok := os.LookupEnv(envInfoEnvName); ok {
		return env, filepath.Join(env, "..", "..", "..", "devfile.yaml"), nil
	}

	if envDir == "" {
		var err error
		envDir, err = os.Getwd()
		if err != nil {
			return "", "", err
		}
	}

	return filepath.Join(envDir, ".odo", "env", envInfoFileName), filepath.Join(envDir, "devfile.yaml"), nil
}

// New returns the EnvSpecificInfo
func New() (*EnvSpecificInfo, error) {
	return NewEnvSpecificInfo("")
}

// NewEnvSpecificInfo retrieves the environment file. If it does not exist, it returns *blank*
func NewEnvSpecificInfo(envDir string) (*EnvSpecificInfo, error) {
	return newEnvSpecificInfo(envDir, filesystem.DefaultFs{})
}

// newEnvSpecificInfo retrieves the env.yaml file, if it does not exist, we return a *BLANK* environment file.
func newEnvSpecificInfo(envDir string, fs filesystem.Filesystem) (*EnvSpecificInfo, error) {
	// Get the path of the environment file
	envInfoFile, devfilePath, err := getEnvInfoFile(envDir)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get the path of the environment file")
	}

	// Organize that information into a struct
	e := EnvSpecificInfo{
		EnvInfo:           NewEnvInfo(),
		devfilePath:       devfilePath,
		Filename:          envInfoFile,
		envinfoFileExists: true,
		fs:                fs,
	}

	// If the env.yaml file does not exist then we simply return and set e.envinfoFileExists as false
	if _, err = e.fs.Stat(envInfoFile); os.IsNotExist(err) {
		e.envinfoFileExists = false
		return &e, nil
	}

	// Retrieve the environment file
	err = getFromFile(&e.EnvInfo, e.Filename)
	if err != nil {
		return nil, err
	}

	return &e, nil
}

func getFromFile(envinfo *EnvInfo, filename string) error {
	proxyei := newProxyEnvInfo()

	err := util.GetFromFile(&proxyei, filename)
	if err != nil {
		return err
	}
	envinfo.componentSettings = proxyei.ComponentSettings
	return nil
}

// NewEnvInfo creates an empty EnvSpecificInfo struct with typeMeta populated
func NewEnvInfo() EnvInfo {
	return EnvInfo{}
}

// newProxyEnvInfo creates an empty ProxyEnvInfo struct with typeMeta populated
func newProxyEnvInfo() proxyEnvInfo {
	return proxyEnvInfo{}
}

// SetConfiguration sets the environment specific info such as the Cluster Host, Name, etc.
// we then **write** this data to the environment yaml file (see envInfoFileName const)
func (esi *EnvSpecificInfo) SetConfiguration(parameter string, value interface{}) (err error) {
	if parameter, ok := asLocallySupportedParameter(parameter); ok {
		switch parameter {
		case "name":
			val := value.(string)
			esi.componentSettings.Name = val
		case "project":
			val := value.(string)
			esi.componentSettings.Project = val
		case "debugport":
			val, err := strconv.Atoi(value.(string))
			if err != nil {
				return errors.Wrap(err, "failed to set debug port")
			}
			esi.componentSettings.DebugPort = &val
		case "url":
			urlValue := value.(localConfigProvider.LocalURL)
			if esi.componentSettings.URL != nil {
				*esi.componentSettings.URL = append(*esi.componentSettings.URL, urlValue)
			} else {
				esi.componentSettings.URL = &[]localConfigProvider.LocalURL{urlValue}
			}

		case "link":
			linkValue := value.(EnvInfoLink)
			if esi.componentSettings.Link != nil {
				*esi.componentSettings.Link = append(*esi.componentSettings.Link, linkValue)
			} else {
				esi.componentSettings.Link = &[]EnvInfoLink{linkValue}
			}
		}

		return esi.writeToFile()
	}
	return errors.Errorf("unknown parameter: %q is not a parameter in the odo environment file, please refer `odo env set --help` to see valid parameters", parameter)

}

// DeleteEnvDirIfEmpty Deletes the env directory if its empty
func (esi *EnvSpecificInfo) DeleteEnvDirIfEmpty() error {
	envDir := filepath.Dir(esi.Filename)
	_, err := esi.fs.Stat(envDir)
	if os.IsNotExist(err) {
		// If the Env dir doesn't exist then we dont mind
		return nil
	} else if err != nil {
		// Possible to not have permission to the dir
		return err
	}
	f, err := esi.fs.Open(envDir)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Readdir(1)

	// If directory is empty we can remove it
	if err == io.EOF {
		klog.V(4).Info("Deleting the env directory as well because its empty")

		return esi.fs.Remove(envDir)
	}
	return err
}

// DeleteEnvInfoFile deletes the envinfo.yaml file if it exists
func (esi *EnvSpecificInfo) DeleteEnvInfoFile() error {
	return util.DeletePath(esi.Filename)
}

// IsSet uses reflection to get the parameter from the envinfo struct, currently
// it only searches the componentSettings
func (esi *EnvSpecificInfo) IsSet(parameter string) bool {
	return util.IsSet(esi.componentSettings, parameter)
}

// Exists returns whether the envinfo file exists or not
func (esi *EnvSpecificInfo) Exists() bool {
	return esi.envinfoFileExists
}

var (
	// Mandatory parameters in the environment file (env.yaml)
	manParams = []string{
		"name",
		"project",
	}
)

// DeleteConfiguration is used to delete environment specific info from local odo envinfo
func (esi *EnvSpecificInfo) DeleteConfiguration(parameter string) error {
	for _, manParam := range manParams {
		if parameter == manParam {
			return errors.Errorf("failed to unset %q: %q is mandatory parameter", parameter, parameter)
		}
	}

	if parameter, ok := asLocallySupportedParameter(parameter); ok {

		switch parameter {
		default:
			if err := util.DeleteConfiguration(&esi.componentSettings, parameter); err != nil {
				return err
			}
		}
		return esi.writeToFile()
	}
	return errors.Errorf("unknown parameter: %q is not a parameter in the odo environment file, please refer `odo env unset --help` to unset a valid parameter", parameter)

}

func (esi *EnvSpecificInfo) DeleteLink(parameter string) error {
	index := -1

	for i, link := range *esi.componentSettings.Link {
		if link.Name == parameter {
			index = i
			break
		}
	}

	if index != -1 {
		s := *esi.componentSettings.Link
		s = append(s[:index], s[index+1:]...)
		esi.componentSettings.Link = &s
		return esi.writeToFile()
	} else {
		return nil
	}
}

// GetComponentSettings returns the componentSettings from envinfo
func (esi *EnvSpecificInfo) GetComponentSettings() ComponentSettings {
	return esi.componentSettings
}

// SetComponentSettings sets the componentSettings from to the envinfo and writes to the file
func (esi *EnvSpecificInfo) SetComponentSettings(cs ComponentSettings) error {
	esi.componentSettings = cs
	return esi.writeToFile()
}

func (esi *EnvSpecificInfo) writeToFile() error {
	proxyei := newProxyEnvInfo()
	proxyei.ComponentSettings = esi.componentSettings

	return util.WriteToFile(&proxyei, esi.Filename)
}

// GetName returns the component name
func (ei *EnvInfo) GetName() string {
	return ei.componentSettings.Name
}

// GetDebugPort returns the DebugPort, returns default if nil
func (ei *EnvInfo) GetDebugPort() int {
	if ei.componentSettings.DebugPort == nil {
		return DefaultDebugPort
	}
	return *ei.componentSettings.DebugPort
}

// GetContainers returns the Container components from the devfile
// returns empty list if nil
func (ei *EnvInfo) GetContainers() ([]localConfigProvider.LocalContainer, error) {
	var localContainers []localConfigProvider.LocalContainer

	devfileComponents, err := ei.devfileObj.Data.GetComponents(common.DevfileOptions{})
	if err != nil {
		return localContainers, err
	}
	for _, component := range devfileComponents {
		if component.Container == nil {
			continue
		}
		localContainers = append(localContainers, localConfigProvider.LocalContainer{
			Name: component.Name,
		})
	}
	return localContainers, nil
}

// IsUserCreatedDevfile returns the UserCreatedDevfile
func (ei *EnvInfo) IsUserCreatedDevfile() bool {
	return ei.componentSettings.UserCreatedDevfile
}

// SetUserCreatedDevfile sets the UserCreatedDevfile and writes to file
func (esi *EnvSpecificInfo) SetUserCreatedDevfile(value bool) error {
	esi.componentSettings.UserCreatedDevfile = value
	return esi.writeToFile()
}

// GetRunMode returns the RunMode, returns default if nil
func (ei *EnvInfo) GetRunMode() RUNMode {
	if ei.componentSettings.RunMode == nil {
		return DefaultRunMode
	}
	return *ei.componentSettings.RunMode
}

// SetRunMode sets the RunMode in the env file
func (esi *EnvSpecificInfo) SetRunMode(runMode RUNMode) error {
	esi.componentSettings.RunMode = &runMode
	return esi.writeToFile()
}

// GetNamespace returns component namespace
func (ei *EnvInfo) GetNamespace() string {
	return ei.componentSettings.Project
}

// GetApplication returns the application name
func (ei *EnvInfo) GetApplication() string {
	return ei.componentSettings.AppName
}

// MatchComponent matches a component information provided by a devfile component with the local env info
func (ei *EnvInfo) MatchComponent(name, app, namespace string) bool {
	return name == ei.GetName() && app == ei.GetApplication() && namespace == ei.GetNamespace()
}

// SetDevfileObj sets the devfileObj for the envinfo
func (ei *EnvInfo) SetDevfileObj(devfileObj parser.DevfileObj) {
	ei.devfileObj = devfileObj
}

// GetDevfileObj returns devfileObj of the envinfo
func (ei *EnvInfo) GetDevfileObj() parser.DevfileObj {
	return ei.devfileObj
}

// SetIsRouteSupported sets the isRouteSupported value for the envinfo
func (ei *EnvInfo) SetIsRouteSupported(isRouteSupported bool) {
	ei.isRouteSupported = isRouteSupported
}

// GetLink returns the EnvInfoLink, returns default if nil
func (ei *EnvInfo) GetLink() []EnvInfoLink {
	if ei.componentSettings.Link == nil {
		return []EnvInfoLink{}
	}
	return *ei.componentSettings.Link
}

// SearchLinkName searches for a Link with given service kind and service name
// and returns its name if found
func (ei *EnvInfo) SearchLinkName(serviceKind, serviceName string) (string, bool) {
	links := ei.GetLink()
	for _, link := range links {
		if link.ServiceKind == serviceKind && link.ServiceName == serviceName {
			return link.Name, true
		}
	}
	return "", false
}

const (
	// Name is the name of the setting controlling the component name
	Name = "Name"
	// NameDescription is the human-readable description for name setting
	NameDescription = "Set this value to user-defined component name to specify the component name"
	// Project is the name of the setting controlling the component project
	Project = "Project"
	// ProjectDescription is the human-readable description for project setting
	ProjectDescription = "Set this value to user-defined project to let the component create under the project"
	// DebugPort is the name of the setting controlling the component debug port
	DebugPort = "DebugPort"
	// DebugPortDescription s the human-readable description for debug port setting
	DebugPortDescription = "Set this value to user-defined debug port to assign the debug port to the component"
	// URL parameter
	URL = "URL"
	// URLDescription is the description of URL
	URLDescription = "URL to access the component"
	// Push parameter
	Push = "PUSH"
	// PushDescription is the description of push parameter
	PushDescription = "Push parameter is the action to write devfile commands to env.yaml"
	// Link parameter
	Link = "LINK"
	// LinkDescription is the description of Link
	LinkDescription = "Link to an Operator backed service"
)

var (
	supportedLocalParameterDescriptions = map[string]string{
		Name:      NameDescription,
		Project:   ProjectDescription,
		DebugPort: DebugPortDescription,
		URL:       URLDescription,
		Push:      PushDescription,
		Link:      LinkDescription,
	}

	lowerCaseLocalParameters = util.GetLowerCaseParameters(GetLocallySupportedParameters())
)

// FormatLocallySupportedParameters outputs supported parameters and their description
func FormatLocallySupportedParameters() (result string) {
	for _, v := range GetLocallySupportedParameters() {
		result = result + " " + v + " - " + supportedLocalParameterDescriptions[v] + "\n"
	}
	return "\nAvailable Parameter in the local Env file:\n" + result
}

func asLocallySupportedParameter(param string) (string, bool) {
	lower := strings.ToLower(param)
	return lower, lowerCaseLocalParameters[lower]
}

// GetLocallySupportedParameters returns the name of the supported global parameters
func GetLocallySupportedParameters() []string {
	return util.GetSortedKeys(supportedLocalParameterDescriptions)
}
