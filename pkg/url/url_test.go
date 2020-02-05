package url

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	appsv1 "github.com/openshift/api/apps/v1"
	routev1 "github.com/openshift/api/route/v1"
	applabels "github.com/openshift/odo/pkg/application/labels"
	componentlabels "github.com/openshift/odo/pkg/component/labels"
	"github.com/openshift/odo/pkg/occlient"
	"github.com/openshift/odo/pkg/url/labels"
	"github.com/openshift/odo/pkg/util"
	"github.com/openshift/odo/pkg/version"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ktesting "k8s.io/client-go/testing"
)

func TestCreate(t *testing.T) {
	type args struct {
		componentName   string
		applicationName string
		urlName         string
		portNumber      int
		secure          bool
	}
	tests := []struct {
		name          string
		args          args
		returnedRoute *routev1.Route
		want          string
		wantErr       bool
	}{
		{
			name: "Case 1: Component name same as urlName",
			args: args{
				componentName:   "nodejs",
				applicationName: "app",
				urlName:         "nodejs",
				portNumber:      8080,
			},
			returnedRoute: &routev1.Route{
				ObjectMeta: metav1.ObjectMeta{
					Name: "nodejs-app",
					Labels: map[string]string{
						"app.kubernetes.io/part-of":  "app",
						"app.kubernetes.io/instance": "nodejs",
						applabels.App:                "app",
						applabels.OdoManagedBy:       "odo",
						applabels.OdoVersion:         version.VERSION,
						"odo.openshift.io/url-name":  "nodejs",
					},
				},
				Spec: routev1.RouteSpec{
					To: routev1.RouteTargetReference{
						Kind: "Service",
						Name: "nodejs-app",
					},
					Port: &routev1.RoutePort{
						TargetPort: intstr.FromInt(8080),
					},
				},
			},
			want:    "http://host",
			wantErr: false,
		},
		{
			name: "Case 2: Component name different than urlName",
			args: args{
				componentName:   "nodejs",
				applicationName: "app",
				urlName:         "example-url",
				portNumber:      9100,
			},
			returnedRoute: &routev1.Route{
				ObjectMeta: metav1.ObjectMeta{
					Name: "example-url-app",
					Labels: map[string]string{
						"app.kubernetes.io/part-of":  "app",
						"app.kubernetes.io/instance": "nodejs",
						applabels.App:                "app",
						applabels.OdoManagedBy:       "odo",
						applabels.OdoVersion:         version.VERSION,
						"odo.openshift.io/url-name":  "example-url",
					},
				},
				Spec: routev1.RouteSpec{
					To: routev1.RouteTargetReference{
						Kind: "Service",
						Name: "nodejs-app",
					},
					Port: &routev1.RoutePort{
						TargetPort: intstr.FromInt(9100),
					},
				},
			},
			want:    "http://host",
			wantErr: false,
		},
		{
			name: "Case 3: a secure URL",
			args: args{
				componentName:   "nodejs",
				applicationName: "app",
				urlName:         "example-url",
				portNumber:      9100,
				secure:          true,
			},
			returnedRoute: &routev1.Route{
				ObjectMeta: metav1.ObjectMeta{
					Name: "example-url-app",
					Labels: map[string]string{
						"app.kubernetes.io/part-of":  "app",
						"app.kubernetes.io/instance": "nodejs",
						applabels.App:                "app",
						applabels.OdoManagedBy:       "odo",
						applabels.OdoVersion:         version.VERSION,
						"odo.openshift.io/url-name":  "example-url",
					},
				},
				Spec: routev1.RouteSpec{
					TLS: &routev1.TLSConfig{
						Termination:                   routev1.TLSTerminationEdge,
						InsecureEdgeTerminationPolicy: routev1.InsecureEdgeTerminationPolicyRedirect,
					},
					To: routev1.RouteTargetReference{
						Kind: "Service",
						Name: "nodejs-app",
					},
					Port: &routev1.RoutePort{
						TargetPort: intstr.FromInt(9100),
					},
				},
			},
			want:    "https://host",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, fakeClientSet := occlient.FakeNew()

			fakeClientSet.RouteClientset.PrependReactor("create", "routes", func(action ktesting.Action) (bool, runtime.Object, error) {
				route := action.(ktesting.CreateAction).GetObject().(*routev1.Route)
				route.Spec.Host = "host"
				return true, route, nil
			})

			serviceName, err := util.NamespaceOpenShiftObject(tt.args.componentName, tt.args.applicationName)
			if err != nil {
				t.Error(err)
			}

			fakeClientSet.AppsClientset.PrependReactor("get", "deploymentconfigs", func(action ktesting.Action) (bool, runtime.Object, error) {
				dc := &appsv1.DeploymentConfig{}
				dc.Name = serviceName
				return true, dc, nil
			})

			got, err := Create(client, tt.args.urlName, tt.args.portNumber, tt.args.secure, tt.args.componentName, tt.args.applicationName)

			if err == nil && !tt.wantErr {
				if len(fakeClientSet.RouteClientset.Actions()) != 1 {
					t.Errorf("expected 1 RouteClientset.Actions() in CreateService, got: %v", fakeClientSet.RouteClientset.Actions())
				}

				createdRoute := fakeClientSet.RouteClientset.Actions()[0].(ktesting.CreateAction).GetObject().(*routev1.Route)
				if !reflect.DeepEqual(createdRoute.Name, tt.returnedRoute.Name) {
					t.Errorf("route name not matching, expected: %s, got %s", tt.returnedRoute.Name, createdRoute.Name)
				}
				if !reflect.DeepEqual(createdRoute.Labels, tt.returnedRoute.Labels) {
					t.Errorf("route name not matching, %v", pretty.Compare(tt.returnedRoute.Labels, createdRoute.Labels))
				}
				if !reflect.DeepEqual(createdRoute.Spec.Port, tt.returnedRoute.Spec.Port) {
					t.Errorf("route name not matching, expected: %s, got %s", tt.returnedRoute.Spec.Port, createdRoute.Spec.Port)
				}
				if !reflect.DeepEqual(createdRoute.Spec.To.Name, tt.returnedRoute.Spec.To.Name) {
					t.Errorf("route name not matching, expected: %s, got %s", tt.returnedRoute.Spec.To.Name, createdRoute.Spec.To.Name)
				}

				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("Create() = %#v, want %#v", got, tt.want)
				}
			} else if err == nil && tt.wantErr {
				t.Error("error was expected, but no error was returned")
			} else if err != nil && !tt.wantErr {
				t.Errorf("test failed, no error was expected, but got unexpected error: %s", err)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	type args struct {
		urlName         string
		applicationName string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "first test",
			args: args{
				urlName:         "component",
				applicationName: "appname",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, fakeClientSet := occlient.FakeNew()

			fakeClientSet.RouteClientset.PrependReactor("delete", "routes", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, nil, nil
			})

			err := Delete(client, tt.args.urlName, tt.args.applicationName)
			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %#v, wantErr %#v", err, tt.wantErr)
				return
			}

			// Check for value with which the function has called
			DeletedURL := fakeClientSet.RouteClientset.Actions()[0].(ktesting.DeleteAction).GetName()
			if !reflect.DeepEqual(DeletedURL, tt.args.urlName+"-"+tt.args.applicationName) {
				t.Errorf("Delete is been called with %#v, expected %#v", DeletedURL, tt.args.urlName+"-"+tt.args.applicationName)
			}
		})
	}
}

