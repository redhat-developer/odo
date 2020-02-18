package routes

import (
	routev1 "github.com/openshift/api/route/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// Generate creates a route for event listener
func Generate() routev1.Route {
	return routev1.Route{
		TypeMeta:   createRouteTypeMeta(),
		ObjectMeta: createRouteObjectMeta("github-webhook-event-listener"),
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

func createRouteTypeMeta() metav1.TypeMeta {
	return metav1.TypeMeta{
		Kind:       "Route",
		APIVersion: "route.openshift.io/v1",
	}
}

func createRouteObjectMeta(name string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:   name,
		Labels: createRouteLabels(),
	}
}

func createRouteLabels() map[string]string {
	return map[string]string{
		"app.kubernetes.io/managed-by": "EventListener",
		"app.kubernetes.io/part-of":    "Triggers",
		"eventlistener":                "cicd-event-listener",
	}
}
