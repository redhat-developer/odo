package init

import (
	"fmt"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	color "github.com/gookit/color"
	"gopkg.in/AlecAivazis/survey.v1"
	"net/url"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
	dfutil "github.com/devfile/library/pkg/util"

	"github.com/redhat-developer/odo/pkg/catalog"
	"github.com/redhat-developer/odo/pkg/init/asker"
	"github.com/redhat-developer/odo/pkg/init/backend"
	"github.com/redhat-developer/odo/pkg/init/registry"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/preference"
	"github.com/redhat-developer/odo/pkg/segment"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
)

type InitClient struct {
	// Backends
	flagsBackend       *backend.FlagsBackend
	interactiveBackend *backend.InteractiveBackend

	// Clients
	fsys             filesystem.Filesystem
	preferenceClient preference.Client
	registryClient   registry.Client
}

func NewInitClient(fsys filesystem.Filesystem, preferenceClient preference.Client, registryClient registry.Client) *InitClient {
	return &InitClient{
		flagsBackend:       backend.NewFlagsBackend(preferenceClient),
		interactiveBackend: backend.NewInteractiveBackend(asker.NewSurveyAsker(), catalog.NewCatalogClient(fsys, preferenceClient)),
		fsys:               fsys,
		preferenceClient:   preferenceClient,
		registryClient:     registryClient,
	}
}

// Validate calls Validate method of the adequate backend
func (o *InitClient) Validate(flags map[string]string) error {
	var backend backend.InitBackend
	if len(flags) == 0 {
		backend = o.interactiveBackend
	} else {
		backend = o.flagsBackend
	}
	return backend.Validate(flags)
}

// SelectDevfile calls SelectDevfile methods of the adequate backend
func (o *InitClient) SelectDevfile(flags map[string]string) (*backend.DevfileLocation, error) {
	var backend backend.InitBackend
	if len(flags) == 0 {
		backend = o.interactiveBackend
	} else {
		backend = o.flagsBackend
	}
	return backend.SelectDevfile(flags)
}

func (o *InitClient) DownloadDevfile(devfileLocation *backend.DevfileLocation, destDir string) (string, error) {
	destDevfile := filepath.Join(destDir, "devfile.yaml")
	if devfileLocation.DevfilePath != "" {
		return destDevfile, o.downloadDirect(devfileLocation.DevfilePath, destDevfile)
	} else {
		return destDevfile, o.downloadFromRegistry(devfileLocation.DevfileRegistry, devfileLocation.Devfile, destDir)
	}
}

// downloadDirect downloads a devfile at the provided URL and saves it in dest
func (o *InitClient) downloadDirect(URL string, dest string) error {
	parsedURL, err := url.Parse(URL)
	if err != nil {
		return err
	}
	if strings.HasPrefix(parsedURL.Scheme, "http") {
		downloadSpinner := log.Spinnerf("Downloading devfile from %q", URL)
		defer downloadSpinner.End(false)
		params := dfutil.HTTPRequestParams{
			URL: URL,
		}
		devfileData, err := o.registryClient.DownloadFileInMemory(params)
		if err != nil {
			return err
		}
		err = o.fsys.WriteFile(dest, devfileData, 0644)
		if err != nil {
			return err
		}
		downloadSpinner.End(true)
	} else {
		downloadSpinner := log.Spinnerf("Copying devfile from %q", URL)
		defer downloadSpinner.End(false)
		content, err := o.fsys.ReadFile(URL)
		if err != nil {
			return err
		}
		info, err := o.fsys.Stat(URL)
		if err != nil {
			return err
		}
		err = o.fsys.WriteFile(dest, content, info.Mode().Perm())
		if err != nil {
			return err
		}
		downloadSpinner.End(true)
	}

	return nil
}

// downloadFromRegistry downloads a devfile from the provided registry and saves it in dest
// If registryName is empty, will try to download the devfile from the list of registries in preferences
func (o *InitClient) downloadFromRegistry(registryName string, devfile string, dest string) error {
	var downloadSpinner *log.Status
	var forceRegistry bool
	if registryName == "" {
		downloadSpinner = log.Spinnerf("Downloading devfile %q", devfile)
		forceRegistry = false
	} else {
		downloadSpinner = log.Spinnerf("Downloading devfile %q from registry %q", devfile, registryName)
		forceRegistry = true
	}
	defer downloadSpinner.End(false)

	registries := o.preferenceClient.RegistryList()
	var reg preference.Registry
	for _, reg = range *registries {
		if forceRegistry && reg.Name == registryName {
			err := o.registryClient.PullStackFromRegistry(reg.URL, devfile, dest, segment.GetRegistryOptions())
			if err != nil {
				return err
			}
			downloadSpinner.End(true)
			return nil
		} else if !forceRegistry {
			err := o.registryClient.PullStackFromRegistry(reg.URL, devfile, dest, segment.GetRegistryOptions())
			if err != nil {
				continue
			}
			downloadSpinner.End(true)
			return nil
		}
	}

	return fmt.Errorf("unable to find the registry with name %q", devfile)
}

