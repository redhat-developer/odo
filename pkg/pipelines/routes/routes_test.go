package routes

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	routev1 "github.com/openshift/api/route/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestGenerate(t *testing.T) {
	want := map[string]interface{}{
		"apiVersion": "route.openshift.io/v1",
		"kind":       "Route",
		"metadata": map[string]interface{}{
			"creationTimestamp": nil,
			"name":              "gitops-webhook-event-listener-route",
			"namespace":         "cicd-environment",
		},
		"spec": map[string]interface{}{
			"host": "",
			"port": map[string]interface{}{"targetPort": float64(8080)},
			"to": map[string]interface{}{
				"kind":   "Service",
				"name":   "el-cicd-event-listener",
				"weight": float64(100),
			},
			"wildcardPolicy": "None",
		},
	}

	route, err := Generate("cicd-environment")
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want, route); diff != "" {
		t.Fatalf("Generate() failed:\n%s", diff)
	}
}

func TestCreateRoute(t *testing.T) {
	weight := int32(100)
	validRoute := routev1.Route{
		TypeMeta: routeTypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:      "gitops-webhook-event-listener-route",
			Namespace: "cicd-environment",
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
	route := createRoute("cicd-environment")
	if diff := cmp.Diff(validRoute, route); diff != "" {
		t.Fatalf("createRoute() failed:\n%s", diff)
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
	weight := int32(100)
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
