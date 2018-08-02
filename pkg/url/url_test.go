package url

import (
	"reflect"
	"testing"

	"fmt"
	routev1 "github.com/openshift/api/route/v1"
	applabels "github.com/redhat-developer/odo/pkg/application/labels"
	componentlabels "github.com/redhat-developer/odo/pkg/component/labels"
	"github.com/redhat-developer/odo/pkg/occlient"
	"github.com/redhat-developer/odo/pkg/url/labels"
	corev1 "k8s.io/api/core/v1"
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
	}
	tests := []struct {
		name    string
		args    args
		want    *URL
		wantErr bool
	}{
		{
			name: "component name same as urlName",
			args: args{
				componentName:   "component",
				applicationName: "application",
				urlName:         "component",
				portNumber:      8080,
			},
			want: &URL{
				Name:     "component",
				Protocol: "http",
				URL:      "host",
				Port:     intstr.FromInt(8080),
			},
			wantErr: false,
		},
		{
			name: "component name different than urlName",
			args: args{
				componentName:   "component",
				applicationName: "application",
				urlName:         "example-url",
				portNumber:      9100,
			},
			want: &URL{
				Name:     "example-url",
				Protocol: "http",
				URL:      "host",
				Port:     intstr.FromInt(9100),
			},
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

			got, err := Create(client, tt.args.urlName, tt.args.portNumber, tt.args.componentName, tt.args.applicationName)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %#v, wantErr %#v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Create() = %#v, want %#v", got, tt.want)
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
			name:            "correct values and URL found",
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
								labels.UrlLabel:                "nodejs",
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
								labels.UrlLabel:                "wildfly",
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
			labelSelector: "app.kubernetes.io/component-name=nodejs,app.kubernetes.io/name=app",
			wantErr:       false,
		},
		{
			name:            "correct values and URL not found",
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
								labels.UrlLabel:                "nodejs",
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
								labels.UrlLabel:                "wildfly",
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
			labelSelector: "app.kubernetes.io/component-name=nodejs,app.kubernetes.io/name=app",
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

func TestGetComponentServicePortNumbers(t *testing.T) {
	type args struct {
		componentName   string
		applicationName string
	}
	tests := []struct {
		name             string
		args             args
		selectors        string
		returnedServices corev1.ServiceList
		wantedPorts      []int
		wantErr          bool
	}{
		{
			name: "case 1: with valid values and one port",
			args: args{
				componentName:   "nodejs",
				applicationName: "app",
			},
			selectors: "app.kubernetes.io/component-name=nodejs,app.kubernetes.io/name=app",
			returnedServices: corev1.ServiceList{
				Items: []corev1.Service{
					{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"app.kubernetes.io/name":           "app",
								"app.kubernetes.io/component-name": "nodejs",
							},
						},
						Spec: corev1.ServiceSpec{
							Ports: []corev1.ServicePort{
								{
									Port: 8080,
								},
							},
						},
					},
				},
			},
			wantedPorts: []int{8080},
			wantErr:     false,
		},
		{
			name: "case 2: with valid values and two ports",
			args: args{
				componentName:   "nodejs",
				applicationName: "app",
			},
			selectors: "app.kubernetes.io/component-name=nodejs,app.kubernetes.io/name=app",
			returnedServices: corev1.ServiceList{
				Items: []corev1.Service{
					{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"app.kubernetes.io/name":           "app",
								"app.kubernetes.io/component-name": "nodejs",
							},
						},
						Spec: corev1.ServiceSpec{
							Ports: []corev1.ServicePort{
								{
									Port: 8080,
								},
							},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"app.kubernetes.io/name":           "app",
								"app.kubernetes.io/component-name": "nodejs",
							},
						},
						Spec: corev1.ServiceSpec{
							Ports: []corev1.ServicePort{
								{
									Port: 9100,
								},
							},
						},
					},
				},
			},
			wantedPorts: []int{8080, 9100},
			wantErr:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, fakeClientSet := occlient.FakeNew()

			fakeClientSet.Kubernetes.PrependReactor("list", "services", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
				selectors := action.(ktesting.ListAction).GetListRestrictions()
				if !reflect.DeepEqual(tt.selectors, selectors.Labels.String()) {
					return true, nil, fmt.Errorf("'list' called with different selector")
				}
				return true, &tt.returnedServices, nil
			})

			ports, err := GetComponentServicePortNumbers(client, tt.args.componentName, tt.args.applicationName)

			if err == nil && !tt.wantErr {
				if len(fakeClientSet.Kubernetes.Actions()) != 1 {
					t.Errorf("expected 1 Kubernetes.Actions() in CreateService, got: %v", fakeClientSet.ImageClientset.Actions())
				}

				if !reflect.DeepEqual(tt.wantedPorts, ports) {
					t.Errorf("the returned ports do not match the expected value, expected: %v, got: %v", tt.wantedPorts, ports)
				}
			} else if err == nil && tt.wantErr {
				t.Error("error was expected, but no error was returned")
			} else if err != nil && !tt.wantErr {
				t.Errorf("test failed, no error was expected, but got unexpected error: %s", err)
			}
		})
	}
}
