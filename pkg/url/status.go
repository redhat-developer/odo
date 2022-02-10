package url

import (
	"github.com/redhat-developer/odo/pkg/localConfigProvider"

	routev1 "github.com/openshift/api/route/v1"
	"github.com/redhat-developer/odo/pkg/kclient"
)

type statusURL struct {
	name   string
	url    string
	port   int
	secure bool
	kind   string
}

func getURLsForKubernetes(client kclient.ClientInterface, lcProvider localConfigProvider.LocalConfigProvider, ignoreUnpushed bool) ([]statusURL, error) {
	var err error
	componentName := lcProvider.GetName()

	routesSupported := false

	if routesSupported, err = client.IsRouteSupported(); err != nil {
		// Fallback to Kubernetes client on error
		routesSupported = false
	}

	urlClient := NewClient(ClientOptions{
		LocalConfigProvider: lcProvider,
		Client:              client,
		IsRouteSupported:    routesSupported,
	})
	urls, err := urlClient.List()

	if err != nil {
		return nil, err
	}
	urlList := []statusURL{}

	for _, u := range urls.Items {

		// Ignore unpushed URLs, they necessarily are unreachable
		if u.Status.State != StateTypePushed && ignoreUnpushed {
			continue
		}

		var properURL, protocol string

		if u.Spec.Kind != localConfigProvider.ROUTE {
			protocol = GetProtocol(routev1.Route{}, ConvertExtensionV1IngressURLToIngress(u, componentName))
			properURL = GetURLString(protocol, "", u.Spec.Host)
		} else {
			protocol = u.Spec.Protocol
			properURL = GetURLString(protocol, u.Spec.Host, "")
		}

		statusURLVal := statusURL{
			name:   u.Name,
			url:    properURL,
			kind:   string(u.Spec.Kind),
			port:   u.Spec.Port,
			secure: protocol == "https",
		}

		urlList = append(urlList, statusURLVal)

	}

	return urlList, nil
}