func TestExists(t *testing.T) {
	tests := []struct {
		name            string
		urlName         string
		componentName   string
		applicationName string
		wantBool        bool
		routes          routev1.RouteList
		labelSelector   string
		wantErr         bool
	}{
		{
			name:            "correct values and Host found",
			urlName:         "nodejs",
			componentName:   "nodejs",
			applicationName: "app",
			routes: routev1.RouteList{
				Items: []routev1.Route{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "nodejs",
							Labels: map[string]string{
								applabels.ApplicationLabel:     "app",
								componentlabels.ComponentLabel: "nodejs",
								applabels.OdoManagedBy:         "odo",
								applabels.OdoVersion:           version.VERSION,
								labels.URLLabel:                "nodejs",
							},
						},
						Spec: routev1.RouteSpec{
							To: routev1.RouteTargetReference{
								Kind: "Service",
								Name: "nodejs-app",
							},
							Port: &routev1.RoutePort{
								TargetPort: intstr.FromInt(8080),
							},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "wildfly",
							Labels: map[string]string{
								applabels.ApplicationLabel:     "app",
								componentlabels.ComponentLabel: "wildfly",
								applabels.OdoManagedBy:         "odo",
								applabels.OdoVersion:           version.VERSION,
								labels.URLLabel:                "wildfly",
							},
						},
						Spec: routev1.RouteSpec{
							To: routev1.RouteTargetReference{
								Kind: "Service",
								Name: "wildfly-app",
							},
							Port: &routev1.RoutePort{
								TargetPort: intstr.FromInt(9100),
							},
						},
					},
				},
			},
			wantBool:      true,
			labelSelector: "app.kubernetes.io/instance=nodejs,app.kubernetes.io/part-of=app",
			wantErr:       false,
		},
		{
			name:            "correct values and Host not found",
			urlName:         "example",
			componentName:   "nodejs",
			applicationName: "app",
			routes: routev1.RouteList{
				Items: []routev1.Route{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "nodejs",
							Labels: map[string]string{
								applabels.ApplicationLabel:     "app",
								componentlabels.ComponentLabel: "nodejs",
								applabels.OdoManagedBy:         "odo",
								applabels.OdoVersion:           version.VERSION,
								labels.URLLabel:                "nodejs",
							},
						},
						Spec: routev1.RouteSpec{
							To: routev1.RouteTargetReference{
								Kind: "Service",
								Name: "nodejs-app",
							},
							Port: &routev1.RoutePort{
								TargetPort: intstr.FromInt(8080),
							},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "wildfly",
							Labels: map[string]string{
								applabels.ApplicationLabel:     "app",
								componentlabels.ComponentLabel: "wildfly",
								applabels.OdoManagedBy:         "odo",
								applabels.OdoVersion:           version.VERSION,
								labels.URLLabel:                "wildfly",
							},
						},
						Spec: routev1.RouteSpec{
							To: routev1.RouteTargetReference{
								Kind: "Service",
								Name: "wildfly-app",
							},
							Port: &routev1.RoutePort{
								TargetPort: intstr.FromInt(9100),
							},
						},
					},
				},
			},
			wantBool:      false,
			labelSelector: "app.kubernetes.io/instance=nodejs,app.kubernetes.io/part-of=app",
			wantErr:       false,
		},
	}
	for _, tt := range tests {
		client, fakeClientSet := occlient.FakeNew()

		fakeClientSet.RouteClientset.PrependReactor("list", "routes", func(action ktesting.Action) (bool, runtime.Object, error) {
			if !reflect.DeepEqual(action.(ktesting.ListAction).GetListRestrictions().Labels.String(), tt.labelSelector) {
				return true, nil, fmt.Errorf("labels not matching with expected values, expected:%s, got:%s", tt.labelSelector, action.(ktesting.ListAction).GetListRestrictions())
			}
			return true, &tt.routes, nil
		})

		exists, err := Exists(client, tt.urlName, tt.componentName, tt.applicationName)
		if err == nil && !tt.wantErr {
			if (len(fakeClientSet.RouteClientset.Actions()) != 1) && (tt.wantErr != true) {
				t.Errorf("expected 1 action in ListRoutes got: %v", fakeClientSet.RouteClientset.Actions())
			}
			if exists != tt.wantBool {
				t.Errorf("expected exists to be:%t, got :%t", tt.wantBool, exists)
			}
		} else if err == nil && tt.wantErr {
			t.Errorf("test failed, expected: %s, got %s", "false", "true")
		} else if err != nil && !tt.wantErr {
			t.Errorf("test failed, expected: %s, got %s", "no error", "error:"+err.Error())
		}
	}
}

