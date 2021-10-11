package envinfo

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	"github.com/openshift/odo/pkg/localConfigProvider"
	"github.com/openshift/odo/pkg/odo/util/validation"
	"github.com/openshift/odo/pkg/util"
	"github.com/pkg/errors"
)

//getPorts gets the ports from devfile
func (ei *EnvInfo) getPorts(container string) ([]string, error) {
	var portList []string
	containerComponents, err := ei.devfileObj.Data.GetDevfileContainerComponents(common.DevfileOptions{})
	if err != nil {
		return nil, err
	}
	containerExists := false
	portMap := make(map[string]bool)
	for _, component := range containerComponents {
		if container == "" || container == component.Name {
			containerExists = true
			for _, endpoint := range component.Container.Endpoints {
				portMap[strconv.FormatInt(int64(endpoint.TargetPort), 10)] = true
			}
		}
	}
	if !containerExists {
		return portList, fmt.Errorf("the container specified: %s does not exist in devfile", container)
	}
	for port := range portMap {
		portList = append(portList, port)
	}
	sort.Strings(portList)
	return portList, nil
}

//GetContainerPorts returns list of the ports of specified container, if it exists
func (ei *EnvInfo) GetContainerPorts(container string) ([]string, error) {
	if container == "" {
		return nil, fmt.Errorf("please provide a container")
	}
	return ei.getPorts(container)
}

//GetComponentPorts returns all unique ports declared in all the containers
func (ei *EnvInfo) GetComponentPorts() ([]string, error) {
	return ei.getPorts("")
}

