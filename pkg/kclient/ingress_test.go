package kclient

import (
	"fmt"
	"testing"

	"github.com/openshift/odo/pkg/kclient/generator"
	"github.com/pkg/errors"
	extensionsv1 "k8s.io/api/extensions/v1beta1"
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

			objectMeta := generator.CreateObjectMeta(tt.ingressName, "default", nil, nil)

			IngressSpec := generator.GenerateIngressSpec(generator.IngressParams{ServiceName: tt.ingressName})
			createdIngress, err := fkclient.CreateIngress(objectMeta, *IngressSpec)

			// Checks for unexpected error cases
			if !tt.wantErr == (err != nil) {
				t.Errorf("fkclient.CreateIngress unexpected error %v, wantErr %v", err, tt.wantErr)
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
		name          string
		labelSelector string
		wantIngress   []extensionsv1.Ingress
	}{
		{
			name:          "Case: one ingress",
			labelSelector: fmt.Sprintf("%v=%v", componentLabel, componentName),
			wantIngress: []extensionsv1.Ingress{
				extensionsv1.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Name: "testIngress1",
						Labels: map[string]string{
							componentLabel: componentName,
						},
					},
				},
			},
		},
		{
			name:          "Case: two ingresses",
			labelSelector: fmt.Sprintf("%v=%v", componentLabel, componentName),
			wantIngress: []extensionsv1.Ingress{
				extensionsv1.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Name: "testIngress1",
						Labels: map[string]string{
							componentLabel: componentName,
						},
					},
				},
				extensionsv1.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Name: "testIngress2",
						Labels: map[string]string{
							componentLabel: componentName,
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// initialising the fakeclient
			fkclient, fkclientset := FakeNew()
			fkclient.Namespace = "default"

			fkclientset.Kubernetes.PrependReactor("list", "ingresses", func(action ktesting.Action) (bool, runtime.Object, error) {
				if tt.labelSelector != action.(ktesting.ListAction).GetListRestrictions().Labels.String() {
					return true, nil, errors.Errorf("selectors are different")
				}
				ingress := extensionsv1.IngressList{
					Items: tt.wantIngress,
				}
				return true, &ingress, nil
			})
			ingresses, err := fkclient.ListIngresses(tt.labelSelector)

			if err != nil {
				t.Errorf("fkclient.ListIngresses unexpected error %v", err)
			}

			if err == nil {
				if len(fkclientset.Kubernetes.Actions()) != 1 {
					t.Errorf("expected 1 action, got: %v", fkclientset.Kubernetes.Actions())
				} else {
					if len(tt.wantIngress) != len(ingresses) {
						t.Errorf("IngressList length is different, expected %v, got %v", len(tt.wantIngress), len(ingresses))
					} else if len(ingresses) == 1 && ingresses[0].Name != tt.wantIngress[0].Name {
						t.Errorf("ingress name does not match the expected name, expected: %s, got %s", tt.wantIngress[0].Name, ingresses[0].Name)
					} else if len(ingresses) == 2 && (ingresses[0].Name != tt.wantIngress[0].Name || ingresses[1].Name != tt.wantIngress[1].Name) {
						t.Errorf("ingress name does not match the expected name, expected: %s and %s, got %s and %s", tt.wantIngress[0].Name, tt.wantIngress[1].Name, ingresses[0].Name, ingresses[1].Name)
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

			err := fkclient.DeleteIngress(tt.ingressName)

			// Checks for unexpected error cases
			if !tt.wantErr == (err != nil) {
				t.Errorf("fkclient.DeleteIngress unexpected error %v, wantErr %v", err, tt.wantErr)
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

			ingress, err := fkclient.GetIngress(tt.ingressName)

			// Checks for unexpected error cases
			if !tt.wantErr == (err != nil) {
				t.Errorf("fkclient.GetIngress unexpected error %v, wantErr %v", err, tt.wantErr)
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
