package kclient

import (
	"testing"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	ktesting "k8s.io/client-go/testing"
)

func TestCreateTLSSecret(t *testing.T) {

	tests := []struct {
		name          string
		secretName    string
		componentName string
		host          string
		wantErr       bool
	}{
		{
			name:          "Case: Valid secret name",
			componentName: "testComponent",
			secretName:    "testComponent-tlssecret",
			host:          "1.2.3.4.nip.io",
			wantErr:       false,
		},
		{
			name:          "Case: Invalid secret name",
			secretName:    "",
			componentName: "testComponent",
			host:          "1.2.3.4.nip.io",
			wantErr:       true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// initialising the fakeclient
			fkclient, fkclientset := FakeNew()
			fkclient.Namespace = "default"

			fkclientset.Kubernetes.PrependReactor("create", "secrets", func(action ktesting.Action) (bool, runtime.Object, error) {
				if tt.secretName == "" {
					return true, nil, errors.Errorf("secret name is empty")
				}
				secret := corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name: tt.secretName,
					},
				}
				return true, &secret, nil
			})
			selfsignedcert, err := GenerateSelfSignedCertificate(tt.host)
			if err != nil {
				t.Errorf("fkclient.GenerateSelfSignedCertificate unexpected error %v", err)
			}
			createdTLSSceret, err := fkclient.CreateTLSSecret(selfsignedcert.CertPem, selfsignedcert.KeyPem, tt.componentName, "")
			// Checks for unexpected error cases
			if !tt.wantErr == (err != nil) {
				t.Errorf("fkclient.CreateIngress unexpected error %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil {
				if len(fkclientset.Kubernetes.Actions()) != 1 {
					t.Errorf("expected 1 action, got: %v", fkclientset.Kubernetes.Actions())
				} else {
					if createdTLSSceret.Name != tt.secretName {
						t.Errorf("secret name does not match the expected name, expected: %s, got %s", tt.secretName, createdTLSSceret.Name)
					}
				}
			}

		})
	}
}