//checkValidPort checks and retrieves valid port from devfile when no port is specified
func (ei *EnvInfo) checkValidPort(url *localConfigProvider.LocalURL, portsOf string, ports []string) (err error) {
	if url.Port == -1 {
		if len(ports) > 1 {
			return fmt.Errorf("port for the %s is required as it exposes %d ports: %s", portsOf, len(ports), strings.Trim(strings.Replace(fmt.Sprint(ports), " ", ",", -1), "[]"))
		} else if len(ports) <= 0 {
			return fmt.Errorf("no port is exposed by the %s, please specify a port", portsOf)
		} else {
			url.Port, err = strconv.Atoi(strings.Split(ports[0], "/")[0])
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// CompleteURL completes the given URL with default values
func (ei *EnvInfo) CompleteURL(url *localConfigProvider.LocalURL) error {
	if url.Kind == "" {
		if !ei.isRouteSupported {
			url.Kind = localConfigProvider.INGRESS
		} else {
			url.Kind = localConfigProvider.ROUTE
		}
	}

	if len(url.Path) > 0 && (strings.HasPrefix(url.Path, "/") || strings.HasPrefix(url.Path, "\\")) {
		if len(url.Path) <= 1 {
			url.Path = ""
		} else {
			// remove the leading / or \ from provided path
			url.Path = string([]rune(url.Path)[1:])
		}
	}
	// add leading / to path, if the path provided is empty, it will be set to / which is the default value of path
	url.Path = "/" + url.Path

	// get the port if not provided
	var ports []string
	var err error
	if url.Container == "" {
		ports, err = ei.GetComponentPorts()
		if err != nil {
			return err
		}
		err = ei.checkValidPort(url, fmt.Sprintf("component %s", ei.GetName()), ports)
		if err != nil {
			return err
		}
	} else {
		ports, err = ei.GetContainerPorts(url.Container)
		if err != nil {
			return err
		}
		err = ei.checkValidPort(url, fmt.Sprintf("container %s", url.Container), ports)
		if err != nil {
			return err
		}
	}

	// get the name for the URL if not provided
	if len(url.Name) == 0 {
		foundURL, err := findInvalidEndpoint(ei, url.Port)
		if err != nil {
			return err
		}

		if foundURL.Name != "" {
			// found an URL that can be overridden or more info can be added to it
			url.Name = foundURL.Name
			ei.updateURL = true
		} else {
			url.Name = util.GetURLName(ei.GetName(), url.Port)
		}
	}

	containerComponents, err := ei.devfileObj.Data.GetDevfileContainerComponents(common.DevfileOptions{})
	if err != nil {
		return err
	}

	if len(containerComponents) == 0 {
		return fmt.Errorf("no valid components found in the devfile")
	}

	// if a container name is provided for the URL, return
	if len(url.Container) > 0 {
		return nil
	}

	containerPortMap := make(map[int]string)
	portMap := make(map[string]bool)

	// if a container name for the URL is not provided
	// use a container which uses the given URL port in one of it's endpoints
	for _, component := range containerComponents {
		for _, endpoint := range component.Container.Endpoints {
			containerPortMap[endpoint.TargetPort] = component.Name
			portMap[strconv.FormatInt(int64(endpoint.TargetPort), 10)] = true
		}
	}
	if containerName, exist := containerPortMap[url.Port]; exist {
		url.Container = containerName
	}

	// container is not provided, or the specified port is not being used under any containers
	// pick the first container to store the new endpoint
	if len(url.Container) == 0 {
		url.Container = containerComponents[0].Name
	}

	return nil
}

// ValidateURL validates the given URL
func (ei *EnvInfo) ValidateURL(url localConfigProvider.LocalURL) error {

	foundContainer := false
	containerComponents, err := ei.devfileObj.Data.GetDevfileContainerComponents(common.DevfileOptions{})
	if err != nil {
		return err
	}

	if len(containerComponents) == 0 {
		return fmt.Errorf("no valid components found in the devfile")
	}

	// map TargetPort with containerName
	containerPortMap := make(map[int]string)
	for _, component := range containerComponents {
		if len(url.Container) > 0 && !foundContainer {
			if component.Name == url.Container {
				foundContainer = true
			}
		}
		for _, endpoint := range component.Container.Endpoints {
			if endpoint.Name == url.Name && !ei.updateURL {
				return fmt.Errorf("url %v already exist in devfile endpoint entry under container %v", url.Name, component.Name)
			}
			containerPortMap[endpoint.TargetPort] = component.Name
		}
	}

	if len(url.Container) > 0 && !foundContainer {
		return fmt.Errorf("the container specified: %v does not exist in devfile", url.Container)
	}
	if containerName, exist := containerPortMap[url.Port]; exist {
		if len(url.Container) > 0 && url.Container != containerName {
			return fmt.Errorf("cannot set URL %v under container %v, TargetPort %v is being used under container %v", url.Name, url.Container, url.Port, containerName)
		}
	}

	errorList := make([]string, 0)
	if url.TLSSecret != "" && (url.Kind != localConfigProvider.INGRESS || !url.Secure) {
		errorList = append(errorList, "TLS secret is only available for secure URLs of Ingress kind")
	}

	// check if a host is provided for route based URLs
	if len(url.Host) > 0 {
		if url.Kind == localConfigProvider.ROUTE {
			errorList = append(errorList, "host is not supported for URLs of Route Kind")
		}
		if err := validation.ValidateHost(url.Host); err != nil {
			errorList = append(errorList, err.Error())
		}
	} else if url.Kind == localConfigProvider.INGRESS {
		errorList = append(errorList, "host must be provided in order to create URLS of Ingress Kind")
	}

	// check the protocol of the URL
	if len(url.Protocol) > 0 {
		switch strings.ToLower(url.Protocol) {
		case string(devfilev1.HTTPEndpointProtocol):
			break
		case string(devfilev1.HTTPSEndpointProtocol):
			break
		case string(devfilev1.WSEndpointProtocol):
			break
		case string(devfilev1.WSSEndpointProtocol):
			break
		case string(devfilev1.TCPEndpointProtocol):
			break
		case string(devfilev1.UDPEndpointProtocol):
			break
		default:
			errorList = append(errorList, fmt.Sprintf("endpoint protocol only supports %v|%v|%v|%v|%v|%v", devfilev1.HTTPEndpointProtocol, devfilev1.HTTPSEndpointProtocol, devfilev1.WSSEndpointProtocol, devfilev1.WSEndpointProtocol, devfilev1.TCPEndpointProtocol, devfilev1.UDPEndpointProtocol))
		}
	}

	if !ei.updateURL {
		urls, err := ei.ListURLs()
		if err != nil {
			return err
		}
		for _, localURL := range urls {
			if url.Name == localURL.Name {
				errorList = append(errorList, fmt.Sprintf("URL %s already exists", url.Name))
			}
		}
	}

	if len(errorList) > 0 {
		return fmt.Errorf(strings.Join(errorList, "\n"))
	}
	return nil
}

// GetURL gets the given url from the env.yaml and devfile
func (ei *EnvInfo) GetURL(name string) (*localConfigProvider.LocalURL, error) {
	urls, err := ei.ListURLs()
	if err != nil {
		return nil, err
	}
	for _, url := range urls {
		if name == url.Name {
			return &url, nil
		}
	}
	return nil, nil
}

// CreateURL write the given url to the env.yaml and devfile
func (esi *EnvSpecificInfo) CreateURL(url localConfigProvider.LocalURL) error {

	if !esi.updateURL {
		newEndpointEntry := devfilev1.Endpoint{
			Name:       url.Name,
			Path:       url.Path,
			Secure:     util.GetBoolPtr(url.Secure),
			Exposure:   devfilev1.PublicEndpointExposure,
			TargetPort: url.Port,
			Protocol:   devfilev1.EndpointProtocol(strings.ToLower(url.Protocol)),
		}

		err := addEndpointInDevfile(esi.devfileObj, newEndpointEntry, url.Container)
		if err != nil {
			return errors.Wrapf(err, "failed to write endpoints information into devfile")
		}
	} else {
		err := updateEndpointInDevfile(esi.devfileObj, url)
		if err != nil {
			return err
		}
	}

	err := esi.SetConfiguration("url", localConfigProvider.LocalURL{Name: url.Name, Host: url.Host, TLSSecret: url.TLSSecret, Kind: url.Kind})
	if err != nil {
		return errors.Wrapf(err, "failed to persist the component settings to env file")
	}
	return nil
}

// ListURLs returns the urls from the env and devfile, returns default if nil
func (ei *EnvInfo) ListURLs() ([]localConfigProvider.LocalURL, error) {

	envMap := make(map[string]localConfigProvider.LocalURL)
	if ei.componentSettings.URL != nil {
		for _, url := range *ei.componentSettings.URL {
			envMap[url.Name] = url
		}
	}

	var urls []localConfigProvider.LocalURL

	if ei.devfileObj.Data == nil {
		return urls, nil
	}

	devfileComponents, err := ei.devfileObj.Data.GetDevfileContainerComponents(common.DevfileOptions{})
	if err != nil {
		return urls, err
	}
	for _, comp := range devfileComponents {
		for _, localEndpoint := range comp.Container.Endpoints {
			// only exposed endpoint will be shown as a URL in `odo url list`
			if localEndpoint.Exposure == devfilev1.NoneEndpointExposure || localEndpoint.Exposure == devfilev1.InternalEndpointExposure {
				continue
			}

			path := "/"
			if localEndpoint.Path != "" {
				path = localEndpoint.Path
			}

			secure := false
			if util.SafeGetBool(localEndpoint.Secure) || localEndpoint.Protocol == "https" || localEndpoint.Protocol == "wss" {
				secure = true
			}

			url := localConfigProvider.LocalURL{
				Name:      localEndpoint.Name,
				Port:      localEndpoint.TargetPort,
				Secure:    secure,
				Path:      path,
				Container: comp.Name,
			}

			if envInfoURL, exist := envMap[localEndpoint.Name]; exist {
				url.Host = envInfoURL.Host
				url.TLSSecret = envInfoURL.TLSSecret
				url.Kind = envInfoURL.Kind
			} else {
				url.Kind = localConfigProvider.ROUTE
			}

			urls = append(urls, url)
		}
	}

	return urls, nil
}

// DeleteURL is used to delete environment specific info for url from envinfo and devfile
func (esi *EnvSpecificInfo) DeleteURL(name string) error {
	err := removeEndpointInDevfile(esi.devfileObj, name)
	if err != nil {
		return errors.Wrap(err, "failed to delete URL")
	}

	if esi.componentSettings.URL == nil {
		return nil
	}
	for i, url := range *esi.componentSettings.URL {
		if url.Name == name {
			s := *esi.componentSettings.URL
			s = append(s[:i], s[i+1:]...)
			esi.componentSettings.URL = &s
		}
	}
	return esi.writeToFile()
}

// addEndpointInDevfile writes the provided endpoint information into devfile
func addEndpointInDevfile(devObj parser.DevfileObj, endpoint devfilev1.Endpoint, container string) error {
	components, err := devObj.Data.GetComponents(common.DevfileOptions{})
	if err != nil {
		return err
	}
	for _, component := range components {
		if component.Container != nil && component.Name == container {
			component.Container.Endpoints = append(component.Container.Endpoints, endpoint)
			err := devObj.Data.UpdateComponent(component)
			if err != nil {
				return err
			}
			break
		}
	}
	return devObj.WriteYamlDevfile()
}

// removeEndpointInDevfile deletes the specific endpoint information from devfile
func removeEndpointInDevfile(devObj parser.DevfileObj, urlName string) error {
	found := false
	devfileComponents, err := devObj.Data.GetDevfileContainerComponents(common.DevfileOptions{})
	if err != nil {
		return err
	}

	for _, component := range devfileComponents {
		for index, enpoint := range component.Container.Endpoints {
			if enpoint.Name == urlName {
				component.Container.Endpoints = append(component.Container.Endpoints[:index], component.Container.Endpoints[index+1:]...)
				err := devObj.Data.UpdateComponent(component)
				if err != nil {
					return err
				}
				found = true
				break
			}
		}
		if found {
			break
		}
	}
	if !found {
		return fmt.Errorf("the URL %s does not exist", urlName)
	}
	return devObj.WriteYamlDevfile()
}

// updateEndpointInDevfile updates the endpoint of the given URL in the devfile
func updateEndpointInDevfile(devObj parser.DevfileObj, url localConfigProvider.LocalURL) error {
	components, err := devObj.Data.GetComponents(common.DevfileOptions{})
	if err != nil {
		return err
	}
	for _, component := range components {
		if component.Container != nil && component.Name == url.Container {
			for j := range component.ComponentUnion.Container.Endpoints {
				endpoint := component.ComponentUnion.Container.Endpoints[j]

				if endpoint.Name == url.Name {
					// fill the default values
					if endpoint.Exposure == "" {
						endpoint.Exposure = devfilev1.PublicEndpointExposure
					}
					if endpoint.Path == "" {
						endpoint.Path = "/"
					}
					if endpoint.Protocol == "" {
						endpoint.Protocol = devfilev1.HTTPEndpointProtocol
					}

					// prevent write unless required
					if endpoint.Exposure != devfilev1.PublicEndpointExposure || url.Secure != util.SafeGetBool(endpoint.Secure) ||
						url.Path != endpoint.Path || url.Protocol != string(endpoint.Protocol) {
						endpoint = devfilev1.Endpoint{
							Name:       url.Name,
							Path:       url.Path,
							Secure:     &url.Secure,
							Exposure:   devfilev1.PublicEndpointExposure,
							TargetPort: url.Port,
							Protocol:   devfilev1.EndpointProtocol(strings.ToLower(url.Protocol)),
						}
						component.ComponentUnion.Container.Endpoints[j] = endpoint
						err := devObj.Data.UpdateComponent(component)
						if err != nil {
							return err
						}
						return devObj.WriteYamlDevfile()
					}
					return nil
				}
			}
		}
	}
	return fmt.Errorf("url %s not found for updating", url.Name)
}

// findInvalidEndpoint finds the URLs which are invalid for the current cluster e.g
// route urls on a vanilla k8s based cluster
// urls with no host information on a vanilla k8s based cluster
func findInvalidEndpoint(ei *EnvInfo, port int) (localConfigProvider.LocalURL, error) {
	urls, err := ei.ListURLs()
	if err != nil {
		return localConfigProvider.LocalURL{}, err
	}
	for _, url := range urls {
		if url.Kind == localConfigProvider.ROUTE && url.Port == port && !ei.isRouteSupported {
			return url, nil
		}
	}
	return localConfigProvider.LocalURL{}, nil
}
