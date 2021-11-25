package occlient

import (
	"context"
	"fmt"

	v1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/generator"
	routev1 "github.com/openshift/api/route/v1"
	"github.com/pkg/errors"
	"github.com/redhat-developer/odo/pkg/kclient"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/klog"
)

// IsRouteSupported checks if route resource type is present on the cluster
func (c *Client) IsRouteSupported() (bool, error) {

	return c.GetKubeClient().IsResourceSupported("route.openshift.io", "v1", "routes")
}

// GetRoute gets the route with the given name
func (c *Client) GetRoute(name string) (*routev1.Route, error) {
	return c.routeClient.Routes(c.Namespace).Get(context.TODO(), name, metav1.GetOptions{})
}

// CreateRoute creates a route object for the given service and with the given labels
// serviceName is the name of the service for the target reference
// portNumber is the target port of the route
// path is the path of the endpoint URL
// secureURL indicates if the route is a secure one or not
func (c *Client) CreateRoute(name string, serviceName string, portNumber intstr.IntOrString, labels map[string]string, secureURL bool, path string, ownerReference metav1.OwnerReference) (*routev1.Route, error) {
	routeParams := generator.RouteParams{
		ObjectMeta: generator.GetObjectMeta(name, c.Namespace, labels, nil),
		RouteSpecParams: generator.RouteSpecParams{
			ServiceName: serviceName,
			PortNumber:  portNumber,
			Secure:      secureURL,
			Path:        path,
		},
	}

	route := generator.GetRoute(v1.Endpoint{}, routeParams)

	route.SetOwnerReferences(append(route.GetOwnerReferences(), ownerReference))

	r, err := c.routeClient.Routes(c.Namespace).Create(context.TODO(), route, metav1.CreateOptions{FieldManager: kclient.FieldManager})
	if err != nil {
		return nil, errors.Wrap(err, "error creating route")
	}
	return r, nil
}

// DeleteRoute deleted the given route
func (c *Client) DeleteRoute(name string) error {
	err := c.routeClient.Routes(c.Namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		return errors.Wrap(err, "unable to delete route")
	}
	return nil
}

// ListRoutes lists all the routes based on the given label selector
func (c *Client) ListRoutes(labelSelector string) ([]routev1.Route, error) {
	klog.V(3).Infof("Listing routes with label selector: %v", labelSelector)
	routeList, err := c.routeClient.Routes(c.Namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get route list")
	}

	return routeList.Items, nil
}

// GetOneRouteFromSelector gets one route with the given selector
// if no or multiple routes are found with the given selector, it throws an error
func (c *Client) GetOneRouteFromSelector(selector string) (*routev1.Route, error) {
	routes, err := c.ListRoutes(selector)
	if err != nil {
		return nil, err
	}

	if num := len(routes); num == 0 {
		return nil, fmt.Errorf("no ingress was found for the selector: %v", selector)
	} else if num > 1 {
		return nil, fmt.Errorf("multiple ingresses exist for the selector: %v. Only one must be present", selector)
	}

	return &routes[0], nil
}