// SelectStarterProject calls SelectStarterProject methods of the adequate backend
func (o *InitClient) SelectStarterProject(devfile parser.DevfileObj, flags map[string]string) (*v1alpha2.StarterProject, error) {
	var backend backend.InitBackend
	if len(flags) == 0 {
		backend = o.interactiveBackend
	} else {
		backend = o.flagsBackend
	}
	return backend.SelectStarterProject(devfile, flags)
}

func (o *InitClient) DownloadStarterProject(starter *v1alpha2.StarterProject, dest string) error {
	downloadSpinner := log.Spinnerf("Downloading starter project %q", starter.Name)
	err := o.registryClient.DownloadStarterProject(starter, "", dest, false)
	if err != nil {
		downloadSpinner.End(false)
		return err
	}
	downloadSpinner.End(true)
	return nil
}

// PersonalizeName calls PersonalizeName methods of the adequate backend
func (o *InitClient) PersonalizeName(devfile parser.DevfileObj, flags map[string]string) error {
	var backend backend.InitBackend
	if len(flags) == 0 {
		backend = o.interactiveBackend
	} else {
		backend = o.flagsBackend
	}
	err := backend.PersonalizeName(devfile, flags)
	return err
}

func (o *InitClient) PersonalizeDevfileConfig(devfileobj parser.DevfileObj) error {
	// var ports = []string{}
	var envs = map[string]string{}
	var portsMap = map[string][]string{}
	var deletePortMessage = "Delete port (container: %q): %q"
	var deleteEnvMessage = "Delete environment variable: %q"
	options := []string{
		"NOTHING - configuration is correct",
		"Add new port",
		"Add new environment variable",
	}
	components, err := devfileobj.Data.GetComponents(common.DevfileOptions{})
	if err != nil {
		return err
	}
	for _, component := range components {
		if component.Container != nil {
			for _, ep := range component.Container.Endpoints {
				if _, ok := portsMap[component.Name]; !ok {
					portsMap[component.Name] = []string{strconv.Itoa(ep.TargetPort)}
				} else {
					portsMap[component.Name] = append(portsMap[component.Name], strconv.Itoa(ep.TargetPort))
				}
				// ports = append(ports, strconv.Itoa(ep.TargetPort))
				options = append(options, fmt.Sprintf(deletePortMessage, component.Name, strconv.Itoa(ep.TargetPort)))
			}
			for _, env := range component.Container.Env {
				envs[env.Name] = env.Value
				options = append(options, fmt.Sprintf(deleteEnvMessage, env.Name))
			}
		}
	}

	var configChangeAnswer string
	for configChangeAnswer != "NOTHING - configuration is correct" {
		printConfiguration(portsMap, envs)

		configChangeQuestion := &survey.Select{
			Message: "What configuration do you want change?",
			Default: options[0],
			Options: options,
		}

		err = survey.AskOne(configChangeQuestion, &configChangeAnswer, nil)
		if err != nil {
			return err
		}

		if strings.HasPrefix(configChangeAnswer, "Delete port") {
			re := regexp.MustCompile("\"(.*?)\"")
			match := re.FindAllStringSubmatch(configChangeAnswer, -1)
			containerName, portToDelete := match[0][1], match[1][1]

			if _, ok := portsMap[containerName]; !ok && !parser.InArray(portsMap[containerName], portToDelete) {
				log.Warningf("unable to delete port %q, not found", portToDelete)
				continue
			}
			// Delete port from the devfile
			// err = RemovePort(devfileobj, portToDelete)
			err = devfileobj.RemovePorts(map[string][]string{containerName: []string{portToDelete}})
			if err != nil {
				return err
			}
			for i, port := range portsMap[containerName] {
				if port == portToDelete {
					portsMap[containerName] = append(portsMap[containerName][:i], portsMap[containerName][i+1:]...)
					break
				}
			}
			// Delete port from the options
			for i, opt := range options {
				if opt == fmt.Sprintf(deletePortMessage, containerName, portToDelete) {
					options = append(options[:i], options[i+1:]...)
					break
				}
			}
		} else if strings.HasPrefix(configChangeAnswer, "Delete environment variable") {
			re := regexp.MustCompile("\"(.*?)\"")
			match := re.FindStringSubmatch(configChangeAnswer)
			envToDelete := match[1]
			if _, ok := envs[envToDelete]; !ok {
				log.Warningf("unable to delete env %q, not found", envToDelete)
			}
			delete(envs, envToDelete)
			err = devfileobj.RemoveEnvVars([]string{envToDelete})
			if err != nil {
				return err
			}
			// Delete env from the options
			for i, opt := range options {
				if opt == fmt.Sprintf(deleteEnvMessage, envToDelete) {
					options = append(options[:i], options[i+1:]...)
					break
				}
			}
		} else if configChangeAnswer == "Add new port" {
			var containers []string
			for containerName, _ := range portsMap {
				containers = append(containers, containerName)
			}
			containerNameQuestion := &survey.Select{
				Message: "Enter container name: ",
				Options: containers,
			}
			var containerNameAnswer string
			survey.AskOne(containerNameQuestion, &containerNameAnswer, survey.Required)

			newPortQuestion := &survey.Input{
				Message: "Enter port number:",
			}
			var newPortAnswer string
			survey.AskOne(newPortQuestion, &newPortAnswer, survey.Required)

			// Ensure the newPortAnswer is not already present; otherwise it will cause a duplicate endpoint error while parsing the devfile
			if parser.InArray(portsMap[containerNameAnswer], newPortAnswer) {
				log.Warningf("Port is %q already present in container %q.", newPortAnswer, containerNameAnswer)
				continue
			}
			portsMap[containerNameAnswer] = append(portsMap[containerNameAnswer], newPortAnswer)
			err = devfileobj.SetPorts(map[string][]string{containerNameAnswer: []string{newPortAnswer}})
			if err != nil {
				return err
			}
			options = append(options, fmt.Sprintf(deletePortMessage, containerNameAnswer, newPortAnswer))
		} else if configChangeAnswer == "Add new environment variable" {
			newEnvNameQuesion := &survey.Input{
				Message: "Enter new environment variable name:",
			}
			// Ask for env name
			var newEnvNameAnswer string
			survey.AskOne(newEnvNameQuesion, &newEnvNameAnswer, survey.Required)
			newEnvValueQuestion := &survey.Input{
				Message: fmt.Sprintf("Enter value for %q environment variable:", newEnvNameAnswer),
			}

			// Ask for env value
			var newEnvValueAnswer string
			survey.AskOne(newEnvValueQuestion, &newEnvValueAnswer, survey.Required)
			envs[newEnvNameAnswer] = newEnvValueAnswer

			// Write the env to devfile
			err = devfileobj.AddEnvVars([]v1alpha2.EnvVar{
				{
					Name:  newEnvNameAnswer,
					Value: newEnvValueAnswer,
				},
			})
			if err != nil {
				return err
			}
			// Append the env to list of options
			options = append(options, fmt.Sprintf(deleteEnvMessage, newEnvNameAnswer))
		} else if configChangeAnswer == "NOTHING - configuration is correct" {
			// nothing to do
		} else {
			return fmt.Errorf("Unknown configuration selected %q", configChangeAnswer)
		}
	}
	return nil
}

func printConfiguration(portsMap map[string][]string, envs map[string]string) {
	color.New(color.Bold, color.FgGreen).Println("Current component configuration:")
	color.Greenln("Opened ports:")
	for containerName, ports := range portsMap {
		color.New(color.Bold, color.FgWhite).Printf(" - Container %q:\n", containerName)
		for _, port := range ports {
			color.New(color.FgWhite).Printf("    · %s\n", port)
		}
	}

	color.Greenln("Environment variables:")
	for key, value := range envs {
		color.New(color.Bold, color.FgWhite).Printf(" - %s = %s\n", key, value)
	}
}

func RemovePort(devfileObj parser.DevfileObj, portToRemove string) error {
	components, err := devfileObj.Data.GetComponents(common.DevfileOptions{})
	if err != nil {
		return err
	}

	for _, component := range components {
		if component.Container != nil {
			for i, ep := range component.Container.Endpoints {
				if strconv.Itoa(ep.TargetPort) == portToRemove {
					component.Container.Endpoints = append(component.Container.Endpoints[:i], component.Container.Endpoints[i+1:]...)
				}
			}
			devfileObj.Data.UpdateComponent(component)
		}
	}
	return devfileObj.WriteYamlDevfile()

}
