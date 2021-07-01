package kclient

import (
	"fmt"
	"github.com/openshift/odo/pkg/unions"
	"testing"

	"github.com/devfile/library/pkg/devfile/generator"
	"github.com/pkg/errors"
	extensionsv1 "k8s.io/api/extensions/v1beta1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ktesting "k8s.io/client-go/testing"
)

func TestCreateIngress(t *testing.T) {

	tests := []struct {
		name        string
		ingressName string
		wantErr     bool
	}{
		{
			name:        "Case: Valid ingress name",
			ingressName: "testIngress",
			wantErr:     false,
		},
		{
			name:        "Case: Invalid ingress name",
			ingressName: "",
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// initialising the fakeclient
			fkclient, fkclientset := FakeNew()
			fkclient.Namespace = "default"

			objectMeta := generator.GetObjectMeta(tt.ingressName, "default", nil, nil)
			ingressParams := generator.IngressParams{
				ObjectMeta:        objectMeta,
				IngressSpecParams: generator.IngressSpecParams{ServiceName: tt.ingressName},
			}
			ingress := generator.GetIngress(ingressParams)
			createdIngress, err := fkclient.CreateIngressExtensionV1(*ingress)

			// Checks for unexpected error cases
			if !tt.wantErr == (err != nil) {
				t.Errorf("fkclient.CreateIngressExtensionV1 unexpected error %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil {
				if len(fkclientset.Kubernetes.Actions()) != 1 {
					t.Errorf("expected 1 action, got: %v", fkclientset.Kubernetes.Actions())
				} else {
					if createdIngress.Name != tt.ingressName {
						t.Errorf("ingress name does not match the expected name, expected: %s, got %s", tt.ingressName, createdIngress.Name)
					}
				}
			}

		})
	}
}

func TestListIngresses(t *testing.T) {
	componentName := "testcomponent"
	componentLabel := "componentName"
	tests := []struct {
		name                    string
		labelSelector           string
		wantIngress             unions.KubernetesIngressList
		isNetworkingV1Supported bool
		isExtensionV1Supported  bool
	}{
		{
			name:          "Case: one ingress",
			labelSelector: fmt.Sprintf("%v=%v", componentLabel, componentName),
			wantIngress: unions.KubernetesIngressList{
				Items: []*unions.KubernetesIngress{
					{
						NetworkingV1Ingress: &networkingv1.Ingress{
							ObjectMeta: metav1.ObjectMeta{
								Name: "testIngress1",
								Labels: map[string]string{
									componentLabel: componentName,
								},
							},
						},
						ExtensionV1Beta1Ingress: nil,
					},
				},
			},
			isNetworkingV1Supported: true,
			isExtensionV1Supported:  false,
		},
		{
			name:          "Case: One extension v1 beta ingress",
			labelSelector: fmt.Sprintf("%v=%v", componentLabel, componentName),
			wantIngress: unions.KubernetesIngressList{
				Items: []*unions.KubernetesIngress{
					{
						NetworkingV1Ingress: nil,
						ExtensionV1Beta1Ingress: &extensionsv1.Ingress{
							ObjectMeta: metav1.ObjectMeta{
								Name: "testIngress1",
								Labels: map[string]string{
									componentLabel: componentName,
								},
							},
						},
					},
				},
			},
			isNetworkingV1Supported: false,
			isExtensionV1Supported:  true,
		},
		{
			name:          "Case: two ingresses",
			labelSelector: fmt.Sprintf("%v=%v", componentLabel, componentName),
			wantIngress: unions.KubernetesIngressList{
				Items: []*unions.KubernetesIngress{
					{
						NetworkingV1Ingress: &networkingv1.Ingress{
							ObjectMeta: metav1.ObjectMeta{
								Name: "testIngress1",
								Labels: map[string]string{
									componentLabel: componentName,
								},
							},
						},
						ExtensionV1Beta1Ingress: nil,
					},
					{
						NetworkingV1Ingress: &networkingv1.Ingress{
							ObjectMeta: metav1.ObjectMeta{
								Name: "testIngress2",
								Labels: map[string]string{
									componentLabel: componentName,
								},
							},
						},
						ExtensionV1Beta1Ingress: nil,
					},
				},
			},
			isNetworkingV1Supported: true,
			isExtensionV1Supported:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// initialising the fakeclient
			fkclient, fkclientset := FakeNewWithIngressSupports(tt.isNetworkingV1Supported, tt.isExtensionV1Supported)
			fkclient.Namespace = "default"

			fkclientset.Kubernetes.PrependReactor("list", "ingresses", func(action ktesting.Action) (bool, runtime.Object, error) {
				if tt.labelSelector != action.(ktesting.ListAction).GetListRestrictions().Labels.String() {
					return true, nil, errors.Errorf("selectors are different")
				}
				if action.GetResource().GroupVersion().Group == "networking.k8s.io" {
					return true, tt.wantIngress.GetNetworkingV1IngressList(true), nil
				}
				ingress := tt.wantIngress.GetExtensionV1Beta1IngresList(true)
				return true, ingress, nil
			})
			ingresses, err := fkclient.ListIngresses(tt.labelSelector)

			if err != nil {
				t.Errorf("fkclient.ListIngressesExtensionV1 unexpected error %v", err)
			}

			if err == nil {
				if len(fkclientset.Kubernetes.Actions()) != 1 {
					t.Errorf("expected 1 action, got: %v", fkclientset.Kubernetes.Actions())
				} else {
					if len(tt.wantIngress.Items) != len(ingresses.Items) {
						t.Errorf("IngressList length is different, expected %v, got %v", len(tt.wantIngress.Items), len(ingresses.Items))
					} else if len(ingresses.Items) == 1 && ingresses.Items[0].GetName() != tt.wantIngress.Items[0].GetName() {
						t.Errorf("ingress name does not match the expected name, expected: %s, got %s", tt.wantIngress.Items[0].GetName(), ingresses.Items[0].GetName())
					} else if len(ingresses.Items) == 2 && (ingresses.Items[0].GetName() != tt.wantIngress.Items[0].GetName() || ingresses.Items[1].GetName() != tt.wantIngress.Items[1].GetName()) {
						t.Errorf("ingress name does not match the expected name, expected: %s and %s, got %s and %s", tt.wantIngress.Items[0].GetName(), tt.wantIngress.Items[1].GetName(), ingresses.Items[0].GetName(), ingresses.Items[1].GetName())
					}
				}
			}

		})
	}

}

