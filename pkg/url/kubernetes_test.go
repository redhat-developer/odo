package url

import (
	"fmt"
	"reflect"
	"testing"

	networkingv1 "k8s.io/api/networking/v1"

	"github.com/devfile/library/pkg/devfile/generator"
	"github.com/golang/mock/gomock"
	"github.com/kylelemons/godebug/pretty"
	routev1 "github.com/openshift/api/route/v1"
	applabels "github.com/openshift/odo/pkg/application/labels"
	"github.com/openshift/odo/pkg/kclient"
	"github.com/openshift/odo/pkg/kclient/fake"
	"github.com/openshift/odo/pkg/localConfigProvider"
	"github.com/openshift/odo/pkg/occlient"
	"github.com/openshift/odo/pkg/testingutil"
	urlLabels "github.com/openshift/odo/pkg/url/labels"
	"github.com/openshift/odo/pkg/version"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	extensionsv1 "k8s.io/api/extensions/v1beta1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
	ktesting "k8s.io/client-go/testing"
)

func getFakeURL(name string, host string, port int, path string, protocol string, kind localConfigProvider.URLKind, urlState StateType) URL {
	return URL{
		TypeMeta: v1.TypeMeta{
			Kind:       "url",
			APIVersion: "odo.dev/v1alpha1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: name,
		},
		Spec: URLSpec{
			Host:     host,
			Protocol: protocol,
			Kind:     kind,
			Path:     path,
			Port:     port,
		},
		Status: URLStatus{
			State: urlState,
		},
	}
}

func Test_kubernetesClient_ListCluster(t *testing.T) {
	componentName := "nodejs"
	appName := "app"
	ingress0 := fake.GetSingleExtensionV1Ingress("testIngress0", componentName, appName)
	ingress1 := fake.GetSingleExtensionV1Ingress("testIngress1", componentName, appName)

	route0 := testingutil.GetSingleRoute("testRoute0", 8080, componentName, appName)
	route1 := testingutil.GetSingleRoute("testRoute1", 8080, componentName, appName)
	routeOwnedByIngress := testingutil.GetSingleRoute("testRoute1-ingress", 8080, componentName, appName)
	routeOwnedByIngress.SetOwnerReferences([]v1.OwnerReference{
		{
			Kind: "Ingress",
		},
	})

	type fields struct {
		generic          generic
		isRouteSupported bool
	}
	tests := []struct {
		name              string
		fields            fields
		returnedIngresses extensionsv1.IngressList
		returnedRoutes    routev1.RouteList
		want              URLList
		wantErr           bool
	}{
		{
			name: "case 1: list ingresses when route resource is not supported",
			fields: fields{
				generic: generic{
					appName:       "app",
					componentName: componentName,
				},
				isRouteSupported: false,
			},
			returnedIngresses: extensionsv1.IngressList{
				Items: []extensionsv1.Ingress{
					*ingress0,
					*ingress1,
				},
			},
			want: getMachineReadableFormatForList([]URL{
				getMachineReadableFormatExtensionV1Ingress(*ingress0),
				getMachineReadableFormatExtensionV1Ingress(*ingress1),
			}),
		},
		{
			name: "case 2: only route based URLs are pushed",
			fields: fields{
				generic: generic{
					appName:       "app",
					componentName: componentName,
				},
				isRouteSupported: true,
			},
			returnedRoutes: routev1.RouteList{
				Items: []routev1.Route{
					route0,
					route1,
				},
			},
			want: getMachineReadableFormatForList([]URL{
				getMachineReadableFormat(route0),
				getMachineReadableFormat(route1)},
			),
		},
		{
			name: "case 3: both route and ingress based URLs are pushed",
			fields: fields{
				generic: generic{
					appName:       "app",
					componentName: componentName,
				},
				isRouteSupported: true,
			},
			returnedRoutes: routev1.RouteList{
				Items: []routev1.Route{
					route0,
					route1,
				},
			},
			returnedIngresses: extensionsv1.IngressList{
				Items: []extensionsv1.Ingress{
					*ingress0,
					*ingress1,
				},
			},
			want: getMachineReadableFormatForList([]URL{
				getMachineReadableFormatExtensionV1Ingress(*ingress0),
				getMachineReadableFormatExtensionV1Ingress(*ingress1),
				getMachineReadableFormat(route0),
				getMachineReadableFormat(route1),
			}),
		},
		{
			name: "case 4: no urls are pushed",
			fields: fields{
				generic: generic{
					appName:       "app",
					componentName: componentName,
				},
				isRouteSupported: true,
			},
			want: getMachineReadableFormatForList(nil),
		},
		{
			name: "case 5: ignore the routes with ingress kind owners",
			fields: fields{
				generic: generic{
					appName:       "app",
					componentName: componentName,
				},
				isRouteSupported: true,
			},
			returnedRoutes: routev1.RouteList{
				Items: []routev1.Route{
					route0,
					routeOwnedByIngress,
				},
			},
			want: getMachineReadableFormatForList([]URL{
				getMachineReadableFormat(route0)},
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fkclient, fkclientset := kclient.FakeNew()
			fkclient.Namespace = "default"

			fkclientset.Kubernetes.PrependReactor("list", "ingresses", func(action ktesting.Action) (bool, runtime.Object, error) {
				if action.GetResource().GroupVersion().Group == "networking.k8s.io" {
					return true, &networkingv1.Ingress{}, nil
				}
				return true, &tt.returnedIngresses, nil
			})

			fkocclient, fkocclientset := occlient.FakeNew()
			fkocclient.SetKubeClient(fkclient)

			fkocclientset.RouteClientset.PrependReactor("list", "routes", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, &tt.returnedRoutes, nil
			})

			k := kubernetesClient{
				generic:          tt.fields.generic,
				isRouteSupported: tt.fields.isRouteSupported,
				client:           *fkocclient,
			}
			got, err := k.ListFromCluster()
			if (err != nil) != tt.wantErr {
				t.Errorf("ListFromCluster() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ListFromCluster() error: %v", pretty.Compare(got, tt.want))
			}
		})
	}
}

