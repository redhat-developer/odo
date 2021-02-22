package url

import (
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/kylelemons/godebug/pretty"
	routev1 "github.com/openshift/api/route/v1"
	"github.com/openshift/odo/pkg/kclient"
	"github.com/openshift/odo/pkg/kclient/fake"
	"github.com/openshift/odo/pkg/localConfigProvider"
	"github.com/openshift/odo/pkg/occlient"
	"github.com/openshift/odo/pkg/testingutil"
	extensionsv1 "k8s.io/api/extensions/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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
	ingress0 := fake.GetSingleIngress("testIngress0", componentName, appName)
	ingress1 := fake.GetSingleIngress("testIngress1", componentName, appName)

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
				getMachineReadableFormatIngress(*ingress0),
				getMachineReadableFormatIngress(*ingress1),
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
				getMachineReadableFormatIngress(*ingress0),
				getMachineReadableFormatIngress(*ingress1),
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

	ingress0 := fake.GetSingleIngress("testIngress0", componentName, appName)

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
				getFakeURL("example-1", "example-1.com", 8080, "", "", localConfigProvider.INGRESS, StateTypeNotPushed),
				getFakeURL("example-2", "example-2.com", 8080, "", "", localConfigProvider.INGRESS, StateTypeNotPushed)}),
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
				getFakeURL("testIngress0", "testIngress0.com", 8080, "/", "", localConfigProvider.INGRESS, StateTypePushed),
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
				getFakeURL("testIngress0", "testIngress0.com", 8080, "/", "", localConfigProvider.INGRESS, StateTypePushed),
				getFakeURL("testRoute0", "", 8080, "/", "", localConfigProvider.ROUTE, StateTypeNotPushed),
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
				getFakeURL("testIngress0", "testIngress0.com", 8080, "/", "", localConfigProvider.INGRESS, StateTypeNotPushed),
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
