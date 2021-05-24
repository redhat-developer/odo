package url

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/devfile/library/pkg/devfile/generator"
	"github.com/golang/mock/gomock"
	"github.com/kylelemons/godebug/pretty"
	appsv1 "github.com/openshift/api/apps/v1"
	routev1 "github.com/openshift/api/route/v1"
	applabels "github.com/openshift/odo/pkg/application/labels"
	componentlabels "github.com/openshift/odo/pkg/component/labels"
	"github.com/openshift/odo/pkg/kclient"
	"github.com/openshift/odo/pkg/kclient/fake"
	"github.com/openshift/odo/pkg/localConfigProvider"
	"github.com/openshift/odo/pkg/occlient"
	"github.com/openshift/odo/pkg/testingutil"
	"github.com/openshift/odo/pkg/url/labels"
	"github.com/openshift/odo/pkg/util"
	v1 "k8s.io/api/core/v1"
	extensionsv1 "k8s.io/api/extensions/v1beta1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/openshift/odo/pkg/version"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ktesting "k8s.io/client-go/testing"
)

func TestCreate(t *testing.T) {
	type args struct {
		componentName    string
		applicationName  string
		urlName          string
		portNumber       int
		secure           bool
		host             string
		urlKind          localConfigProvider.URLKind
		isRouteSupported bool
		isS2I            bool
		tlsSecret        string
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
				isS2I:            true,
				urlKind:          localConfigProvider.ROUTE,
			},
			returnedRoute: &routev1.Route{
				ObjectMeta: metav1.ObjectMeta{
					Name: "nodejs-app",
					Labels: map[string]string{
						"app.kubernetes.io/part-of":  "app",
						"app.kubernetes.io/instance": "nodejs",
						applabels.App:                "app",
						applabels.ManagedBy:          "odo",
						applabels.ManagerVersion:     version.VERSION,
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
				isS2I:            true,
				urlKind:          localConfigProvider.ROUTE,
			},
			returnedRoute: &routev1.Route{
				ObjectMeta: metav1.ObjectMeta{
					Name: "example-url-app",
					Labels: map[string]string{
						"app.kubernetes.io/part-of":  "app",
						"app.kubernetes.io/instance": "nodejs",
						applabels.App:                "app",
						applabels.ManagedBy:          "odo",
						applabels.ManagerVersion:     version.VERSION,
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
				isS2I:            true,
				urlKind:          localConfigProvider.ROUTE,
			},
			returnedRoute: &routev1.Route{
				ObjectMeta: metav1.ObjectMeta{
					Name: "example-url-app",
					Labels: map[string]string{
						"app.kubernetes.io/part-of":  "app",
						"app.kubernetes.io/instance": "nodejs",
						applabels.App:                "app",
						applabels.ManagedBy:          "odo",
						applabels.ManagerVersion:     version.VERSION,
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
				componentName:    "nodejs",
				applicationName:  "app",
				urlName:          "nodejs",
				portNumber:       8080,
				host:             "com",
				isRouteSupported: true,
				urlKind:          localConfigProvider.INGRESS,
			},
			returnedIngress: fake.GetSingleIngress("nodejs-nodejs", "nodejs", "app"),
			want:            "http://nodejs.com",
			wantErr:         false,
		},
		{
			name: "Case 5: Create a ingress, with different name as component,instead of route on openshift cluster",
			args: args{
				componentName:    "nodejs",
				applicationName:  "app",
				urlName:          "example",
				portNumber:       8080,
				host:             "com",
				isRouteSupported: true,
				urlKind:          localConfigProvider.INGRESS,
			},
			returnedRoute: &routev1.Route{
				ObjectMeta: metav1.ObjectMeta{
					Name: "nodejs-app",
					Labels: map[string]string{
						"app.kubernetes.io/part-of":  "app",
						"app.kubernetes.io/instance": "nodejs",
						applabels.App:                "app",
						applabels.ManagedBy:          "odo",
						applabels.ManagerVersion:     version.VERSION,
						"odo.openshift.io/url-name":  "nodejs-nodejs",
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
			returnedIngress: fake.GetSingleIngress("example-nodejs", "nodejs", "app"),
			want:            "http://example.com",
			wantErr:         false,
		},
		{
			name: "Case 6: Create a secure ingress, instead of route on openshift cluster, default tls exists",
			args: args{
				componentName:    "nodejs",
				applicationName:  "app",
				urlName:          "example",
				portNumber:       8080,
				host:             "com",
				isRouteSupported: true,
				secure:           true,
				urlKind:          localConfigProvider.INGRESS,
			},
			returnedIngress:  fake.GetSingleIngress("example-nodejs", "nodejs", "app"),
			defaultTLSExists: true,
			want:             "https://example.com",
			wantErr:          false,
		},
		{
			name: "Case 7: Create a secure ingress, instead of route on openshift cluster and default tls doesn't exist",
			args: args{
				componentName:    "nodejs",
				applicationName:  "app",
				urlName:          "example",
				portNumber:       8080,
				host:             "com",
				isRouteSupported: true,
				secure:           true,
				urlKind:          localConfigProvider.INGRESS,
			},
			returnedIngress:  fake.GetSingleIngress("example-nodejs", "nodejs", "app"),
			defaultTLSExists: false,
			want:             "https://example.com",
			wantErr:          false,
		},
		{
			name: "Case 8: Fail when while creating ingress when user given tls secret doesn't exists",
			args: args{
				applicationName:  "app",
				componentName:    "nodejs",
				urlName:          "example",
				portNumber:       8080,
				host:             "com",
				isRouteSupported: true,
				secure:           true,
				tlsSecret:        "user-secret",
				urlKind:          localConfigProvider.INGRESS,
			},
			returnedIngress:    fake.GetSingleIngress("example", "nodejs", "app"),
			defaultTLSExists:   false,
			userGivenTLSExists: false,
			want:               "http://example.com",
			wantErr:            true,
		},
		{
			name: "Case 9: Create a secure ingress, instead of route on openshift cluster, user tls secret does exists",
			args: args{
				applicationName:  "app",
				componentName:    "nodejs",
				urlName:          "example",
				portNumber:       8080,
				host:             "com",
				isRouteSupported: true,
				secure:           true,
				tlsSecret:        "user-secret",
				urlKind:          localConfigProvider.INGRESS,
			},
			returnedIngress:    fake.GetSingleIngress("example-nodejs", "nodejs", "app"),
			defaultTLSExists:   false,
			userGivenTLSExists: true,
			want:               "https://example.com",
			wantErr:            false,
		},

		{
			name: "Case 10: invalid url kind",
			args: args{
				applicationName:  "app",
				componentName:    "nodejs",
				urlName:          "example",
				portNumber:       8080,
				host:             "com",
				isRouteSupported: true,
				secure:           true,
				tlsSecret:        "user-secret",
				urlKind:          "blah",
			},
			returnedIngress:    fake.GetSingleIngress("example-nodejs", "nodejs", "app"),
			defaultTLSExists:   false,
			userGivenTLSExists: true,
			want:               "",
			wantErr:            true,
		},
		{
			name: "Case 11: route is not supported on the cluster",
			args: args{
				componentName:    "nodejs",
				applicationName:  "app",
				urlName:          "example",
				isRouteSupported: false,
				urlKind:          localConfigProvider.ROUTE,
			},
			returnedIngress:    fake.GetSingleIngress("example", "nodejs", "app"),
			defaultTLSExists:   false,
			userGivenTLSExists: true,
			want:               "",
			wantErr:            true,
		},
		{
			name: "Case 11: secretName used without secure flag",
			args: args{
				componentName:    "nodejs",
				applicationName:  "app",
				urlName:          "example",
				isRouteSupported: false,
				tlsSecret:        "secret",
				urlKind:          localConfigProvider.ROUTE,
			},
			returnedIngress:    fake.GetSingleIngress("example", "nodejs", "app"),
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
			if tt.args.urlKind == localConfigProvider.INGRESS {
				serviceName = tt.args.componentName

			} else if tt.args.urlKind == localConfigProvider.ROUTE {
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

			urlCreateParameters := CreateParameters{
				urlName:         tt.args.urlName,
				portNumber:      tt.args.portNumber,
				secureURL:       tt.args.secure,
				componentName:   tt.args.componentName,
				applicationName: tt.args.applicationName,
				host:            tt.args.host,
				secretName:      tt.args.tlsSecret,
				urlKind:         tt.args.urlKind,
			}

			got, err := Create(client, fakeKClient, urlCreateParameters, tt.args.isRouteSupported, tt.args.isS2I)

			if err == nil && !tt.wantErr {
				if tt.args.urlKind == localConfigProvider.INGRESS {
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
					tt.returnedIngress.Labels["odo.openshift.io/url-name"] = tt.args.urlName
					if !reflect.DeepEqual(createdIngress.Name, tt.returnedIngress.Name) {
						t.Errorf("ingress name not matching, expected: %s, got %s", tt.returnedIngress.Name, createdIngress.Name)
					}
					if !reflect.DeepEqual(createdIngress.Labels, tt.returnedIngress.Labels) {
						t.Errorf("ingress labels not matching, %v", pretty.Compare(tt.returnedIngress.Labels, createdIngress.Labels))
					}

					wantedIngressSpecParams := generator.IngressSpecParams{
						ServiceName:   serviceName,
						IngressDomain: tt.args.host,
						PortNumber:    intstr.FromInt(tt.args.portNumber),
						TLSSecretName: tt.args.tlsSecret,
					}

					if !reflect.DeepEqual(createdIngress.Spec.Rules[0].HTTP.Paths[0].Backend.ServicePort.IntVal, wantedIngressSpecParams.PortNumber.IntVal) {
						t.Errorf("ingress port not matching, expected: %s, got %s", tt.returnedRoute.Spec.Port, createdIngress.Spec.Rules[0].HTTP.Paths[0].Backend.ServicePort.StrVal)
					}
					if tt.args.secure {
						if wantedIngressSpecParams.TLSSecretName == "" {
							wantedIngressSpecParams.TLSSecretName = tt.args.componentName + "-tlssecret"
						}
						if !reflect.DeepEqual(createdIngress.Spec.TLS[0].SecretName, wantedIngressSpecParams.TLSSecretName) {
							t.Errorf("ingress tls name not matching, expected: %s, got %s", wantedIngressSpecParams.TLSSecretName, createdIngress.Spec.TLS)
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
								applabels.ManagedBy:            "odo",
								applabels.ManagerVersion:       version.VERSION,
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
								applabels.ManagedBy:            "odo",
								applabels.ManagerVersion:       version.VERSION,
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
								applabels.ManagedBy:            "odo",
								applabels.ManagerVersion:       version.VERSION,
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
								applabels.ManagedBy:            "odo",
								applabels.ManagerVersion:       version.VERSION,
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

func TestPush(t *testing.T) {
	type args struct {
		isRouteSupported bool
		isS2I            bool
	}
	tests := []struct {
		name                string
		args                args
		componentName       string
		applicationName     string
		existingLocalURLs   []localConfigProvider.LocalURL
		existingClusterURLs URLList
		deletedURLs         []URL
		createdURLs         []URL
		wantErr             bool
	}{
		{
			name: "no urls on local config and cluster",
			args: args{
				isRouteSupported: true,
				isS2I:            true,
			},
			componentName:   "nodejs",
			applicationName: "app",
		},
		{
			name:            "2 urls on local config and 0 on openshift cluster",
			componentName:   "nodejs",
			applicationName: "app",
			args: args{
				isRouteSupported: true,
				isS2I:            true,
			},
			existingLocalURLs: []localConfigProvider.LocalURL{
				{
					Name:   "example",
					Port:   8080,
					Secure: false,
					Kind:   localConfigProvider.ROUTE,
				},
				{
					Name:   "example-1",
					Port:   9090,
					Secure: false,
					Kind:   localConfigProvider.ROUTE,
				},
			},
			createdURLs: []URL{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "example-app",
					},
					Spec: URLSpec{
						Port:   8080,
						Secure: false,
						Kind:   localConfigProvider.ROUTE,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "example-1-app",
					},
					Spec: URLSpec{
						Port:   9090,
						Secure: false,
						Kind:   localConfigProvider.ROUTE,
					},
				},
			},
		},
		{
			name:            "0 url on local config and 2 on openshift cluster",
			componentName:   "wildfly",
			applicationName: "app",
			args:            args{isRouteSupported: true, isS2I: true},
			existingClusterURLs: getMachineReadableFormatForList([]URL{
				getMachineReadableFormat(testingutil.GetSingleRoute("example", 8080, "wildfly", "app")),
				getMachineReadableFormat(testingutil.GetSingleRoute("example-1", 9100, "wildfly", "app")),
			}),
			deletedURLs: []URL{
				getMachineReadableFormat(testingutil.GetSingleRoute("example-app", 8080, "nodejs", "app")),
				getMachineReadableFormat(testingutil.GetSingleRoute("example-1-app", 9100, "nodejs", "app")),
			},
		},
		{
			name:            "2 url on local config and 2 on openshift cluster, but they are different",
			componentName:   "nodejs",
			applicationName: "app",
			args:            args{isRouteSupported: true, isS2I: true},
			existingLocalURLs: []localConfigProvider.LocalURL{
				{
					Name:   "example-local-0",
					Port:   8080,
					Secure: false,
					Kind:   localConfigProvider.ROUTE,
				},
				{
					Name:   "example-local-1",
					Port:   9090,
					Secure: false,
					Kind:   localConfigProvider.ROUTE,
				},
			},
			existingClusterURLs: getMachineReadableFormatForList([]URL{
				getMachineReadableFormat(testingutil.GetSingleRoute("example", 8080, "wildfly", "app")),
				getMachineReadableFormat(testingutil.GetSingleRoute("example-1", 9100, "wildfly", "app")),
			}),
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
						Port:   8080,
						Secure: false,
						Kind:   localConfigProvider.ROUTE,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "example-local-1-app",
					},
					Spec: URLSpec{
						Port:   9090,
						Secure: false,
						Kind:   localConfigProvider.ROUTE,
					},
				},
			},
		},
		{
			name:            "2 url on local config and openshift cluster are in sync",
			componentName:   "nodejs",
			applicationName: "app",
			args:            args{isRouteSupported: true, isS2I: true},
			existingLocalURLs: []localConfigProvider.LocalURL{
				{
					Name:   "example",
					Port:   8080,
					Secure: false,
					Path:   "/",
					Kind:   localConfigProvider.ROUTE,
				},
				{
					Name:   "example-1",
					Port:   9100,
					Secure: false,
					Path:   "/",
					Kind:   localConfigProvider.ROUTE,
				},
			},
			existingClusterURLs: getMachineReadableFormatForList([]URL{
				getMachineReadableFormat(testingutil.GetSingleRoute("example", 8080, "wildfly", "app")),
				getMachineReadableFormat(testingutil.GetSingleRoute("example-1", 9100, "wildfly", "app")),
			}),
			deletedURLs: []URL{},
			createdURLs: []URL{},
		},
		{
			name:              "0 urls on env file and cluster",
			componentName:     "nodejs",
			applicationName:   "app",
			args:              args{isRouteSupported: true},
			existingLocalURLs: []localConfigProvider.LocalURL{},
		},
		{
			name:            "2 urls on env file and 0 on openshift cluster",
			componentName:   "nodejs",
			applicationName: "app",
			args:            args{isRouteSupported: true},
			existingLocalURLs: []localConfigProvider.LocalURL{
				{
					Name: "example",
					Host: "com",
					Port: 8080,
					Kind: localConfigProvider.INGRESS,
				},
				{
					Name: "example-1",
					Host: "com",
					Port: 9090,
					Kind: localConfigProvider.INGRESS,
				},
			},
			createdURLs: []URL{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "example-nodejs",
					},
					Spec: URLSpec{
						Port:   8080,
						Secure: false,
						Host:   "com",
						Kind:   localConfigProvider.INGRESS,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "example-1-nodejs",
					},
					Spec: URLSpec{
						Port:   9090,
						Secure: false,
						Host:   "com",
						Kind:   localConfigProvider.INGRESS,
					},
				},
			},
		},
		{
			name:              "0 urls on env file and 2 on openshift cluster",
			componentName:     "nodejs",
			applicationName:   "app",
			args:              args{isRouteSupported: true},
			existingLocalURLs: []localConfigProvider.LocalURL{},
			existingClusterURLs: getMachineReadableFormatForList([]URL{
				getMachineReadableFormatIngress(*fake.GetSingleIngress("example-0", "nodejs", "app")),
				getMachineReadableFormatIngress(*fake.GetSingleIngress("example-1", "nodejs", "app")),
			}),
			deletedURLs: []URL{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "example-0-nodejs",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "example-1-nodejs",
					},
				},
			},
		},
		{
			name:            "2 urls on env file and 2 on openshift cluster, but they are different",
			componentName:   "wildfly",
			applicationName: "app",
			args:            args{isRouteSupported: true},
			existingLocalURLs: []localConfigProvider.LocalURL{
				{
					Name: "example-local-0",
					Host: "com",
					Port: 8080,
					Kind: localConfigProvider.INGRESS,
				},
				{
					Name: "example-local-1",
					Host: "com",
					Port: 9090,
					Kind: localConfigProvider.INGRESS,
				},
			},
			existingClusterURLs: getMachineReadableFormatForList([]URL{
				getMachineReadableFormatIngress(*fake.GetSingleIngress("example-0", "nodejs", "app")),
				getMachineReadableFormatIngress(*fake.GetSingleIngress("example-1", "nodejs", "app")),
			}),
			createdURLs: []URL{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "example-local-0-wildfly",
					},
					Spec: URLSpec{
						Port:   8080,
						Secure: false,
						Host:   "com",
						Kind:   localConfigProvider.INGRESS,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "example-local-1-wildfly",
					},
					Spec: URLSpec{
						Port:   9090,
						Secure: false,
						Host:   "com",
						Kind:   localConfigProvider.INGRESS,
					},
				},
			},
			deletedURLs: []URL{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "example-0-wildfly",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "example-1-wildfly",
					},
				},
			},
		},
		{
			name:            "2 urls on env file and openshift cluster are in sync",
			componentName:   "wildfly",
			applicationName: "app",
			args:            args{isRouteSupported: true},
			existingLocalURLs: []localConfigProvider.LocalURL{
				{
					Name: "example-0",
					Host: "com",
					Port: 8080,
					Kind: localConfigProvider.INGRESS,
				},
				{
					Name: "example-1",
					Host: "com",
					Port: 9090,
					Kind: localConfigProvider.INGRESS,
				},
			},
			existingClusterURLs: getMachineReadableFormatForList([]URL{
				getMachineReadableFormatIngress(*fake.GetSingleIngress("example-0", "wildfly", "app")),
				getMachineReadableFormatIngress(*fake.GetSingleIngress("example-1", "wildfly", "app")),
			}),
			createdURLs: []URL{},
			deletedURLs: []URL{},
		},
		{
			name:            "2 (1 ingress,1 route) urls on env file and 2 on openshift cluster (1 ingress,1 route), but they are different",
			componentName:   "nodejs",
			applicationName: "app",
			args:            args{isRouteSupported: true},
			existingLocalURLs: []localConfigProvider.LocalURL{
				{
					Name: "example-local-0",
					Port: 8080,
					Kind: localConfigProvider.ROUTE,
				},
				{
					Name: "example-local-1",
					Host: "com",
					Port: 9090,
					Kind: localConfigProvider.INGRESS,
				},
			},
			existingClusterURLs: getMachineReadableFormatForList([]URL{
				getMachineReadableFormatIngress(*fake.GetSingleIngress("example-0", "nodejs", "app")),
				getMachineReadableFormatIngress(*fake.GetSingleIngress("example-1", "nodejs", "app")),
			}),
			createdURLs: []URL{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "example-local-0-nodejs",
					},
					Spec: URLSpec{
						Port:   8080,
						Secure: false,
						Kind:   localConfigProvider.ROUTE,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "example-local-1-nodejs",
					},
					Spec: URLSpec{
						Port:   9090,
						Secure: false,
						Host:   "com",
						Kind:   localConfigProvider.INGRESS,
					},
				},
			},
			deletedURLs: []URL{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "example-0-nodejs",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "example-1-nodejs",
					},
				},
			},
		},
		{
			name:            "create a ingress on a kubernetes cluster",
			componentName:   "nodejs",
			applicationName: "app",
			args:            args{isRouteSupported: false},
			existingLocalURLs: []localConfigProvider.LocalURL{
				{
					Name:      "example",
					Host:      "com",
					TLSSecret: "secret",
					Port:      8080,
					Secure:    true,
					Kind:      localConfigProvider.INGRESS,
				},
			},
			createdURLs: []URL{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "example-nodejs",
					},
					Spec: URLSpec{
						Port:      8080,
						Secure:    true,
						Host:      "com",
						TLSSecret: "secret",
						Kind:      localConfigProvider.INGRESS,
					},
				},
			},
		},
		{
			name:            "url with same name exists on env and cluster but with different specs",
			componentName:   "nodejs",
			applicationName: "app",
			args: args{
				isRouteSupported: true,
			},
			existingLocalURLs: []localConfigProvider.LocalURL{
				{
					Name: "example-local-0",
					Port: 8080,
					Kind: localConfigProvider.ROUTE,
				},
			},
			existingClusterURLs: getMachineReadableFormatForList([]URL{
				getMachineReadableFormatIngress(*fake.GetSingleIngress("example-local-0", "nodejs", "app")),
			}),
			createdURLs: []URL{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "example-local-0-nodejs",
					},
					Spec: URLSpec{
						Port:   8080,
						Secure: false,
						Kind:   localConfigProvider.ROUTE,
					},
				},
			},
			deletedURLs: []URL{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "example-local-0-nodejs",
					},
				},
			},
			wantErr: false,
		},
		{
			name:            "url with same name exists on config and cluster but with different specs",
			componentName:   "nodejs",
			applicationName: "app",
			args:            args{isRouteSupported: true, isS2I: true},
			existingLocalURLs: []localConfigProvider.LocalURL{
				{
					Name:   "example-local-0",
					Port:   8080,
					Secure: false,
					Kind:   localConfigProvider.ROUTE,
				},
			},
			existingClusterURLs: getMachineReadableFormatForList([]URL{
				getMachineReadableFormat(testingutil.GetSingleRoute("example-local-0", 9090, "nodejs", "app")),
			}),
			createdURLs: []URL{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "example-local-0-app",
					},
					Spec: URLSpec{
						Port:   8080,
						Secure: false,
						Kind:   localConfigProvider.ROUTE,
					},
				},
			},
			deletedURLs: []URL{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "example-local-0-app",
					},
				},
			},
			wantErr: false,
		},
		{
			name:            "create a secure route url",
			componentName:   "nodejs",
			applicationName: "app",
			args:            args{isRouteSupported: true},
			existingLocalURLs: []localConfigProvider.LocalURL{
				{
					Name:   "example",
					Port:   8080,
					Secure: true,
					Kind:   localConfigProvider.ROUTE,
				},
			},
			createdURLs: []URL{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "example-nodejs",
					},
					Spec: URLSpec{
						Port:   8080,
						Secure: true,
						Kind:   localConfigProvider.ROUTE,
					},
				},
			},
		},
		{
			name:            "create a secure ingress url with empty user given tls secret",
			componentName:   "nodejs",
			applicationName: "app",
			args:            args{isRouteSupported: true},
			existingLocalURLs: []localConfigProvider.LocalURL{
				{
					Name:   "example",
					Host:   "com",
					Secure: true,
					Port:   8080,
					Kind:   localConfigProvider.INGRESS,
				},
			},
			createdURLs: []URL{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "example-nodejs",
					},
					Spec: URLSpec{
						Port:   8080,
						Secure: true,
						Host:   "com",
						Kind:   localConfigProvider.INGRESS,
					},
				},
			},
		},
		{
			name:            "create a secure ingress url with user given tls secret",
			componentName:   "nodejs",
			applicationName: "app",
			args:            args{isRouteSupported: true},
			existingLocalURLs: []localConfigProvider.LocalURL{
				{
					Name:      "example",
					Host:      "com",
					TLSSecret: "secret",
					Port:      8080,
					Secure:    true,
					Kind:      localConfigProvider.INGRESS,
				},
			},
			createdURLs: []URL{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "example-nodejs",
					},
					Spec: URLSpec{
						Port:      8080,
						Secure:    true,
						Host:      "com",
						TLSSecret: "secret",
						Kind:      localConfigProvider.INGRESS,
					},
				},
			},
		},
		{
			name:          "no host defined for ingress should not create any URL",
			componentName: "nodejs",
			args:          args{isRouteSupported: false},
			existingLocalURLs: []localConfigProvider.LocalURL{
				{
					Name: "example",
					Port: 8080,
					Kind: localConfigProvider.ROUTE,
				},
			},
			wantErr:     false,
			createdURLs: []URL{},
		},
		{
			name:            "should create route in openshift cluster if endpoint is defined in devfile",
			componentName:   "nodejs",
			applicationName: "app",
			args:            args{isRouteSupported: true},
			existingLocalURLs: []localConfigProvider.LocalURL{
				{
					Name:   "example",
					Port:   8080,
					Kind:   localConfigProvider.ROUTE,
					Secure: false,
				},
			},
			wantErr: false,
			createdURLs: []URL{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "example-nodejs",
					},
					Spec: URLSpec{
						Port:   8080,
						Secure: false,
						Kind:   localConfigProvider.ROUTE,
						Path:   "/",
					},
				},
			},
		},
		{
			name:            "should create ingress if endpoint is defined in devfile",
			componentName:   "nodejs",
			applicationName: "app",
			args:            args{isRouteSupported: true},
			existingLocalURLs: []localConfigProvider.LocalURL{
				{
					Name: "example",
					Host: "com",
					Port: 8080,
					Kind: localConfigProvider.INGRESS,
				},
			},
			wantErr: false,
			createdURLs: []URL{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "example-nodejs",
					},
					Spec: URLSpec{
						Port:   8080,
						Secure: false,
						Host:   "com",
						Kind:   localConfigProvider.INGRESS,
						Path:   "/",
					},
				},
			},
		},
		{
			name:            "should create route in openshift cluster with path defined in devfile",
			componentName:   "nodejs",
			applicationName: "app",
			args:            args{isRouteSupported: true},
			existingLocalURLs: []localConfigProvider.LocalURL{
				{
					Name:   "example",
					Port:   8080,
					Secure: false,
					Path:   "/testpath",
					Kind:   localConfigProvider.ROUTE,
				},
			},
			wantErr: false,
			createdURLs: []URL{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "example-nodejs",
					},
					Spec: URLSpec{
						Port:   8080,
						Secure: false,
						Kind:   localConfigProvider.ROUTE,
						Path:   "/testpath",
					},
				},
			},
		},
		{
			name:            "should create ingress with path defined in devfile",
			componentName:   "nodejs",
			applicationName: "app",
			args:            args{isRouteSupported: true},
			existingLocalURLs: []localConfigProvider.LocalURL{
				{
					Name:   "example",
					Host:   "com",
					Port:   8080,
					Secure: false,
					Path:   "/testpath",
					Kind:   localConfigProvider.INGRESS,
				},
			},
			wantErr: false,
			createdURLs: []URL{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "example-nodejs",
					},
					Spec: URLSpec{
						Port:   8080,
						Secure: false,
						Host:   "com",
						Kind:   localConfigProvider.INGRESS,
						Path:   "/testpath",
					},
				},
			},
		},
	}
	for testNum, tt := range tests {
		tt.name = fmt.Sprintf("case %d: ", testNum+1) + tt.name
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLocalConfigProvider := localConfigProvider.NewMockLocalConfigProvider(ctrl)
			mockLocalConfigProvider.EXPECT().GetName().Return(tt.componentName).AnyTimes()
			mockLocalConfigProvider.EXPECT().GetApplication().Return(tt.applicationName).AnyTimes()
			mockLocalConfigProvider.EXPECT().ListURLs().Return(tt.existingLocalURLs, nil)

			mockURLClient := NewMockClient(ctrl)
			mockURLClient.EXPECT().ListFromCluster().Return(tt.existingClusterURLs, nil)

			fakeClient, fakeClientSet := occlient.FakeNew()
			fakeKClient, fakeKClientSet := kclient.FakeNew()

			fakeClient.SetKubeClient(fakeKClient)

			fakeKClientSet.Kubernetes.PrependReactor("delete", "ingresses", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, nil, nil
			})

			fakeClientSet.RouteClientset.PrependReactor("delete", "routes", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, nil, nil
			})

			fakeKClientSet.Kubernetes.PrependReactor("get", "secrets", func(action ktesting.Action) (bool, runtime.Object, error) {
				if tt.existingLocalURLs[0].TLSSecret != "" {
					return true, fake.GetSecret(tt.existingLocalURLs[0].TLSSecret), nil
				}
				return true, fake.GetSecret(tt.componentName + "-tlssecret"), nil
			})

			fakeClientSet.AppsClientset.PrependReactor("get", "deploymentconfigs", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, testingutil.OneFakeDeploymentConfigWithMounts(tt.componentName, "local", tt.applicationName, map[string]*v1.PersistentVolumeClaim{}), nil
			})

			fakeKClientSet.Kubernetes.PrependReactor("get", "deployments", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, testingutil.CreateFakeDeployment(tt.componentName), nil
			})

			if err := Push(fakeClient, PushParameters{
				LocalConfig:      mockLocalConfigProvider,
				URLClient:        mockURLClient,
				IsRouteSupported: tt.args.isRouteSupported,
				IsS2I:            tt.args.isS2I,
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
								expectedHost := fmt.Sprintf("%v.%v", strings.Split(url.Name, "-"+tt.componentName)[0], url.Spec.Host)
								if createdObject.Name == url.Name &&
									(createdObject.Spec.TLS != nil) == url.Spec.Secure &&
									int(createdObject.Spec.Rules[0].HTTP.Paths[0].Backend.ServicePort.IntVal) == url.Spec.Port &&
									localConfigProvider.INGRESS == url.Spec.Kind &&
									expectedHost == createdObject.Spec.Rules[0].Host {

									if url.Spec.Secure {
										secretName := tt.componentName + "-tlssecret"
										if url.Spec.TLSSecret != "" {
											secretName = url.Spec.TLSSecret
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
									localConfigProvider.ROUTE == url.Spec.Kind {
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
					t.Errorf("number of created urls is different, want: %d,got: %d", len(tt.createdURLs), len(createdURLMap))
				}

				if !tt.args.isRouteSupported {
					if len(fakeClientSet.RouteClientset.Actions()) > 0 {
						t.Errorf("route is not supproted, total actions on the routeClient should be 0")
					}
				}

				if len(tt.createdURLs) == 0 && len(tt.deletedURLs) == 0 {
					if len(fakeClientSet.RouteClientset.Actions()) > 1 {
						t.Errorf("when urls are in sync, total action for route client set should be less than 1")
					}

					if len(fakeClientSet.Kubernetes.Actions()) > 1 {
						t.Errorf("when urls are in snyc, total action for kubernetes client set should be less than 1")
					}
				}
			}
		})
	}
}

