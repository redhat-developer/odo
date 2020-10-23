package kclient

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	ktesting "k8s.io/client-go/testing"
)

func TestCreateTLSSecret(t *testing.T) {
	tests := []struct {
		name       string
		objectMeta metav1.ObjectMeta
		host       string
		wantErr    bool
	}{
		{
			name: "Case: Valid Secret name",
			objectMeta: metav1.ObjectMeta{
				Name: "testComponent-tlssecret",
				OwnerReferences: []v1.OwnerReference{
					metav1.OwnerReference{
						APIVersion: "1",
						Kind:       "fakeOwnerReference",
						Name:       "testDeployment",
					},
				},
			},
			host:    "1.2.3.4.nip.io",
			wantErr: false,
		},
		{
			name: "Case: Invalid Secret name",
			objectMeta: metav1.ObjectMeta{
				Name: "",
				OwnerReferences: []v1.OwnerReference{
					metav1.OwnerReference{
						APIVersion: "1",
						Kind:       "fakeOwnerReference",
						Name:       "testDeployment",
					},
				},
			},
			host:    "1.2.3.4.nip.io",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// initialising the fakeclient
			fkclient, fkclientset := FakeNew()
			fkclient.Namespace = "default"

			fkclientset.Kubernetes.PrependReactor("create", "secrets", func(action ktesting.Action) (bool, runtime.Object, error) {
				secret := corev1.Secret{
					ObjectMeta: tt.objectMeta,
				}
				return true, &secret, nil
			})
			selfsignedcert, err := GenerateSelfSignedCertificate(tt.host)
			if err != nil {
				t.Errorf("fkclient.GenerateSelfSignedCertificate unexpected error %v", err)
			}
			createdTLSSceret, err := fkclient.CreateTLSSecret(selfsignedcert.CertPem, selfsignedcert.KeyPem, tt.objectMeta)
			// Checks for unexpected error cases
			if !tt.wantErr == (err != nil) {
				t.Errorf("fkclient.CreateIngress unexpected error %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil {
				if len(fkclientset.Kubernetes.Actions()) != 1 {
					t.Errorf("expected 1 action, got: %v", fkclientset.Kubernetes.Actions())
				} else {
					if createdTLSSceret.Name != tt.objectMeta.Name {
						t.Errorf("secret name does not match the expected name, expected: %s, got %s", tt.objectMeta.Name, createdTLSSceret.Name)
					}
				}
			}

		})
	}
}

func TestGenerateSelfSignedCertificate(t *testing.T) {

	tests := []struct {
		name string
		host string
	}{
		{
			name: "test1",
			host: "1.2.3.4.nip.io",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			cert, err := GenerateSelfSignedCertificate(tt.host)
			if err != nil {
				t.Errorf("Unexpected error %v", err)
			}
			if cert.CertPem == nil || cert.KeyPem == nil || len(cert.CertPem) == 0 || len(cert.KeyPem) == 0 {
				t.Errorf("Invalid certificate created")
			}

		})
	}
}
