package webhook

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	routeclientset "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	fakeKubeClientset "k8s.io/client-go/kubernetes/fake"

	routev1 "github.com/openshift/api/route/v1"
	fakeRouteClientset "github.com/openshift/client-go/route/clientset/versioned/fake"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ktesting "k8s.io/client-go/testing"
)

func TestGetRouteHost(t *testing.T) {

	reouteClientset := fakeRouteClientset.NewSimpleClientset()
	reouteClientset.PrependReactor("get", "routes", func(action ktesting.Action) (bool, runtime.Object, error) {
		if action.GetNamespace() != "tst-cicd" {
			return true, nil, fmt.Errorf("'get' called with a different namespace %s", action.GetNamespace())
		}

		route := &routev1.Route{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "gitops-webhook-event-listener-route",
				Namespace: "tst-cicd",
			},
			Spec: routev1.RouteSpec{
				Host: "devcluster.openshift.com",
				Port: &routev1.RoutePort{
					TargetPort: intstr.IntOrString{
						IntVal: 8080,
						StrVal: "8080",
					},
				},
			},
		}
		return true, route, nil
	})
	resources := fakeNewResources(reouteClientset.Route(), nil)

	hasTLS, host, err := resources.getListenerAddress("tst-cicd", "gitops-webhook-event-listener-route")
	if err != nil {
		t.Fatal(err)
	}

	if hasTLS {
		t.Error("hasTLS is expected to be false.")
	}

	if diff := cmp.Diff(host, "devcluster.openshift.com"); diff != "" {
		t.Errorf("host mismatch got\n%s", diff)
	}

}

func TestGetSecret(t *testing.T) {

	kubeClient := fakeKubeClientset.NewSimpleClientset()

	kubeClient.PrependReactor("get", "secrets", func(action ktesting.Action) (bool, runtime.Object, error) {
		if action.GetNamespace() != "tst-cicd" {
			return true, nil, fmt.Errorf("'get' called with a different namespace %s", action.GetNamespace())
		}

		return true, &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "gitops-webhook-secret",
				Namespace: "tst-cicd",
			},
			Data: map[string][]byte{
				"webhook-secret-key": []byte("testing"),
			},
		}, nil
	})

	resources := fakeNewResources(nil, kubeClient)

	secret, err := resources.getWebhookSecret("tst-cicd", "gitops-webhook-secret", "webhook-secret-key")
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(secret, "testing"); diff != "" {
		t.Errorf("secret value errMsg mismatch got\n%s", diff)
	}
}

func fakeNewResources(routeClient routeclientset.RouteV1Interface,

	kubeClient kubernetes.Interface) *resources {

	return &resources{
		routeClient: routeClient,
		kubeClient:  kubeClient,
	}
}