func TestConvertEnvinfoURL(t *testing.T) {
	serviceName := "testService"
	urlName := "testURL"
	host := "com"
	secretName := "test-tls-secret"
	tests := []struct {
		name       string
		envInfoURL localConfigProvider.LocalURL
		wantURL    URL
	}{
		{
			name: "Case 1: insecure URL",
			envInfoURL: localConfigProvider.LocalURL{
				Name:   urlName,
				Host:   host,
				Port:   8080,
				Secure: false,
				Kind:   localConfigProvider.INGRESS,
			},
			wantURL: URL{
				TypeMeta:   metav1.TypeMeta{Kind: "url", APIVersion: "odo.dev/v1alpha1"},
				ObjectMeta: metav1.ObjectMeta{Name: urlName},
				Spec:       URLSpec{Host: fmt.Sprintf("%s.%s", urlName, host), Port: 8080, Secure: false, Kind: localConfigProvider.INGRESS},
			},
		},
		{
			name: "Case 2: secure Ingress URL without tls secret defined",
			envInfoURL: localConfigProvider.LocalURL{
				Name:   urlName,
				Host:   host,
				Port:   8080,
				Secure: true,
				Kind:   localConfigProvider.INGRESS,
			},
			wantURL: URL{
				TypeMeta:   metav1.TypeMeta{Kind: "url", APIVersion: "odo.dev/v1alpha1"},
				ObjectMeta: metav1.ObjectMeta{Name: urlName},
				Spec:       URLSpec{Host: fmt.Sprintf("%s.%s", urlName, host), Port: 8080, Secure: true, TLSSecret: fmt.Sprintf("%s-tlssecret", serviceName), Kind: localConfigProvider.INGRESS},
			},
		},
		{
			name: "Case 3: secure Ingress URL with tls secret defined",
			envInfoURL: localConfigProvider.LocalURL{
				Name:      urlName,
				Host:      host,
				Port:      8080,
				Secure:    true,
				TLSSecret: secretName,
				Kind:      localConfigProvider.INGRESS,
			},
			wantURL: URL{
				TypeMeta:   metav1.TypeMeta{Kind: "url", APIVersion: "odo.dev/v1alpha1"},
				ObjectMeta: metav1.ObjectMeta{Name: urlName},
				Spec:       URLSpec{Host: fmt.Sprintf("%s.%s", urlName, host), Port: 8080, Secure: true, TLSSecret: secretName, Kind: localConfigProvider.INGRESS},
			},
		},
		{
			name: "Case 4: Insecure route URL",
			envInfoURL: localConfigProvider.LocalURL{
				Name: urlName,
				Port: 8080,
				Kind: localConfigProvider.ROUTE,
			},
			wantURL: URL{
				TypeMeta:   metav1.TypeMeta{Kind: "url", APIVersion: "odo.dev/v1alpha1"},
				ObjectMeta: metav1.ObjectMeta{Name: urlName},
				Spec:       URLSpec{Port: 8080, Secure: false, Kind: localConfigProvider.ROUTE},
			},
		},
		{
			name: "Case 4: Secure route URL",
			envInfoURL: localConfigProvider.LocalURL{
				Name:   urlName,
				Port:   8080,
				Secure: true,
				Kind:   localConfigProvider.ROUTE,
			},
			wantURL: URL{
				TypeMeta:   metav1.TypeMeta{Kind: "url", APIVersion: "odo.dev/v1alpha1"},
				ObjectMeta: metav1.ObjectMeta{Name: urlName},
				Spec:       URLSpec{Port: 8080, Secure: true, Kind: localConfigProvider.ROUTE},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := ConvertEnvinfoURL(tt.envInfoURL, serviceName)
			if !reflect.DeepEqual(url, tt.wantURL) {
				t.Errorf("Expected %v, got %v", tt.wantURL, url)
			}
		})
	}
}

func TestGetURLString(t *testing.T) {
	cases := []struct {
		name          string
		protocol      string
		URL           string
		ingressDomain string
		isS2I         bool
		expected      string
	}{
		{
			name:          "simple s2i case",
			protocol:      "http",
			URL:           "example.com",
			ingressDomain: "",
			isS2I:         true,
			expected:      "http://example.com",
		},
		{
			name:          "all blank with s2i",
			protocol:      "",
			URL:           "",
			ingressDomain: "",
			isS2I:         true,
			expected:      "",
		},
		{
			name:          "all blank without s2i",
			protocol:      "",
			URL:           "",
			ingressDomain: "",
			isS2I:         false,
			expected:      "",
		},
		{
			name:          "devfile case",
			protocol:      "http",
			URL:           "",
			ingressDomain: "spring-8080.192.168.39.247.nip.io",
			isS2I:         false,
			expected:      "http://spring-8080.192.168.39.247.nip.io",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			output := GetURLString(testCase.protocol, testCase.URL, testCase.ingressDomain, testCase.isS2I)
			if output != testCase.expected {
				t.Errorf("Expected: %v, got %v", testCase.expected, output)

			}
		})
	}
}