func Test_kubernetesClient_List(t *testing.T) {
	componentName := "nodejs"
	appName := "app"

	route0 := testingutil.GetSingleRoute("testRoute0", 8080, componentName, appName)
	route1 := testingutil.GetSingleRoute("testRoute1", 8080, componentName, appName)

	ingress0 := fake.GetSingleExtensionV1Ingress("testIngress0", componentName, appName)

	type fields struct {
		generic          generic
		isRouteSupported bool
	}
	tests := []struct {
		name              string
		fields            fields
		returnedRoutes    routev1.RouteList
		returnedIngress   extensionsv1.IngressList
		returnedLocalURLs []localConfigProvider.LocalURL
		want              URLList
		wantErr           bool
	}{
		{
			name: "case 1: two urls in local config and none pushed",
			fields: fields{
				generic: generic{
					appName:       appName,
					componentName: componentName,
				},
				isRouteSupported: true,
			},
			returnedLocalURLs: []localConfigProvider.LocalURL{
				{
					Name:   "example-1",
					Port:   8080,
					Secure: false,
					Host:   "com",
					Kind:   localConfigProvider.INGRESS,
				},
				{
					Name:   "example-2",
					Port:   8080,
					Secure: false,
					Host:   "com",
					Kind:   localConfigProvider.INGRESS,
				},
			},
			want: getMachineReadableFormatForList([]URL{
				getFakeURL("example-1", "example-1.com", 8080, "", "http", localConfigProvider.INGRESS, StateTypeNotPushed),
				getFakeURL("example-2", "example-2.com", 8080, "", "http", localConfigProvider.INGRESS, StateTypeNotPushed)}),
		},
		{
			name: "case 2: two urls pushed but are deleted locally",
			fields: fields{
				generic: generic{
					appName:       appName,
					componentName: componentName,
				},
				isRouteSupported: true,
			},
			returnedRoutes: routev1.RouteList{
				Items: []routev1.Route{
					route0,
					route1,
				},
			},
			returnedLocalURLs: []localConfigProvider.LocalURL{},

			want: getMachineReadableFormatForList([]URL{
				getFakeURL("testRoute0", "", 8080, "/", "http", localConfigProvider.ROUTE, StateTypeLocallyDeleted),
				getFakeURL("testRoute1", "", 8080, "/", "http", localConfigProvider.ROUTE, StateTypeLocallyDeleted)}),
		},
		{
			name: "case 3: two urls which are pushed",
			fields: fields{
				generic: generic{
					appName:       appName,
					componentName: componentName,
				},
				isRouteSupported: true,
			},
			returnedRoutes: routev1.RouteList{
				Items: []routev1.Route{
					route0,
				},
			},
			returnedIngress: extensionsv1.IngressList{
				Items: []extensionsv1.Ingress{
					*ingress0,
				},
			},
			returnedLocalURLs: []localConfigProvider.LocalURL{
				{
					Name:     "testRoute0",
					Port:     8080,
					Secure:   false,
					Path:     "/",
					Protocol: "http",
					Kind:     localConfigProvider.ROUTE,
				},
				{
					Name:   "testIngress0",
					Port:   8080,
					Secure: false,
					Host:   "com",
					Kind:   localConfigProvider.INGRESS,
				},
			},
			want: getMachineReadableFormatForList([]URL{
				getFakeURL("testIngress0", "testIngress0.com", 8080, "/", "http", localConfigProvider.INGRESS, StateTypePushed),
				getFakeURL("testRoute0", "", 8080, "/", "http", localConfigProvider.ROUTE, StateTypePushed),
			}),
		},
		{
			name: "case 4: three URLs with mixed states",
			fields: fields{
				generic: generic{
					appName:       appName,
					componentName: componentName,
				},
				isRouteSupported: true,
			},
			returnedRoutes: routev1.RouteList{
				Items: []routev1.Route{
					route1,
				},
			},
			returnedIngress: extensionsv1.IngressList{
				Items: []extensionsv1.Ingress{
					*ingress0,
				},
			},
			returnedLocalURLs: []localConfigProvider.LocalURL{
				{
					Name:   "testRoute0",
					Port:   8080,
					Secure: false,
					Path:   "/",
					Kind:   localConfigProvider.ROUTE,
				},
				{
					Name:   "testIngress0",
					Port:   8080,
					Secure: false,
					Host:   "com",
					Kind:   localConfigProvider.INGRESS,
				},
			},

			want: getMachineReadableFormatForList([]URL{
				getFakeURL("testIngress0", "testIngress0.com", 8080, "/", "http", localConfigProvider.INGRESS, StateTypePushed),
				getFakeURL("testRoute0", "", 8080, "/", "http", localConfigProvider.ROUTE, StateTypeNotPushed),
				getFakeURL("testRoute1", "", 8080, "/", "http", localConfigProvider.ROUTE, StateTypeLocallyDeleted),
			}),
		},
		{
			name: "case 5: ignore routes when route resources are not supported",
			fields: fields{
				generic: generic{
					appName:       appName,
					componentName: componentName,
				},
				isRouteSupported: false,
			},
			returnedLocalURLs: []localConfigProvider.LocalURL{
				{
					Name:   "testRoute0",
					Port:   8080,
					Secure: false,
					Host:   "com",
					Kind:   localConfigProvider.ROUTE,
				},
				{
					Name:   "testIngress0",
					Port:   8080,
					Secure: false,
					Host:   "com",
					Path:   "/",
					Kind:   localConfigProvider.INGRESS,
				},
			},
			want: getMachineReadableFormatForList([]URL{
				getFakeURL("testIngress0", "testIngress0.com", 8080, "/", "http", localConfigProvider.INGRESS, StateTypeNotPushed),
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLocalConfig := localConfigProvider.NewMockLocalConfigProvider(ctrl)
			mockLocalConfig.EXPECT().ListURLs().Return(tt.returnedLocalURLs, nil)

			fkclient, fkclientset := kclient.FakeNew()
			fkclient.Namespace = "default"

			fkclientset.Kubernetes.PrependReactor("list", "ingresses", func(action ktesting.Action) (bool, runtime.Object, error) {
				if action.GetResource().GroupVersion().Group == "networking.k8s.io" {
					return true, &networkingv1.Ingress{}, nil
				}
				return true, &tt.returnedIngress, nil
			})

			fkocclient, fkocclientset := occlient.FakeNew()
			fkocclient.SetKubeClient(fkclient)

			fkocclientset.RouteClientset.PrependReactor("list", "routes", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, &tt.returnedRoutes, nil
			})

			tt.fields.generic.localConfig = mockLocalConfig
			k := kubernetesClient{
				generic:          tt.fields.generic,
				isRouteSupported: tt.fields.isRouteSupported,
				client:           *fkocclient,
			}
			got, err := k.List()
			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("List() error: %v", pretty.Compare(got, tt.want))
			}
		})
	}
}

