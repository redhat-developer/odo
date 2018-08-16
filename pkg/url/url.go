package url

import (
	"fmt"

	routev1 "github.com/openshift/api/route/v1"
	"github.com/pkg/errors"
	applabels "github.com/redhat-developer/odo/pkg/application/labels"
	componentlabels "github.com/redhat-developer/odo/pkg/component/labels"
	"github.com/redhat-developer/odo/pkg/occlient"
	urlLabels "github.com/redhat-developer/odo/pkg/url/labels"
	"github.com/redhat-developer/odo/pkg/util"

	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/intstr"
	"strings"
)

type URL struct {
	Name     string
	URL      string
	Protocol string
	Port     intstr.IntOrString
}

// Delete deletes a URL
func Delete(client *occlient.Client, urlName string, applicationName string) error {

	// Namespace the URL name
	namespacedOpenShiftObject, err := util.NamespaceOpenShiftObject(urlName, applicationName)
	if err != nil {
		return errors.Wrapf(err, "unable to create namespaced name")
	}

	return client.DeleteRoute(namespacedOpenShiftObject)
}

// Create creates a URL
// portNumber is the target port number for the route and is -1 in case no port number is specified
func Create(client *occlient.Client, urlName string, portNumber int, componentName, applicationName string) (*URL, error) {
	labels := urlLabels.GetLabels(urlName, componentName, applicationName, false)

	serviceName, err := util.NamespaceOpenShiftObject(componentName, applicationName)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to create namespaced name")
	}

	componentPorts, err := GetComponentServicePortNumbers(client, componentName, applicationName)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get component exposed ports for component %s", componentName)
	}

	var portFound bool

	if portNumber == -1 {
		if len(componentPorts) > 1 {
			return nil, errors.Errorf("'port' is required as the component %s exposes %d ports: %s", componentName, len(componentPorts), strings.Trim(strings.Replace(fmt.Sprint(componentPorts), " ", ",", -1), "[]"))
		} else if len(componentPorts) == 1 {
			portNumber = componentPorts[0]
		} else {
			return nil, errors.Errorf("no port is exposed by the component %s", componentName)
		}
	} else {
		for _, port := range componentPorts {
			if portNumber == port {
				portFound = true
			}
		}

		if !portFound {
			return nil, errors.Errorf("port %d is not exposed by the component", portNumber)
		}
	}

	urlName, err = util.NamespaceOpenShiftObject(urlName, applicationName)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to create namespaced name")
	}

	// Pass in the namespace name, link to the service (componentName) and labels to create a route
	route, err := client.CreateRoute(urlName, serviceName, intstr.FromInt(portNumber), labels)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create route")
	}
	return &URL{
		Name:     route.Labels[urlLabels.UrlLabel],
		URL:      route.Spec.Host,
		Protocol: getProtocol(*route),
		Port:     route.Spec.Port.TargetPort,
	}, nil
}

// List lists the URLs in an application. The results can further be narrowed
// down if a component name is provided, which will only list URLs for the
// given component
func List(client *occlient.Client, componentName string, applicationName string) ([]URL, error) {

	labelSelector := fmt.Sprintf("%v=%v", applabels.ApplicationLabel, applicationName)

	if componentName != "" {
		labelSelector = labelSelector + fmt.Sprintf(",%v=%v", componentlabels.ComponentLabel, componentName)
	}

	log.Debugf("Listing routes with label selector: %v", labelSelector)
	routes, err := client.ListRoutes(labelSelector)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list route names")
	}

	var urls []URL
	for _, r := range routes {
		urls = append(urls, URL{
			Name:     r.Labels[urlLabels.UrlLabel],
			URL:      r.Spec.Host,
			Protocol: getProtocol(r),
			Port:     r.Spec.Port.TargetPort,
		})
	}

	return urls, nil
}

func getProtocol(route routev1.Route) string {
	if route.Spec.TLS != nil {
		return "https"
	} else {
		return "http"
	}
}

func GetUrlString(url URL) string {
	return url.Protocol + "://" + url.URL
}

// Exists checks if the url exists in the component or not
// urlName is the name of the url for checking
// componentName is the name of the component to which the url's existence is checked
// applicationName is the name of the application to which the url's existence is checked
func Exists(client *occlient.Client, urlName string, componentName string, applicationName string) (bool, error) {
	urls, err := List(client, componentName, applicationName)
	if err != nil {
		return false, errors.Wrap(err, "unable to list the urls")
	}

	for _, url := range urls {
		if url.Name == urlName {
			return true, nil
		}
	}
	return false, nil
}

// GetComponentServicePortNumbers returns the port numbers exposed by the service of the component
// componentName is the name of the component
// applicationName is the name of the application
func GetComponentServicePortNumbers(client *occlient.Client, componentName string, applicationName string) ([]int, error) {
	componentLabels := componentlabels.GetLabels(componentName, applicationName, false)
	componentSelector := util.ConvertLabelsToSelector(componentLabels)

	services, err := client.GetServicesFromSelector(componentSelector)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get the service")
	}

	var ports []int

	for _, service := range services {
		for _, port := range service.Spec.Ports {
			ports = append(ports, int(port.Port))
		}
	}

	return ports, nil
}
