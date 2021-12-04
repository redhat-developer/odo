package secret

import (
	"testing"

	"github.com/redhat-developer/odo/pkg/kclient"

	applabels "github.com/redhat-developer/odo/pkg/application/labels"
	componentlabels "github.com/redhat-developer/odo/pkg/component/labels"
	"github.com/redhat-developer/odo/pkg/occlient"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ktesting "k8s.io/client-go/testing"
)

func TestGetServiceInstanceList(t *testing.T) {

	tests := []struct {
		name            string
		componentName   string
		applicationName string
		port            string
		secretList      corev1.SecretList
		want            string
		wantErr         bool
	}{
		{
			name:            "Case 1: No secrets returned",
			componentName:   "backend",
			applicationName: "app",
			port:            "",
			secretList: corev1.SecretList{
				Items: []corev1.Secret{
					{},
				},
			},
			want:    "",
			wantErr: true,
		},
		{
			name:            "Case 2: No (matching) secrets returned",
			componentName:   "backend",
			applicationName: "app",
			port:            "",
			secretList: corev1.SecretList{
				Items: []corev1.Secret{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "other-8080",
							Labels: map[string]string{
								applabels.ApplicationLabel:         "app",
								componentlabels.ComponentLabel:     "other",
								componentlabels.ComponentTypeLabel: "java",
							},
							Annotations: map[string]string{
								kclient.ComponentPortAnnotationName: "8080",
							},
						},
					},
				},
			},
			want:    "",
			wantErr: true,
		},
		{
			name:            "Case 3: Single (matching) secret returned and no port is specified",
			componentName:   "backend",
			applicationName: "app",
			port:            "",
			secretList: corev1.SecretList{
				Items: []corev1.Secret{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "other-8080",
							Labels: map[string]string{
								applabels.ApplicationLabel:         "app",
								componentlabels.ComponentLabel:     "other",
								componentlabels.ComponentTypeLabel: "java",
							},
							Annotations: map[string]string{
								kclient.ComponentPortAnnotationName: "8080",
							},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "backend-8080",
							Labels: map[string]string{
								applabels.ApplicationLabel:         "app",
								componentlabels.ComponentLabel:     "backend",
								componentlabels.ComponentTypeLabel: "java",
							},
							Annotations: map[string]string{
								kclient.ComponentPortAnnotationName: "8080",
							},
						},
					},
				},
			},
			want:    "backend-8080",
			wantErr: false,
		},
		{
			name:            "Case 4: Multiple secrets returned and no port is specified",
			componentName:   "backend",
			applicationName: "app",
			port:            "",
			secretList: corev1.SecretList{
				Items: []corev1.Secret{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "backend-8080",
							Labels: map[string]string{
								applabels.ApplicationLabel:         "app",
								componentlabels.ComponentLabel:     "backend",
								componentlabels.ComponentTypeLabel: "java",
							},
							Annotations: map[string]string{
								kclient.ComponentPortAnnotationName: "8080",
							},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "backend-8443",
							Labels: map[string]string{
								applabels.ApplicationLabel:         "app",
								componentlabels.ComponentLabel:     "backend",
								componentlabels.ComponentTypeLabel: "java",
							},
							Annotations: map[string]string{
								kclient.ComponentPortAnnotationName: "8443",
							},
						},
					},
				},
			},
			want:    "",
			wantErr: true,
		},
		{
			name:            "Case 5: Multiple secrets returned and non-matching port is specified",
			componentName:   "backend",
			applicationName: "app",
			port:            "9090",
			secretList: corev1.SecretList{
				Items: []corev1.Secret{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "backend-8080",
							Labels: map[string]string{
								applabels.ApplicationLabel:         "app",
								componentlabels.ComponentLabel:     "backend",
								componentlabels.ComponentTypeLabel: "java",
							},
							Annotations: map[string]string{
								kclient.ComponentPortAnnotationName: "8080",
							},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "backend-8443",
							Labels: map[string]string{
								applabels.ApplicationLabel:         "app",
								componentlabels.ComponentLabel:     "backend",
								componentlabels.ComponentTypeLabel: "java",
							},
							Annotations: map[string]string{
								kclient.ComponentPortAnnotationName: "8443",
							},
						},
					},
				},
			},
			want:    "",
			wantErr: true,
		},
		{
			name:            "Case 6: Multiple secrets returned and matching port is specified",
			componentName:   "backend",
			applicationName: "app",
			port:            "8080",
			secretList: corev1.SecretList{
				Items: []corev1.Secret{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "backend-8443",
							Labels: map[string]string{
								applabels.ApplicationLabel:         "app",
								componentlabels.ComponentLabel:     "backend",
								componentlabels.ComponentTypeLabel: "java",
							},
							Annotations: map[string]string{
								kclient.ComponentPortAnnotationName: "8443",
							},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "backend-8080",
							Labels: map[string]string{
								applabels.ApplicationLabel:         "app",
								componentlabels.ComponentLabel:     "backend",
								componentlabels.ComponentTypeLabel: "java",
							},
							Annotations: map[string]string{
								kclient.ComponentPortAnnotationName: "8080",
							},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "backend-8779",
							Labels: map[string]string{
								applabels.ApplicationLabel:         "app",
								componentlabels.ComponentLabel:     "backend",
								componentlabels.ComponentTypeLabel: "java",
							},
							Annotations: map[string]string{
								kclient.ComponentPortAnnotationName: "8779",
							},
						},
					},
				},
			},
			want:    "backend-8080",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		client, fakeClientSet := occlient.FakeNew()

		//fake the secrets
		fakeClientSet.Kubernetes.PrependReactor("list", "secrets", func(action ktesting.Action) (bool, runtime.Object, error) {
			return true, &tt.secretList, nil
		})

		secretName, err := DetermineSecretName(client, tt.componentName, tt.applicationName, tt.port)

		if !tt.wantErr == (err != nil) {
			t.Errorf("client.GetExposedPorts(imagestream imageTag) unexpected error %v, wantErr %v", err, tt.wantErr)
		}
		if secretName != tt.want {
			t.Errorf("Expected service name '%s', got '%s'", tt.want, secretName)
		}
	}
}
