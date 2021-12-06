package kclient

import (
	"reflect"
	"testing"

	appsV1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	routev1 "github.com/openshift/api/route/v1"
	"github.com/redhat-developer/odo/pkg/testingutil"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ktesting "k8s.io/client-go/testing"
)

// GenerateOwnerReference generates an ownerReference which can then be set as
// owner for various OpenShift objects and ensure that when the owner object is
// deleted from the cluster, all other objects are automatically removed by
// OpenShift garbage collector
func GenerateOwnerReference(deployment *appsV1.Deployment) metav1.OwnerReference {

	ownerReference := metav1.OwnerReference{
		APIVersion: "apps.openshift.io/v1",
		Kind:       "Deployment",
		Name:       deployment.Name,
		UID:        deployment.UID,
	}

	return ownerReference
}

func TestCreateRoute(t *testing.T) {
	tests := []struct {
		name               string
		urlName            string
		service            string
		portNumber         intstr.IntOrString
		labels             map[string]string
		wantErr            bool
		existingDeployment *appsV1.Deployment
		secureURL          bool
		path               string
	}{
		{
			name:       "Case : mailserver",
			urlName:    "mailserver",
			service:    "mailserver",
			portNumber: intstr.FromInt(8080),
			labels: map[string]string{
				"SLA":                        "High",
				"app.kubernetes.io/instance": "backend",
				"app.kubernetes.io/name":     "python",
			},
			wantErr:            false,
			existingDeployment: testingutil.CreateFakeDeployment("mailserver"),
		},

		{
			name:       "Case : blog (urlName is different than service)",
			urlName:    "example",
			service:    "blog",
			portNumber: intstr.FromInt(9100),
			labels: map[string]string{
				"SLA":                        "High",
				"app.kubernetes.io/instance": "backend",
				"app.kubernetes.io/name":     "golang",
			},
			wantErr:            false,
			existingDeployment: testingutil.CreateFakeDeployment("blog"),
		},

		{
			name:       "Case : secure url",
			urlName:    "example",
			service:    "blog",
			portNumber: intstr.FromInt(9100),
			labels: map[string]string{
				"SLA":                        "High",
				"app.kubernetes.io/instance": "backend",
				"app.kubernetes.io/name":     "golang",
			},
			wantErr:            false,
			existingDeployment: testingutil.CreateFakeDeployment("blog"),
			secureURL:          true,
		},

		{
			name:       "Case : specify a path",
			urlName:    "example",
			service:    "blog",
			portNumber: intstr.FromInt(9100),
			labels: map[string]string{
				"SLA":                        "High",
				"app.kubernetes.io/instance": "backend",
				"app.kubernetes.io/name":     "golang",
			},
			wantErr:            false,
			existingDeployment: testingutil.CreateFakeDeployment("blog"),
			path:               "/testpath",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// initialising the fakeclient
			fkclient, fkclientset := FakeNew()

			ownerReferences := GenerateOwnerReference(tt.existingDeployment)

			fkclientset.AppsClientset.PrependReactor("get", "deployments", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, tt.existingDeployment, nil
			})

			createdRoute, err := fkclient.CreateRoute(tt.urlName, tt.service, tt.portNumber, tt.labels, tt.secureURL, tt.path, ownerReferences)

			if tt.secureURL {
				wantedTLSConfig := &routev1.TLSConfig{
					Termination:                   routev1.TLSTerminationEdge,
					InsecureEdgeTerminationPolicy: routev1.InsecureEdgeTerminationPolicyRedirect,
				}
				if !reflect.DeepEqual(createdRoute.Spec.TLS, wantedTLSConfig) {
					t.Errorf("tls config is different, wanted %v, got %v", wantedTLSConfig, createdRoute.Spec.TLS)
				}
			} else {
				if createdRoute.Spec.TLS != nil {
					t.Errorf("tls config is set for a non secure url")
				}
				if tt.path == "" && createdRoute.Spec.Path != "/" {
					t.Errorf("expect path: /, but got path: %v", createdRoute.Spec.Path)
				} else if tt.path != "" && createdRoute.Spec.Path != tt.path {
					t.Errorf("expect path: %v, but got path: %v", tt.path, createdRoute.Spec.Path)
				}
			}

			// Checks for error in positive cases
			if !tt.wantErr == (err != nil) {
				t.Errorf(" client.CreateRoute(string, labels) unexpected error %v, wantErr %v", err, tt.wantErr)
			}

			// Check for validating actions performed
			if len(fkclientset.RouteClientset.Actions()) != 1 {
				t.Errorf("expected 1 action in CreateRoute got: %v", fkclientset.RouteClientset.Actions())
			}
			// Checks for return values in positive cases
			if err == nil {
				createdRoute := fkclientset.RouteClientset.Actions()[0].(ktesting.CreateAction).GetObject().(*routev1.Route)
				// created route should be labeled with labels passed to CreateRoute
				if !reflect.DeepEqual(createdRoute.Labels, tt.labels) {
					t.Errorf("labels in created route is not matching expected labels, expected: %v, got: %v", tt.labels, createdRoute.Labels)
				}
				// route name and service that route is pointg to should match
				if createdRoute.Spec.To.Name != tt.service {
					t.Errorf("route is not matching to expected service name, expected: %s, got %s", tt.service, createdRoute)
				}
				if createdRoute.Name != tt.urlName {
					t.Errorf("route name is not matching to expected route name, expected: %s, got %s", tt.urlName, createdRoute.Name)
				}
				if createdRoute.Spec.To.Name != tt.service {
					t.Errorf("service name is not matching to expected service name, expected: %s, got %s", tt.service, createdRoute.Spec.To.Name)
				}
				if createdRoute.Spec.Port.TargetPort != tt.portNumber {
					t.Errorf("port number is not matching to expected port number, expected: %v, got %v", tt.portNumber, createdRoute.Spec.Port.TargetPort)
				}
			}
		})
	}
}

