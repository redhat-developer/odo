package routes

import (
	"log"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	routev1 "github.com/openshift/odo/pkg/pipelines/routes/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/yaml"
)

func TestGenerateRoute(t *testing.T) {
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
	route := Generate("cicd-environment")
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

func TestMarshalRoute(t *testing.T) {
	weight := int32(100)
	r := routev1.Route{
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

	b, err := yaml.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	log.Printf("KEVIN!!!! %s\n", b)
	if strings.Index(strings.ToLower(string(b)), "ingress") > -1 {
		t.Fatal("the marshaled output contains the ingress field")
	}
}