func Test_kubernetesClient_createIngress(t *testing.T) {
	type fields struct {
		generic          generic
		isRouteSupported bool
	}
	type args struct {
		url URL
	}
	tests := []struct {
		name               string
		fields             fields
		args               args
		createdIngress     *extensionsv1.Ingress
		defaultTLSExists   bool
		userGivenTLSExists bool
		want               string
		wantErr            bool
	}{
		{
			name:   "Case 1: Create a ingress, with same name as component",
			fields: fields{generic: generic{componentName: "nodejs", appName: "app"}},
			args: args{
				url: getFakeURL("nodejs", "com", 8080, "/", "http", localConfigProvider.INGRESS, StateTypeNotPushed),
			},
			createdIngress: fake.GetSingleExtensionV1Ingress("nodejs-nodejs-app", "nodejs", "app"),
			want:           "http://nodejs.com",
			wantErr:        false,
		},
		{
			name:   "Case 2: Create a ingress, with different name as component",
			fields: fields{generic: generic{componentName: "nodejs", appName: "app"}},
			args: args{
				url: getFakeURL("example", "com", 8080, "/", "http", localConfigProvider.INGRESS, StateTypeNotPushed),
			},
			createdIngress: fake.GetSingleExtensionV1Ingress("example-nodejs-app", "nodejs", "app"),
			want:           "http://example.com",
			wantErr:        false,
		},
		{
			name:   "Case 3: Create a secure ingress, default tls exists",
			fields: fields{generic: generic{componentName: "nodejs", appName: "app"}},
			args: args{
				url: func() URL {
					url := getFakeURL("example", "com", 8080, "/", "http", localConfigProvider.INGRESS, StateTypeNotPushed)
					url.Spec.Secure = true
					return url
				}(),
			},
			createdIngress:   fake.GetSingleExtensionV1Ingress("example-nodejs-app", "nodejs", "app"),
			defaultTLSExists: true,
			want:             "https://example.com",
			wantErr:          false,
		},
		{
			name:   "Case 4: Create a secure ingress and default tls doesn't exist",
			fields: fields{generic: generic{componentName: "nodejs", appName: "app"}},
			args: args{
				url: func() URL {
					url := getFakeURL("example", "com", 8080, "/", "http", localConfigProvider.INGRESS, StateTypeNotPushed)
					url.Spec.Secure = true
					return url
				}(),
			},
			createdIngress:   fake.GetSingleExtensionV1Ingress("example-nodejs-app", "nodejs", "app"),
			defaultTLSExists: false,
			want:             "https://example.com",
			wantErr:          false,
		},
		{
			name:   "Case 5: Fail when while creating ingress when user given tls secret doesn't exists",
			fields: fields{generic: generic{componentName: "nodejs", appName: "app"}},
			args: args{
				url: func() URL {
					url := getFakeURL("example", "com", 8080, "/", "http", localConfigProvider.INGRESS, StateTypeNotPushed)
					url.Spec.Secure = true
					url.Spec.TLSSecret = "user-secret"
					return url
				}(),
			},
			defaultTLSExists:   false,
			userGivenTLSExists: false,
			want:               "http://example.com",
			wantErr:            true,
		},
		{
			name:   "Case 6: Create a secure ingress, user tls secret does exists",
			fields: fields{generic: generic{componentName: "nodejs", appName: "app"}},
			args: args{
				url: func() URL {
					url := getFakeURL("example", "com", 8080, "/", "http", localConfigProvider.INGRESS, StateTypeNotPushed)
					url.Spec.Secure = true
					url.Spec.TLSSecret = "user-secret"
					return url
				}(),
			},
			createdIngress:     fake.GetSingleExtensionV1Ingress("example-nodejs-app", "nodejs", "app"),
			defaultTLSExists:   false,
			userGivenTLSExists: true,
			want:               "https://example.com",
			wantErr:            false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var serviceName string
			if tt.args.url.Spec.Kind == localConfigProvider.INGRESS {
				serviceName = tt.fields.generic.componentName

			}

			client, fakeClientSet := occlient.FakeNew()
			fakeKClient, fakeKClientSet := kclient.FakeNew()
			client.SetKubeClient(fakeKClient)

			k := kubernetesClient{
				generic:          tt.fields.generic,
				isRouteSupported: tt.fields.isRouteSupported,
				client:           *client,
			}

			fakeKClientSet.Kubernetes.PrependReactor("get", "secrets", func(action ktesting.Action) (bool, runtime.Object, error) {
				var secretName string
				if tt.args.url.Spec.TLSSecret == "" {
					secretName = tt.fields.generic.componentName + "-tlssecret"
					if action.(ktesting.GetAction).GetName() != secretName {
						return true, nil, fmt.Errorf("get for secrets called with invalid name, want: %s,got: %s", secretName, action.(ktesting.GetAction).GetName())
					}
				} else {
					secretName = tt.args.url.Spec.TLSSecret
					if action.(ktesting.GetAction).GetName() != tt.args.url.Spec.TLSSecret {
						return true, nil, fmt.Errorf("get for secrets called with invalid name, want: %s,got: %s", tt.args.url.Spec.TLSSecret, action.(ktesting.GetAction).GetName())
					}
				}
				if tt.args.url.Spec.TLSSecret != "" {
					if !tt.userGivenTLSExists {
						return true, nil, kerrors.NewNotFound(schema.GroupResource{}, "")
					}
				} else if !tt.defaultTLSExists {
					return true, nil, kerrors.NewNotFound(schema.GroupResource{}, "")
				}
				return true, fake.GetSecret(secretName), nil
			})

			fakeKClientSet.Kubernetes.PrependReactor("list", "deployments", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, &appsv1.DeploymentList{Items: []appsv1.Deployment{*testingutil.CreateFakeDeployment("nodejs")}}, nil
			})

			got, err := k.createIngress(tt.args.url, urlLabels.GetLabels(tt.args.url.Name, k.componentName, k.appName, true))
			if (err != nil) != tt.wantErr {
				t.Errorf("createIngress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return
			}

			if got != tt.want {
				t.Errorf("createIngress() got = %v, want %v", got, tt.want)
			}

			wantKubernetesActionLength := 0
			if !tt.args.url.Spec.Secure {
				wantKubernetesActionLength = 2
			} else {
				if tt.args.url.Spec.TLSSecret != "" && tt.userGivenTLSExists {
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
			if !tt.args.url.Spec.Secure {
				createIngressActionNo = 1
			} else {
				if tt.args.url.Spec.TLSSecret != "" {
					createIngressActionNo = 2
				} else if !tt.defaultTLSExists {
					createdDefaultTLS := fakeKClientSet.Kubernetes.Actions()[2].(ktesting.CreateAction).GetObject().(*corev1.Secret)
					if createdDefaultTLS.Name != tt.fields.generic.componentName+"-tlssecret" {
						t.Errorf("default tls created with different name, want: %s,got: %s", tt.fields.generic.componentName+"-tlssecret", createdDefaultTLS.Name)
					}
					createIngressActionNo = 3
				} else {
					createIngressActionNo = 2
				}
			}
			createdIngress = fakeKClientSet.Kubernetes.Actions()[createIngressActionNo].(ktesting.CreateAction).GetObject().(*extensionsv1.Ingress)
			tt.createdIngress.Labels["odo.openshift.io/url-name"] = tt.args.url.Name
			if !reflect.DeepEqual(createdIngress.Name, tt.createdIngress.Name) {
				t.Errorf("ingress name not matching, expected: %s, got %s", tt.createdIngress.Name, createdIngress.Name)
			}
			if !reflect.DeepEqual(createdIngress.Labels, tt.createdIngress.Labels) {
				t.Errorf("ingress labels not matching, %v", pretty.Compare(tt.createdIngress.Labels, createdIngress.Labels))
			}

			wantedIngressSpecParams := generator.IngressSpecParams{
				ServiceName:   serviceName,
				IngressDomain: tt.args.url.Spec.Host,
				PortNumber:    intstr.FromInt(tt.args.url.Spec.Port),
				TLSSecretName: tt.args.url.Spec.TLSSecret,
			}

			if tt.args.url.Spec.Secure {
				if wantedIngressSpecParams.TLSSecretName == "" {
					wantedIngressSpecParams.TLSSecretName = tt.fields.generic.componentName + "-tlssecret"
				}
				if !reflect.DeepEqual(createdIngress.Spec.TLS[0].SecretName, wantedIngressSpecParams.TLSSecretName) {
					t.Errorf("ingress tls name not matching, expected: %s, got %s", wantedIngressSpecParams.TLSSecretName, createdIngress.Spec.TLS)
				}
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Create() = %#v, want %#v", got, tt.want)
			}

		})
	}
}

func Test_kubernetesClient_createRoute(t *testing.T) {
	type fields struct {
		generic          generic
		isRouteSupported bool
	}
	type args struct {
		url URL
	}
	tests := []struct {
		name          string
		fields        fields
		args          args
		returnedRoute *routev1.Route
		want          string
		wantErr       bool
	}{
		{
			name:   "Case 1: Component name same as urlName",
			fields: fields{generic: generic{componentName: "nodejs", appName: "app"}},
			args: args{
				url: getFakeURL("example", "com", 8080, "/", "http", localConfigProvider.ROUTE, StateTypeNotPushed),
			},
			returnedRoute: &routev1.Route{
				ObjectMeta: v1.ObjectMeta{
					Name: "example-nodejs",
					Labels: map[string]string{
						"app.kubernetes.io/part-of":  "app",
						"app.kubernetes.io/instance": "nodejs",
						applabels.App:                "app",
						applabels.ManagedBy:          "odo",
						applabels.ManagerVersion:     version.VERSION,
						"odo.openshift.io/url-name":  "example",
					},
				},
				Spec: routev1.RouteSpec{
					To: routev1.RouteTargetReference{
						Kind: "Service",
						Name: "nodejs",
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
			name:   "Case 2: Component name different than urlName",
			fields: fields{generic: generic{componentName: "nodejs", appName: "app"}},
			args: args{
				url: getFakeURL("example-url", "com", 9100, "/", "http", localConfigProvider.ROUTE, StateTypeNotPushed),
			},
			returnedRoute: &routev1.Route{
				ObjectMeta: v1.ObjectMeta{
					Name: "example-url-nodejs",
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
						Name: "nodejs",
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
			name:   "Case 3: a secure URL",
			fields: fields{generic: generic{componentName: "nodejs", appName: "app"}},
			args: args{
				url: func() URL {
					url := getFakeURL("example-url", "com", 9100, "/", "http", localConfigProvider.ROUTE, StateTypeNotPushed)
					url.Spec.Secure = true
					return url
				}(),
			},
			returnedRoute: &routev1.Route{
				ObjectMeta: v1.ObjectMeta{
					Name: "example-url-nodejs",
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
						Name: "nodejs",
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
			fakeKClient, fakeKClientSet := kclient.FakeNew()
			client.SetKubeClient(fakeKClient)

			fakeClientSet.RouteClientset.PrependReactor("create", "routes", func(action ktesting.Action) (bool, runtime.Object, error) {
				route := action.(ktesting.CreateAction).GetObject().(*routev1.Route)
				route.Spec.Host = "host"
				return true, route, nil
			})

			fakeKClientSet.Kubernetes.PrependReactor("list", "deployments", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, &appsv1.DeploymentList{Items: []appsv1.Deployment{*testingutil.CreateFakeDeployment("nodejs")}}, nil
			})

			k := kubernetesClient{
				generic:          tt.fields.generic,
				isRouteSupported: tt.fields.isRouteSupported,
				client:           *client,
			}
			got, err := k.createRoute(tt.args.url, urlLabels.GetLabels(tt.args.url.Name, k.componentName, k.appName, true))
			if (err != nil) != tt.wantErr {
				t.Errorf("createRoute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("createRoute() got = %v, want %v", got, tt.want)
			}

			if len(fakeClientSet.RouteClientset.Actions()) != 1 {
				t.Errorf("expected 1 RouteClientset.Actions() in CreateService, got: %v", fakeClientSet.RouteClientset.Actions())
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

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Create() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func Test_kubernetesClient_Create(t *testing.T) {
	type fields struct {
		generic          generic
		isRouteSupported bool
	}
	type args struct {
		url URL
	}
	tests := []struct {
		name            string
		fields          fields
		args            args
		returnedIngress *extensionsv1.Ingress
		want            string
		wantErr         bool
	}{
		{
			name:   "Case 1: invalid url kind",
			fields: fields{generic: generic{componentName: "nodejs", appName: "app"}},
			args: args{
				url: getFakeURL("nodejs", "com", 8080, "/", "http", "blah", StateTypeNotPushed),
			},
			wantErr: true,
		},
		{
			name:   "Case 2: route is not supported on the cluster",
			fields: fields{generic: generic{componentName: "nodejs", appName: "app"}, isRouteSupported: false},
			args: args{
				url: getFakeURL("example", "com", 8080, "/", "http", localConfigProvider.ROUTE, StateTypeNotPushed),
			},
			wantErr: true,
		},
		{
			name:   "Case 3: secretName used without secure flag",
			fields: fields{generic: generic{componentName: "nodejs", appName: "app"}, isRouteSupported: false},
			args: args{
				url: func() URL {
					url := getFakeURL("example", "com", 8080, "/", "http", localConfigProvider.ROUTE, StateTypeNotPushed)
					url.Spec.TLSSecret = "secret"
					return url
				}(),
			},
			wantErr: true,
		},
		{
			name:   "Case 4: create a route",
			fields: fields{generic: generic{componentName: "nodejs", appName: "app"}, isRouteSupported: true},
			args: args{
				url: func() URL {
					url := getFakeURL("example", "com", 8080, "/", "http", localConfigProvider.ROUTE, StateTypeNotPushed)
					return url
				}(),
			},
			want:    "http://host",
			wantErr: false,
		},
		{
			name:   "Case 5: create a ingress",
			fields: fields{generic: generic{componentName: "nodejs", appName: "app"}, isRouteSupported: true},
			args: args{
				url: func() URL {
					url := getFakeURL("example", "com", 8080, "/", "http", localConfigProvider.INGRESS, StateTypeNotPushed)
					return url
				}(),
			},
			want:    "http://example.com",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, fakeClientSet := occlient.FakeNew()
			fakeKClient, fakeKClientSet := kclient.FakeNew()
			client.SetKubeClient(fakeKClient)

			fakeKClientSet.Kubernetes.PrependReactor("list", "deployments", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, &appsv1.DeploymentList{Items: []appsv1.Deployment{*testingutil.CreateFakeDeployment("nodejs")}}, nil
			})

			fakeClientSet.RouteClientset.PrependReactor("create", "routes", func(action ktesting.Action) (bool, runtime.Object, error) {
				route := action.(ktesting.CreateAction).GetObject().(*routev1.Route)
				route.Spec.Host = "host"
				return true, route, nil
			})

			k := kubernetesClient{
				generic:          tt.fields.generic,
				isRouteSupported: tt.fields.isRouteSupported,
				client:           *client,
			}
			got, err := k.Create(tt.args.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return
			}

			if got != tt.want {
				t.Errorf("Create() got = %v, want %v", got, tt.want)
			}

			if tt.args.url.Spec.Kind == localConfigProvider.INGRESS {
				requiredIngress := fake.GetSingleExtensionV1Ingress(tt.args.url.Name, tt.fields.generic.componentName, tt.fields.generic.appName)

				createdIngress := fakeKClientSet.Kubernetes.Actions()[1].(ktesting.CreateAction).GetObject().(*extensionsv1.Ingress)
				requiredIngress.Labels["odo.openshift.io/url-name"] = tt.args.url.Name
				if !reflect.DeepEqual(createdIngress.Labels, requiredIngress.Labels) {
					t.Errorf("ingress name not matching, expected: %s, got %s", requiredIngress.Labels, createdIngress.Labels)
				}
			} else if tt.args.url.Spec.Kind == localConfigProvider.ROUTE {
				requiredRoute := testingutil.GetSingleRoute(tt.args.url.Name, tt.args.url.Spec.Port, tt.fields.generic.componentName, tt.fields.generic.appName)
				requiredRoute.Labels["app"] = tt.fields.generic.appName

				createdRoute := fakeClientSet.RouteClientset.Actions()[0].(ktesting.CreateAction).GetObject().(*routev1.Route)
				if !reflect.DeepEqual(createdRoute.Labels, requiredRoute.Labels) {
					t.Errorf("route labels not matching, %v", pretty.Compare(requiredRoute.Labels, createdRoute.Labels))
				}
			}
		})
	}
}
