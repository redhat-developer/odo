package testingutil

import (
	"fmt"

	routev1 "github.com/openshift/api/route/v1"
	applabels "github.com/openshift/odo/pkg/application/labels"
	componentlabels "github.com/openshift/odo/pkg/component/labels"
	"github.com/openshift/odo/pkg/url/labels"
	"github.com/openshift/odo/pkg/version"
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
				applabels.ApplicationLabel:     applicationName,
				componentlabels.ComponentLabel: componentName,
				applabels.ManagedBy:            "odo",
				applabels.ManagerVersion:       version.VERSION,
				labels.URLLabel:                urlName,
			},
		},
		Spec: routev1.RouteSpec{
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
