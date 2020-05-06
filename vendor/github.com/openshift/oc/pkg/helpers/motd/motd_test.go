package motd

import (
	"bytes"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

func configmap(namespace, name, key, content string) *v1.ConfigMap {
	return &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: map[string]string{
			key: content,
		},
	}
}

func TestDisplayMOTD(t *testing.T) {
	var tests = []struct {
		description string
		expected    string
		obj         runtime.Object
	}{
		{"Correctly displays message", "\nThis is a legal notice\n", configmap("openshift", "motd", "message", "This is a legal notice")},
		{"No message because of malformed notice (misspelled key)", "", configmap("openshift", "motd", "mesage", "This is a legal notice")},
		{"No message because of misconfigured notice (wrong name)", "", configmap("openshift", "bad-name", "message", "This is a legal notice")},
		{"No message because notice not found in namespace", "", configmap("default", "motd", "message", "This is a legal notice")},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			clientSet := fake.NewSimpleClientset(test.obj)
			var output bytes.Buffer
			if err := DisplayMOTD(clientSet.CoreV1(), &output); err != nil {
				t.Errorf("Unexpected error: %s", err)
				return
			}
			if output.String() != test.expected {
				t.Errorf("The output is not correct! Got: %s -- Wanted: %s", output.String(), test.expected)
				return
			}
		})
	}
}
