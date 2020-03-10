package kclient

import (
	"testing"

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

			objectMeta := CreateObjectMeta(tt.ingressName, "default", nil, nil)

			fkclientset.Kubernetes.PrependReactor("create", "ingresses", func(action ktesting.Action) (bool, runtime.Object, error) {
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

			IngressSpec := GenerateIngressSpec(IngressParameter{ServiceName: tt.ingressName})
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

			fkclientset.Kubernetes.PrependReactor("list", "ingresses", func(action ktesting.Action) (bool, runtime.Object, error) {
				if tt.ingressName == "" {
					return true, nil, errors.Errorf("ingress name is empty")
				}
				ingress := extensionsv1.IngressList{
					Items: []extensionsv1.Ingress{
						extensionsv1.Ingress{
							ObjectMeta: metav1.ObjectMeta{
								Name: tt.ingressName,
							},
						},
					},
				}
				return true, &ingress, nil
			})

			ingresses, err := fkclient.ListIngresses("")

			// Checks for unexpected error cases
			if !tt.wantErr == (err != nil) {
				t.Errorf("fkclient.ListIngresses unexpected error %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil {
				if len(fkclientset.Kubernetes.Actions()) != 1 {
					t.Errorf("expected 1 action, got: %v", fkclientset.Kubernetes.Actions())
				} else {
					if ingresses[0].Name != tt.ingressName {
						t.Errorf("ingress name does not match the expected name, expected: %s, got %s", tt.ingressName, ingresses[0].Name)
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
			name:        "delet test",
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