func TestDeleteIngress(t *testing.T) {

	tests := []struct {
		name        string
		ingressName string
		wantErr     bool
	}{
		{
			name:        "delete test",
			ingressName: "testIngress",
			wantErr:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// initialising the fakeclient
			fkclient, fkclientset := FakeNew()
			fkclient.Namespace = "default"

			fkclientset.Kubernetes.PrependReactor("delete", "ingresses", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, nil, nil
			})

			err := fkclient.DeleteIngressExtensionV1(tt.ingressName)

			// Checks for unexpected error cases
			if !tt.wantErr == (err != nil) {
				t.Errorf("fkclient.DeleteIngressExtensionV1 unexpected error %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil {
				if len(fkclientset.Kubernetes.Actions()) != 1 {
					t.Errorf("expected 1 action, got: %v", fkclientset.Kubernetes.Actions())
				} else {
					DeletedIngress := fkclientset.Kubernetes.Actions()[0].(ktesting.DeleteAction).GetName()
					if DeletedIngress != tt.ingressName {
						t.Errorf("Delete action is performed with wrong ingress name, expected: %s, got %s", tt.ingressName, DeletedIngress)
					}
				}
			}

		})
	}

}

func TestGetIngresses(t *testing.T) {
	tests := []struct {
		name        string
		ingressName string
		wantErr     bool
	}{
		{
			name:        "Case: Valid ingress name",
			ingressName: "testIngress",
			wantErr:     false,
		},
		{
			name:        "Case: Invalid ingress name",
			ingressName: "",
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// initialising the fakeclient
			fkclient, fkclientset := FakeNew()
			fkclient.Namespace = "default"

			fkclientset.Kubernetes.PrependReactor("get", "ingresses", func(action ktesting.Action) (bool, runtime.Object, error) {
				if tt.ingressName == "" {
					return true, nil, errors.Errorf("ingress name is empty")
				}
				ingress := extensionsv1.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Name: tt.ingressName,
					},
				}
				return true, &ingress, nil
			})

			ingress, err := fkclient.GetIngressExtensionV1(tt.ingressName)

			// Checks for unexpected error cases
			if !tt.wantErr == (err != nil) {
				t.Errorf("fkclient.GetIngressExtensionV1 unexpected error %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil {
				if len(fkclientset.Kubernetes.Actions()) != 1 {
					t.Errorf("expected 1 action, got: %v", fkclientset.Kubernetes.Actions())
				} else {
					if ingress.Name != tt.ingressName {
						t.Errorf("ingress name does not match the expected name, expected: %s, got %s", tt.ingressName, ingress.Name)
					}
				}
			}

		})
	}

}