func TestGetValidPortNumber(t *testing.T) {
	type args struct {
		portNumber    int
		componentName string
		portList      []string
	}
	tests := []struct {
		name       string
		args       args
		wantedPort int
		wantErr    bool
	}{
		{
			name: "test case 1: component with one port and port number provided",
			args: args{
				componentName: "nodejs",
				portNumber:    8080,
				portList:      []string{"8080/TCP"},
			},
			wantedPort: 8080,
			wantErr:    false,
		},
		{
			name: "test case 2: component with two ports and port number provided",
			args: args{
				componentName: "nodejs",
				portNumber:    8080,
				portList:      []string{"8080/TCP", "8081/TCP"},
			},
			wantedPort: 8080,
			wantErr:    false,
		},
		{
			name: "test case 3: service with two ports and no port number provided",
			args: args{
				componentName: "nodejs",
				portNumber:    -1,
				portList:      []string{"8080/TCP", "8081/TCP"},
			},

			wantErr: true,
		},
		{
			name: "test case 4: component with one port and no port number provided",
			args: args{
				componentName: "nodejs",
				portNumber:    -1,
				portList:      []string{"8080/TCP"},
			},
			wantedPort: 8080,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			gotPortNumber, err := GetValidPortNumber(tt.args.componentName, tt.args.portNumber, tt.args.portList)

			if err == nil && !tt.wantErr {

				if !reflect.DeepEqual(gotPortNumber, tt.wantedPort) {
					t.Errorf("Create() = %#v, want %#v", gotPortNumber, tt.wantedPort)
				}
			} else if err == nil && tt.wantErr {
				t.Error("error was expected, but no error was returned")
			} else if err != nil && !tt.wantErr {
				t.Errorf("test failed, no error was expected, but got unexpected error: %s", err)
			}
		})
	}
}