func TestListRoutes(t *testing.T) {
	type args struct {
		labelSelector string
	}
	tests := []struct {
		name           string
		args           args
		returnedRoutes routev1.RouteList
		want           []routev1.Route
		wantErr        bool
	}{
		{
			name: "case 1: list multiple routes",
			args: args{
				labelSelector: "app.kubernetes.io/instance",
			},
			returnedRoutes: *testingutil.GetRouteListWithMultiple("nodejs", "app"),
			want: []routev1.Route{
				testingutil.GetSingleRoute("example", 8080, "nodejs", "app"),
				testingutil.GetSingleRoute("example-1", 9100, "nodejs", "app"),
			},
		},
		{
			name: "case 2: no routes returned",
			args: args{
				labelSelector: "app.kubernetes.io/instance",
			},
			returnedRoutes: routev1.RouteList{},
			want:           nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// initialising the fakeclient
			fkclient, fkclientset := FakeNew()

			fkclientset.RouteClientset.PrependReactor("list", "routes", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, &tt.returnedRoutes, nil
			})

			got, err := fkclient.ListRoutes(tt.args.labelSelector)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListRoutes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ListRoutes() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetRoute(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name          string
		args          args
		returnedRoute routev1.Route
		want          routev1.Route
		wantErr       bool
	}{
		{
			name: "case 1: existing route returned",
			args: args{
				name: "example",
			},
			returnedRoute: testingutil.GetSingleRoute("example", 8080, "nodejs", "app"),
			want:          testingutil.GetSingleRoute("example", 8080, "nodejs", "app"),
		},
		{
			name: "case 2: no existing route returned",
			args: args{
				name: "example",
			},
			returnedRoute: routev1.Route{},
			want:          routev1.Route{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// initialising the fakeclient
			fkclient, fkclientset := FakeNew()

			fkclientset.RouteClientset.PrependReactor("get", "routes", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, &tt.returnedRoute, nil
			})

			got, err := fkclient.GetRoute(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRoute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, &tt.want) {
				t.Errorf("GetRoute() got = %v, want %v", got, tt.want)
			}
		})
	}
}
