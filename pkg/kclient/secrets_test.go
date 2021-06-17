package kclient

import (
	"fmt"
	"reflect"
	"testing"

	"k8s.io/apimachinery/pkg/watch"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
				OwnerReferences: []metav1.OwnerReference{
					{
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
				OwnerReferences: []metav1.OwnerReference{
					{
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
				t.Errorf("fkclient.CreateIngressExtensionV1 unexpected error %v, wantErr %v", err, tt.wantErr)
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

func TestGetSecret(t *testing.T) {
	tests := []struct {
		name       string
		secretNS   string
		secretName string
		wantErr    bool
		want       *corev1.Secret
	}{
		{
			name:       "Case: Valid request for retrieving a secret",
			secretNS:   "",
			secretName: "foo",
			want: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "foo",
				},
			},
			wantErr: false,
		},
		{
			name:       "Case: Invalid request for retrieving a secret",
			secretNS:   "",
			secretName: "foo2",
			want: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "foo",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient, fakeClientSet := FakeNew()

			// Fake getting Secret
			fakeClientSet.Kubernetes.PrependReactor("get", "secrets", func(action ktesting.Action) (bool, runtime.Object, error) {
				if tt.want.Name != tt.secretName {
					return true, nil, fmt.Errorf("'get' called with a different secret name")
				}
				return true, tt.want, nil
			})

			returnValue, err := fakeClient.GetSecret(tt.secretName, tt.secretNS)

			// Check for validating return value
			if err == nil && returnValue != tt.want {
				t.Errorf("error in return value got: %v, expected %v", returnValue, tt.want)
			}

			if !tt.wantErr == (err != nil) {
				t.Errorf("\nclient.GetSecret(secretNS, secretName) unexpected error %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestListSecrets(t *testing.T) {

	tests := []struct {
		name       string
		secretList corev1.SecretList
		output     []corev1.Secret
		wantErr    bool
	}{
		{
			name: "Case 1: Ensure secrets are properly listed",
			secretList: corev1.SecretList{
				Items: []corev1.Secret{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "secret1",
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "secret2",
						},
					},
				},
			},
			output: []corev1.Secret{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "secret1",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "secret2",
					},
				},
			},

			wantErr: false,
		},
	}

	for _, tt := range tests {
		client, fakeClientSet := FakeNew()

		fakeClientSet.Kubernetes.PrependReactor("list", "secrets", func(action ktesting.Action) (bool, runtime.Object, error) {
			return true, &tt.secretList, nil
		})

		secretsList, err := client.ListSecrets("")

		if !reflect.DeepEqual(tt.output, secretsList) {
			t.Errorf("expected output: %#v,got: %#v", tt.secretList, secretsList)
		}

		if err == nil && !tt.wantErr {
			if len(fakeClientSet.Kubernetes.Actions()) != 1 {
				t.Errorf("expected 1 action in ListSecrets got: %v", fakeClientSet.Kubernetes.Actions())
			}
		} else if err == nil && tt.wantErr {
			t.Error("test failed, expected: false, got true")
		} else if err != nil && !tt.wantErr {
			t.Errorf("test failed, expected: no error, got error: %s", err.Error())
		}
	}
}

func TestWaitAndGetSecret(t *testing.T) {

	tests := []struct {
		name       string
		secretName string
		namespace  string
		wantErr    bool
	}{
		{
			name:       "Case 1: no error expected",
			secretName: "ruby",
			namespace:  "dummy",
			wantErr:    false,
		},

		{
			name:       "Case 2: error expected",
			secretName: "",
			namespace:  "dummy",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			fkclient, fkclientset := FakeNew()
			fkWatch := watch.NewFake()

			// Change the status
			go func() {
				fkWatch.Modify(&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name: tt.secretName,
					},
				})
			}()

			fkclientset.Kubernetes.PrependWatchReactor("secrets", func(action ktesting.Action) (handled bool, ret watch.Interface, err error) {
				if len(tt.secretName) == 0 {
					return true, nil, fmt.Errorf("error watching secret")
				}
				return true, fkWatch, nil
			})

			pod, err := fkclient.WaitAndGetSecret(tt.secretName, tt.namespace)

			if !tt.wantErr == (err != nil) {
				t.Errorf(" client.WaitAndGetSecret(string, string) unexpected error %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(fkclientset.Kubernetes.Actions()) != 1 {
				t.Errorf("expected 1 action in WaitAndGetSecret got: %v", fkclientset.Kubernetes.Actions())
			}

			if err == nil {
				if pod.Name != tt.secretName {
					t.Errorf("secret name is not matching to expected name, expected: %s, got %s", tt.secretName, pod.Name)
				}
			}
		})
	}
}
