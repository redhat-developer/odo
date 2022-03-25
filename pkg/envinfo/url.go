package envinfo

import (
	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	"github.com/redhat-developer/odo/pkg/localConfigProvider"
	"github.com/redhat-developer/odo/pkg/util"
)

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
