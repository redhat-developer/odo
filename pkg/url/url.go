package url

import (
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	routev1 "github.com/openshift/api/route/v1"
	"github.com/pkg/errors"
	applabels "github.com/redhat-developer/odo/pkg/application/labels"
	componentlabels "github.com/redhat-developer/odo/pkg/component/labels"
	"github.com/redhat-developer/odo/pkg/occlient"
	urlLabels "github.com/redhat-developer/odo/pkg/url/labels"
	"github.com/redhat-developer/odo/pkg/util"

	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// machine readable struct
type MachineURL struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              UrlSpec `json:"spec,omitempty"`
}

// PodList is a list of Pods.
type MachineUrlList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds
	// +optional
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// List of pods.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md
	Items []UrlSpec `json:"items"`
}

type UrlSpec struct {
	Name     string `json:"name,omitempty"`
	URL      string `json:"url,omitempty"`
	Protocol string `json:"-"`
	Port     int    `json:"-"`
}

// Delete deletes a UrlSpec
func Delete(client *occlient.Client, urlName string, applicationName string) error {

	// Namespace the UrlSpec name
	namespacedOpenShiftObject, err := util.NamespaceOpenShiftObject(urlName, applicationName)
	if err != nil {
		return errors.Wrapf(err, "unable to create namespaced name")
	}

	return client.DeleteRoute(namespacedOpenShiftObject)
}

// Create creates a UrlSpec
// portNumber is the target port number for the route and is -1 in case no port number is specified in which case it is automatically detected for components which expose only one service port)
func Create(client *occlient.Client, urlName string, portNumber int, componentName, applicationName string) (*UrlSpec, error) {
	labels := urlLabels.GetLabels(urlName, componentName, applicationName, false)

	serviceName, err := util.NamespaceOpenShiftObject(componentName, applicationName)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to create namespaced name")
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
	return &UrlSpec{
		Name:     route.Labels[urlLabels.URLLabel],
		URL:      route.Spec.Host,
		Protocol: getProtocol(*route),
		Port:     route.Spec.Port.TargetPort.IntValue(),
	}, nil
}

// List lists the URLs in an application. The results can further be narrowed
// down if a component name is provided, which will only list URLs for the
// given component
func List(client *occlient.Client, componentName string, applicationName string) ([]UrlSpec, error) {

	labelSelector := fmt.Sprintf("%v=%v", applabels.ApplicationLabel, applicationName)

	if componentName != "" {
		labelSelector = labelSelector + fmt.Sprintf(",%v=%v", componentlabels.ComponentLabel, componentName)
	}

	glog.V(4).Infof("Listing routes with label selector: %v", labelSelector)
	routes, err := client.ListRoutes(labelSelector)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list route names")
	}

	var urls []UrlSpec
	for _, r := range routes {
		urls = append(urls, UrlSpec{
			Name:     r.Labels[urlLabels.URLLabel],
			URL:      r.Spec.Host,
			Protocol: getProtocol(r),
			Port:     r.Spec.Port.TargetPort.IntValue(),
		})
	}

	return urls, nil
}

func getProtocol(route routev1.Route) string {
	if route.Spec.TLS != nil {
		return "https"
	}
	return "http"

}

// GetURLString returns a string representation of given url
func GetURLString(url UrlSpec) string {
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

// GetURLName returns a url name from the component name and the given port number
func GetURLName(componentName string, componentPort int) string {
	if componentPort == -1 {
		return componentName
	}
	return fmt.Sprintf("%v-%v", componentName, componentPort)
}

// GetValidPortNumber checks if the given port number is a valid component port or not
// if port number is not provided and the component is a single port component, the component port is returned
// port number is -1 if the user does not specify any port
func GetValidPortNumber(client *occlient.Client, portNumber int, componentName string, applicationName string) (int, error) {
	componentPorts, err := GetComponentServicePortNumbers(client, componentName, applicationName)
	if err != nil {
		return portNumber, errors.Wrapf(err, "unable to get exposed ports for component %s", componentName)
	}

	// port number will be -1 if the user doesn't specify any port
	if portNumber == -1 {

		switch {
		case len(componentPorts) > 1:
			return portNumber, errors.Errorf("port for the component %s is required as it exposes %d ports: %s", componentName, len(componentPorts), strings.Trim(strings.Replace(fmt.Sprint(componentPorts), " ", ",", -1), "[]"))
		case len(componentPorts) == 1:
			return componentPorts[0], nil
		default:
			return portNumber, errors.Errorf("no port is exposed by the component %s", componentName)
		}
	} else {
		for _, port := range componentPorts {
			if portNumber == port {
				return portNumber, nil
			}
		}
	}

	return portNumber, nil
}
