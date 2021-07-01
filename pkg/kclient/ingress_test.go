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

func ingressNameMatchError(expected, got []string) string {
	if len(expected) != len(got) {
		return "invalid error string format expected and got length should be the same"
	}
	val := "ingress name does not match the expected name"
	for i := 0; i < len(expected); i++ {
		val = fmt.Sprintf("%s, expected %s, got %s", val, expected[i], got[i])
	}
	return fmt.Sprintf("ingress name does not match the expected name, expected: %s, got %s", expected, got)
}

func TestCreateIngress(t *testing.T) {

	tests := []struct {
		name                    string
		ingressName             string
		wantErr                 bool
		isNetworkingV1Supported bool
		ieExtensionV1Supported  bool
	}{
		{
			name:                    "Case: Valid networking v1 ingress name",
			ingressName:             "testIngress",
			wantErr:                 false,
			isNetworkingV1Supported: true,
			ieExtensionV1Supported:  false,
		},
		{
			name:                    "Case: Invalid networking v1 ingress name",
			ingressName:             "",
			wantErr:                 true,
			isNetworkingV1Supported: true,
			ieExtensionV1Supported:  false,
		},
		{
			name:                    "Case: Valid extensions v1 beta1 ingress name",
			ingressName:             "testIngress",
			wantErr:                 false,
			isNetworkingV1Supported: false,
			ieExtensionV1Supported:  true,
		},
		{
			name:                    "Case: Invalid extensions v1 beta1 ingress name",
			ingressName:             "",
			wantErr:                 true,
			isNetworkingV1Supported: false,
			ieExtensionV1Supported:  true,
		},
		{
			name:                    "Case: fail if neither is supported",
			ingressName:             "testIngress",
			wantErr:                 true,
			isNetworkingV1Supported: false,
			ieExtensionV1Supported:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// initialising the fakeclient
			fkclient, fkclientset := FakeNewWithIngressSupports(tt.isNetworkingV1Supported, tt.ieExtensionV1Supported)
			fkclient.Namespace = "default"

			objectMeta := generator.GetObjectMeta(tt.ingressName, "default", nil, nil)
			ingressParams := generator.IngressParams{
				ObjectMeta:        objectMeta,
				IngressSpecParams: generator.IngressSpecParams{ServiceName: tt.ingressName},
			}
			ingress := unions.NewKubernetesIngressFromParams(ingressParams)
			createdIngress, err := fkclient.CreateIngress(*ingress)

			// Checks for unexpected error cases
			if !tt.wantErr == (err != nil) {
				t.Errorf("fkclient.CreateIngress unexpected error %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil {
				if len(fkclientset.Kubernetes.Actions()) != 1 {
					t.Errorf("expected 1 action, got: %v", fkclientset.Kubernetes.Actions())
				} else {
					if createdIngress.GetName() != tt.ingressName {
						t.Errorf(ingressNameMatchError([]string{tt.ingressName}, []string{createdIngress.GetName()}))
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
		wantErr                 bool
	}{
		{
			name:          "Case: one networking v1 ingress",
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
			wantErr:                 false,
		},
		{
			name:          "Case: One extension v1 beta1 ingress",
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
			wantErr:                 false,
		},
		{
			name:          "Case: two networking v1 ingresses",
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
			wantErr:                 false,
		},
		{
			name:                    "Case: fails if none of the ingresses are supported",
			labelSelector:           fmt.Sprintf("%v=%v", componentLabel, componentName),
			wantIngress:             unions.KubernetesIngressList{},
			isNetworkingV1Supported: false,
			isExtensionV1Supported:  false,
			wantErr:                 true,
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

			if tt.wantErr && err == nil {
				t.Errorf("fkclient.ListIngress expected err got %s", err)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("fkclient.ListIngresses unexpected error %v", err)
			}

			if err == nil {
				if len(fkclientset.Kubernetes.Actions()) != 1 {
					t.Errorf("expected 1 action, got: %v", fkclientset.Kubernetes.Actions())
				} else {
					if len(tt.wantIngress.Items) != len(ingresses.Items) {
						t.Errorf("IngressList length is different, expected %v, got %v", len(tt.wantIngress.Items), len(ingresses.Items))
					} else if len(ingresses.Items) == 1 && ingresses.Items[0].GetName() != tt.wantIngress.Items[0].GetName() {
						t.Errorf(ingressNameMatchError([]string{tt.wantIngress.Items[0].GetName()}, []string{ingresses.Items[0].GetName()}))
					} else if len(ingresses.Items) == 2 && (ingresses.Items[0].GetName() != tt.wantIngress.Items[0].GetName() || ingresses.Items[1].GetName() != tt.wantIngress.Items[1].GetName()) {
						t.Errorf(ingressNameMatchError([]string{tt.wantIngress.Items[0].GetName(), tt.wantIngress.Items[1].GetName()}, []string{ingresses.Items[0].GetName(), ingresses.Items[1].GetName()}))
					}
				}
			}

		})
	}

}

func TestDeleteIngress(t *testing.T) {

	tests := []struct {
		name                    string
		ingressName             string
		wantErr                 bool
		isNetworkingV1Supported bool
		isExtensionV1Supported  bool
	}{
		{
			name:                    "delete networking v1 test",
			ingressName:             "testIngress",
			wantErr:                 false,
			isNetworkingV1Supported: true,
			isExtensionV1Supported:  false,
		},
		{
			name:                    "delete extension v1 beta1 test",
			ingressName:             "testIngress",
			wantErr:                 false,
			isNetworkingV1Supported: false,
			isExtensionV1Supported:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// initialising the fakeclient
			fkclient, fkclientset := FakeNewWithIngressSupports(tt.isNetworkingV1Supported, tt.isExtensionV1Supported)
			fkclient.Namespace = "default"

			fkclientset.Kubernetes.PrependReactor("delete", "ingresses", func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, nil, nil
			})

			err := fkclient.DeleteIngress(tt.ingressName)

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
		name                           string
		ingressName                    string
		wantErr                        bool
		isNetworkingV1IngressSupported bool
		isExtensionV1IngressSupported  bool
	}{
		{
			name:                           "Case: Valid ingress name",
			ingressName:                    "testIngress",
			wantErr:                        false,
			isNetworkingV1IngressSupported: true,
			isExtensionV1IngressSupported:  false,
		},
		{
			name:                           "Case: Invalid ingress name",
			ingressName:                    "",
			wantErr:                        true,
			isNetworkingV1IngressSupported: true,
			isExtensionV1IngressSupported:  false,
		},
		{
			name:                           "Case: valid extension v1 ingress name",
			ingressName:                    "testIngress",
			wantErr:                        false,
			isExtensionV1IngressSupported:  true,
			isNetworkingV1IngressSupported: false,
		},
		{
			name:                           "Case: invalid extension v1 ingress name",
			ingressName:                    "",
			wantErr:                        true,
			isExtensionV1IngressSupported:  true,
			isNetworkingV1IngressSupported: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// initialising the fakeclient
			fkclient, fkclientset := FakeNewWithIngressSupports(tt.isNetworkingV1IngressSupported, tt.isExtensionV1IngressSupported)
			fkclient.Namespace = "default"

			fkclientset.Kubernetes.PrependReactor("get", "ingresses", func(action ktesting.Action) (bool, runtime.Object, error) {
				if tt.ingressName == "" {
					return true, nil, errors.Errorf("ingress name is empty")
				}
				if action.GetResource().Group == "networking.k8s.io" {
					ingress := networkingv1.Ingress{
						ObjectMeta: metav1.ObjectMeta{
							Name: tt.ingressName,
						},
					}
					return true, &ingress, nil
				}
				ingress := extensionsv1.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Name: tt.ingressName,
					},
				}
				return true, &ingress, nil
			})

			ingress, err := fkclient.GetIngress(tt.ingressName)

			// Checks for unexpected error cases
			if !tt.wantErr == (err != nil) {
				t.Errorf("fkclient.GetIngressExtensionV1 unexpected error %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil {
				if len(fkclientset.Kubernetes.Actions()) != 1 {
					t.Errorf("expected 1 action, got: %v", fkclientset.Kubernetes.Actions())
				} else {
					if ingress.GetName() != tt.ingressName {
						t.Errorf(ingressNameMatchError([]string{tt.ingressName}, []string{ingress.GetName()}))
					}
				}
			}

		})
	}

}
