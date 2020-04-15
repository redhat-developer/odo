package url

import (
	"fmt"
	"github.com/kylelemons/godebug/pretty"
	appsv1 "github.com/openshift/api/apps/v1"
	routev1 "github.com/openshift/api/route/v1"
	applabels "github.com/openshift/odo/pkg/application/labels"
	componentlabels "github.com/openshift/odo/pkg/component/labels"
	"github.com/openshift/odo/pkg/config"
	"github.com/openshift/odo/pkg/envinfo"
	"github.com/openshift/odo/pkg/kclient"
	"github.com/openshift/odo/pkg/kclient/fake"
	"github.com/openshift/odo/pkg/occlient"
	"github.com/openshift/odo/pkg/testingutil"
	"github.com/openshift/odo/pkg/url/labels"
	"github.com/openshift/odo/pkg/util"
	"k8s.io/api/core/v1"
	extensionsv1 "k8s.io/api/extensions/v1beta1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"reflect"
	"testing"
	//"github.com/openshift/odo/pkg/util"
	"github.com/openshift/odo/pkg/version"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ktesting "k8s.io/client-go/testing"
)

func TestCreate(t *testing.T) {
	type args struct {
		componentName             string
		applicationName           string
		urlName                   string
		portNumber                int
		secure                    bool
		host                      string
		urlKind                   envinfo.URLKind
		isRouteSupported          bool
		isExperimentalModeEnabled bool
		tlsSecret                 string
	}
	tests := []struct {
		name               string
		args               args
		returnedRoute      *routev1.Route
		returnedIngress    *extensionsv1.Ingress
		defaultTLSExists   bool
		userGivenTLSExists bool
		want               string
		wantErr            bool
	}{
		{
			name: "Case 1: Component name same as urlName",
			args: args{
				componentName:    "nodejs",
				applicationName:  "app",
				urlName:          "nodejs",
				portNumber:       8080,
				isRouteSupported: true,
				urlKind:          envinfo.ROUTE,
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
				componentName:    "nodejs",
				applicationName:  "app",
				urlName:          "example-url",
				portNumber:       9100,
				isRouteSupported: true,
				urlKind:          envinfo.ROUTE,
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
				componentName:    "nodejs",
				applicationName:  "app",
				urlName:          "example-url",
				portNumber:       9100,
				secure:           true,
				isRouteSupported: true,
				urlKind:          envinfo.ROUTE,
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

		{
			name: "Case 4: Create a ingress, with same name as component,instead of route on openshift cluster",
			args: args{
				componentName:             "nodejs",
				urlName:                   "nodejs",
				portNumber:                8080,
				host:                      "com",
				isRouteSupported:          true,
				isExperimentalModeEnabled: true,
				urlKind:                   envinfo.INGRESS,
			},
			returnedIngress: fake.GetSingleIngress("nodejs", "nodejs"),
			want:            "http://nodejs.com",
			wantErr:         false,
		},
		{
			name: "Case 5: Create a ingress, with different name as component,instead of route on openshift cluster",
			args: args{
				componentName:             "nodejs",
				urlName:                   "example",
				portNumber:                8080,
				host:                      "com",
				isRouteSupported:          true,
				isExperimentalModeEnabled: true,
				urlKind:                   envinfo.INGRESS,
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
			returnedIngress: fake.GetSingleIngress("example", "nodejs"),
			want:            "http://example.com",
			wantErr:         false,
		},
		{
			name: "Case 6: Create a secure ingress, instead of route on openshift cluster, default tls exists",
			args: args{
				componentName:             "nodejs",
				urlName:                   "example",
				portNumber:                8080,
				host:                      "com",
				isRouteSupported:          true,
				isExperimentalModeEnabled: true,
				secure:                    true,
				urlKind:                   envinfo.INGRESS,
			},
			returnedIngress:  fake.GetSingleIngress("example", "nodejs"),
			defaultTLSExists: true,
			want:             "https://example.com",
			wantErr:          false,
		},
		{
			name: "Case 7: Create a secure ingress, instead of route on openshift cluster and default tls doesn't exist",
			args: args{
				componentName:             "nodejs",
				urlName:                   "example",
				portNumber:                8080,
				host:                      "com",
				isRouteSupported:          true,
				isExperimentalModeEnabled: true,
				secure:                    true,
				urlKind:                   envinfo.INGRESS,
			},
			returnedIngress:  fake.GetSingleIngress("example", "nodejs"),
			defaultTLSExists: false,
			want:             "https://example.com",
			wantErr:          false,
		},
		{
			name: "Case 8: Fail when while creating ingress when user given tls secret doesn't exists",
			args: args{
				componentName:             "nodejs",
				urlName:                   "example",
				portNumber:                8080,
				host:                      "com",
				isRouteSupported:          true,
				isExperimentalModeEnabled: true,
				secure:                    true,
				tlsSecret:                 "user-secret",
				urlKind:                   envinfo.INGRESS,
			},
			returnedIngress:    fake.GetSingleIngress("example", "nodejs"),
			defaultTLSExists:   false,
			userGivenTLSExists: false,
			want:               "http://example.com",
			wantErr:            true,
		},
		{
			name: "Case 9: Create a secure ingress, instead of route on openshift cluster, user tls secret does exists",
			args: args{
				componentName:             "nodejs",
				urlName:                   "example",
				portNumber:                8080,
				host:                      "com",
				isRouteSupported:          true,
				isExperimentalModeEnabled: true,
				secure:                    true,
				tlsSecret:                 "user-secret",
				urlKind:                   envinfo.INGRESS,
			},
			returnedIngress:    fake.GetSingleIngress("example", "nodejs"),
			defaultTLSExists:   false,
			userGivenTLSExists: true,
			want:               "https://example.com",
			wantErr:            false,
		},

		{
			name: "Case 10: invalid url kind",
			args: args{
				componentName:             "nodejs",
				urlName:                   "example",
				portNumber:                8080,
				host:                      "com",
				isRouteSupported:          true,
				isExperimentalModeEnabled: true,
				secure:                    true,
				tlsSecret:                 "user-secret",
				urlKind:                   "blah",
			},
			returnedIngress:    fake.GetSingleIngress("example", "nodejs"),
			defaultTLSExists:   false,
			userGivenTLSExists: true,
			want:               "",
			wantErr:            true,
		},
		{
			name: "Case 11: route is not supported on the cluster",
			args: args{
				componentName:             "nodejs",
				applicationName:           "app",
				urlName:                   "example",
				isRouteSupported:          false,
				isExperimentalModeEnabled: true,
				urlKind:                   envinfo.ROUTE,
			},
			returnedIngress:    fake.GetSingleIngress("example", "nodejs"),
			defaultTLSExists:   false,
			userGivenTLSExists: true,
			want:               "",
			wantErr:            true,
		},
		{
			name: "Case 11: secretName used without secure flag",
			args: args{
				componentName:             "nodejs",
				applicationName:           "app",
				urlName:                   "example",
				isRouteSupported:          false,
				isExperimentalModeEnabled: true,
				tlsSecret:                 "secret",
				urlKind:                   envinfo.ROUTE,
			},
			returnedIngress:    fake.GetSingleIngress("example", "nodejs"),
			defaultTLSExists:   false,
			userGivenTLSExists: true,
			want:               "",
			wantErr:            true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, fakeClientSet := occlient.FakeNew()
			fakeKClient, fakeKClientSet := kclient.FakeNew()

			fakeClientSet.RouteClientset.PrependReactor("create", "routes", func(action ktesting.Action) (bool, runtime.Object, error) {
				route := action.(ktesting.CreateAction).GetObject().(*routev1.Route)
				route.Spec.Host = "host"
				return true, route, nil
			})

			fakeKClientSet.Kubernetes.PrependReactor("get", "secrets", func(action ktesting.Action) (bool, runtime.Object, error) {
				var secretName string
				if tt.args.tlsSecret == "" {
					secretName = tt.args.componentName + "-tlssecret"
					if action.(ktesting.GetAction).GetName() != secretName {
						return true, nil, fmt.Errorf("get for secrets called with invalid name, want: %s,got: %s", secretName, action.(ktesting.GetAction).GetName())
					}
				} else {
					secretName = tt.args.tlsSecret
					if action.(ktesting.GetAction).GetName() != tt.args.tlsSecret {
						return true, nil, fmt.Errorf("get for secrets called with invalid name, want: %s,got: %s", tt.args.tlsSecret, action.(ktesting.GetAction).GetName())
					}
				}
				if tt.args.tlsSecret != "" {
					if !tt.userGivenTLSExists {
						return true, nil, kerrors.NewNotFound(schema.GroupResource{}, "")
					}
				} else if !tt.defaultTLSExists {
					return true, nil, kerrors.NewNotFound(schema.GroupResource{}, "")
				}
				return true, fake.GetSecret(secretName), nil
			})

			var serviceName string
			if tt.args.urlKind == envinfo.INGRESS {
				serviceName = tt.args.componentName

			} else if tt.args.urlKind == envinfo.ROUTE {
				var err error
				serviceName, err = util.NamespaceOpenShiftObject(tt.args.componentName, tt.args.applicationName)
				if err != nil {
					t.Error(err)
				}
			}

			fakeClientSet.AppsClientset.PrependReactor("get", "deploymentconfigs", func(action ktesting.Action) (bool, runtime.Object, error) {
				dc := &appsv1.DeploymentConfig{}
				dc.Name = serviceName
				return true, dc, nil
			})

			fakeKClientSet.Kubernetes.PrependReactor("get", "deployments", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, testingutil.CreateFakeDeployment("nodejs"), nil
			})

			got, err := Create(client, fakeKClient, tt.args.urlName, tt.args.portNumber, tt.args.secure, tt.args.componentName, tt.args.applicationName, tt.args.host, tt.args.tlsSecret, tt.args.urlKind, tt.args.isRouteSupported, tt.args.isExperimentalModeEnabled)

			if err == nil && !tt.wantErr {
				if tt.args.urlKind == envinfo.INGRESS {
					wantKubernetesActionLength := 0
					if !tt.args.secure {
						wantKubernetesActionLength = 2
					} else {
						if tt.args.tlsSecret != "" && tt.userGivenTLSExists {
							wantKubernetesActionLength = 3
						} else if !tt.defaultTLSExists {
							wantKubernetesActionLength = 4
						} else {
							wantKubernetesActionLength = 3
						}
					}
					if len(fakeKClientSet.Kubernetes.Actions()) != wantKubernetesActionLength {
						t.Errorf("expected %v Kubernetes.Actions() in Create, got: %v", wantKubernetesActionLength, len(fakeKClientSet.Kubernetes.Actions()))
					}

					if len(fakeClientSet.RouteClientset.Actions()) != 0 {
						t.Errorf("expected 0 RouteClientset.Actions() in CreateService, got: %v", fakeClientSet.RouteClientset.Actions())
					}

					var createdIngress *extensionsv1.Ingress
					createIngressActionNo := 0
					if !tt.args.secure {
						createIngressActionNo = 1
					} else {
						if tt.args.tlsSecret != "" {
							createIngressActionNo = 2
						} else if !tt.defaultTLSExists {
							createdDefaultTLS := fakeKClientSet.Kubernetes.Actions()[2].(ktesting.CreateAction).GetObject().(*v1.Secret)
							if createdDefaultTLS.Name != tt.args.componentName+"-tlssecret" {
								t.Errorf("default tls created with different name, want: %s,got: %s", tt.args.componentName+"-tlssecret", createdDefaultTLS.Name)
							}
							createIngressActionNo = 3
						} else {
							createIngressActionNo = 2
						}
					}
					createdIngress = fakeKClientSet.Kubernetes.Actions()[createIngressActionNo].(ktesting.CreateAction).GetObject().(*extensionsv1.Ingress)

					if !reflect.DeepEqual(createdIngress.Name, tt.returnedIngress.Name) {
						t.Errorf("ingress name not matching, expected: %s, got %s", tt.returnedRoute.Name, createdIngress.Name)
					}
					if !reflect.DeepEqual(createdIngress.Labels, tt.returnedIngress.Labels) {
						t.Errorf("ingress labels not matching, %v", pretty.Compare(tt.returnedIngress.Labels, createdIngress.Labels))
					}

					wantedIngressParams := kclient.IngressParameter{
						ServiceName:   serviceName,
						IngressDomain: tt.args.host,
						PortNumber:    intstr.FromInt(tt.args.portNumber),
						TLSSecretName: tt.args.tlsSecret,
					}

					if !reflect.DeepEqual(createdIngress.Spec.Rules[0].HTTP.Paths[0].Backend.ServicePort.IntVal, wantedIngressParams.PortNumber.IntVal) {
						t.Errorf("ingress port not matching, expected: %s, got %s", tt.returnedRoute.Spec.Port, createdIngress.Spec.Rules[0].HTTP.Paths[0].Backend.ServicePort.StrVal)
					}
					if tt.args.secure {
						if wantedIngressParams.TLSSecretName == "" {
							wantedIngressParams.TLSSecretName = tt.args.componentName + "-tlssecret"
						}
						if !reflect.DeepEqual(createdIngress.Spec.TLS[0].SecretName, wantedIngressParams.TLSSecretName) {
							t.Errorf("ingress tls name not matching, expected: %s, got %s", wantedIngressParams.TLSSecretName, createdIngress.Spec.TLS)
						}
					}

				} else {
					if len(fakeClientSet.RouteClientset.Actions()) != 1 {
						t.Errorf("expected 1 RouteClientset.Actions() in CreateService, got: %v", fakeClientSet.RouteClientset.Actions())
					}

					if len(fakeKClientSet.Kubernetes.Actions()) != 0 {
						t.Errorf("expected 0 Kubernetes.Actions() in CreateService, got: %v", len(fakeKClientSet.Kubernetes.Actions()))
					}

					createdRoute := fakeClientSet.RouteClientset.Actions()[0].(ktesting.CreateAction).GetObject().(*routev1.Route)
					if !reflect.DeepEqual(createdRoute.Name, tt.returnedRoute.Name) {
						t.Errorf("route name not matching, expected: %s, got %s", tt.returnedRoute.Name, createdRoute.Name)
					}
					if !reflect.DeepEqual(createdRoute.Labels, tt.returnedRoute.Labels) {
						t.Errorf("route labels not matching, %v", pretty.Compare(tt.returnedRoute.Labels, createdRoute.Labels))
					}
					if !reflect.DeepEqual(createdRoute.Spec.Port, tt.returnedRoute.Spec.Port) {
						t.Errorf("route port not matching, expected: %s, got %s", tt.returnedRoute.Spec.Port, createdRoute.Spec.Port)
					}
					if !reflect.DeepEqual(createdRoute.Spec.To.Name, tt.returnedRoute.Spec.To.Name) {
						t.Errorf("route spec not matching, expected: %s, got %s", tt.returnedRoute.Spec.To.Name, createdRoute.Spec.To.Name)
					}

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

func TestPush(t *testing.T) {
	type args struct {
		isRouteSupported          bool
		isExperimentalModeEnabled bool
	}
	tests := []struct {
		name                string
		args                args
		componentName       string
		applicationName     string
		existingConfigURLs  []config.ConfigURL
		existingEnvInfoURLs []envinfo.EnvInfoURL
		returnedRoutes      *routev1.RouteList
		returnedIngress     *extensionsv1.IngressList
		deletedURLs         []URL
		createdURLs         []URL
		wantErr             bool
	}{
		{
			name: "no urls on local config and cluster",
			args: args{
				isRouteSupported: true,
			},
			componentName:   "nodejs",
			applicationName: "app",
			returnedRoutes:  &routev1.RouteList{},
		},
		{
			name:            "2 urls on local config and 0 on openshift cluster",
			componentName:   "nodejs",
			applicationName: "app",
			args:            args{isRouteSupported: true},
			existingConfigURLs: []config.ConfigURL{
				{
					Name:   "example",
					Port:   8080,
					Secure: false,
				},
				{
					Name:   "example-1",
					Port:   8080,
					Secure: false,
				},
			},
			returnedRoutes: &routev1.RouteList{},
		},
		{
			name:            "0 url on local config and 2 on openshift cluster",
			componentName:   "wildfly",
			applicationName: "app",
			args:            args{isRouteSupported: true},
			returnedRoutes:  testingutil.GetRouteListWithMultiple("wildfly", "app"),
			deletedURLs: []URL{
				getMachineReadableFormat(testingutil.GetSingleRoute("example-app", 8080, "nodejs", "app")),
				getMachineReadableFormat(testingutil.GetSingleRoute("example-1-app", 9100, "nodejs", "app")),
			},
		},
		{
			name:            "2 url on local config and 2 on openshift cluster, but they are different",
			componentName:   "nodejs",
			applicationName: "app",
			args:            args{isRouteSupported: true},
			existingConfigURLs: []config.ConfigURL{
				{
					Name:   "example-local-0",
					Port:   8080,
					Secure: false,
				},
				{
					Name:   "example-local-1",
					Port:   9090,
					Secure: false,
				},
			},
			returnedRoutes: testingutil.GetRouteListWithMultiple("nodejs", "app"),
			deletedURLs: []URL{
				getMachineReadableFormat(testingutil.GetSingleRoute("example-app", 8080, "nodejs", "app")),
				getMachineReadableFormat(testingutil.GetSingleRoute("example-1-app", 9100, "nodejs", "app")),
			},
			createdURLs: []URL{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "example-local-0-app",
					},
					Spec: URLSpec{
						Port:    8080,
						Secure:  false,
						urlKind: envinfo.ROUTE,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "example-local-1-app",
					},
					Spec: URLSpec{
						Port:    9090,
						Secure:  false,
						urlKind: envinfo.ROUTE,
					},
				},
			},
		},

		{
			name:                "0 urls on env file and cluster",
			componentName:       "nodejs",
			args:                args{isRouteSupported: true, isExperimentalModeEnabled: true},
			existingEnvInfoURLs: []envinfo.EnvInfoURL{},
			returnedRoutes:      &routev1.RouteList{},
			returnedIngress:     &extensionsv1.IngressList{},
		},
		{
			name:          "2 urls on env file and 0 on openshift cluster",
			componentName: "nodejs",
			args:          args{isRouteSupported: true, isExperimentalModeEnabled: true},
			existingEnvInfoURLs: []envinfo.EnvInfoURL{
				{
					Name:   "example",
					Port:   8080,
					Secure: false,
					Host:   "com",
					Kind:   envinfo.INGRESS,
				},
				{
					Name:   "example-1",
					Port:   9090,
					Secure: false,
					Host:   "com",
					Kind:   envinfo.INGRESS,
				},
			},
			returnedRoutes:  &routev1.RouteList{},
			returnedIngress: &extensionsv1.IngressList{},
			createdURLs: []URL{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "example",
					},
					Spec: URLSpec{
						Port:    8080,
						Secure:  false,
						Host:    "com",
						urlKind: envinfo.INGRESS,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "example-1",
					},
					Spec: URLSpec{
						Port:    9090,
						Secure:  false,
						Host:    "com",
						urlKind: envinfo.INGRESS,
					},
				},
			},
		},
		{
			name:                "0 urls on env file and 2 on openshift cluster",
			componentName:       "nodejs",
			args:                args{isRouteSupported: true, isExperimentalModeEnabled: true},
			existingEnvInfoURLs: []envinfo.EnvInfoURL{},
			returnedRoutes:      &routev1.RouteList{},
			returnedIngress:     fake.GetIngressListWithMultiple("nodejs"),
			deletedURLs: []URL{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "example-0",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "example-1",
					},
				},
			},
		},
		{
			name:          "2 urls on env file and 2 on openshift cluster, but they are different",
			componentName: "wildfly",
			args:          args{isRouteSupported: true, isExperimentalModeEnabled: true},
			existingEnvInfoURLs: []envinfo.EnvInfoURL{
				{
					Name:   "example-local-0",
					Port:   8080,
					Secure: false,
					Host:   "com",
					Kind:   envinfo.INGRESS,
				},
				{
					Name:   "example-local-1",
					Port:   9090,
					Secure: false,
					Host:   "com",
					Kind:   envinfo.INGRESS,
				},
			},
			returnedRoutes:  &routev1.RouteList{},
			returnedIngress: fake.GetIngressListWithMultiple("wildfly"),
			createdURLs: []URL{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "example-local-0",
					},
					Spec: URLSpec{
						Port:    8080,
						Secure:  false,
						Host:    "com",
						urlKind: envinfo.INGRESS,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "example-local-1",
					},
					Spec: URLSpec{
						Port:    9090,
						Secure:  false,
						Host:    "com",
						urlKind: envinfo.INGRESS,
					},
				},
			},
			deletedURLs: []URL{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "example-0",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "example-1",
					},
				},
			},
		},
		{
			name:          "2 (1 ingress,1 route) urls on env file and 2 on openshift cluster (1 ingress,1 route), but they are different",
			componentName: "nodejs",
			args:          args{isRouteSupported: true, isExperimentalModeEnabled: true},
			existingEnvInfoURLs: []envinfo.EnvInfoURL{
				{
					Name:   "example-local-0",
					Port:   8080,
					Secure: false,
					Kind:   envinfo.ROUTE,
				},
				{
					Name:   "example-local-1",
					Port:   9090,
					Secure: false,
					Host:   "com",
					Kind:   envinfo.INGRESS,
				},
			},
			returnedRoutes:  &routev1.RouteList{},
			returnedIngress: fake.GetIngressListWithMultiple("nodejs"),
			createdURLs: []URL{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "example-local-0",
					},
					Spec: URLSpec{
						Port:    8080,
						Secure:  false,
						urlKind: envinfo.ROUTE,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "example-local-1",
					},
					Spec: URLSpec{
						Port:    9090,
						Secure:  false,
						Host:    "com",
						urlKind: envinfo.INGRESS,
					},
				},
			},
			deletedURLs: []URL{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "example-0",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "example-1",
					},
				},
			},
		},
		{
			name:          "create a ingress on a kubernetes cluster",
			componentName: "nodejs",
			args:          args{isRouteSupported: false, isExperimentalModeEnabled: true},
			existingEnvInfoURLs: []envinfo.EnvInfoURL{
				{
					Name:      "example",
					Port:      8080,
					Secure:    true,
					Host:      "com",
					TLSSecret: "secret",
					Kind:      envinfo.INGRESS,
				},
			},
			returnedRoutes:  &routev1.RouteList{},
			returnedIngress: &extensionsv1.IngressList{},
			createdURLs: []URL{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "example",
					},
					Spec: URLSpec{
						Port:      8080,
						Secure:    true,
						Host:      "com",
						tLSSecret: "secret",
						urlKind:   envinfo.INGRESS,
					},
				},
			},
		},

		{
			name:          "url with same name exists on env and cluster but with different specs",
			componentName: "nodejs",
			args:          args{isRouteSupported: true, isExperimentalModeEnabled: true},
			existingEnvInfoURLs: []envinfo.EnvInfoURL{
				{
					Name:   "example-local-0",
					Port:   8080,
					Secure: false,
					Kind:   envinfo.ROUTE,
				},
			},
			returnedRoutes: &routev1.RouteList{},
			returnedIngress: &extensionsv1.IngressList{
				Items: []extensionsv1.Ingress{
					*fake.GetSingleIngress("example-local-0", "nodejs"),
				},
			},
			createdURLs: []URL{},
			deletedURLs: []URL{},
			wantErr:     true,
		},
		{
			name:            "url with same name exists on config and cluster but with different specs",
			componentName:   "nodejs",
			applicationName: "app",
			args:            args{isRouteSupported: true, isExperimentalModeEnabled: false},
			existingConfigURLs: []config.ConfigURL{
				{
					Name:   "example-local-0",
					Port:   8080,
					Secure: false,
				},
			},
			returnedRoutes: &routev1.RouteList{
				Items: []routev1.Route{
					testingutil.GetSingleRoute("example-local-0", 9090, "nodejs", "app"),
				},
			},
			returnedIngress: &extensionsv1.IngressList{},
			createdURLs:     []URL{},
			deletedURLs:     []URL{},
			wantErr:         true,
		},

		{
			name:          "create a secure route url",
			componentName: "nodejs",
			args:          args{isRouteSupported: true, isExperimentalModeEnabled: true},
			existingEnvInfoURLs: []envinfo.EnvInfoURL{
				{
					Name:   "example",
					Port:   8080,
					Secure: true,
					Kind:   envinfo.ROUTE,
				},
			},
			returnedRoutes:  &routev1.RouteList{},
			returnedIngress: &extensionsv1.IngressList{},
			createdURLs: []URL{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "example",
					},
					Spec: URLSpec{
						Port:    8080,
						Secure:  true,
						urlKind: envinfo.ROUTE,
					},
				},
			},
		},
		{
			name:          "create a secure ingress url with empty user given tls secret",
			componentName: "nodejs",
			args:          args{isRouteSupported: true, isExperimentalModeEnabled: true},
			existingEnvInfoURLs: []envinfo.EnvInfoURL{
				{
					Name:   "example",
					Port:   8080,
					Secure: true,
					Host:   "com",
					Kind:   envinfo.INGRESS,
				},
			},
			returnedRoutes:  &routev1.RouteList{},
			returnedIngress: &extensionsv1.IngressList{},
			createdURLs: []URL{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "example",
					},
					Spec: URLSpec{
						Port:    8080,
						Secure:  true,
						Host:    "com",
						urlKind: envinfo.INGRESS,
					},
				},
			},
		},
		{
			name:          "create a secure ingress url with user given tls secret",
			componentName: "nodejs",
			args:          args{isRouteSupported: true, isExperimentalModeEnabled: true},
			existingEnvInfoURLs: []envinfo.EnvInfoURL{
				{
					Name:      "example",
					Port:      8080,
					Secure:    true,
					Host:      "com",
					TLSSecret: "secret",
					Kind:      envinfo.INGRESS,
				},
			},
			returnedRoutes:  &routev1.RouteList{},
			returnedIngress: &extensionsv1.IngressList{},
			createdURLs: []URL{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "example",
					},
					Spec: URLSpec{
						Port:      8080,
						Secure:    true,
						Host:      "com",
						tLSSecret: "secret",
						urlKind:   envinfo.INGRESS,
					},
				},
			},
		},
	}
	for testNum, tt := range tests {
		tt.name = fmt.Sprintf("case %d: ", testNum+1) + tt.name
		t.Run(tt.name, func(t *testing.T) {
			fakeClient, fakeClientSet := occlient.FakeNew()
			fakeKClient, fakeKClientSet := kclient.FakeNew()

			fakeKClientSet.Kubernetes.PrependReactor("list", "ingresses", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, tt.returnedIngress, nil
			})

			fakeKClientSet.Kubernetes.PrependReactor("delete", "ingresses", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, nil, nil
			})

			fakeClientSet.RouteClientset.PrependReactor("list", "routes", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, tt.returnedRoutes, nil
			})

			fakeClientSet.RouteClientset.PrependReactor("delete", "routes", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, nil, nil
			})

			fakeKClientSet.Kubernetes.PrependReactor("get", "secrets", func(action ktesting.Action) (bool, runtime.Object, error) {
				if tt.existingEnvInfoURLs[0].TLSSecret != "" {
					return true, fake.GetSecret(tt.existingEnvInfoURLs[0].TLSSecret), nil
				}
				return true, fake.GetSecret(tt.componentName + "-tlssecret"), nil
			})

			fakeClientSet.AppsClientset.PrependReactor("get", "deploymentconfigs", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, testingutil.OneFakeDeploymentConfigWithMounts(tt.componentName, "local", tt.applicationName, map[string]*v1.PersistentVolumeClaim{}), nil
			})

			fakeKClientSet.Kubernetes.PrependReactor("get", "deployments", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, testingutil.CreateFakeDeployment(tt.componentName), nil
			})

			if err := Push(fakeClient, fakeKClient, PushParameters{
				ComponentName:             tt.componentName,
				ApplicationName:           tt.applicationName,
				ConfigURLs:                tt.existingConfigURLs,
				EnvURLS:                   tt.existingEnvInfoURLs,
				IsRouteSupported:          tt.args.isRouteSupported,
				IsExperimentalModeEnabled: tt.args.isExperimentalModeEnabled,
			}); (err != nil) != tt.wantErr {
				t.Errorf("Push() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				deletedURLMap := make(map[string]bool)
				for _, url := range tt.deletedURLs {
					found := false
					for _, action := range fakeKClientSet.Kubernetes.Actions() {
						value, ok := action.(ktesting.DeleteAction)
						if ok && value.GetVerb() == "delete" {
							deletedURLMap[value.GetName()] = true
							if value.GetName() == url.Name {
								found = true
								break
							}
						}
					}

					for _, action := range fakeClientSet.RouteClientset.Actions() {
						value, ok := action.(ktesting.DeleteAction)
						if ok && value.GetVerb() == "delete" {
							deletedURLMap[value.GetName()] = true
							if value.GetName() == url.Name {
								found = true
								break
							}
						}
					}
					if !found {
						t.Errorf("the url %s was not deleted", url.Name)
					}
				}

				if len(deletedURLMap) != len(tt.deletedURLs) {
					t.Errorf("number of deleted urls is different, want: %d,got: %d", len(tt.deletedURLs), len(deletedURLMap))
				}

				createdURLMap := make(map[string]bool)
				for _, url := range tt.createdURLs {
					found := false
					for _, action := range fakeKClientSet.Kubernetes.Actions() {
						value, ok := action.(ktesting.CreateAction)
						if ok {
							createdObject, ok := value.GetObject().(*extensionsv1.Ingress)
							if ok {
								createdURLMap[createdObject.Name] = true
								if createdObject.Name == url.Name &&
									(createdObject.Spec.TLS != nil) == url.Spec.Secure &&
									int(createdObject.Spec.Rules[0].HTTP.Paths[0].Backend.ServicePort.IntVal) == url.Spec.Port &&
									envinfo.INGRESS == url.Spec.urlKind &&
									fmt.Sprintf("%v.%v", url.Name, url.Spec.Host) == createdObject.Spec.Rules[0].Host {

									if url.Spec.Secure {
										secretName := tt.componentName + "-tlssecret"
										if url.Spec.tLSSecret != "" {
											secretName = url.Spec.tLSSecret
										}
										if createdObject.Spec.TLS[0].SecretName == secretName {
											found = true
											break
										}
									} else {
										found = true
										break
									}
								}
							}
						}
					}

					for _, action := range fakeClientSet.RouteClientset.Actions() {
						value, ok := action.(ktesting.CreateAction)
						if ok {
							createdObject, ok := value.GetObject().(*routev1.Route)
							if ok {
								createdURLMap[createdObject.Name] = true
								if createdObject.Name == url.Name &&
									(createdObject.Spec.TLS != nil) == url.Spec.Secure &&
									int(createdObject.Spec.Port.TargetPort.IntVal) == url.Spec.Port &&
									envinfo.ROUTE == url.Spec.urlKind {
									found = true
									break
								}
							}
						}
					}
					if !found {
						t.Errorf("the url %s was not created with proper specs", url.Name)
					}
				}

				if len(createdURLMap) != len(tt.createdURLs) {
					t.Errorf("number of created urls is different, want: %d,got: %d", len(tt.deletedURLs), len(deletedURLMap))
				}

				if !tt.args.isRouteSupported {
					if len(fakeClientSet.RouteClientset.Actions()) > 0 {
						t.Errorf("route is not supproted, total actions on the routeClient should be 0")
					}
				}
			}
		})
	}
}
