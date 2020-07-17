package envinfo

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/openshift/odo/pkg/testingutil/filesystem"

	"github.com/pkg/errors"
	"k8s.io/klog"

	"github.com/openshift/odo/pkg/util"
)

// ComponentSettings holds all component related information
type ComponentSettings struct {
	Name        string              `yaml:"Name,omitempty"`
	Namespace   string              `yaml:"Namespace,omitempty"`
	URL         *[]EnvInfoURL       `yaml:"Url,omitempty"`
	PushCommand *EnvInfoPushCommand `yaml:"PushCommand,omitempty"`

	// DebugPort controls the port used by the pod to run the debugging agent on
	DebugPort *int `yaml:"DebugPort,omitempty"`
}

// URLKind is an enum to indicate the type of the URL i.e ingress/route
type URLKind string

const (
	DOCKER          URLKind = "docker"
	INGRESS         URLKind = "ingress"
	ROUTE           URLKind = "route"
	envInfoEnvName          = "ENVINFO"
	envInfoFileName         = "env.yaml"

	// DefaultDebugPort is the default port used for debugging on remote pod
	DefaultDebugPort = 5858
)

// EnvInfoURL holds URL related information
type EnvInfoURL struct {
	// Name of the URL
	Name string `yaml:"Name,omitempty"`
	// Port number for the url of the component, required in case of components which expose more than one service port
	Port int `yaml:"Port,omitempty"`
	// Indicates if the URL should be a secure https one
	Secure bool `yaml:"Secure,omitempty"`
	// Cluster host
	Host string `yaml:"Host,omitempty"`
	// TLS secret name to create ingress to provide a secure URL
	TLSSecret string `yaml:"TLSSecret,omitempty"`
	// Exposed port number for docker container, required for local scenarios
	ExposedPort int `yaml:"ExposedPort,omitempty"`
	// Kind is the kind of the URL
	Kind URLKind `yaml:"Kind,omitempty"`
}

// EnvInfoPushCommand holds the devfile push commands for the component
type EnvInfoPushCommand struct {
	Init  string `yaml:"Init,omitempty"`
	Build string `yaml:"Build,omitempty"`
	Run   string `yaml:"Run,omitempty"`
}

// EnvInfo holds all the env specific information relavent to a specific Component.
type EnvInfo struct {
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
	Filename          string `yaml:"FileName,omitempty"`
	fs                filesystem.Filesystem
	EnvInfo           `yaml:",omitempty"`
	envinfoFileExists bool
}

func getEnvInfoFile(envDir string) (string, error) {
	if env, ok := os.LookupEnv(envInfoEnvName); ok {
		return env, nil
	}

	if envDir == "" {
		var err error
		envDir, err = os.Getwd()
		if err != nil {
			return "", err
		}
	}

	return filepath.Join(envDir, ".odo", "env", envInfoFileName), nil
}

// New returns the EnvSpecificInfo
func New() (*EnvSpecificInfo, error) {
	return NewEnvSpecificInfo("")
}

// NewEnvSpecificInfo gets the EnvSpecificInfo from envinfo file and creates the envinfo file in case it's
// not present then it
func NewEnvSpecificInfo(envDir string) (*EnvSpecificInfo, error) {
	return newEnvSpecificInfo(envDir, filesystem.DefaultFs{})
}

func newEnvSpecificInfo(envDir string, fs filesystem.Filesystem) (*EnvSpecificInfo, error) {
	envInfoFile, err := getEnvInfoFile(envDir)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get odo envinfo file")
	}
	e := EnvSpecificInfo{
		EnvInfo:           NewEnvInfo(),
		Filename:          envInfoFile,
		envinfoFileExists: true,
		fs:                fs,
	}

	// if the env.yaml file doesn't exist then we dont worry about it and return
	if _, err = e.fs.Stat(envInfoFile); os.IsNotExist(err) {
		e.envinfoFileExists = false
		return &e, nil
	}

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

