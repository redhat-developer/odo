package testingutil

import (
	"fmt"

	v1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/generator"
	routev1 "github.com/openshift/api/route/v1"
	applabels "github.com/redhat-developer/odo/pkg/application/labels"
	componentlabels "github.com/redhat-developer/odo/pkg/component/labels"
	"github.com/redhat-developer/odo/pkg/url/labels"
	"github.com/redhat-developer/odo/pkg/version"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func GetRouteListWithMultiple(componentName, applicationName string) *routev1.RouteList {
	return &routev1.RouteList{
		Items: []routev1.Route{
			GetSingleRoute("example", 8080, componentName, applicationName),
			GetSingleRoute("example-1", 9100, componentName, applicationName),
		},
	}
}

func GetSingleRoute(urlName string, port int, componentName, applicationName string) routev1.Route {
	return routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name: urlName,
			Labels: map[string]string{
				applabels.ApplicationLabel:                       applicationName,
				componentlabels.ComponentKubernetesInstanceLabel: componentName,
				applabels.ManagedBy:                              "odo",
				applabels.ManagerVersion:                         version.VERSION,
				labels.URLLabel:                                  urlName,
			},
		},
		Spec: routev1.RouteSpec{
			Host: "example.com",
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: fmt.Sprintf("%s-%s", componentName, applicationName),
			},
			Port: &routev1.RoutePort{
				TargetPort: intstr.FromInt(port),
			},
			Path: "/",
		},
	}
}

// GetSingleSecureRoute returns a secure route generated with the given parameters
func GetSingleSecureRoute(urlName string, port int, componentName, applicationName string) routev1.Route {
	generatedRoute := *generator.GetRoute(v1.Endpoint{}, generator.RouteParams{
		ObjectMeta: metav1.ObjectMeta{
			Name: urlName,
			Labels: map[string]string{
				applabels.ApplicationLabel:                       applicationName,
				componentlabels.ComponentKubernetesInstanceLabel: componentName,
				applabels.ManagedBy:                              "odo",
				applabels.ManagerVersion:                         version.VERSION,
				labels.URLLabel:                                  urlName,
				applabels.App:                                    applicationName,
			},
		},
		RouteSpecParams: generator.RouteSpecParams{
			ServiceName: componentName,
			PortNumber:  intstr.FromInt(port),
			Secure:      true,
		},
	})
	generatedRoute.Spec.Host = "example.com"
	return generatedRoute
}
