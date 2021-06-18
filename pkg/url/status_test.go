package url

import (
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	routev1 "github.com/openshift/api/route/v1"
	"github.com/openshift/odo/pkg/kclient"
	"github.com/openshift/odo/pkg/localConfigProvider"
	"github.com/openshift/odo/pkg/occlient"
	"github.com/openshift/odo/pkg/testingutil"

	extensionsv1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kclient_fake "github.com/openshift/odo/pkg/kclient/fake"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery/fake"
	ktesting "k8s.io/client-go/testing"
)

type fakeDiscovery struct {
	*fake.FakeDiscovery
}

var fakeDiscoveryWithProject = &fakeDiscovery{
	FakeDiscovery: &fake.FakeDiscovery{
		Fake: &ktesting.Fake{
			Resources: []*metav1.APIResourceList{
				{
					GroupVersion: "route.openshift.io/v1",
					APIResources: []metav1.APIResource{{
						Name:         "routes",
						SingularName: "route",
						Namespaced:   false,
						Kind:         "Route",
						ShortNames:   []string{"route"},
					}},
				},
			},
		},
	},
}

func TestGetURLsForKubernetes(t *testing.T) {
	componentName := "my-component"

	testURL1 := localConfigProvider.LocalURL{Name: "example-1", Port: 9090, Host: "com", Kind: "ingress", Secure: true}
	testURL2 := localConfigProvider.LocalURL{Name: "example-2", Port: 9090, Host: "com", Kind: "ingress", Secure: false}
	testURL3 := localConfigProvider.LocalURL{Name: "routeurl2", Port: 8080, Kind: "route"}
	testURL4 := localConfigProvider.LocalURL{Name: "example", Port: 8080, Kind: "route"}

	tests := []struct {
		name              string
		envURLs           []localConfigProvider.LocalURL
		routeList         *routev1.RouteList
		ingressList       *extensionsv1.IngressList
		expectedStatusURL statusURL
	}{
		{
			name:    "1) Cluster with https URL defined in env info",
			envURLs: []localConfigProvider.LocalURL{testURL1},
			ingressList: &extensionsv1.IngressList{
				Items: []extensionsv1.Ingress{},
			},
			expectedStatusURL: statusURL{
				name:   testURL1.Name,
				kind:   "ingress",
				port:   testURL1.Port,
				secure: testURL1.Secure,
				url:    "https://example-1.com",
			},
			routeList: &routev1.RouteList{
				Items: []routev1.Route{},
			},
		},
		{
			name:    "2) Cluster with https URL defined in env info",
			envURLs: []localConfigProvider.LocalURL{testURL2},
			ingressList: &extensionsv1.IngressList{
				Items: []extensionsv1.Ingress{},
			},
			expectedStatusURL: statusURL{
				name:   testURL2.Name,
				kind:   "ingress",
				port:   testURL2.Port,
				secure: testURL2.Secure,
				url:    "http://example-2.com",
			},
			routeList: &routev1.RouteList{
				Items: []routev1.Route{},
			},
		},
		{
			name:    "3) Cluster with route defined in env info",
			envURLs: []localConfigProvider.LocalURL{testURL3},
			ingressList: &extensionsv1.IngressList{
				Items: []extensionsv1.Ingress{},
			},
			expectedStatusURL: statusURL{
				name:   testURL3.Name,
				kind:   "route",
				port:   testURL3.Port,
				secure: false,
				url:    "",
			},

			routeList: &routev1.RouteList{
				Items: []routev1.Route{},
			},
		},

		{
			name:    "4) Cluster with route defined",
			envURLs: []localConfigProvider.LocalURL{},
			ingressList: &extensionsv1.IngressList{
				Items: []extensionsv1.Ingress{},
			},
			expectedStatusURL: statusURL{
				name:   testURL4.Name,
				kind:   "route",
				port:   testURL4.Port,
				secure: false,
				url:    "http://example.com",
			},
			routeList: &routev1.RouteList{
				Items: []routev1.Route{
					testingutil.GetSingleRoute(testURL4.Name, testURL4.Port, componentName, ""),
				},
			},
		},
		{
			name:    "5) Cluster with ingress defined",
			envURLs: []localConfigProvider.LocalURL{},

			ingressList: &extensionsv1.IngressList{
				Items: []extensionsv1.Ingress{
					kclient_fake.GetIngressListWithMultiple(componentName, "app").Items[0],
				},
			},
			routeList: &routev1.RouteList{
				Items: []routev1.Route{},
			},
			expectedStatusURL: statusURL{
				name:   "example-0",
				kind:   "ingress",
				port:   8080,
				secure: false,
				url:    "http://example-0.com",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLocalConfig := localConfigProvider.NewMockLocalConfigProvider(ctrl)
			mockLocalConfig.EXPECT().GetName().Return(componentName).AnyTimes()
			mockLocalConfig.EXPECT().GetApplication().Return("")
			mockLocalConfig.EXPECT().ListURLs().Return(tt.envURLs, nil)

			// Initialising the fakeclient
			fkclient, fkclientset := kclient.FakeNew()
			fkclient.Namespace = "default"

			// Return the test's ingress list when requested
			fkclientset.Kubernetes.PrependReactor("list", "ingresses", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, tt.ingressList, nil
			})

			// Initializing the fake occlient
			fkoclient, fakeoclientSet := occlient.FakeNew()
			fkoclient.Namespace = "default"
			fkoclient.SetKubeClient(fkclient)
			fkoclient.GetKubeClient().SetDiscoveryInterface(fakeDiscoveryWithProject)
			fkoclient.SetKubeClient(fkclient)

			// Return the test's route list when requested
			fakeoclientSet.RouteClientset.PrependReactor("list", "routes", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, tt.routeList, nil
			})

			statusUrls, err := getURLsForKubernetes(fkoclient, fkclient, mockLocalConfig, false)

			if err != nil {
				t.Fatalf("Error occurred: %v", err)
			}

			if len(statusUrls) == 0 {
				t.Fatalf("statusURLs has unexpected size 0, must be 1")
			}

			if !reflect.DeepEqual(tt.expectedStatusURL, statusUrls[0]) {
				t.Fatalf("Mismatching status URL - expected: %v,  actual: %v", tt.expectedStatusURL, statusUrls[0])
			}
		})
	}
}
