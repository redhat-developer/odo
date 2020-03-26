package routes

import (
	routev1 "github.com/openshift/api/route/v1"
	"github.com/openshift/odo/pkg/pipelines/meta"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var (
	routeTypeMeta = meta.TypeMeta("Route", "route.openshift.io/v1")
)

func Generate(ns string) routev1.Route {
	return routev1.Route{
		TypeMeta:   routeTypeMeta,
		ObjectMeta: meta.ObjectMeta(meta.NamespacedName(ns, "gitops-webhook-event-listener-route")),
		Spec: routev1.RouteSpec{
			To: creatRouteTargetReference(
				"Service",
				"el-cicd-event-listener",
				100,
			),
			Port:           createRoutePort(8080),
			WildcardPolicy: routev1.WildcardPolicyNone,
		},
	}
}

func createRoutePort(port int32) *routev1.RoutePort {
	return &routev1.RoutePort{
		TargetPort: intstr.IntOrString{
			IntVal: 8080,
		},
	}
}

func creatRouteTargetReference(kind string, name string, weight int32) routev1.RouteTargetReference {
	return routev1.RouteTargetReference{
		Kind:   kind,
		Name:   name,
		Weight: &weight,
	}
}
