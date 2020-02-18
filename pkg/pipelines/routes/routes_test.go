package routes

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	routev1 "github.com/openshift/api/route/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestGenerateRoute(t *testing.T) {
	var weight int32
	weight = 100
	validRoute := routev1.Route{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Route",
			APIVersion: "route.openshift.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "github-webhook-event-listener",
			Labels: map[string]string{
				"app.kubernetes.io/managed-by": "EventListener",
				"app.kubernetes.io/part-of":    "Triggers",
				"eventlistener":                "cicd-event-listener",
			},
		},
		Spec: routev1.RouteSpec{
			To: routev1.RouteTargetReference{
				Kind:   "Service",
				Name:   "el-cicd-event-listener",
				Weight: &weight,
			},
			Port: &routev1.RoutePort{
				TargetPort: intstr.IntOrString{
					IntVal: 8080,
				},
			},
			WildcardPolicy: routev1.WildcardPolicyNone,
		},
	}
	route := Generate()
	if diff := cmp.Diff(validRoute, route); diff != "" {
		t.Fatalf("Generate() failed:\n%s", diff)
	}
}

func TestCreateRoutePort(t *testing.T) {
	validRoutePort := &routev1.RoutePort{
		TargetPort: intstr.IntOrString{
			IntVal: 8080,
		},
	}
	routePort := createRoutePort(8080)
	if diff := cmp.Diff(routePort, validRoutePort); diff != "" {
		t.Fatalf("createRoutePort() failed:\n%s", diff)
	}
}

func TestCreatRouteTargetReference(t *testing.T) {
	var weight int32
	weight = 100
	validRouteTargetReference := routev1.RouteTargetReference{
		Kind:   "Service",
		Name:   "el-cicd-event-listener",
		Weight: &weight,
	}
	routeTargetReference := creatRouteTargetReference("Service", "el-cicd-event-listener", 100)
	if diff := cmp.Diff(validRouteTargetReference, routeTargetReference); diff != "" {
		t.Fatalf("creatRouteTargetReference() failed:\n%s", diff)
	}
}

func TestCreateRouteObjectMeta(t *testing.T) {
	validObjectMeta := metav1.ObjectMeta{
		Name: "sampleName",
		Labels: map[string]string{
			"app.kubernetes.io/managed-by": "EventListener",
			"app.kubernetes.io/part-of":    "Triggers",
			"eventlistener":                "cicd-event-listener",
		},
	}
	objectMeta := createRouteObjectMeta("sampleName")
	if diff := cmp.Diff(validObjectMeta, objectMeta); diff != "" {
		t.Fatalf("createRouteObjectMeta() failed:\n%s", diff)
	}
}
