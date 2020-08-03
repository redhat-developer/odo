package kclient

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	ktesting "k8s.io/client-go/testing"
)

func TestWaitForServiceAccountInNamespace(t *testing.T) {
	tests := []struct {
		name               string
		namespace          string
		serviceAccountName string
		wantErr            bool
	}{
		{
			name:               "Test case 1: with valid namespace and serviceAccountName",
			namespace:          "test-1",
			serviceAccountName: "default",
			wantErr:            false,
		},
		{
			name:               "Test case 2: with no namespace and serviceAccountName",
			namespace:          "",
			serviceAccountName: "",
			wantErr:            true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Fake the client with the appropriate arguments
			client, fakeClientSet := FakeNew()
			fkWatch := watch.NewFake()

			go func() {
				fkWatch.Add(&corev1.ServiceAccount{
					ObjectMeta: metav1.ObjectMeta{
						Name: tt.serviceAccountName,
					},
				})
			}()

			fakeClientSet.Kubernetes.PrependWatchReactor("serviceaccounts", func(action ktesting.Action) (handled bool, ret watch.Interface, err error) {
				return true, fkWatch, nil
			})

			err := client.WaitForServiceAccountInNamespace(tt.namespace, tt.serviceAccountName)
			if err == nil && !tt.wantErr {
				if len(fakeClientSet.Kubernetes.Actions()) != 1 {
					t.Errorf("expected 1 Kubernetes.Actions() in ServiceAccountName wait, got: %v", len(fakeClientSet.Kubernetes.Actions()))
				}
			}

			// Checks for error in positive cases
			if !tt.wantErr == (err != nil) {
				t.Errorf("unexpected error %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