// SetConfiguration sets the environment specific info like cluster host etc.
func (esi *EnvSpecificInfo) SetConfiguration(parameter string, value interface{}) (err error) {
	if parameter, ok := asLocallySupportedParameter(parameter); ok {
		switch parameter {
		case "url":
			urlValue := value.(EnvInfoURL)
			if esi.componentSettings.URL != nil {
				*esi.componentSettings.URL = append(*esi.componentSettings.URL, urlValue)
			} else {
				esi.componentSettings.URL = &[]EnvInfoURL{urlValue}
			}
		case "push":
			pushCommandValue := value.(EnvInfoPushCommand)
			esi.componentSettings.PushCommand = &pushCommandValue
		}

		return esi.writeToFile()
	}
	return errors.Errorf("unknown parameter :'%s' is not a parameter in envinfo", parameter)

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

// EnvInfoFileExists if the envinfo file exists or not
func (esi *EnvSpecificInfo) EnvInfoFileExists() bool {
	return esi.envinfoFileExists
}

// DeleteConfiguration is used to delete environment specific info from local odo envinfo
func (esi *EnvSpecificInfo) DeleteConfiguration(parameter string) error {
	if parameter, ok := asLocallySupportedParameter(parameter); ok {

		switch parameter {
		default:
			if err := util.DeleteConfiguration(&esi.componentSettings, parameter); err != nil {
				return err
			}
		}
		return esi.writeToFile()
	}
	return errors.Errorf("unknown parameter :'%s' is not a parameter in envinfo", parameter)

}

// DeleteURL is used to delete environment specific info for url from envinfo
func (esi *EnvSpecificInfo) DeleteURL(parameter string) error {
	for i, url := range *esi.componentSettings.URL {
		if url.Name == parameter {
			s := *esi.componentSettings.URL
			s = append(s[:i], s[i+1:]...)
			esi.componentSettings.URL = &s
		}
	}
	return esi.writeToFile()
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

// GetURL returns the EnvInfoURL, returns default if nil
func (ei *EnvInfo) GetURL() []EnvInfoURL {
	if ei.componentSettings.URL == nil {
		return []EnvInfoURL{}
	}
	return *ei.componentSettings.URL
}

// GetPushCommand returns the EnvInfoPushCommand, returns default if nil
func (ei *EnvInfo) GetPushCommand() EnvInfoPushCommand {
	if ei.componentSettings.PushCommand == nil {
		return EnvInfoPushCommand{}
	}
	return *ei.componentSettings.PushCommand
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

// GetPortByURLKind returns the Port of a specific URL type, returns 0 if nil
func (ei *EnvInfo) GetPortByURLKind(urlKind URLKind) (int, error) {
	for _, localURL := range ei.GetURL() {
		if localURL.Kind == urlKind {
			return localURL.Port, nil
		}
	}
	return 0, errors.New(fmt.Sprintf("unable to find port for URL of kind: '%s'", urlKind))
}

// GetNamespace returns component namespace
func (ei *EnvInfo) GetNamespace() string {
	return ei.componentSettings.Namespace
}

const (
	// URL parameter
	URL = "URL"
	// URLDescription is the description of URL
	URLDescription = "URL to access the component"
	// Push parameter
	Push = "PUSH"
	// PushDescription is the description of URL
	PushDescription = "Push parameter is the action to write devfile commands to env.yaml"
)

var (
	supportedLocalParameterDescriptions = map[string]string{
		URL:  URLDescription,
		Push: PushDescription,
	}

	lowerCaseLocalParameters = util.GetLowerCaseParameters(GetLocallySupportedParameters())
)

// FormatLocallySupportedParameters outputs supported parameters and their description
func FormatLocallySupportedParameters() (result string) {
	for _, v := range GetLocallySupportedParameters() {
		result = result + v + " - " + supportedLocalParameterDescriptions[v] + "\n"
	}
	return "\nAvailable Local Parameters:\n" + result
}

func asLocallySupportedParameter(param string) (string, bool) {
	lower := strings.ToLower(param)
	return lower, lowerCaseLocalParameters[lower]
}

// GetLocallySupportedParameters returns the name of the supported global parameters
func GetLocallySupportedParameters() []string {
	return util.GetSortedKeys(supportedLocalParameterDescriptions)
}
