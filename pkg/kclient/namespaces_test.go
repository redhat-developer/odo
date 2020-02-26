package kclient

import (
	"testing"
)

func TestSetNamespace(t *testing.T) {

	tests := []struct {
		name          string
		namespaceName string
		wantErr       bool
	}{
		{
			name:          "Case: Valid, different namespace name",
			namespaceName: "test-namespace",
			wantErr:       false,
		},
		{
			name:          "Case: Same namespace name",
			namespaceName: "default",
			wantErr:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// initialising the fakeclient
			fkclient, _ := FakeNew()

			fkclient.Namespace = "default"

			err := fkclient.SetCurrentNamespace(tt.namespaceName)

			// Checks for unexpected error cases
			if !tt.wantErr == (err != nil) {
				t.Errorf("fkclient.CreateDeployment(pod) unexpected error %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil {
				rawConfig, _ := fkclient.KubeConfig.RawConfig()
				configNamespace := rawConfig.Contexts[rawConfig.CurrentContext].Namespace
				if configNamespace != tt.namespaceName {
					t.Errorf("current namespace does not match expected namespace: %s, got %s", tt.namespaceName, fkclient.Namespace)
				}
				if fkclient.Namespace != tt.namespaceName {
					t.Errorf("current namespace does not match expected namespace: %s, got %s", tt.namespaceName, fkclient.Namespace)
				}

			}

		})
	}
}
